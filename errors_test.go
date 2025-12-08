package drivefs_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Jumpaku/go-drivefs"
)

func TestErrVars_IsAndMessage(t *testing.T) {
	cases := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrInvalidPath", drivefs.ErrInvalidPath, "invalid path"},
		{"ErrDriveError", drivefs.ErrDriveError, "drive error"},
		{"ErrDriveError2", drivefs.NewDriveError("", fmt.Errorf("")), "drive error"},
		{"ErrIOError", drivefs.ErrIOError, "io error"},
		{"ErrIOError2", drivefs.NewIOError("", fmt.Errorf("")), "io error"},
		{"ErrNotFound", drivefs.ErrNotFound, "not found"},
		{"ErrAlreadyExists", drivefs.ErrAlreadyExists, "already exists"},
		{"ErrMultiParentsNotSupported", drivefs.ErrMultiParentsNotSupported, "multi parents not supported"},
		{"ErrNotReadable", drivefs.ErrNotReadable, "not readable"},
		{"ErrNotRemovable", drivefs.ErrNotRemovable, "not removable"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name+"/IsWrapped", func(t *testing.T) {
			wrapped := fmt.Errorf("higher: %w", c.err)
			if !errors.Is(wrapped, c.err) {
				t.Fatalf("errors.Is(wrapped, %s) = false, want true", c.name)
			}
		})

		t.Run(c.name+"/Message", func(t *testing.T) {
			wrapped := fmt.Errorf("higher: %w", c.err)
			if !strings.Contains(wrapped.Error(), c.msg) {
				t.Fatalf("%s.Error() = %q does not contain %q", c.name, wrapped.Error(), c.msg)
			}
		})
	}
}
