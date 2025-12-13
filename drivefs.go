// Package drivefs provides a file system-like interface for Google Drive operations.
//
// This package wraps the google.golang.org/api/drive/v3 package and offers familiar
// filesystem operations such as creating, reading, writing, copying, renaming, moving,
// and deleting files and directories. The package fully supports both My Drive and Shared Drives.
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

// DriveFS provides file system-like operations for Google Drive.
// It wraps a drive.Service and provides high-level methods for managing files and directories.
type DriveFS struct {
	service *drive.Service
}

// New creates a new DriveFS instance with the given drive.Service.
// The service should be properly authenticated before being passed to this function.
func New(service *drive.Service) *DriveFS {
	return &DriveFS{service: service}
}

// PermList lists all permissions for the file or directory with the given fileID.
// Returns a slice of Permission objects representing the access permissions.
func (s *DriveFS) PermList(fileID FileID) (permissions []Permission, err error) {
	perms, err := listPermissions(s.service, string(fileID))
	if err != nil {
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}
	return newPermissions(perms), nil
}

// PermSet sets a permission for the file or directory with the given fileID.
// If a permission for the same grantee already exists, it will be updated.
// Otherwise, a new permission will be created.
// Returns all permissions after the operation.
func (s *DriveFS) PermSet(fileID FileID, permission Permission) (permissions []Permission, err error) {
	perms, err := listPermissions(s.service, string(fileID))
	if err != nil {
		return nil, fmt.Errorf("failed to set permissions: %w", err)
	}

	var updated bool
	for _, perm := range perms {
		if granteeMatch(perm, permission.Grantee()) {
			updated = true
			perm.AllowFileDiscovery = permission.AllowFileDiscovery()
			perm.Role = string(permission.Role())
			err := updatePermissions(s.service, string(fileID), perm)
			if err != nil {
				return nil, newDriveError("failed to set permission", err)
			}
		}
	}

	if !updated {
		var email, domain, granteeType string
		switch grantee := permission.Grantee().(type) {
		case GranteeUser:
			email, granteeType = grantee.Email, granteeTypeUser
		case GranteeGroup:
			email, granteeType = grantee.Email, granteeTypeGroup
		case GranteeDomain:
			domain, granteeType = grantee.Domain, granteeTypeDomain
		case GranteeAnyone:
			granteeType = granteeTypeAnyone
		}
		perm, err := createPermissions(s.service, string(fileID), &drive.Permission{
			AllowFileDiscovery: permission.AllowFileDiscovery(),
			EmailAddress:       email,
			Domain:             domain,
			Id:                 string(permission.ID()),
			Role:               string(permission.Role()),
			Type:               granteeType,
		})
		if err != nil {
			return nil, newDriveError("failed to set permission", err)
		}
		perms = append(perms, perm)
	}

	return newPermissions(perms), nil
}

// PermDel deletes all permissions matching the given grantee for the file or directory with the given fileID.
// Returns all remaining permissions after the operation.
func (s *DriveFS) PermDel(fileID FileID, grantee Grantee) (permissions []Permission, err error) {
	perms, err := listPermissions(s.service, string(fileID))
	if err != nil {
		return nil, fmt.Errorf("failed to delete permissions: %w", err)
	}

	remainedPermissions := []*drive.Permission{}
	for _, perm := range perms {
		if granteeMatch(perm, grantee) {
			err := deletePermissions(s.service, string(fileID), perm.Id)
			if err != nil {
				return nil, newDriveError("failed to delete permission", err)
			}
		} else {
			remainedPermissions = append(remainedPermissions, perm)
		}
	}

	return newPermissions(perms), nil
}

// MkdirAll creates all directories along the given path if they do not already exist.
// The path must be absolute (starting with '/') and is resolved from the specified rootID.
// Returns the FileInfo of the final directory in the path.
// If two or more directories with the same name exist at any level, returns ErrAlreadyExists.
func (s *DriveFS) MkdirAll(rootID FileID, path Path) (info FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return FileInfo{}, fmt.Errorf("path validation failed: %w", err)
	}
	currentID := string(rootID)
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
			return FileInfo{}, fmt.Errorf("multiple directories '%s' already exist in '%s': %w", p, currentID, ErrAlreadyExists)
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

