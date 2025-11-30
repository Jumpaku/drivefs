package drivefs

import (
	"fmt"
	"io"
	"io/fs"
	"sync"

	"google.golang.org/api/drive/v3"
)

// DriveDir implements fs.File and fs.ReadDirFile for a Google Drive directory.
// DriveDir's ReadDir method is protected by a mutex for concurrent use.
type DriveDir struct {
	file    *drive.File
	entries []fs.DirEntry
	offset  int
	mu      sync.Mutex
}

// Verify interface implementations at compile time.
var _ fs.ReadDirFile = (*DriveDir)(nil)

// Stat returns the directory info.
func (d *DriveDir) Stat() (fs.FileInfo, error) {
	modTime, err := parseModTime(d.file.ModifiedTime)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: d.file.Name, Err: fmt.Errorf("invalid modification time: %w", err)}
	}

	return &DriveFileInfo{
		file:    d.file,
		modTime: modTime,
	}, nil
}

// Read returns an error because directories cannot be read.
func (d *DriveDir) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.file.Name, Err: fs.ErrInvalid}
}

// Close closes the directory.
func (d *DriveDir) Close() error {
	return nil
}

// ReadDir reads the directory entries.
func (d *DriveDir) ReadDir(n int) ([]fs.DirEntry, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

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
