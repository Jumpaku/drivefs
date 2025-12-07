package drivefs

type PermissionID string

type Permission interface {
	ID() PermissionID
	Grantee() Grantee
	Role() Role
	AllowFileDiscovery() bool
	doNotImplement(Permission)
}

type permission struct {
	grantee            Grantee
	role               Role
	id                 PermissionID
	allowFileDiscovery bool
}

func UserPermission(email string, role Role, allowFileDiscovery bool) Permission {
	return permission{grantee: User(email), role: role, allowFileDiscovery: allowFileDiscovery}
}

func GroupPermission(email string, role Role, allowFileDiscovery bool) Permission {
	return permission{grantee: Group(email), role: role, allowFileDiscovery: allowFileDiscovery}
}

func DomainPermission(domain string, role Role, allowFileDiscovery bool) Permission {
	return permission{grantee: Domain(domain), role: role, allowFileDiscovery: allowFileDiscovery}
}

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
