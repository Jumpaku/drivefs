package drivefs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

type DriveFS struct {
	service *drive.Service
	rootID  string
}

// New creates a new DriveFS instance with the given drive.Service.
func New(service *drive.Service, rootID FileID) (*DriveFS, error) {
	if rootID == "" {
		rootID = "root"
	}
	if rootID == "root" {
		f, _, err := findByID(service, string(rootID))
		if err != nil {
			return nil, fmt.Errorf("failed to find root directory: %w", err)
		}
		rootID = FileID(f.Id)
	}
	return &DriveFS{service: service, rootID: string(rootID)}, nil
}

// MkdirAll creates all directories along the given path if they do not already exist and returns the ID of the last created directory.
func (s *DriveFS) MkdirAll(path Path) (info FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return FileInfo{}, fmt.Errorf("path validation failed: %w", err)
	}
	currentID := s.rootID
	file, found, err := findByID(s.service, currentID)
	if err != nil {
		return FileInfo{}, err
	}
	if !found {
		return FileInfo{}, fmt.Errorf("root not found: %s: %w", currentID, ErrNotFound)
	}
	for _, p := range parts {
		files, err := findAllByNameIn(s.service, currentID, p)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to find directory '%s' in '%s': %w", p, currentID, err)
		}
		if len(files) > 1 {
			return FileInfo{}, fmt.Errorf("multiple directory '%s' already exists in '%s': %w", p, currentID, ErrAlreadyExists)
		}
		if len(files) == 1 {
			file = files[0]
			currentID = file.Id
			continue
		}
		file, err = createDirIn(s.service, currentID, p)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to create directory '%s' in '%s': %w", p, currentID, err)
		}
		currentID = file.Id
	}
	return newFileInfo(file)
}

// Mkdir creates a directory with the given name and returns the ID of the created directory.
// If errorOnDuplicate is true, an error is returned if a directory with the same name already exists in the parent directory.
// Otherwise, a new directory is created.
func (s *DriveFS) Mkdir(parentID FileID, name string, errorOnDuplicate bool) (info FileInfo, err error) {
	alreadyExists, err := existsByNameIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to find parent directory '%s': %w", parentID, err)
	}
	if errorOnDuplicate && alreadyExists {
		return FileInfo{}, fmt.Errorf("directory with name '%s' already exists in directory '%s': %w", name, parentID, ErrAlreadyExists)
	}
	f, err := createDirIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create directory: %w", err)
	}
	return newFileInfo(f)
}

// ReadFile reads the file with the given fileID and returns its contents as a byte slice.
func (s *DriveFS) ReadFile(fileID FileID) (data []byte, err error) {
	return downloadFile(s.service, string(fileID))
}

// Remove moves the file or directory with the given fileID to the trash.
func (s *DriveFS) Remove(fileID FileID, trash bool) (err error) {
	file, found, err := findByID(s.service, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}
	if !found {
		return nil
	}
	if file.MimeType == mimeTypeGoogleAppFolder {
		exists, err := existsIn(s.service, string(fileID))
		if err != nil {
			return fmt.Errorf("failed to check if directory is empty: %w", err)
		}
		if exists {
			return fmt.Errorf("directory '%s' is not empty: %w", fileID, ErrNotRemovable)
		}
	}

	return s.RemoveAll(fileID, trash)
}

// RemoveAll moves the file or directory with the given fileID to the trash or deletes it permanently.
func (s *DriveFS) RemoveAll(fileID FileID, trash bool) (err error) {
	if trash {
		_, err := s.service.Files.Update(string(fileID), &drive.File{Trashed: true}).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return newDriveError("failed to trash file", err)
		}
		return nil
	} else {
		err := s.service.Files.Delete(string(fileID)).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return newDriveError("failed to delete file", err)
		}
		return nil
	}
}

// Move moves the file or directory at fileID to the new parent directory specified by newParentID.
func (s *DriveFS) Move(fileID, newParentID FileID) (err error) {
	f, found, err := findByID(s.service, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}
	if !found {
		return fmt.Errorf("file '%s' not found: %w", fileID, ErrNotFound)
	}
	_, err = s.service.Files.Update(string(fileID), &drive.File{}).
		SupportsAllDrives(true).
		RemoveParents(strings.Join(f.Parents, ",")).
		AddParents(string(newParentID)).
		Do()
	if err != nil {
		return newDriveError("failed to move file", err)
	}
	return nil
}

// WriteFile writes the provided data to the file with the given fileID. If the file exists, it will be overwritten. Returns an error on failure.
func (s *DriveFS) WriteFile(fileID FileID, data []byte) (err error) {
	return uploadFile(s.service, string(fileID), data)
}

// ReadDir lists the contents of the directory with the given fileID.
func (s *DriveFS) ReadDir(fileID FileID) (children []FileInfo, err error) {
	l, err := findAllIn(s.service, string(fileID))
	if err != nil {
		return nil, fmt.Errorf("failed to list directory contents: %w", err)
	}
	for _, f := range l {
		child, err := newFileInfo(f)
		if err != nil {
			return nil, fmt.Errorf("failed to create FileInfo: %w", err)
		}
		children = append(children, child)
	}
	return children, nil
}

