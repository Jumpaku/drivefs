# drivefs

[![Go Reference](https://pkg.go.dev/badge/github.com/Jumpaku/drivefs.svg)](https://pkg.go.dev/github.com/Jumpaku/drivefs)
[![License: BSD-2-Clause](https://img.shields.io/badge/License-BSD_2--Clause-blue.svg)](https://opensource.org/licenses/BSD-2-Clause)

A Go module that provides a file system-like interface for Google Drive operations.

## Overview

`drivefs` provides a convenient and intuitive Go API for managing files and directories in Google Drive. It wraps the `google.golang.org/api/drive/v3` package and offers familiar filesystem operations such as creating, reading, writing, copying, renaming, moving, and deleting files and directories. The package fully supports both My Drive and Shared Drives.

## Installation

```bash
go get github.com/Jumpaku/drivefs
```

**Requirements:**
- Go 1.24 or later
- An authenticated Google Drive API `*drive.Service`

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/Jumpaku/drivefs"
    "google.golang.org/api/drive/v3"
)

func main() {
    // Assuming you have an authenticated drive.Service
    // See Authentication section for setup details
    var service *drive.Service // Your authenticated service
    
    // Create DriveFS instance
    driveFS, err := drivefs.New(service, "root")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a directory
    dirInfo, err := driveFS.MkdirAll("/my-project/data")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create and write a file
    fileInfo, err := driveFS.Create(dirInfo.ID, "notes.txt")
    if err != nil {
        log.Fatal(err)
    }
    
    err = driveFS.WriteFile(fileInfo.ID, []byte("Hello from drivefs!"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Read the file back
    data, err := driveFS.ReadFile(fileInfo.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(data)) // Output: Hello from drivefs!
}
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Jumpaku/drivefs"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Create a drive.Service (authentication setup required)
	// For authentication examples, see:
	// https://developers.google.com/drive/api/v3/quickstart/go
	service, err := drive.NewService(ctx /* add your auth options here */)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new DriveFS instance with a root folder ID
	// Use "root" or "" for the user's My Drive root, or provide a specific folder ID
	// for a Shared Drive or subdirectory
	driveFS, err := drivefs.New(service, "root")
	if err != nil {
		log.Fatal(err)
	}

	// Create a directory structure
	// MkdirAll creates all directories along the path if they don't exist
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
	fmt.Printf("Created file: %s (ID: %s)\n", fileInfo.Name, fileInfo.ID)

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
	fmt.Println("File content:", string(data))

	// Get file metadata
	info, err := driveFS.Info(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File: %s, Size: %d bytes, Modified: %v\n", info.Name, info.Size, info.ModTime)

	// List directory contents
	entries, err := driveFS.ReadDir(dirInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Printf("- %s (folder: %v, ID: %s)\n", entry.Name, entry.IsFolder(), entry.ID)
	}

	// Find files by path
	// Returns all matching files (multiple if duplicates exist at any level)
	resolvedInfos, err := driveFS.FindByPath("/path/to/directory/example.txt")
	if err != nil {
		log.Fatal(err)
	}
	for _, info := range resolvedInfos {
		fmt.Printf("Resolved: %s (ID: %s)\n", info.Name, info.ID)
	}

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
	fmt.Println("File moved successfully")

	// Walk the file tree
	// Traverses all files and directories recursively
	err = driveFS.Walk(dirInfo.ID, func(path drivefs.Path, info drivefs.FileInfo) error {
		fmt.Printf("%s: %s (ID: %s)\n", path, info.Name, info.ID)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Manage permissions
	// List current permissions
	permissions, err := driveFS.PermList(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, perm := range permissions {
		fmt.Printf("Permission ID: %s, Role: %s\n", perm.ID(), perm.Role())
	}

	// Grant read access to a user
	_, err = driveFS.PermSet(fileInfo.ID, drivefs.UserPermission("user@example.com", drivefs.RoleReader, false))
	if err != nil {
		log.Fatal(err)
	}

	// Grant write access to a group
	_, err = driveFS.PermSet(fileInfo.ID, drivefs.GroupPermission("group@example.com", drivefs.RoleWriter, true))
	if err != nil {
		log.Fatal(err)
	}

	// Grant read access to anyone with the link
	_, err = driveFS.PermSet(fileInfo.ID, drivefs.AnyonePermission(drivefs.RoleReader, false))
	if err != nil {
		log.Fatal(err)
	}

	// Remove a user's permission
	_, err = driveFS.PermDel(fileInfo.ID, drivefs.User("user@example.com"))
	if err != nil {
		log.Fatal(err)
	}

	// Remove a file (move to trash)
	err = driveFS.Remove(fileInfo.ID, true)
	if err != nil {
		log.Fatal(err)
	}

	// Permanently delete a file (skip trash)
	err = driveFS.Remove(copiedFileInfo.ID, false)
	if err != nil {
		log.Fatal(err)
	}

	// Remove a directory and all its contents
	err = driveFS.RemoveAll(dirInfo.ID, true)
	if err != nil {
		log.Fatal(err)
	}
}
```

## API Reference

### DriveFS

The main type for interacting with Google Drive.

#### Constructor

```go
func New(service *drive.Service, rootID FileID) (*DriveFS, error)
```

Creates a new DriveFS instance with the specified root folder ID.
- Use `"root"` or `""` (empty string) for the user's My Drive root
- Use a specific folder ID for a Shared Drive or to use a subdirectory as root
- Returns an error if the root directory cannot be accessed
- When `"root"` is provided, it is automatically resolved to the actual My Drive root ID

#### Directory Operations

```go
func (s *DriveFS) MkdirAll(path Path) (FileInfo, error)
```

Creates all directories along the given path if they do not already exist.
- Path must be absolute (start with `/`)
- Returns the FileInfo of the final directory in the path
- If multiple directories with the same name exist at any level, returns `ErrAlreadyExists`

```go
func (s *DriveFS) Mkdir(parentID FileID, name string) (FileInfo, error)
```

Creates a single directory with the given name under the specified parent.
- Creates a new directory even if one with the same name already exists (Google Drive allows duplicates)

#### File Operations

```go
func (s *DriveFS) Create(parentID FileID, name string) (FileInfo, error)
```

Creates a new empty file in the specified parent directory.
- Creates a new file even if one with the same name already exists (Google Drive allows duplicates)

```go
func (s *DriveFS) ReadFile(fileID FileID) ([]byte, error)
```

Reads and returns the entire contents of a file.
- Returns `ErrNotReadable` for Google Apps files (Docs, Sheets, Slides, etc.)
- For Google Apps files, use the Drive API export functionality instead

```go
func (s *DriveFS) WriteFile(fileID FileID, data []byte) error
```

Writes data to an existing file, completely replacing its contents.

```go
func (s *DriveFS) Shortcut(parentID FileID, name string, targetID FileID) (FileInfo, error)
```

Creates a shortcut (link) to another file or directory.
- `parentID`: The directory where the shortcut will be created
- `name`: The name of the shortcut
- `targetID`: The ID of the file or directory that the shortcut points to
- Returns the FileInfo of the created shortcut

#### Metadata and Navigation

```go
func (s *DriveFS) Info(fileID FileID) (FileInfo, error)
```

Returns the FileInfo (metadata) for the file or directory with the given ID.

```go
func (s *DriveFS) ReadDir(fileID FileID) ([]FileInfo, error)
```

Lists all files and directories directly within the specified directory.
- Does not include trashed items
- Returns only immediate children (not recursive)

```go
func (s *DriveFS) FindByPath(path Path) ([]FileInfo, error)
```

Resolves an absolute path (relative to the root) and returns all matching FileInfo objects.
- Path must be absolute (start with `/`)
- Returns multiple results if there are duplicate files/folders with the same name at any level
- Returns an empty slice if the path does not exist

```go
func (s *DriveFS) ResolvePath(fileID FileID) (Path, error)
```

Returns the absolute path from the root directory to the file or directory with the given ID.
- Returns `ErrMultiParentsNotSupported` if the file has multiple parents
- Path is relative to the DriveFS root

#### File System Manipulation

```go
func (s *DriveFS) Copy(fileID, newParentID FileID, newName string) (FileInfo, error)
```

Creates a copy of the file in the specified new parent directory with the provided new name.
- Creates a new file with a new ID
- Original file remains unchanged

```go
func (s *DriveFS) Rename(fileID FileID, newName string) (FileInfo, error)
```

Renames the file or directory to the specified new name.
- Does not change the file's parent or location

```go
func (s *DriveFS) Move(fileID, newParentID FileID) error
```

Moves a file or directory to a new parent directory.
- Removes all existing parents and sets the new parent
- Does not change the file's name

```go
func (s *DriveFS) Remove(fileID FileID, moveToTrash bool) error
```

Removes a file or directory.
- If `moveToTrash` is `true`, the item is moved to Google Drive trash
- If `moveToTrash` is `false`, the item is permanently deleted
- Returns `ErrNotRemovable` if attempting to remove a non-empty directory
- For non-empty directories, use `RemoveAll` instead

```go
func (s *DriveFS) RemoveAll(fileID FileID, moveToTrash bool) error
```

Removes a file or directory and all its contents recursively.
- If `moveToTrash` is `true`, items are moved to Google Drive trash
- If `moveToTrash` is `false`, items are permanently deleted
- Safe to use on both files and directories

#### Tree Walking

```go
func (s *DriveFS) Walk(fileID FileID, f func(Path, FileInfo) error) error
```

Walks the file tree rooted at the specified file or directory, calling the provided function for each item.
- Includes the root item itself
- The function receives both the path and FileInfo for each item
- If the callback function returns an error, walking stops and that error is returned

#### Permission Management

```go
func (s *DriveFS) PermList(fileID FileID) ([]Permission, error)
```

Lists all permissions for the file or directory with the given ID.
- Returns a slice of Permission objects representing all users, groups, domains, or anyone who has access
- Each Permission contains information about the grantee, role, and whether file discovery is allowed

```go
func (s *DriveFS) PermSet(fileID FileID, permission Permission) ([]Permission, error)
```

Sets or updates a permission for the file or directory with the given ID.
- If a permission for the specified grantee already exists, it will be updated
- If no permission exists for the grantee, a new one will be created
- Returns the updated list of all permissions for the file
- Use helper functions like `UserPermission()`, `GroupPermission()`, `DomainPermission()`, or `AnyonePermission()` to create Permission objects

```go
func (s *DriveFS) PermDel(fileID FileID, grantee Grantee) ([]Permission, error)
```

Deletes all permissions matching the specified grantee for the file or directory.
- Removes permissions for the specified user, group, domain, or anyone access
- Returns the updated list of remaining permissions for the file
- Use helper functions like `User()`, `Group()`, `Domain()`, or `Anyone()` to create Grantee objects

### Types

#### FileID

```go
type FileID string
```

A unique identifier for a file or directory in Google Drive. These IDs are assigned by Google Drive and remain stable across renames and moves.

#### Path

```go
type Path string
```

An absolute path string representing a location in the Drive filesystem.
- Must start with `/` (absolute path required)
- Relative path components (`.` and `..`) are not allowed
- Path separators are forward slashes (`/`)
- Example: `"/folder/subfolder/file.txt"`

#### FileInfo

```go
type FileInfo struct {
    Name           string    // File or directory name
    ID             FileID    // Unique Google Drive ID
    Size           int64     // File size in bytes (0 for directories)
    Mime           string    // MIME type (e.g., "text/plain", "application/vnd.google-apps.folder")
    ModTime        time.Time // Last modification time
    ShortcutTarget FileID    // Target file ID (for shortcuts only, empty otherwise)
    WebViewLink    string    // URL to view the file in the Google Drive web interface
}
```

Contains metadata about a file or directory.

**Methods:**

```go
func (i FileInfo) IsFolder() bool
```
Returns `true` if the item is a folder/directory.

```go
func (i FileInfo) IsShortcut() bool
```
Returns `true` if the item is a shortcut to another file or directory.
The target file ID can be found in the `ShortcutTarget` field.

```go
func (i FileInfo) IsAppFile() bool
```
Returns `true` if the item is a Google Apps file (e.g., Google Docs, Sheets, Slides).
Google Apps files cannot be read with `ReadFile()` and must be exported using the Drive API's export functionality.

#### Permission

```go
type Permission interface {
    ID() PermissionID
    Grantee() Grantee
    Role() Role
    AllowFileDiscovery() bool
}
```

Represents a permission granted to a user, group, domain, or anyone for a file or directory.

**Helper Functions to Create Permissions:**

```go
func UserPermission(email string, role Role, allowFileDiscovery bool) Permission
```
Creates a permission for a specific user identified by email address.

```go
func GroupPermission(email string, role Role, allowFileDiscovery bool) Permission
```
Creates a permission for a Google Group identified by email address.

```go
func DomainPermission(domain string, role Role, allowFileDiscovery bool) Permission
```
Creates a permission for an entire domain (e.g., "example.com").

```go
func AnyonePermission(role Role, allowFileDiscovery bool) Permission
```
Creates a permission that grants access to anyone with the link.

#### Grantee

```go
type Grantee interface {
    // Sealed interface - cannot be implemented outside the package
}
```

Represents the recipient of a permission (user, group, domain, or anyone).

**Helper Functions to Create Grantees:**

```go
func User(email string) Grantee
```
Creates a grantee representing a specific user.

```go
func Group(email string) Grantee
```
Creates a grantee representing a Google Group.

```go
func Domain(domain string) Grantee
```
Creates a grantee representing an entire domain.

```go
func Anyone() Grantee
```
Creates a grantee representing anyone with the link.

**Concrete Grantee Types:**

- `GranteeUser` - Represents a user with an `Email` field
- `GranteeGroup` - Represents a group with an `Email` field
- `GranteeDomain` - Represents a domain with a `Domain` field
- `GranteeAnyone` - Represents public access (anyone with the link)

#### Role

```go
type Role string
```

Represents the level of access granted by a permission.

**Available Roles:**

```go
const (
    RoleOwner         Role = "owner"         // Full ownership with ability to delete
    RoleOrganizer     Role = "organizer"     // Can organize files in Shared Drives
    RoleFileOrganizer Role = "fileOrganizer" // Can organize files
    RoleWriter        Role = "writer"        // Can edit files and add comments
    RoleCommenter     Role = "commenter"     // Can view and add comments
    RoleReader        Role = "reader"        // Can only view files
)
```

#### PermissionID

```go
type PermissionID string
```

A unique identifier for a permission. These IDs are assigned by Google Drive and are used to update or delete specific permissions.

### Errors

The package defines the following error constants that can be checked using `errors.Is()`:

```go
var (
    ErrInvalidPath              error // Invalid path format
    ErrDriveError               error // Google Drive API error
    ErrIOError                  error // I/O operation error
    ErrNotFound                 error // File or directory not found
    ErrAlreadyExists            error // File or directory already exists
    ErrMultiParentsNotSupported error // File has multiple parents
    ErrNotReadable              error // File cannot be read (e.g., Google Apps files)
    ErrNotRemovable             error // Directory not empty or cannot be removed
)
```

**Error Descriptions:**

- **`ErrInvalidPath`** - Returned when a path is invalid (e.g., not absolute, contains `.` or `..` components, or is empty)
- **`ErrDriveError`** - Returned when a Google Drive API call fails (wraps the underlying API error)
- **`ErrIOError`** - Returned when an I/O operation fails (e.g., reading response body)
- **`ErrNotFound`** - Returned when a requested file or directory is not found
- **`ErrAlreadyExists`** - Returned when `MkdirAll` encounters multiple directories with the same name at any level in the path
- **`ErrMultiParentsNotSupported`** - Returned by `ResolvePath` when attempting to resolve the path of a file that has multiple parents (Google Drive allows files to have multiple parents, but this library doesn't support path resolution for such files)
- **`ErrNotReadable`** - Returned by `ReadFile` when attempting to read a Google Apps file (Docs, Sheets, Slides, etc.), which cannot be downloaded as raw bytes
- **`ErrNotRemovable`** - Returned by `Remove` when attempting to remove a non-empty directory (use `RemoveAll` instead)

**Error Handling Example:**

```go
import (
    "errors"
    "github.com/Jumpaku/drivefs"
)

data, err := driveFS.ReadFile(fileID)
if err != nil {
    if errors.Is(err, drivefs.ErrNotReadable) {
        // Handle Google Apps files differently
        fmt.Println("This is a Google Apps file, use export instead")
    } else if errors.Is(err, drivefs.ErrNotFound) {
        fmt.Println("File not found")
    } else {
        log.Fatal(err)
    }
}
```

## Features

- ✅ **File and Directory Operations**: Create, read, write, copy, rename, move, and delete files and directories
- ✅ **Permission Management**: List, set, and delete permissions for users, groups, domains, and public access
- ✅ **Shortcut Support**: Create shortcuts (links) to files and directories
- ✅ **Path-Based Operations**: Use familiar path strings like `/folder/subfolder/file.txt`
- ✅ **Path Resolution**: Convert between file IDs and absolute paths
- ✅ **Tree Walking**: Recursively traverse directory structures with the `Walk` function
- ✅ **Shared Drive Support**: Full support for both My Drive and Shared Drives
- ✅ **Comprehensive Error Handling**: Well-defined error constants that can be checked with `errors.Is()`
- ✅ **Google Apps File Detection**: Identify Google Docs, Sheets, Slides, and other Apps files
- ✅ **Trash Support**: Choose between moving items to trash or permanently deleting them

## Authentication

This package requires an authenticated `*drive.Service` instance from the Google Drive API. 

### Setup Steps

1. **Create a Google Cloud project**
   - Visit the [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one

2. **Enable the Google Drive API**
   - In your project, navigate to "APIs & Services" > "Library"
   - Search for "Google Drive API" and enable it

3. **Create credentials**
   - Go to "APIs & Services" > "Credentials"
   - Create OAuth 2.0 credentials (for user access) or a Service Account (for server-to-server access)
   - Download the credentials JSON file

4. **Authenticate and create a `drive.Service`**

See the [Google Drive API Go Quickstart](https://developers.google.com/drive/api/v3/quickstart/go) for detailed authentication instructions.

### Required Scopes

- **Full read-write access**: `drive.DriveScope` (`https://www.googleapis.com/auth/drive`)
- **Read-only access**: `drive.DriveReadonlyScope` (`https://www.googleapis.com/auth/drive.readonly`)

**Example Authentication (OAuth 2.0):**

```go
import (
    "context"
    "log"
    "os"
    
    "golang.org/x/oauth2/google"
    "google.golang.org/api/drive/v3"
    "google.golang.org/api/option"
)

// Load credentials from a JSON file
b, err := os.ReadFile("credentials.json")
if err != nil {
    log.Fatalf("Unable to read credentials file: %v", err)
}

config, err := google.ConfigFromJSON(b, drive.DriveScope)
if err != nil {
    log.Fatalf("Unable to parse credentials: %v", err)
}

// Get an OAuth 2.0 token (implement token retrieval as needed)
token, err := getToken(config)
if err != nil {
    log.Fatalf("Unable to retrieve token: %v", err)
}

client := config.Client(context.Background(), token)
service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
if err != nil {
    log.Fatalf("Unable to create Drive service: %v", err)
}

// Now use the service with drivefs
driveFS, err := drivefs.New(service, "root")
```

## Important Notes

### Path Requirements

- **Absolute Paths Only**: All path strings must be absolute and start with `/`
- **No Relative Components**: Paths cannot contain `.` (current directory) or `..` (parent directory) components
- **Forward Slashes**: Use `/` as the path separator (Unix-style)

### Duplicate File Names

Google Drive's file system differs from traditional filesystems in that it allows multiple files or folders with the same name in the same parent directory.

- `Create()` and `Mkdir()` will create new items even if items with the same name already exist
- To avoid duplicates, check existing items with `ReadDir()` before creating
- `FindByPath()` returns **all** matching items when duplicates exist
- `MkdirAll()` returns an error if it encounters multiple directories with the same name at any level

### Google Apps Files

Google Apps files (Docs, Sheets, Slides, Forms, etc.) are special:

- They cannot be read with `ReadFile()` - this will return `ErrNotReadable`
- They have MIME types starting with `application/vnd.google-apps.`
- Use `FileInfo.IsAppFile()` to detect them
- To access their content, use the Google Drive API's [export functionality](https://developers.google.com/drive/api/guides/manage-downloads#download_a_document)

### Multiple Parents

Google Drive allows files to have multiple parent directories:

- This library primarily supports single-parent files
- `ResolvePath()` will return `ErrMultiParentsNotSupported` for files with multiple parents
- `Move()` removes all existing parents and sets a single new parent

### Trashed Items

- Trashed items are automatically excluded from `ReadDir()` and path resolution operations
- Use the `moveToTrash` parameter in `Remove()` and `RemoveAll()`:
  - `moveToTrash=true`: Items can be restored from Google Drive trash
  - `moveToTrash=false`: Items are permanently deleted

### Shared Drives

- Full support for Shared Drives (formerly Team Drives)
- All API calls use `SupportsAllDrives(true)` and `IncludeItemsFromAllDrives(true)`
- Create a DriveFS instance with a Shared Drive root ID to work within that Shared Drive

## License

BSD 2-Clause License. See [LICENSE](LICENSE) for details.
