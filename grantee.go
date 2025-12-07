package drivefs

const (
	granteeTypeUser   = "user"
	granteeTypeGroup  = "group"
	granteeTypeDomain = "domain"
	granteeTypeAnyone = "anyone"
)

type Grantee interface {
	doNotImplement(Grantee)
}

func User(email string) Grantee {
	return GranteeUser{Email: email}
}

func Group(email string) Grantee {
	return GranteeGroup{Email: email}
}

func Domain(domain string) Grantee {
	return GranteeDomain{Domain: domain}
}

func Anyone() Grantee {
	return GranteeAnyone{}
}

type GranteeUser struct {
	Email string
}

func (GranteeUser) doNotImplement(Grantee) {}

type GranteeGroup struct {
	Email string
}

func (GranteeGroup) doNotImplement(Grantee) {}

type GranteeDomain struct {
	Domain string
}

func (GranteeDomain) doNotImplement(Grantee) {}

type GranteeAnyone struct{}

func (GranteeAnyone) doNotImplement(Grantee) {}
