package drivefs

import (
	"strings"
	"time"
)

const (
	mimeTypeGoogleAppFolder   = "application/vnd.google-apps.folder"
	mimeTypeGoogleAppShortcut = "application/vnd.google-apps.shortcut"
	mimeTypePrefixGoogleApp   = "application/vnd.google-apps."
)

// FileID represents a unique identifier for a file or directory in Google Drive.
type FileID string

// FileInfo contains metadata about a file or directory in Google Drive.
type FileInfo struct {
	// Name is the file or directory name.
	Name string

	// ID is the unique identifier for this file or directory.
	ID FileID

	// Size is the file size in bytes. For directories and Google Apps files, this is 0.
	Size int64

	// Mime is the MIME type of the file.
	Mime string

	// ModTime is the last modification time.
	ModTime time.Time

	// ShortcutTarget is the ID of the target file if this is a shortcut, empty otherwise.
	ShortcutTarget FileID

	// WebViewLink is the URL to view the file in a web browser.
	WebViewLink string
}

// IsFolder returns true if this FileInfo represents a directory.
func (i FileInfo) IsFolder() bool {
	return i.Mime == mimeTypeGoogleAppFolder
}

// IsShortcut returns true if this FileInfo represents a shortcut.
func (i FileInfo) IsShortcut() bool {
	return i.Mime == mimeTypeGoogleAppShortcut
}

// IsAppFile returns true if this FileInfo represents a Google Apps file
// (e.g., Google Docs, Sheets, Slides).
func (i FileInfo) IsAppFile() bool {
	return strings.HasPrefix(i.Mime, mimeTypePrefixGoogleApp)
}
