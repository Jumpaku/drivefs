package drivefs

// PermissionID represents a unique identifier for a permission.
type PermissionID string

// Permission represents access permissions for a file or directory.
// This is a sealed interface - use the constructor functions UserPermission, GroupPermission,
// DomainPermission, or AnyonePermission.
type Permission interface {
	// ID returns the unique identifier for this permission.
	ID() PermissionID

	// Grantee returns the entity that has been granted this permission.
	Grantee() Grantee

	// Role returns the access level granted by this permission.
	Role() Role

	// AllowFileDiscovery returns true if the file can be discovered through search
	// by users who have this permission (applicable to domain and anyone permissions).
	AllowFileDiscovery() bool

	doNotImplement(Permission)
}

type permission struct {
	grantee            Grantee
	role               Role
	id                 PermissionID
	allowFileDiscovery bool
}

// UserPermission creates a Permission for a specific user identified by email.
func UserPermission(email string, role Role) Permission {
	return permission{grantee: User(email), role: role}
}

// GroupPermission creates a Permission for a Google Group identified by email.
func GroupPermission(email string, role Role) Permission {
	return permission{grantee: Group(email), role: role}
}

// DomainPermission creates a Permission for all users in a Google Workspace domain.
// The allowFileDiscovery parameter controls whether users can find the file through search.
func DomainPermission(domain string, role Role, allowFileDiscovery bool) Permission {
	return permission{grantee: Domain(domain), role: role, allowFileDiscovery: allowFileDiscovery}
}

// AnyonePermission creates a Permission for all users (public access).
// The allowFileDiscovery parameter controls whether anyone can find the file through search.
func AnyonePermission(role Role, allowFileDiscovery bool) Permission {
	return permission{grantee: Anyone(), role: role, allowFileDiscovery: allowFileDiscovery}
}

func (p permission) Grantee() Grantee {
	return p.grantee
}

func (p permission) Role() Role {
	return p.role
}

func (p permission) ID() PermissionID {
	return p.id
}

func (p permission) AllowFileDiscovery() bool {
	return p.allowFileDiscovery
}

func (p permission) doNotImplement(Permission) {}
