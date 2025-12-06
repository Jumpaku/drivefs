package drivefs

import (
	"errors"
)

var (
	ErrInvalidPath   = errors.New("invalid path")
	ErrDriveError    = errors.New("drive error")
	ErrIOError       = errors.New("io error")
	ErrNotExist      = errors.New("not exist")
	ErrNotReadable   = errors.New("not readable")
	ErrAlreadyExists = errors.New("already exists")
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
