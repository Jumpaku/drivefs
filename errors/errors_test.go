package errors_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	. "github.com/Jumpaku/go-drivefs/errors"
)

func TestErrVars_IsAndMessage(t *testing.T) {
	cases := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrInvalidPath", ErrInvalidPath, "invalid path"},
		{"ErrAPIError", ErrAPIError, "api error"},
		{"ErrAPIError2", NewAPIError("", fmt.Errorf("")), "api error"},
		{"ErrIOError", ErrIOError, "io error"},
		{"ErrIOError2", NewIOError("", fmt.Errorf("")), "io error"},
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrAlreadyExists", ErrAlreadyExists, "already exists"},
		{"ErrMultiParentsNotSupported", ErrMultiParentsNotSupported, "multi parents not supported"},
		{"ErrNotReadable", ErrNotReadable, "not readable"},
		{"ErrNotRemovable", ErrNotRemovable, "not removable"},
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