// Create creates a new file with the given name in the directory with the given parentID.
// If errorOnDuplicate is true, an error is returned if a file with the same name already exists in the directory.
// Otherwise, a new file is created.
func (s *DriveFS) Create(parentID FileID, name string, errorOnDuplicate bool) (info FileInfo, err error) {
	alreadyExists, err := existsByNameIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to find parent directory '%s': %w", parentID, err)
	}
	if errorOnDuplicate && alreadyExists {
		return FileInfo{}, fmt.Errorf("file with name '%s' already exists in directory '%s': %w", name, parentID, ErrAlreadyExists)
	}
	f, err := createFileIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create file: %w", err)
	}
	return newFileInfo(f)
}

// Stat returns the FileInfo for the file with the given fileID.
func (s *DriveFS) Stat(fileID FileID) (info FileInfo, err error) {
	f, found, err := findByID(s.service, string(fileID))
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info '%s': %w", fileID, err)
	}
	if !found {
		return FileInfo{}, fmt.Errorf("file not found: %s: %w", fileID, ErrNotFound)
	}
	return newFileInfo(f)
}

// Copy creates a copy of the file with the given fileID in the specified new parent directory with the provided new name.
func (s *DriveFS) Copy(fileID, newParentID FileID, newName string) (info FileInfo, err error) {
	f, err := s.service.Files.Copy(string(fileID), &drive.File{
		Name:    newName,
		Parents: []string{string(newParentID)},
	}).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return FileInfo{}, newDriveError("failed to copy file", err)
	}
	return newFileInfo(f)
}

// Rename renames the file with the given fileID to the specified new name.
func (s *DriveFS) Rename(fileID FileID, newName string) (info FileInfo, err error) {
	f, err := s.service.Files.Update(string(fileID), &drive.File{Name: newName}).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return FileInfo{}, newDriveError("failed to copy file", err)
	}
	return newFileInfo(f)
}

// ResolveFilesByPath resolves the given absolute path from the root directory and returns a slice of FileInfo for all files and directories that match the path.
func (s *DriveFS) ResolveFilesByPath(path Path) (info []FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}
	file, found, err := findByID(s.service, s.rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to find root directory: %w", err)
	}
	if !found {
		return nil, nil
	}
	err = dfsResolveFilesByPath(s, file, 0, parts, func(i FileInfo) error {
		info = append(info, i)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}
	return info, nil
}

// ResolvePath returns the absolute path from the root directory to the file with the given fileID.
// The returned path is a slash-separated string (e.g., "/folder/subfolder/file").
func (s *DriveFS) ResolvePath(fileID FileID) (path Path, err error) {
	parts, err := resolvePathParts(s, fileID)
	return Path("/" + strings.Join(parts, "/")), nil
}
func resolvePathParts(s *DriveFS, fileID FileID) (parts []string, err error) {
	currentID := string(fileID)
	for {
		if currentID == s.rootID {
			break
		}
		f, found, err := findByID(s.service, currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}
		if !found {
			return nil, fmt.Errorf("file not found: %s: %w", currentID, ErrNotFound)
		}
		parts = append(parts, f.Name)
		if len(f.Parents) == 0 {
			break
		}
		if len(f.Parents) > 1 {
			return nil, fmt.Errorf("failed to resolve path with multiple parents not supported: %w", ErrMultiParentsNotSupported)
		}
		currentID = f.Parents[0]
	}
	slices.Reverse(parts)
	return parts, nil
}

