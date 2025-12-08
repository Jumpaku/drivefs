package drivefs_test

import (
	"reflect"
	"testing"

	"github.com/Jumpaku/go-drivefs"
)

func TestGrantee_ConstructorsReturnExpectedConcreteTypes(t *testing.T) {
	cases := []struct {
		name string
		got  drivefs.Grantee
		want drivefs.Grantee
	}{
		{"User", drivefs.User("alice@example.com"), drivefs.GranteeUser{Email: "alice@example.com"}},
		{"Group", drivefs.Group("team@example.com"), drivefs.GranteeGroup{Email: "team@example.com"}},
		{"Domain", drivefs.Domain("example.com"), drivefs.GranteeDomain{Domain: "example.com"}},
		{"Anyone", drivefs.Anyone(), drivefs.GranteeAnyone{}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if !reflect.DeepEqual(c.got, c.want) {
				t.Fatalf("mismatch for %s: got (%T) %#v, want (%T) %#v", c.name, c.got, c.got, c.want, c.want)
			}
		})
	}
}
