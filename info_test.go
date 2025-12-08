package drivefs_test

import (
	"testing"

	drivefs "github.com/Jumpaku/go-drivefs"
)

func TestFileInfo_Types(t *testing.T) {
	cases := []struct {
		name     string
		mime     string
		isFolder bool
		isShort  bool
		isApp    bool
	}{
		{"folder", "application/vnd.google-apps.folder", true, false, true},
		{"shortcut", "application/vnd.google-apps.shortcut", false, true, true},
		{"app-file", "application/vnd.google-apps.document", false, false, true},
		{"plain", "text/plain", false, false, false},
		{"empty", "", false, false, false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			i := drivefs.FileInfo{Mime: c.mime}
			if got := i.IsFolder(); got != c.isFolder {
				t.Fatalf("IsFolder() = %v, want %v for mime %q", got, c.isFolder, c.mime)
			}
			if got := i.IsShortcut(); got != c.isShort {
				t.Fatalf("IsShortcut() = %v, want %v for mime %q", got, c.isShort, c.mime)
			}
			if got := i.IsAppFile(); got != c.isApp {
				t.Fatalf("IsAppFile() = %v, want %v for mime %q", got, c.isApp, c.mime)
			}
		})
	}
}
