package drivefs

import (
	"bytes"
	"io/fs"
	"testing"
	"time"

	"google.golang.org/api/drive/v3"
)

// TestDriveFileInfo tests the DriveFileInfo implementation.
func TestDriveFileInfo(t *testing.T) {
	modTime := time.Now()
	fi := &DriveFileInfo{
		name:    "test.txt",
		size:    1024,
		modTime: modTime,
		isDir:   false,
	}

	if fi.Name() != "test.txt" {
		t.Errorf("Name() = %q, want %q", fi.Name(), "test.txt")
	}

	if fi.Size() != 1024 {
		t.Errorf("Size() = %d, want %d", fi.Size(), 1024)
	}

	if fi.Mode() != 0444 {
		t.Errorf("Mode() = %v, want %v", fi.Mode(), fs.FileMode(0444))
	}

	if !fi.ModTime().Equal(modTime) {
		t.Errorf("ModTime() = %v, want %v", fi.ModTime(), modTime)
	}

	if fi.IsDir() {
		t.Error("IsDir() = true, want false")
	}

	if fi.Sys() != nil {
		t.Error("Sys() != nil, want nil")
	}
}

// TestDriveFileInfoDir tests the DriveFileInfo implementation for directories.
func TestDriveFileInfoDir(t *testing.T) {
	fi := &DriveFileInfo{
		name:  "testdir",
		isDir: true,
	}

	if fi.Name() != "testdir" {
		t.Errorf("Name() = %q, want %q", fi.Name(), "testdir")
	}

	if fi.Size() != 0 {
		t.Errorf("Size() = %d, want %d", fi.Size(), 0)
	}

	expectedMode := fs.ModeDir | 0555
	if fi.Mode() != expectedMode {
		t.Errorf("Mode() = %v, want %v", fi.Mode(), expectedMode)
	}

	if !fi.IsDir() {
		t.Error("IsDir() = false, want true")
	}
}

// TestDriveFileRead tests the DriveFile Read implementation.
func TestDriveFileRead(t *testing.T) {
	content := []byte("Hello, World!")
	f := &DriveFile{
		name:    "hello.txt",
		content: bytes.NewReader(content),
		size:    int64(len(content)),
		modTime: time.Now(),
	}

	buf := make([]byte, len(content))
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != len(content) {
		t.Errorf("Read() = %d, want %d", n, len(content))
	}
	if string(buf) != string(content) {
		t.Errorf("Read() content = %q, want %q", string(buf), string(content))
	}
}

// TestDriveFileStat tests the DriveFile Stat implementation.
func TestDriveFileStat(t *testing.T) {
	modTime := time.Now()
	f := &DriveFile{
		name:    "test.txt",
		size:    100,
		modTime: modTime,
	}

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if fi.Name() != "test.txt" {
		t.Errorf("Stat().Name() = %q, want %q", fi.Name(), "test.txt")
	}

	if fi.Size() != 100 {
		t.Errorf("Stat().Size() = %d, want %d", fi.Size(), 100)
	}

	if fi.IsDir() {
		t.Error("Stat().IsDir() = true, want false")
	}
}

