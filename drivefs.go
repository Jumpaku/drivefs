package drivefs

import (
	"bytes"
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
func New(service *drive.Service, rootID FileID) *DriveFS {
	return &DriveFS{service: service, rootID: string(rootID)}
}

// MkdirAll creates all directories along the given path if they do not already exist and returns the ID of the last created directory.
func (s *DriveFS) MkdirAll(path Path) (info FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return FileInfo{}, fmt.Errorf("path validation failed: %w", err)
	}
	currentID := s.rootID
	file, found, err := findByID(s, currentID)
	if err != nil {
		return FileInfo{}, err
	}
	if !found {
		return FileInfo{}, fmt.Errorf("root not found: %s: %w", currentID, ErrNotExist)
	}
	for _, p := range parts {
		file, found, err = findByNameIn(s, currentID, p)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to find directory '%s' in '%s': %w", p, currentID, err)
		}
		if found {
			currentID = file.Id
			continue
		}
		file, err = createDirIn(s, currentID, p)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to create directory '%s' in '%s': %w", p, currentID, err)
		}
		currentID = file.Id
	}
	return newFileInfo(file)
}

// Mkdir creates a directory with the given name and returns the ID of the created directory.
func (s *DriveFS) Mkdir(parentID FileID, name string) (info FileInfo, err error) {
	f, err := createDirIn(s, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create directory: %w", err)
	}
	return newFileInfo(f)
}

// ReadFile reads the file with the given fileID and returns its contents as a byte slice.
func (s *DriveFS) ReadFile(fileID FileID) (data []byte, err error) {
	return downloadFile(s, string(fileID))
}

// Remove moves the file or directory at path to the trash.
func (s *DriveFS) Remove(fileID FileID, trash bool) (err error) {
	file, found, err := findByID(s, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}
	if !found {
		return nil
	}
	if file.MimeType == mimeTypeGoogleAppFolder {
		exists, err := existsIn(s, string(fileID))
		if err != nil {
			return fmt.Errorf("failed to check if directory is empty: %w", err)
		}
		if exists {
			return fmt.Errorf("directory '%s' is not empty: %w", fileID, ErrAlreadyExists)
		}
	}

	return s.RemoveAll(fileID, trash)
}

// RemoveAll moves all files and directories under the specified path to the trash.
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
	f, found, err := findByID(s, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}
	if !found {
		return fmt.Errorf("file '%s' not found: %w", fileID, ErrNotExist)
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

// WriteFile writes the provided data to the specified path. If the file exists, it will be overwritten. Returns an error on failure.
func (s *DriveFS) WriteFile(fileID FileID, data []byte) (err error) {
	return uploadFile(s, string(fileID), data)
}

// ReadDir lists the contents of the directory at the specified path.
func (s *DriveFS) ReadDir(fileID FileID) (children []FileInfo, err error) {
	l, err := findAllIn(s, string(fileID))
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

// Create creates a new file at the specified parent directory with the given name.
func (s *DriveFS) Create(parentID FileID, name string) (info FileInfo, err error) {
	f, found, err := findByNameIn(s, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to find parent directory '%s': %w", parentID, err)
	}
	if found {
		if err := uploadFile(s, f.Id, []byte("")); err != nil {
			return FileInfo{}, fmt.Errorf("failed to truncate file: %w", err)
		}
		return newFileInfo(f)
	} else {
		f, err := createFileIn(s, string(parentID), name)
		if err != nil {
			return FileInfo{}, fmt.Errorf("failed to create file: %w", err)
		}
		return newFileInfo(f)
	}
}

// Stat returns the FileInfo for the given path.
func (s *DriveFS) Stat(fileID FileID) (info FileInfo, err error) {
	f, found, err := findByID(s, string(fileID))
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info '%s': %w", fileID, err)
	}
	if !found {
		return FileInfo{}, fmt.Errorf("file not found: %s: %w", fileID, ErrNotExist)
	}
	return newFileInfo(f)
}

// ResolveFileID returns the FileID for the given path.
func (s *DriveFS) ResolveFileID(path Path) (info FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return FileInfo{}, fmt.Errorf("path validation failed: %w", err)
	}
	currentID := s.rootID
	file, found, err := findByID(s, currentID)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to find root directory: %w", err)
	}
	if !found {
		return FileInfo{}, fmt.Errorf("root directory not found: %s: %w", currentID, ErrNotExist)
	}
	for _, p := range parts {
		file, found, err = findByNameIn(s, currentID, p)
		if err != nil {
			return FileInfo{}, err
		}
		if !found {
			return FileInfo{}, fmt.Errorf("path does not exist: %s: %w", path, ErrNotExist)
		}
		currentID = file.Id
	}
	return newFileInfo(file)
}

