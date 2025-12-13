package drivefs

const (
	granteeTypeUser   = "user"
	granteeTypeGroup  = "group"
	granteeTypeDomain = "domain"
	granteeTypeAnyone = "anyone"
)

// Grantee represents an entity that can be granted permission to access a file or directory.
// This is a sealed interface - use the constructor functions User, Group, Domain, or Anyone.
type Grantee interface {
	doNotImplement(Grantee)
}

// User creates a Grantee representing a specific user identified by email address.
func User(email string) Grantee {
	return GranteeUser{Email: email}
}

// Group creates a Grantee representing a Google Group identified by email address.
func Group(email string) Grantee {
	return GranteeGroup{Email: email}
}

// Domain creates a Grantee representing all users in a Google Workspace domain.
func Domain(domain string) Grantee {
	return GranteeDomain{Domain: domain}
}

// Anyone creates a Grantee representing all users (public access).
func Anyone() Grantee {
	return GranteeAnyone{}
}

// GranteeUser represents a specific user identified by email address.
type GranteeUser struct {
	Email string
}

func (GranteeUser) doNotImplement(Grantee) {}

// GranteeGroup represents a Google Group identified by email address.
type GranteeGroup struct {
	Email string
}

func (GranteeGroup) doNotImplement(Grantee) {}

// GranteeDomain represents all users in a Google Workspace domain.
type GranteeDomain struct {
	Domain string
}

func (GranteeDomain) doNotImplement(Grantee) {}

// GranteeAnyone represents all users (public access).
type GranteeAnyone struct{}

func (GranteeAnyone) doNotImplement(Grantee) {}
