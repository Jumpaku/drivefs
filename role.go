package drivefs

type Role string

const (
	RoleOwner         Role = "owner"
	RoleOrganizer     Role = "organizer"
	RoleFileOrganizer Role = "fileOrganizer"
	RoleWriter        Role = "writer"
	RoleCommenter     Role = "commenter"
	RoleReader        Role = "reader"
)
