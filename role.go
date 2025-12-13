package drivefs

// Role represents the level of access granted by a permission.
type Role string

const (
	// RoleOwner grants full ownership including the ability to delete the file and modify permissions.
	RoleOwner Role = "owner"

	// RoleOrganizer grants the ability to organize files in shared drives (shared drives only).
	RoleOrganizer Role = "organizer"

	// RoleFileOrganizer grants the ability to organize files in shared drives (shared drives only).
	RoleFileOrganizer Role = "fileOrganizer"

	// RoleWriter grants the ability to read and modify the file.
	RoleWriter Role = "writer"

	// RoleCommenter grants the ability to read and comment on the file.
	RoleCommenter Role = "commenter"

	// RoleReader grants read-only access to the file.
	RoleReader Role = "reader"
)
