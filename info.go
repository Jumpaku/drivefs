package drivefs

import (
	"strings"
	"time"
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