// ResolvePath returns the Path for the given fileID.
func (s *DriveFS) ResolvePath(fileID FileID) (path Path, err error) {
	currentID := string(fileID)
	var parts []string
	for {
		if currentID == s.rootID {
			break
		}
		f, found, err := findByID(s, currentID)
		if err != nil {
			return "", fmt.Errorf("failed to get file info: %w", err)
		}
		if !found {
			return "", fmt.Errorf("file not found: %s: %w", currentID, ErrNotExist)
		}
		parts = append(parts, f.Name)
		if len(f.Parents) == 0 {
			break
		}
		currentID = f.Parents[0]
	}
	slices.Reverse(parts)
	return Path("/" + strings.Join(parts, "/")), nil
}

// Walk walks the file tree rooted at fileID, calling f for each file or directory in the tree, including fileID itself.
func (s *DriveFS) Walk(fileID FileID, f func(FileInfo) error) (err error) {
	file, found, err := findByID(s, string(fileID))
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if !found {
		return fmt.Errorf("file not found: %s: %w", fileID, ErrNotExist)
	}
	return walk(s, file, f)
}

func walk(s *DriveFS, file *drive.File, f func(FileInfo) error) (err error) {
	info, err := newFileInfo(file)
	if err != nil {
		return fmt.Errorf("failed to create FileInfo: %w", err)
	}
	if err := f(info); err != nil {
		return err
	}
	if file.MimeType != mimeTypeGoogleAppFolder {
		return nil
	}
	files, err := findAllIn(s, file.Id)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	for _, file := range files {
		if err := walk(s, file, f); err != nil {
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

func findByNameIn(s *DriveFS, parentID string, name string) (file *drive.File, found bool, err error) {
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(name), parentID)
	res, err := s.service.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(q).
		Fields(driveFileFields).
		PageSize(1).
		Do()
	if err != nil {
		return nil, false, newDriveError("failed to list files", err)
	}
	if len(res.Files) == 0 {
		return nil, false, nil
	}
	return res.Files[0], true, nil
}

func existsIn(s *DriveFS, parentID string) (found bool, err error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", parentID)
	res, err := s.service.Files.List().
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

func findByID(s *DriveFS, fileID string) (file *drive.File, found bool, err error) {
	file, err = s.service.Files.Get(fileID).
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

func findAllIn(s *DriveFS, parentID string) (files []*drive.File, err error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", parentID)
	err = s.service.Files.List().
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

func createDirIn(s *DriveFS, parentID, name string) (file *drive.File, err error) {
	file, err = s.service.Files.Create(&drive.File{
		Name:     name,
		MimeType: mimeTypeGoogleAppFolder,
		Parents:  []string{parentID},
	}).
		Fields(driveFileFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to create directory", err)
	}
	return file, nil
}

func createFileIn(s *DriveFS, parentID, name string) (file *drive.File, err error) {
	file, err = s.service.Files.Create(&drive.File{
		Name:    name,
		Parents: []string{parentID},
	}).
		Fields(driveFileFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to create file", err)
	}
	return file, nil
}

func downloadFile(s *DriveFS, fileID string) (data []byte, err error) {
	file, err := s.service.Files.Get(fileID).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, newDriveError("failed to get file", err)
	}

	if strings.HasPrefix(file.MimeType, mimeTypePrefixGoogleApp) {
		return nil, fmt.Errorf("cannot download google-apps file: %w", ErrNotReadable)
	}

	resp, err := s.service.Files.Get(fileID).
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

func uploadFile(s *DriveFS, fileID string, data []byte) (err error) {
	_, err = s.service.Files.Update(fileID, &drive.File{}).Media(bytes.NewBuffer(data)).Do()
	if err != nil {
		return newDriveError("failed to upload file", err)
	}
	return nil
}
