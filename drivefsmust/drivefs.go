package drivefsmust

import (
	"github.com/Jumpaku/go-drivefs"
	"google.golang.org/api/drive/v3"
)

// DriveFS provides file system-like operations for Google Drive.
// It wraps a drive.Service and provides high-level methods for managing files and directories.
type DriveFS struct {
	driveFS *drivefs.DriveFS
}

// New creates a new DriveFS instance with the given drive.Service.
// The service should be properly authenticated before being passed to this function.
func New(service *drive.Service) *DriveFS {
	return &DriveFS{driveFS: drivefs.New(service)}
}

// PermList lists all permissions for the file or directory with the given fileID.
// Returns a slice of Permission objects representing the access permissions.
func (s *DriveFS) PermList(fileID drivefs.FileID) (permissions []drivefs.Permission) {
	return must1(s.driveFS.PermList(fileID))
}

// PermSet sets a permission for the file or directory with the given fileID.
// If a permission for the same grantee already exists, it will be updated.
// Otherwise, a new permission will be created.
// Returns all permissions after the operation.
func (s *DriveFS) PermSet(fileID drivefs.FileID, permission drivefs.Permission) (permissions []drivefs.Permission) {
	return must1(s.driveFS.PermSet(fileID, permission))
}

// PermDel deletes all permissions matching the given grantee for the file or directory with the given fileID.
// Returns all remaining permissions after the operation.
func (s *DriveFS) PermDel(fileID drivefs.FileID, grantee drivefs.Grantee) (permissions []drivefs.Permission) {
	return must1(s.driveFS.PermDel(fileID, grantee))
}

// MkdirAll creates all directories along the given path if they do not already exist.
// The path must be absolute (starting with '/') and is resolved from the specified rootID.
// Returns the FileInfo of the final directory in the path.
// If two or more directories with the same name exist at any level, returns ErrAlreadyExists.
func (s *DriveFS) MkdirAll(rootID drivefs.FileID, path drivefs.Path) (info drivefs.FileInfo) {
	return must1(s.driveFS.MkdirAll(rootID, path))
}

// Mkdir creates a single directory with the given name in the specified parent directory.
// Returns the FileInfo of the created directory.
func (s *DriveFS) Mkdir(parentID drivefs.FileID, name string) (info drivefs.FileInfo) {
	return must1(s.driveFS.Mkdir(parentID, name))
}

// ReadFile reads the entire contents of the file with the given fileID.
// Returns the file data as a byte slice.
// Returns ErrNotReadable for Google Apps files (Docs, Sheets, etc.) that cannot be directly downloaded.
func (s *DriveFS) ReadFile(fileID drivefs.FileID) (data []byte) {
	return must1(s.driveFS.ReadFile(fileID))
}

// Remove deletes the file or directory with the given fileID.
// For directories, only empty directories can be removed; otherwise returns ErrNotRemovable.
// If moveToTrash is true, the file is moved to trash; otherwise it is permanently deleted.
func (s *DriveFS) Remove(fileID drivefs.FileID, moveToTrash bool) {
	must0(s.driveFS.Remove(fileID, moveToTrash))
}

// RemoveAll deletes the file or directory with the given fileID, including all children if it's a directory.
// If moveToTrash is true, the file is moved to trash; otherwise it is permanently deleted.
func (s *DriveFS) RemoveAll(fileID drivefs.FileID, moveToTrash bool) {
	must0(s.driveFS.RemoveAll(fileID, moveToTrash))
}

// Move moves the file or directory with the given fileID to a new parent directory.
// Returns ErrNotFound if the file does not exist.
func (s *DriveFS) Move(fileID, newParentID drivefs.FileID) {
	must0(s.driveFS.Move(fileID, newParentID))
}

// WriteFile writes data to the file with the given fileID, overwriting any existing content.
func (s *DriveFS) WriteFile(fileID drivefs.FileID, data []byte) {
	must0(s.driveFS.WriteFile(fileID, data))
}

// ReadDir reads the directory with the given fileID and returns a slice of FileInfo
// for all files and subdirectories within it. Does not include trashed items.
func (s *DriveFS) ReadDir(fileID drivefs.FileID) (children []drivefs.FileInfo) {
	return must1(s.driveFS.ReadDir(fileID))
}

// Create creates a new empty file with the given name in the specified parent directory.
// Returns the FileInfo of the created file.
func (s *DriveFS) Create(parentID drivefs.FileID, name string) (info drivefs.FileInfo) {
	return must1(s.driveFS.Create(parentID, name))
}

// Shortcut creates a new shortcut with the given name that points to the target file.
// The shortcut is created in the specified parent directory.
// Returns the FileInfo of the created shortcut.
func (s *DriveFS) Shortcut(parentID drivefs.FileID, name string, targetID drivefs.FileID) (info drivefs.FileInfo) {
	return must1(s.driveFS.Shortcut(parentID, name, targetID))
}

// Info retrieves metadata for the file or directory with the given fileID.
// Returns ErrNotFound if the file does not exist.
func (s *DriveFS) Info(fileID drivefs.FileID) (info drivefs.FileInfo) {
	return must1(s.driveFS.Info(fileID))
}

// Copy creates a copy of the file with the given fileID.
// The copy is placed in the specified parent directory with the given name.
// Returns the FileInfo of the copied file.
func (s *DriveFS) Copy(fileID, newParentID drivefs.FileID, newName string) (info drivefs.FileInfo) {
	return must1(s.driveFS.Copy(fileID, newParentID, newName))
}

// Rename changes the name of the file or directory with the given fileID.
// Returns the updated FileInfo.
func (s *DriveFS) Rename(fileID drivefs.FileID, newName string) (info drivefs.FileInfo) {
	return must1(s.driveFS.Rename(fileID, newName))
}

// Query executes a Google Drive API search query and returns matching files.
// The query uses Google Drive's query syntax.
// See https://developers.google.com/drive/api/guides/search-files for query syntax.
func (s *DriveFS) Query(query string) (results []drivefs.FileInfo) {
	return must1(s.driveFS.Query(query))
}

// FindByPath resolves the given absolute path from the specified root directory.
// Returns all files matching the path (multiple results if duplicates exist at any level).
// The path must be absolute (starting with '/').
func (s *DriveFS) FindByPath(rootID drivefs.FileID, path drivefs.Path) (info []drivefs.FileInfo) {
	return must1(s.driveFS.FindByPath(rootID, path))
}

// ResolvePath returns the absolute path from the root to the file with the given fileID.
// The returned path is a slash-separated string (e.g., "/folder/subfolder/file").
// Returns ErrMultiParentsNotSupported if the file has multiple parents.
func (s *DriveFS) ResolvePath(fileID drivefs.FileID) (path drivefs.Path) {
	return must1(s.driveFS.ResolvePath(fileID))
}

// Walk traverses the file tree rooted at the given fileID.
// For each file or directory (including the root), it calls the provided function with
// the relative path and FileInfo. If the function returns an error, walking stops.
func (s *DriveFS) Walk(rootID drivefs.FileID, f func(drivefs.Path, drivefs.FileInfo) error) {
	must0(s.driveFS.Walk(rootID, f))
}
