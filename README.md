# drivefs

[![Go Reference](https://pkg.go.dev/badge/github.com/Jumpaku/drivefs.svg)](https://pkg.go.dev/github.com/Jumpaku/drivefs)
[![License: BSD-2-Clause](https://img.shields.io/badge/License-BSD_2--Clause-blue.svg)](https://opensource.org/licenses/BSD-2-Clause)

A Go module that implements standard Go filesystem interfaces for Google Drive.

## Overview

`drivefs` provides a read-only filesystem interface for accessing Google Drive contents using the standard Go `fs` package interfaces. It wraps the `google.golang.org/api/drive/v3` package to offer a familiar filesystem-like experience.

### Implemented Interfaces

- `fs.FS` - Filesystem interface for opening files
- `fs.ReadDirFS` - Interface for reading directory contents
- `fs.File` - Interface for file operations (read, stat, close)
- `fs.ReadDirFile` - Interface for reading directory entries
- `fs.DirEntry` - Interface for directory entry information
- `fs.FileInfo` - Interface for file metadata

## Installation

```bash
go get github.com/Jumpaku/drivefs
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Jumpaku/drivefs"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Create a drive.Service (authentication setup required)
	// Example: Load credentials from a JSON file
	// creds, err := google.CredentialsFromJSON(ctx, jsonContent, drive.DriveReadonlyScope)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// See: https://developers.google.com/drive/api/v3/quickstart/go
	service, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new DriveFS instance
	driveFS := drivefs.New(service)

	// Open a file (uses background context)
	file, err := driveFS.Open("path/to/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Or use OpenContext for context control
	file, err = driveFS.OpenContext(ctx, "path/to/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read file contents
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(content))

	// Read directory contents
	entries, err := driveFS.ReadDir("path/to/directory")
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Printf("%s (dir: %v)\n", entry.Name(), entry.IsDir())
	}
}
```

### Using a Different Root Folder

By default, `DriveFS` uses the user's root folder ("My Drive"). You can specify a different root folder by its ID:

```go
// Use a specific folder as the root
driveFS := drivefs.New(service).WithRootID("your-folder-id")
```

## API Reference

### DriveFS

The main filesystem type that implements `fs.FS` and `fs.ReadDirFS`.

- `New(service *drive.Service) *DriveFS` - Creates a new DriveFS instance
- `WithRootID(rootID string) *DriveFS` - Returns a copy with a different root folder ID (shares the same service)
- `Open(name string) (fs.File, error)` - Opens a file or directory (uses background context)
- `OpenContext(ctx context.Context, name string) (fs.File, error)` - Opens a file or directory with context
- `ReadDir(name string) ([]fs.DirEntry, error)` - Reads directory contents (uses background context)
- `ReadDirContext(ctx context.Context, name string) ([]fs.DirEntry, error)` - Reads directory contents with context

### DriveFile

Implements `fs.File` for regular files.

- `Read(b []byte) (int, error)` - Reads file content
- `Stat() (fs.FileInfo, error)` - Returns file information
- `Close() error` - Closes the file

### DriveDir

Implements `fs.File` and `fs.ReadDirFile` for directories.

- `ReadDir(n int) ([]fs.DirEntry, error)` - Reads directory entries
- `Stat() (fs.FileInfo, error)` - Returns directory information
- `Close() error` - Closes the directory

### DriveDirEntry

Implements `fs.DirEntry` for directory entries.

- `Name() string` - Returns the entry name
- `IsDir() bool` - Returns true if the entry is a directory
- `Type() fs.FileMode` - Returns the file mode bits
- `Info() (fs.FileInfo, error)` - Returns the file info

### DriveFileInfo

Implements `fs.FileInfo` for file metadata.

- `Name() string` - Returns the base name
- `Size() int64` - Returns the file size in bytes
- `Mode() fs.FileMode` - Returns the file mode
- `ModTime() time.Time` - Returns the modification time
- `IsDir() bool` - Returns true if it's a directory
- `Sys() any` - Returns nil (no underlying data)

## Authentication

This package requires an authenticated `drive.Service`. You'll need to:

1. Create a Google Cloud project
2. Enable the Google Drive API
3. Create OAuth 2.0 credentials
4. Authenticate and create a `drive.Service`

See the [Google Drive API Go Quickstart](https://developers.google.com/drive/api/v3/quickstart/go) for detailed instructions.

## License

BSD 2-Clause License. See [LICENSE](LICENSE) for details.