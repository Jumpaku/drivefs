package drivefs

import (
	"fmt"
	"io/fs"

	"google.golang.org/api/drive/v3"
)

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
	return e.file.MimeType == MimeTypeDriveGoogleAppsFolder
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
	modTime, err := parseModTime(e.file.ModifiedTime)
	if err != nil {
		return nil, fmt.Errorf("invalid modification time for file %q: %w", e.file.Name, err)
	}
	return &DriveFileInfo{
		file:    e.file,
		modTime: modTime,
	}, nil
}
