// Package drivefs provides a read-only filesystem interface for Google Drive.
// It implements standard Go filesystem interfaces (fs.FS, fs.File, fs.DirEntry, fs.FileInfo)
// for accessing Google Drive contents using the google.golang.org/api/drive/v3 package.
//
// Note: The openFile method loads entire file content into memory. This can be problematic
// for large files as it may consume excessive memory. This limitation should be considered
// when working with large files.
package drivefs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

const MimeTypeDriveGoogleAppsFolder = "application/vnd.google-apps.folder"
const MimeTypePrefixGoogleApps = "application/vnd.google-apps."

// DriveFS implements fs.FS for Google Drive.
// It provides a read-only filesystem view of Google Drive contents.
type DriveFS struct {
	service *drive.Service
	rootID  string // ID of the root folder (default: "root")
}

// Verify interface implementations at compile time.
var _ fs.ReadDirFS = (*DriveFS)(nil)

// New creates a new DriveFS instance with the given drive.Service.
// The service should be authenticated with appropriate scopes for reading files.
// The rootID specifies the ID of a drive or a root folder.
func New(service *drive.Service, rootID string) *DriveFS {
	return &DriveFS{
		service: service,
		rootID:  rootID,
	}
}

// Open opens the named file from Google Drive using a background context.
// The name must be an absolute path (cannot be ".").
// For context control, use OpenContext instead.
func (d *DriveFS) Open(name string) (fs.File, error) {
	return d.OpenContext(context.Background(), name)
}

// OpenContext opens the named file from Google Drive with the given context.
// The name must be an absolute path (must start with '/') and must not contain
// any '.' or '..' components.
func (d *DriveFS) OpenContext(ctx context.Context, name string) (fs.File, error) {
	path, err := newPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	// Resolve the path to a file ID
	fileID, err := d.resolveFileIDFromPath(ctx, path)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	// Get file metadata
	file, err := d.service.Files.Get(fileID).
		Context(ctx).
		Fields("id,name,mimeType,size,modifiedTime,createdTime").
		Do()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	// Check if it's a directory (folder)
	if file.MimeType == MimeTypeDriveGoogleAppsFolder {
		return d.openDir(ctx, file)
	}

	// Check if it's a directory (folder)
	if strings.HasPrefix(file.MimeType, MimeTypePrefixGoogleApps) {
		return d.openDir(ctx, file)
	}

	return d.openFile(ctx, file)
}

// ReadDir reads the named directory and returns a list of directory entries
// using a background context.
// The name must be an absolute path (must start with '/') and must not contain
// any '.' or '..' components.
// For context control, use ReadDirContext instead.
func (d *DriveFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return d.ReadDirContext(context.Background(), name)
}

// ReadDirContext reads the named directory and returns a list of directory entries
// with the given context.
// The name must be an absolute path (must start with '/') and must not contain
// any '.' or '..' components.
func (d *DriveFS) ReadDirContext(ctx context.Context, name string) ([]fs.DirEntry, error) {
	path, err := newPath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	folderID, err := d.resolveFileIDFromPath(ctx, path)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	entries, err := d.listDir(ctx, folderID)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	return entries, nil
}

// validateAbsolutePath ensures the provided path is absolute (starts with '/')
// and does not contain any relative components like '.' or '..' or empty segments
// (except the leading empty segment caused by the starting '/'). The root path
// of "/" is considered valid.
func validateAbsolutePath(name string) error {
	if name == "" {
		return fmt.Errorf("empty path")
	}
	if !strings.HasPrefix(name, "/") {
		return fmt.Errorf("path must be absolute and start with '/'")
	}
	if name == "/" {
		return nil
	}

	parts := strings.Split(name, "/")
	for i, p := range parts {
		if i == 0 {
			// leading empty element caused by the initial '/'
			continue
		}
		if p == "" {
			// disallow '//' anywhere in the path
			return fmt.Errorf("invalid empty path element")
		}
		if p == "." || p == ".." {
			return fmt.Errorf("relative path components are not allowed")
		}
	}
	return nil
}

// newPath splits a path into its components and returns them as a slice.
func newPath(name string) (path []string, err error) {
	if err := validateAbsolutePath(name); err != nil {
		return nil, err
	}
	for _, v := range strings.Split(name, "/") {
		if v != "" {
			path = append(path, v)
		}
	}
	return
}

// resolveFileIDFromPath resolves a path to a Google Drive file ID.
func (d *DriveFS) resolveFileIDFromPath(ctx context.Context, path []string) (id string, err error) {
	currentID := d.rootID
	for _, part := range path {
		// Search for the file in the current folder
		query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false", escapeQuery(part), currentID)

		fileList, err := d.service.Files.List().
			Context(ctx).
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Q(query).
			Fields("files(id,name,mimeType)").
			Do()
		if err != nil {
			return "", err
		}

		if len(fileList.Files) == 0 {
			return "", fs.ErrNotExist
		}

		currentID = fileList.Files[0].Id
	}

	return currentID, nil
}

// listDir lists the contents of a directory.
func (d *DriveFS) listDir(ctx context.Context, folderID string) (entries []fs.DirEntry, err error) {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	err = d.service.Files.List().
		Context(ctx).
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Q(query).
		Fields("nextPageToken,files(id,name,mimeType,size,modifiedTime)").
		Pages(ctx, func(page *drive.FileList) error {
			for _, file := range page.Files {
				entries = append(entries, &DriveDirEntry{file: file})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// openDir opens a directory for reading.
func (d *DriveFS) openDir(ctx context.Context, file *drive.File) (*DriveDir, error) {
	entries, err := d.listDir(ctx, file.Id)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: file.Name, Err: err}
	}

	return &DriveDir{
		file:    file,
		entries: entries,
	}, nil
}

// openFile opens a file for reading.
// Note: This method loads the entire file content into memory.
func (d *DriveFS) openFile(ctx context.Context, file *drive.File) (*DriveFile, error) {
	if strings.HasPrefix(file.MimeType, MimeTypePrefixGoogleApps) {
		return &DriveFile{
			file:    file,
			content: bytes.NewReader(nil),
		}, nil
	}

	// Download file content
	resp, err := d.service.Files.Get(file.Id).
		Context(ctx).
		Download()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: file.Name, Err: err}
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &fs.PathError{Op: "read", Path: file.Name, Err: err}
	}

	return &DriveFile{
		file:    file,
		content: bytes.NewReader(content),
	}, nil
}

// escapeQuery escapes special characters in a query string.
func escapeQuery(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", `\'`)
	return s
}

// parseModTime parses a modification time string in RFC3339 format.
func parseModTime(modifiedTime string) (time.Time, error) {
	return time.Parse(time.RFC3339, modifiedTime)
}
