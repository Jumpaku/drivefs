package drivefs

import (
	"strings"
	"time"
)

const (
	mimeTypeGoogleAppFolder = "application/vnd.google-apps.folder"
	mimeTypePrefixGoogleApp = "application/vnd.google-apps."
)

type FileInfo struct {
	Name    string
	ID      string
	Size    int64
	Mime    string
	ModTime time.Time
}

func (i FileInfo) IsFolder() bool {
	return i.Mime == mimeTypeGoogleAppFolder
}

func (i FileInfo) IsAppFile() bool {
	return strings.HasPrefix(i.Mime, mimeTypePrefixGoogleApp)
}
