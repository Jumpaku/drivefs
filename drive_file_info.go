package drivefs

import (
	"io/fs"
	"time"

	"google.golang.org/api/drive/v3"
)

// DriveFileInfo implements fs.FileInfo for a Google Drive file or folder.
type DriveFileInfo struct {
	file    *drive.File
	modTime time.Time
}

// Verify interface implementation at compile time.
var _ fs.FileInfo = (*DriveFileInfo)(nil)

// Name returns the base name of the file.
func (fi *DriveFileInfo) Name() string {
	return fi.file.Name
}

// Size returns the size of the file in bytes.
func (fi *DriveFileInfo) Size() int64 {
	return fi.file.Size
}

// Mode returns the file mode bits.
func (fi *DriveFileInfo) Mode() fs.FileMode {
	if fi.IsDir() {
		return fs.ModeDir | 0444
	}
	return 0444
}

// ModTime returns the modification time.
func (fi *DriveFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir reports whether the file is a directory.
func (fi *DriveFileInfo) IsDir() bool {
	return fi.file.MimeType == "application/vnd.google-apps.folder"
}

// Sys returns the underlying data source (*drive.File).
func (fi *DriveFileInfo) Sys() any {
	return fi.file
}