// Walk walks the file tree rooted at fileID, calling f for each file or directory in the tree, including fileID itself.
func (s *DriveFS) Walk(fileID FileID, f func(Path, FileInfo) error) (err error) {
	file, found, err := findByID(s.service, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if !found {
		return fmt.Errorf("file not found: %s: %w", fileID, ErrNotFound)
	}
	parts, err := resolvePathParts(s, fileID)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	return walk(s, parts, file, f)
}

func dfsResolveFilesByPath(s *DriveFS, file *drive.File, partIndex int, parts []string, onPathMatch func(FileInfo) error) (err error) {
	info, err := newFileInfo(file)
	if err != nil {
		return fmt.Errorf("failed to create FileInfo: %w", err)
	}
	if partIndex == len(parts) {
		return onPathMatch(info)
	}
	if file.MimeType != mimeTypeGoogleAppFolder {
		return nil
	}
	files, err := findAllByNameIn(s.service, file.Id, parts[partIndex])
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	for _, file := range files {
		if err := dfsResolveFilesByPath(s, file, partIndex+1, parts, onPathMatch); err != nil {
			return err
		}
	}
	return nil
}

func walk(s *DriveFS, path []string, file *drive.File, f func(Path, FileInfo) error) (err error) {
	info, err := newFileInfo(file)
	if err != nil {
		return fmt.Errorf("failed to create FileInfo: %w", err)
	}
	if err := f(Path("/"+strings.Join(path, "/")), info); err != nil {
		return err
	}
	if file.MimeType != mimeTypeGoogleAppFolder {
		return nil
	}
	files, err := findAllIn(s.service, file.Id)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	for _, file := range files {
		if err := walk(s, append(append([]string{}, path...), file.Name), file, f); err != nil {
			return err
		}
	}
	return nil
}

func validateAndSplitPath(path string) (parts []string, err error) {
	if path == "" {
		return nil, fmt.Errorf("empty path: %w", ErrInvalidPath)
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must be absolute and start with '/': %w", ErrInvalidPath)
	}

	for _, p := range strings.Split(path, "/") {
		if p == "." || p == ".." {
			return nil, fmt.Errorf("relative path components are not allowed: %w", ErrInvalidPath)
		}
		if p == "" {
			continue
		}
		parts = append(parts, p)
	}

	return parts, nil
}

func escapeQuery(s string) string {
	s = strings.ReplaceAll(s, "'", `\'`)
	s = strings.ReplaceAll(s, `\`, `\\`)
	return s
}

const (
	driveFileFields  = "parents,id,name,mimeType,size,modifiedTime"
	driveFilesFields = "nextPageToken,files(parents,id,name,mimeType,size,modifiedTime)"
)

func newFileInfo(f *drive.File) (FileInfo, error) {
	modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
	return FileInfo{
		Name:    f.Name,
		ID:      FileID(f.Id),
		Size:    f.Size,
		Mime:    f.MimeType,
		ModTime: modTime,
	}, nil
}

func findAllByNameIn(s *drive.Service, parentID string, name string) (files []*drive.File, err error) {
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(name), parentID)
	err = s.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(q).
		Fields(driveFilesFields).
		Pages(context.Background(), func(list *drive.FileList) error {
			files = append(files, list.Files...)
			return nil
		})
	if err != nil {
		return nil, newDriveError("failed to list files", err)
	}
	return files, nil
}

func existsByNameIn(s *drive.Service, parentID string, name string) (exists bool, err error) {
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(name), parentID)
	resp, err := s.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(q).
		Fields(driveFilesFields).
		PageSize(1).
		Do()
	if err != nil {
		return false, newDriveError("failed to list files", err)
	}
	return len(resp.Files) != 0, nil
}

func existsIn(s *drive.Service, parentID string) (found bool, err error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", parentID)
	res, err := s.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(q).
		Fields(driveFileFields).
		PageSize(1).
		Do()
	if err != nil {
		return false, newDriveError("failed to list files", err)
	}
	return len(res.Files) != 0, nil
}

func findByID(s *drive.Service, fileID string) (file *drive.File, found bool, err error) {
	file, err = s.Files.Get(fileID).
		SupportsAllDrives(true).
		Fields(driveFileFields).
		Do()
	if err != nil {
		var gErr *googleapi.Error
		if errors.As(err, &gErr) {
			if gErr.Code == 404 {
				return nil, false, nil
			}
		}
		return nil, false, newDriveError("failed to get files", err)
	}
	return file, true, nil
}

func findAllIn(s *drive.Service, parentID string) (files []*drive.File, err error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", parentID)
	err = s.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(q).
		Fields(driveFilesFields).
		Pages(nil, func(page *drive.FileList) error {
			files = append(files, page.Files...)
			return nil
		})
	if err != nil {
		return nil, newDriveError("failed to list files", err)
	}
	return files, nil
}

func createDirIn(s *drive.Service, parentID, name string) (file *drive.File, err error) {
	file, err = s.Files.Create(&drive.File{
		Name:     name,
		MimeType: mimeTypeGoogleAppFolder,
		Parents:  []string{parentID},
	}).
		SupportsAllDrives(true).
		Fields(driveFileFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to create directory", err)
	}
	return file, nil
}

func createFileIn(s *drive.Service, parentID, name string) (file *drive.File, err error) {
	file, err = s.Files.Create(&drive.File{
		Name:    name,
		Parents: []string{parentID},
	}).
		SupportsAllDrives(true).
		Fields(driveFileFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to create file", err)
	}
	return file, nil
}

func downloadFile(s *drive.Service, fileID string) (data []byte, err error) {
	file, err := s.Files.Get(fileID).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, newDriveError("failed to get file", err)
	}

	if strings.HasPrefix(file.MimeType, mimeTypePrefixGoogleApp) {
		return nil, fmt.Errorf("cannot download google-apps file: %w", ErrNotReadable)
	}

	resp, err := s.Files.Get(fileID).
		SupportsAllDrives(true).
		Download()
	if err != nil {
		return nil, newDriveError("failed to download file", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			closeErr = newIOError("failed to close file body", closeErr)
		}
		err = errors.Join(err, closeErr)
	}()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, newIOError("failed to read file body", err)
	}
	return data, nil
}

func uploadFile(s *drive.Service, fileID string, data []byte) (err error) {
	_, err = s.Files.Update(fileID, &drive.File{}).
		SupportsAllDrives(true).
		Media(bytes.NewBuffer(data)).
		Do()
	if err != nil {
		return newDriveError("failed to upload file", err)
	}
	return nil
}