// TestDriveFileClose tests the DriveFile Close implementation.
func TestDriveFileClose(t *testing.T) {
	f := &DriveFile{name: "test.txt"}
	if err := f.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestDriveDirStat tests the DriveDir Stat implementation.
func TestDriveDirStat(t *testing.T) {
	d := &DriveDir{name: "testdir"}

	fi, err := d.Stat()
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if fi.Name() != "testdir" {
		t.Errorf("Stat().Name() = %q, want %q", fi.Name(), "testdir")
	}

	if !fi.IsDir() {
		t.Error("Stat().IsDir() = false, want true")
	}
}

// TestDriveDirReadDir tests the DriveDir ReadDir implementation.
func TestDriveDirReadDir(t *testing.T) {
	entries := []fs.DirEntry{
		&DriveDirEntry{file: &drive.File{Name: "file1.txt", MimeType: "text/plain"}},
		&DriveDirEntry{file: &drive.File{Name: "file2.txt", MimeType: "text/plain"}},
		&DriveDirEntry{file: &drive.File{Name: "subdir", MimeType: "application/vnd.google-apps.folder"}},
	}

	d := &DriveDir{
		name:    "testdir",
		entries: entries,
	}

	// Read all entries
	result, err := d.ReadDir(-1)
	if err != nil {
		t.Fatalf("ReadDir(-1) error = %v", err)
	}
	if len(result) != 3 {
		t.Errorf("ReadDir(-1) returned %d entries, want 3", len(result))
	}
}

// TestDriveDirReadDirN tests the DriveDir ReadDir implementation with n > 0.
func TestDriveDirReadDirN(t *testing.T) {
	entries := []fs.DirEntry{
		&DriveDirEntry{file: &drive.File{Name: "file1.txt", MimeType: "text/plain"}},
		&DriveDirEntry{file: &drive.File{Name: "file2.txt", MimeType: "text/plain"}},
		&DriveDirEntry{file: &drive.File{Name: "file3.txt", MimeType: "text/plain"}},
	}

	d := &DriveDir{
		name:    "testdir",
		entries: entries,
	}

	// Read 2 entries
	result, err := d.ReadDir(2)
	if err != nil {
		t.Fatalf("ReadDir(2) error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ReadDir(2) returned %d entries, want 2", len(result))
	}

	// Read remaining entry
	result, err = d.ReadDir(2)
	if err == nil {
		t.Error("ReadDir(2) error = nil, want io.EOF")
	}
	if len(result) != 1 {
		t.Errorf("ReadDir(2) returned %d entries, want 1", len(result))
	}
}

// TestDriveDirEntryName tests the DriveDirEntry Name implementation.
func TestDriveDirEntryName(t *testing.T) {
	e := &DriveDirEntry{file: &drive.File{Name: "test.txt"}}
	if e.Name() != "test.txt" {
		t.Errorf("Name() = %q, want %q", e.Name(), "test.txt")
	}
}

// TestDriveDirEntryIsDir tests the DriveDirEntry IsDir implementation.
func TestDriveDirEntryIsDir(t *testing.T) {
	// File
	e := &DriveDirEntry{file: &drive.File{Name: "test.txt", MimeType: "text/plain"}}
	if e.IsDir() {
		t.Error("IsDir() = true, want false for file")
	}

	// Directory
	e = &DriveDirEntry{file: &drive.File{Name: "dir", MimeType: "application/vnd.google-apps.folder"}}
	if !e.IsDir() {
		t.Error("IsDir() = false, want true for folder")
	}
}

// TestDriveDirEntryType tests the DriveDirEntry Type implementation.
func TestDriveDirEntryType(t *testing.T) {
	// File
	e := &DriveDirEntry{file: &drive.File{Name: "test.txt", MimeType: "text/plain"}}
	if e.Type() != 0 {
		t.Errorf("Type() = %v, want 0 for file", e.Type())
	}

	// Directory
	e = &DriveDirEntry{file: &drive.File{Name: "dir", MimeType: "application/vnd.google-apps.folder"}}
	if e.Type() != fs.ModeDir {
		t.Errorf("Type() = %v, want %v for folder", e.Type(), fs.ModeDir)
	}
}

// TestDriveDirEntryInfo tests the DriveDirEntry Info implementation.
func TestDriveDirEntryInfo(t *testing.T) {
	e := &DriveDirEntry{file: &drive.File{
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         512,
		ModifiedTime: "2024-01-15T10:30:00Z",
	}}

	fi, err := e.Info()
	if err != nil {
		t.Fatalf("Info() error = %v", err)
	}

	if fi.Name() != "test.txt" {
		t.Errorf("Info().Name() = %q, want %q", fi.Name(), "test.txt")
	}

	if fi.Size() != 512 {
		t.Errorf("Info().Size() = %d, want %d", fi.Size(), 512)
	}

	if fi.IsDir() {
		t.Error("Info().IsDir() = true, want false")
	}
}

// TestEscapeQuery tests the escapeQuery function.
func TestEscapeQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with'quote", "with\\'quote"},
		{"with\\backslash", "with\\\\backslash"},
		{"mixed'and\\special", "mixed\\'and\\\\special"},
	}

	for _, tt := range tests {
		result := escapeQuery(tt.input)
		if result != tt.expected {
			t.Errorf("escapeQuery(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestInterfaceCompliance verifies interface compliance at compile time.
func TestInterfaceCompliance(t *testing.T) {
	// This test ensures that our types implement the expected interfaces.
	// The actual verification is done at compile time with the var _ = statements.
	var _ fs.FS = (*DriveFS)(nil)
	var _ fs.ReadDirFS = (*DriveFS)(nil)
	var _ fs.File = (*DriveFile)(nil)
	var _ fs.File = (*DriveDir)(nil)
	var _ fs.ReadDirFile = (*DriveDir)(nil)
	var _ fs.DirEntry = (*DriveDirEntry)(nil)
	var _ fs.FileInfo = (*DriveFileInfo)(nil)
}
