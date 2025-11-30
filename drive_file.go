package drivefs

import (
	"bytes"
	"fmt"
	"io/fs"

	"google.golang.org/api/drive/v3"
)

// DriveFile implements fs.File for a Google Drive file.
type DriveFile struct {
	file    *drive.File
	content *bytes.Reader
}

// Verify interface implementation at compile time.
var _ fs.File = (*DriveFile)(nil)

// Stat returns the file info.
func (f *DriveFile) Stat() (fs.FileInfo, error) {
	modTime, err := parseModTime(f.file.ModifiedTime)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: f.file.Name, Err: fmt.Errorf("invalid modification time: %w", err)}
	}

	return &DriveFileInfo{
		file:    f.file,
		modTime: modTime,
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
