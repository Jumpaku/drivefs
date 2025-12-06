# drivefs

[![Go Reference](https://pkg.go.dev/badge/github.com/Jumpaku/drivefs.svg)](https://pkg.go.dev/github.com/Jumpaku/drivefs)
[![License: BSD-2-Clause](https://img.shields.io/badge/License-BSD_2--Clause-blue.svg)](https://opensource.org/licenses/BSD-2-Clause)

A Go module that provides a simple API wrapper for Google Drive operations.

## Overview

`drivefs` provides a convenient Go API for managing files and directories in Google Drive. It wraps the `google.golang.org/api/drive/v3` package and supports operations like creating, reading, writing, moving, and deleting files and directories.

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
	// creds, err := google.CredentialsFromJSON(ctx, jsonContent, drive.DriveScope)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// See: https://developers.google.com/drive/api/v3/quickstart/go
	service, err := drive.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new DriveFS instance with a root folder ID
	// Use "root" for the user's root folder ("My Drive")
	driveFS := drivefs.New(service, "root")

	// Create a directory structure
	dirInfo, err := driveFS.MkdirAll("/path/to/directory")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created directory: %s (ID: %s)\n", dirInfo.Name, dirInfo.ID)

	// Create a file in the directory
	fileInfo, err := driveFS.Create(dirInfo.ID, "example.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the file
	err = driveFS.WriteFile(fileInfo.ID, []byte("Hello, Google Drive!"))
	if err != nil {
		log.Fatal(err)
	}

	// Read the file content
	data, err := driveFS.ReadFile(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// List directory contents
	entries, err := driveFS.ReadDir(dirInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Printf("%s (folder: %v, ID: %s)\n", entry.Name, entry.IsFolder(), entry.ID)
	}

	// Resolve a path to get FileInfo
	resolvedInfo, err := driveFS.ResolveFileID("/path/to/directory/example.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Resolved: %s\n", resolvedInfo.Name)

	// Get the full path from a file ID
	path, err := driveFS.ResolvePath(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Path: %s\n", path)

	// Copy a file to a different directory
	newParentInfo, err := driveFS.MkdirAll("/new/location")
	if err != nil {
		log.Fatal(err)
	}
	copiedFileInfo, err := driveFS.Copy(fileInfo.ID, newParentInfo.ID, "example_copy.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Copied file: %s (ID: %s)\n", copiedFileInfo.Name, copiedFileInfo.ID)

	// Rename a file
	renamedFileInfo, err := driveFS.Rename(fileInfo.ID, "renamed_example.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Renamed file: %s\n", renamedFileInfo.Name)

	// Move a file to a different directory
	err = driveFS.Move(fileInfo.ID, newParentInfo.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Delete a file (move to trash)
	err = driveFS.Remove(fileInfo.ID, true)
	if err != nil {
		log.Fatal(err)
	}
}
```

## API Reference

### DriveFS

The main type for interacting with Google Drive.

#### Constructor

- `New(service *drive.Service, rootID FileID) *DriveFS` - Creates a new DriveFS instance with the specified root folder ID.

#### Directory Operations

- `MkdirAll(path Path) (FileInfo, error)` - Creates all directories along the given path if they do not already exist, and returns the FileInfo of the last created directory.
- `Mkdir(parentID FileID, name string) (FileInfo, error)` - Creates a single directory with the given name under the specified parent directory.

#### File Operations

- `Create(parentID FileID, name string) (FileInfo, error)` - Creates a new file in the specified parent directory. If the file already exists, it will be truncated.
- `ReadFile(fileID FileID) ([]byte, error)` - Reads and returns the entire contents of a file.
- `WriteFile(fileID FileID, data []byte) error` - Writes data to an existing file, overwriting its contents.

#### Metadata and Navigation

- `Stat(fileID FileID) (FileInfo, error)` - Returns the FileInfo for the file or directory with the given ID.
- `ReadDir(fileID FileID) ([]FileInfo, error)` - Lists all files and directories within the specified directory.
- `ResolveFileID(path Path) (FileInfo, error)` - Resolves an absolute path (relative to the root) and returns the corresponding FileInfo.
- `ResolvePath(fileID FileID) (Path, error)` - Returns the absolute path from the root directory to the file or directory with the given ID.

#### File System Manipulation

- `Copy(fileID, newParentID FileID, newName string) (FileInfo, error)` - Creates a copy of the file in the specified new parent directory with the provided new name.
- `Rename(fileID FileID, newName string) (FileInfo, error)` - Renames the file to the specified new name.
- `Move(fileID, newParentID FileID) error` - Moves a file or directory to a new parent directory.
- `Remove(fileID FileID, trash bool) error` - Removes a file or empty directory. If `trash` is true, the item is moved to trash; otherwise, it is permanently deleted. Returns an error if trying to remove a non-empty directory.
- `RemoveAll(fileID FileID, trash bool) error` - Removes a file or directory and all its contents. If `trash` is true, items are moved to trash; otherwise, they are permanently deleted.

#### Tree Walking

- `Walk(fileID FileID, f func(FileInfo) error) error` - Walks the file tree rooted at the specified file or directory, calling the provided function for each item (including the root).

### Types

#### FileID

`type FileID string`

A unique identifier for a file or directory in Google Drive.

#### Path

`type Path string`

An absolute path string that must start with `/`. Relative path components (`.` and `..`) are not allowed.

#### FileInfo

```go
type FileInfo struct {
    Name    string
    ID      FileID
    Size    int64
    Mime    string
    ModTime time.Time
}
```

Contains metadata about a file or directory.

**Methods:**

- `IsFolder() bool` - Returns true if the item is a folder.
- `IsAppFile() bool` - Returns true if the item is a Google Apps file (e.g., Google Docs, Sheets, etc.).

## Authentication

This package requires an authenticated `drive.Service`. You'll need to:

1. Create a Google Cloud project
2. Enable the Google Drive API
3. Create OAuth 2.0 credentials
4. Authenticate and create a `drive.Service`

See the [Google Drive API Go Quickstart](https://developers.google.com/drive/api/v3/quickstart/go) for detailed instructions.

**Note:** For full read-write access, use `drive.DriveScope`. For read-only access, use `drive.DriveReadonlyScope`.

## License

BSD 2-Clause License. See [LICENSE](LICENSE) for details.