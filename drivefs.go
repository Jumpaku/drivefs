// Package drivefs provides a read-only filesystem interface for Google Drive.
// It implements standard Go filesystem interfaces (fs.FS, fs.File, fs.DirEntry, fs.FileInfo)
// for accessing Google Drive contents using the google.golang.org/api/drive/v3 package.
package drivefs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

// DriveFS implements fs.FS for Google Drive.
// It provides a read-only filesystem view of Google Drive contents.
type DriveFS struct {
	service *drive.Service
	ctx     context.Context
	rootID  string // ID of the root folder (default: "root")
}

// Verify interface implementations at compile time.
var (
	_ fs.FS        = (*DriveFS)(nil)
	_ fs.ReadDirFS = (*DriveFS)(nil)
)

// New creates a new DriveFS instance with the given drive.Service.
// The service should be authenticated with appropriate scopes for reading files.
func New(ctx context.Context, service *drive.Service) *DriveFS {
	return &DriveFS{
		service: service,
		ctx:     ctx,
		rootID:  "root",
	}
}

// WithRootID returns a copy of DriveFS with a different root folder ID.
func (dfs *DriveFS) WithRootID(rootID string) *DriveFS {
	return &DriveFS{
		service: dfs.service,
		ctx:     dfs.ctx,
		rootID:  rootID,
	}
}

// Open opens the named file from Google Drive.
// The name should be a slash-separated path relative to the root folder.
func (dfs *DriveFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	// Handle root directory
	if name == "." {
		return dfs.openDir(dfs.rootID, ".")
	}

	// Resolve the path to a file ID
	fileID, err := dfs.resolvePath(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	// Get file metadata
	file, err := dfs.service.Files.Get(fileID).
		Context(dfs.ctx).
		Fields("id, name, mimeType, size, modifiedTime, createdTime").
		Do()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	// Check if it's a directory (folder)
	if file.MimeType == "application/vnd.google-apps.folder" {
		return dfs.openDir(fileID, name)
	}

	return dfs.openFile(file, name)
}

// ReadDir reads the named directory and returns a list of directory entries.
func (dfs *DriveFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}

	var folderID string
	if name == "." {
		folderID = dfs.rootID
	} else {
		var err error
		folderID, err = dfs.resolvePath(name)
		if err != nil {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
		}
	}

	return dfs.listDir(folderID)
}