// Mkdir creates a single directory with the given name in the specified parent directory.
// Returns the FileInfo of the created directory.
func (s *DriveFS) Mkdir(parentID FileID, name string) (info FileInfo, err error) {
	f, err := createDirIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create directory: %w", err)
	}
	return newFileInfo(f)
}

// ReadFile reads the entire contents of the file with the given fileID.
// Returns the file data as a byte slice.
// Returns ErrNotReadable for Google Apps files (Docs, Sheets, etc.) that cannot be directly downloaded.
func (s *DriveFS) ReadFile(fileID FileID) (data []byte, err error) {
	return downloadFile(s.service, string(fileID))
}

// Remove deletes the file or directory with the given fileID.
// For directories, only empty directories can be removed; otherwise returns ErrNotRemovable.
// If moveToTrash is true, the file is moved to trash; otherwise it is permanently deleted.
func (s *DriveFS) Remove(fileID FileID, moveToTrash bool) (err error) {
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

	return s.RemoveAll(fileID, moveToTrash)
}

// RemoveAll deletes the file or directory with the given fileID, including all children if it's a directory.
// If moveToTrash is true, the file is moved to trash; otherwise it is permanently deleted.
func (s *DriveFS) RemoveAll(fileID FileID, moveToTrash bool) (err error) {
	if moveToTrash {
		_, err := s.service.Files.Update(string(fileID), &drive.File{Trashed: true}).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return newDriveError("failed to move file to trash", err)
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

// Move moves the file or directory with the given fileID to a new parent directory.
// Returns ErrNotFound if the file does not exist.
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

// WriteFile writes data to the file with the given fileID, overwriting any existing content.
func (s *DriveFS) WriteFile(fileID FileID, data []byte) (err error) {
	return uploadFile(s.service, string(fileID), data)
}

// ReadDir reads the directory with the given fileID and returns a slice of FileInfo
// for all files and subdirectories within it. Does not include trashed items.
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

// Create creates a new empty file with the given name in the specified parent directory.
// Returns the FileInfo of the created file.
func (s *DriveFS) Create(parentID FileID, name string) (info FileInfo, err error) {
	f, err := createFileIn(s.service, string(parentID), name)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create file: %w", err)
	}
	return newFileInfo(f)
}

// Shortcut creates a new shortcut with the given name that points to the target file.
// The shortcut is created in the specified parent directory.
// Returns the FileInfo of the created shortcut.
func (s *DriveFS) Shortcut(parentID FileID, name string, targetID FileID) (info FileInfo, err error) {
	f, err := createShortcutIn(s.service, string(parentID), name, string(targetID))
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to create shortcut: %w", err)
	}
	return newFileInfo(f)
}

// Info retrieves metadata for the file or directory with the given fileID.
// Returns ErrNotFound if the file does not exist.
func (s *DriveFS) Info(fileID FileID) (info FileInfo, err error) {
	f, found, err := findByID(s.service, string(fileID))
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to get file info '%s': %w", fileID, err)
	}
	if !found {
		return FileInfo{}, fmt.Errorf("file not found: %s: %w", fileID, ErrNotFound)
	}
	return newFileInfo(f)
}

// Copy creates a copy of the file with the given fileID.
// The copy is placed in the specified parent directory with the given name.
// Returns the FileInfo of the copied file.
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

// Rename changes the name of the file or directory with the given fileID.
// Returns the updated FileInfo.
func (s *DriveFS) Rename(fileID FileID, newName string) (info FileInfo, err error) {
	f, err := s.service.Files.Update(string(fileID), &drive.File{Name: newName}).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return FileInfo{}, newDriveError("failed to copy file", err)
	}
	return newFileInfo(f)
}

