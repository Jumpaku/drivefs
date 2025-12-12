package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Jumpaku/go-drivefs"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func newDriveFS() *drivefs.DriveFS {
	ctx := context.Background()

	client, err := google.DefaultClient(ctx,
		drive.DriveScope,
	)
	if err != nil {
		log.Panic(err)
	}

	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Panic(err)
	}
	return drivefs.New(driveService)
}

var sc = func() *bufio.Scanner {
	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)
	return sc
}()

func step() {
	sc.Scan()
}

func main() {
	// Create a new DriveFS instance with a root folder ID
	driveFS := newDriveFS()

	// walk through the directory structure
	type r struct {
		path drivefs.Path
		info drivefs.FileInfo
	}
	files := []r{}
	_ = driveFS.Walk("0ADHyXmFLm9riUk9PVA", func(path drivefs.Path, fileInfo drivefs.FileInfo) error {
		files = append(files, r{path, fileInfo})
		return nil
	})
	for _, file := range files {
		fmt.Printf("%s (ID: %s)\n", file.path, file.info.ID)
	}

	// Create a directory structure
	step()
	dirInfo, err := driveFS.MkdirAll("0ADHyXmFLm9riUk9PVA", "/path/to/directory")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created directory: %s (ID: %s)\n", dirInfo.Name, dirInfo.ID)

	// Create a file in the directory
	step()
	fileInfo, err := driveFS.Create(dirInfo.ID, "example.txt")
	if err != nil {
		log.Fatal(err)
	}

	// Write content to the file
	step()
	err = driveFS.WriteFile(fileInfo.ID, []byte("Hello, Google Drive!"))
	if err != nil {
		log.Fatal(err)
	}

	{
		// Manage permissions
		// List current permissions
		step()
		permissions, err := driveFS.PermList(fileInfo.ID)
		if err != nil {
			log.Fatal(err)
		}
		for _, perm := range permissions {
			fmt.Printf("Permission ID: %s, Role: %s\n", perm.ID(), perm.Role())
		}

		// Grant read access to a user
		step()
		_, err = driveFS.PermSet(fileInfo.ID, drivefs.UserPermission("user@example.com", drivefs.RoleReader))
		if err != nil {
			log.Fatal(err)
		}

		// Grant write access to a group
		step()
		_, err = driveFS.PermSet(fileInfo.ID, drivefs.GroupPermission("group@example.com", drivefs.RoleWriter))
		if err != nil {
			log.Fatal(err)
		}

		// Grant read access to anyone with the link
		step()
		_, err = driveFS.PermSet(fileInfo.ID, drivefs.AnyonePermission(drivefs.RoleReader, false))
		if err != nil {
			log.Fatal(err)
		}

		// Remove a user's permission
		step()
		_, err = driveFS.PermDel(fileInfo.ID, drivefs.User("user@example.com"))
		if err != nil {
			log.Fatal(err)
		}
		_, err = driveFS.PermDel(fileInfo.ID, drivefs.Group("group@example.com"))
		if err != nil {
			log.Fatal(err)
		}
		_, err = driveFS.PermDel(fileInfo.ID, drivefs.Anyone())
		if err != nil {
			log.Fatal(err)
		}
	}
	// Read the file content
	step()
	data, err := driveFS.ReadFile(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// List directory contents
	step()
	entries, err := driveFS.ReadDir(dirInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Printf("%s (folder: %v, ID: %s)\n", entry.Name, entry.IsFolder(), entry.ID)
	}

	// Get the full path from a file ID
	step()
	path, err := driveFS.ResolvePath(fileInfo.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Path: %s\n", path)

	// Copy a file to a different directory
	step()
	newParentInfo, err := driveFS.MkdirAll("0ADHyXmFLm9riUk9PVA", "/new/location")
	if err != nil {
		log.Fatal(err)
	}
	copiedFileInfo, err := driveFS.Copy(fileInfo.ID, newParentInfo.ID, "example_copy.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Copied file: %s (ID: %s)\n", copiedFileInfo.Name, copiedFileInfo.ID)

	// Resolve a path to get FileInfo
	step()
	resolvedInfo, err := driveFS.FindByPath("0ADHyXmFLm9riUk9PVA", "/new/location/example_copy.txt")
	if err != nil {
		log.Fatal(err)
	}
	for _, resolvedInfo := range resolvedInfo {
		fmt.Printf("Resolved: %s\n", resolvedInfo.ID)
	}

	// Rename a file
	step()
	renamedFileInfo, err := driveFS.Rename(fileInfo.ID, "renamed_example.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Renamed file: %s\n", renamedFileInfo.Name)

	// Move a file to a different directory
	step()
	err = driveFS.Move(fileInfo.ID, newParentInfo.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Delete a file (move to trash)
	step()
	err = driveFS.Remove(fileInfo.ID, true)
	if err != nil {
		log.Fatal(err)
	}
}