// resolvePath resolves a path to a Google Drive file ID.
func (dfs *DriveFS) resolvePath(name string) (string, error) {
	parts := strings.Split(name, "/")
	currentID := dfs.rootID

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// Search for the file in the current folder
		query := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false",
			escapeQuery(part), currentID)

		fileList, err := dfs.service.Files.List().
			Context(dfs.ctx).
			Q(query).
			Fields("files(id, name, mimeType)").
			PageSize(1).
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
func (dfs *DriveFS) listDir(folderID string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	var pageToken string

	for {
		query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
		call := dfs.service.Files.List().
			Context(dfs.ctx).
			Q(query).
			Fields("nextPageToken, files(id, name, mimeType, size, modifiedTime)").
			OrderBy("name").
			PageSize(100)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		fileList, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, f := range fileList.Files {
			entries = append(entries, &DriveDirEntry{file: f})
		}

		pageToken = fileList.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return entries, nil
}

// openDir opens a directory for reading.
func (dfs *DriveFS) openDir(folderID, name string) (*DriveDir, error) {
	entries, err := dfs.listDir(folderID)
	if err != nil {
		return nil, err
	}

	return &DriveDir{
		name:    path.Base(name),
		entries: entries,
	}, nil
}

// openFile opens a file for reading.
func (dfs *DriveFS) openFile(file *drive.File, name string) (*DriveFile, error) {
	// Download file content
	resp, err := dfs.service.Files.Get(file.Id).
		Context(dfs.ctx).
		Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	modTime, _ := time.Parse(time.RFC3339, file.ModifiedTime)

	return &DriveFile{
		name:    path.Base(name),
		content: bytes.NewReader(content),
		size:    file.Size,
		modTime: modTime,
	}, nil
}

// escapeQuery escapes special characters in a query string.
func escapeQuery(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// DriveFile implements fs.File for a Google Drive file.
type DriveFile struct {
	name    string
	content *bytes.Reader
	size    int64
	modTime time.Time
}

// Verify interface implementation at compile time.
var _ fs.File = (*DriveFile)(nil)

// Stat returns the file info.
func (f *DriveFile) Stat() (fs.FileInfo, error) {
	return &DriveFileInfo{
		name:    f.name,
		size:    f.size,
		modTime: f.modTime,
		isDir:   false,
	}, nil
}

// Read reads from the file.
func (f *DriveFile) Read(b []byte) (int, error) {
	return f.content.Read(b)
}

// Close closes the file.
func (f *DriveFile) Close() error {
	return nil
}

// DriveDir implements fs.File and fs.ReadDirFile for a Google Drive directory.
type DriveDir struct {
	name    string
	entries []fs.DirEntry
	offset  int
}

// Verify interface implementations at compile time.
var (
	_ fs.File        = (*DriveDir)(nil)
	_ fs.ReadDirFile = (*DriveDir)(nil)
)

// Stat returns the directory info.
func (d *DriveDir) Stat() (fs.FileInfo, error) {
	return &DriveFileInfo{
		name:  d.name,
		isDir: true,
	}, nil
}

// Read returns an error because directories cannot be read.
func (d *DriveDir) Read(b []byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: fs.ErrInvalid}
}

// Close closes the directory.
func (d *DriveDir) Close() error {
	return nil
}

// ReadDir reads the directory entries.
func (d *DriveDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if n <= 0 {
		entries := d.entries[d.offset:]
		d.offset = len(d.entries)
		return entries, nil
	}

	if d.offset >= len(d.entries) {
		return nil, io.EOF
	}

	end := d.offset + n
	if end > len(d.entries) {
		end = len(d.entries)
	}

	entries := d.entries[d.offset:end]
	d.offset = end

	if d.offset >= len(d.entries) {
		return entries, io.EOF
	}
	return entries, nil
}

// DriveDirEntry implements fs.DirEntry for a Google Drive file or folder.
type DriveDirEntry struct {
	file *drive.File
}

// Verify interface implementation at compile time.
var _ fs.DirEntry = (*DriveDirEntry)(nil)

// Name returns the name of the entry.
func (e *DriveDirEntry) Name() string {
	return e.file.Name
}

// IsDir reports whether the entry is a directory.
func (e *DriveDirEntry) IsDir() bool {
	return e.file.MimeType == "application/vnd.google-apps.folder"
}

// Type returns the file mode bits.
func (e *DriveDirEntry) Type() fs.FileMode {
	if e.IsDir() {
		return fs.ModeDir
	}
	return 0
}

// Info returns the file info.
func (e *DriveDirEntry) Info() (fs.FileInfo, error) {
	modTime, _ := time.Parse(time.RFC3339, e.file.ModifiedTime)
	return &DriveFileInfo{
		name:    e.file.Name,
		size:    e.file.Size,
		modTime: modTime,
		isDir:   e.IsDir(),
	}, nil
}

// DriveFileInfo implements fs.FileInfo for a Google Drive file or folder.
type DriveFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

// Verify interface implementation at compile time.
var _ fs.FileInfo = (*DriveFileInfo)(nil)

// Name returns the base name of the file.
func (fi *DriveFileInfo) Name() string {
	return fi.name
}

// Size returns the size of the file in bytes.
func (fi *DriveFileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file mode bits.
func (fi *DriveFileInfo) Mode() fs.FileMode {
	if fi.isDir {
		return fs.ModeDir | 0555
	}
	return 0444
}

// ModTime returns the modification time.
func (fi *DriveFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir reports whether the file is a directory.
func (fi *DriveFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns the underlying data source (always nil for DriveFileInfo).
func (fi *DriveFileInfo) Sys() any {
	return nil
}
