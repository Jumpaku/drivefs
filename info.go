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

type FileID string

type FileInfo struct {
	Name           string
	ID             FileID
	Size           int64
	Mime           string
	ModTime        time.Time
	ShortcutTarget FileID
	WebViewLink    string
}

func (i FileInfo) IsFolder() bool {
	return i.Mime == mimeTypeGoogleAppFolder
}

func (i FileInfo) IsShortcut() bool {
	return i.Mime == mimeTypeGoogleAppShortcut
}

func (i FileInfo) IsAppFile() bool {
	return strings.HasPrefix(i.Mime, mimeTypePrefixGoogleApp)
}