// Query executes a Google Drive API search query and returns matching files.
// The query uses Google Drive's query syntax.
// See https://developers.google.com/drive/api/guides/search-files for query syntax.
func (s *DriveFS) Query(query string) (results []FileInfo, err error) {
	files, err := queryFileInfo(s.service, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	for _, f := range files {
		info, err := newFileInfo(f)
		if err != nil {
			return nil, fmt.Errorf("failed to create FileInfo: %w", err)
		}
		results = append(results, info)
	}
	return results, nil
}

// FindByPath resolves the given absolute path from the specified root directory.
// Returns all files matching the path (multiple results if duplicates exist at any level).
// The path must be absolute (starting with '/').
func (s *DriveFS) FindByPath(rootID FileID, path Path) (info []FileInfo, err error) {
	parts, err := validateAndSplitPath(string(path))
	if err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}
	file, found, err := findByID(s.service, string(rootID))
	if err != nil {
		return nil, fmt.Errorf("failed to find root directory: %w", err)
	}
	if !found {
		return nil, nil
	}
	err = dfsFindByPath(s.service, file, 0, parts, func(i FileInfo) error {
		info = append(info, i)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}
	return info, nil
}

// ResolvePath returns the absolute path from the root to the file with the given fileID.
// The returned path is a slash-separated string (e.g., "/folder/subfolder/file").
// Returns ErrMultiParentsNotSupported if the file has multiple parents.
func (s *DriveFS) ResolvePath(fileID FileID) (path Path, err error) {
	parts, err := resolvePathParts(s, fileID)
	return Path("/" + strings.Join(parts, "/")), nil
}

// Walk traverses the file tree rooted at the given fileID.
// For each file or directory (including the root), it calls the provided function with
// the relative path and FileInfo. If the function returns an error, walking stops.
func (s *DriveFS) Walk(rootID FileID, f func(Path, FileInfo) error) (err error) {
	file, found, err := findByID(s.service, string(rootID))
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if !found {
		return fmt.Errorf("file not found: %s: %w", rootID, ErrNotFound)
	}
	return walk(s, []string{}, file, f)
}

func resolvePathParts(s *DriveFS, fileID FileID) (parts []string, err error) {
	currentID := string(fileID)
	for {
		f, found, err := findByID(s.service, currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}
		if !found {
			return nil, fmt.Errorf("file not found: %s: %w", currentID, ErrNotFound)
		}
		if len(f.Parents) == 0 {
			break
		}
		if len(f.Parents) > 1 {
			return nil, fmt.Errorf("failed to resolve path with multiple parents not supported: %w", ErrMultiParentsNotSupported)
		}
		parts = append(parts, f.Name)
		currentID = f.Parents[0]
	}
	slices.Reverse(parts)
	return parts, nil
}

func queryFileInfo(s *drive.Service, query string) (results []*drive.File, err error) {
	err = s.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(query).
		Fields(driveFilesFields).
		Pages(context.Background(), func(list *drive.FileList) error {
			results = append(results, list.Files...)
			return nil
		})
	if err != nil {
		return nil, newDriveError("failed to query files", err)
	}
	return results, nil
}

func dfsFindByPath(s *drive.Service, file *drive.File, partIndex int, parts []string, onPathMatch func(FileInfo) error) (err error) {
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
	files, err := findAllByNameIn(s, file.Id, parts[partIndex])
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	for _, file := range files {
		if err := dfsFindByPath(s, file, partIndex+1, parts, onPathMatch); err != nil {
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
	driveFileFields        = "parents,id,name,mimeType,size,modifiedTime,shortcutDetails,webViewLink"
	driveFilesFields       = "nextPageToken,files(parents,id,name,mimeType,size,modifiedTime,shortcutDetails,webViewLink)"
	drivePermissionFields  = "id,type,emailAddress,domain,role,allowFileDiscovery"
	drivePermissionsFields = "nextPageToken,permissions(id,type,emailAddress,domain,role,allowFileDiscovery)"
)

func newFileInfo(f *drive.File) (FileInfo, error) {
	modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
	var shortcutTarget FileID
	if f.ShortcutDetails != nil {
		shortcutTarget = FileID(f.ShortcutDetails.TargetId)
	}
	return FileInfo{
		Name:           f.Name,
		ID:             FileID(f.Id),
		Size:           f.Size,
		Mime:           f.MimeType,
		ModTime:        modTime,
		ShortcutTarget: shortcutTarget,
		WebViewLink:    f.WebViewLink,
	}, nil
}

func findAllByNameIn(s *drive.Service, parentID string, name string) (files []*drive.File, err error) {
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(name), parentID)
	return queryFileInfo(s, q)
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
	return queryFileInfo(s, q)
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

func createShortcutIn(s *drive.Service, parentID, name, targetID string) (file *drive.File, err error) {
	file, err = s.Files.Create(&drive.File{
		Name:            name,
		MimeType:        mimeTypeGoogleAppShortcut,
		Parents:         []string{parentID},
		ShortcutDetails: &drive.FileShortcutDetails{TargetId: targetID},
	}).
		SupportsAllDrives(true).
		Fields(driveFileFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to create shortcut", err)
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

func newPermissions(perms []*drive.Permission) (permissions []Permission) {
	for _, perm := range perms {
		var grantee Grantee
		switch perm.Type {
		case granteeTypeUser:
			grantee = User(perm.EmailAddress)
		case granteeTypeGroup:
			grantee = Group(perm.EmailAddress)
		case granteeTypeDomain:
			grantee = Domain(perm.EmailAddress)
		case granteeTypeAnyone:
			grantee = Anyone()
		}
		permissions = append(permissions, permission{
			grantee:            grantee,
			role:               Role(perm.Role),
			id:                 PermissionID(perm.Id),
			allowFileDiscovery: perm.AllowFileDiscovery,
		})
	}
	return permissions
}

func granteeMatch(perm *drive.Permission, grantee Grantee) bool {
	switch grantee := grantee.(type) {
	case GranteeUser:
		return perm.Type == granteeTypeUser && perm.EmailAddress == grantee.Email
	case GranteeGroup:
		return perm.Type == granteeTypeGroup && perm.EmailAddress == grantee.Email
	case GranteeDomain:
		return perm.Type == granteeTypeDomain && perm.Domain == grantee.Domain
	case GranteeAnyone:
		return perm.Type == granteeTypeAnyone
	}
	return false
}

func listPermissions(service *drive.Service, fileID string) ([]*drive.Permission, error) {
	var permissions []*drive.Permission
	err := service.Permissions.List(fileID).
		SupportsAllDrives(true).
		Fields(drivePermissionsFields).
		Pages(context.Background(), func(list *drive.PermissionList) error {
			permissions = append(permissions, list.Permissions...)
			return nil
		})
	if err != nil {
		return nil, newDriveError("failed to list permissions", err)
	}
	return permissions, nil
}

func updatePermissions(s *drive.Service, fileID string, perm *drive.Permission) (err error) {
	_, err = s.Permissions.Update(fileID, perm.Id, perm).
		SupportsAllDrives(true).
		Fields(drivePermissionFields).
		Do()
	if err != nil {
		return newDriveError("failed to set permission", err)
	}
	return nil
}

func createPermissions(s *drive.Service, fileID string, perm *drive.Permission) (permission *drive.Permission, err error) {
	permission, err = s.Permissions.Create(fileID, perm).
		SupportsAllDrives(true).
		Fields(drivePermissionFields).
		Do()
	if err != nil {
		return nil, newDriveError("failed to set permission", err)
	}
	return permission, nil
}

func deletePermissions(s *drive.Service, fileID, permID string) (err error) {
	err = s.Permissions.Delete(fileID, permID).
		SupportsAllDrives(true).
		Fields(drivePermissionFields).
		Do()
	if err != nil {
		return newDriveError("failed to set permission", err)
	}
	return nil
}
