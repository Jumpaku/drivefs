package drivefs

import (
	"errors"
)

// Common errors returned by DriveFS operations.
var (
	// ErrInvalidPath is returned when a path is malformed or uses relative path components.
	ErrInvalidPath = errors.New("invalid path")

	// ErrDriveError is the underlying error for all Google Drive API errors.
	ErrDriveError = errors.New("drive error")

	// ErrIOError is the underlying error for all I/O errors.
	ErrIOError = errors.New("io error")

	// ErrNotFound is returned when a requested file or directory does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when attempting to create a file or directory that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrMultiParentsNotSupported is returned when an operation encounters a file with multiple parents.
	ErrMultiParentsNotSupported = errors.New("multi parents not supported")

	// ErrNotReadable is returned when attempting to read a file that cannot be downloaded (e.g., Google Apps files).
	ErrNotReadable = errors.New("not readable")

	// ErrNotRemovable is returned when attempting to remove a non-empty directory.
	ErrNotRemovable = errors.New("not removable")
)

type wrapError struct {
	underlying error
	msg        string
	cause      error
}

var _ error = (*wrapError)(nil)

func newDriveError(msg string, cause error) error {
	return &wrapError{
		underlying: ErrDriveError,
		msg:        msg,
		cause:      cause,
	}
}

func newIOError(msg string, cause error) error {
	return &wrapError{
		underlying: ErrIOError,
		msg:        msg,
		cause:      cause,
	}
}

func (err *wrapError) Error() string {
	if err == nil {
		return "(*wrapError)(nil)"
	}
	message := err.underlying.Error() + ": " + err.msg
	if err.cause != nil {
		message += ": " + err.cause.Error()
	}
	return message
}

func (err *wrapError) Unwrap() []error {
	if err.cause == nil {
		return []error{err.underlying}
	}
	return []error{err.underlying, err.cause}
}
