package drivefs_test

import (
	"reflect"
	"testing"

	"github.com/Jumpaku/go-drivefs"
)

func TestPermission_ConstructorsAndAccessors(t *testing.T) {
	cases := []struct {
		name                   string
		got                    drivefs.Permission
		wantGrantee            drivefs.Grantee
		wantRole               drivefs.Role
		wantAllowFileDiscovery bool
	}{
		{"UserPermission", drivefs.UserPermission("alice@example.com", drivefs.RoleWriter), drivefs.GranteeUser{Email: "alice@example.com"}, drivefs.RoleWriter, false},
		{"GroupPermission", drivefs.GroupPermission("team@example.com", drivefs.RoleCommenter), drivefs.GranteeGroup{Email: "team@example.com"}, drivefs.RoleCommenter, false},
		{"DomainPermission_allow_false", drivefs.DomainPermission("example.com", drivefs.RoleReader, false), drivefs.GranteeDomain{Domain: "example.com"}, drivefs.RoleReader, false},
		{"DomainPermission_allow_true", drivefs.DomainPermission("example.com", drivefs.RoleOrganizer, true), drivefs.GranteeDomain{Domain: "example.com"}, drivefs.RoleOrganizer, true},
		{"AnyonePermission_allow_false", drivefs.AnyonePermission(drivefs.RoleFileOrganizer, false), drivefs.GranteeAnyone{}, drivefs.RoleFileOrganizer, false},
		{"AnyonePermission_allow_true", drivefs.AnyonePermission(drivefs.RoleOwner, true), drivefs.GranteeAnyone{}, drivefs.RoleOwner, true},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			p := c.got

			// Role
			if got := p.Role(); got != c.wantRole {
				t.Fatalf("Role() = %v, want %v", got, c.wantRole)
			}

			// ID should be zero value by default
			if got := p.ID(); got != "" {
				t.Fatalf("ID() = %q, want empty string", got)
			}

			// AllowFileDiscovery
			if got := p.AllowFileDiscovery(); got != c.wantAllowFileDiscovery {
				t.Fatalf("AllowFileDiscovery() = %v, want %v", got, c.wantAllowFileDiscovery)
			}

			// Grantee concrete type and value
			if got := p.Grantee(); !reflect.DeepEqual(got, c.wantGrantee) {
				t.Fatalf("Grantee() = (%T) %#v, want (%T) %#v", got, got, c.wantGrantee, c.wantGrantee)
			}
		})
	}
}
