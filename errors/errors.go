package errors

import (
	"errors"
)

var (
	ErrInvalidPath              = errors.New("invalid path")
	ErrAPIError                 = errors.New("api error")
	ErrIOError                  = errors.New("io error")
	ErrNotFound                 = errors.New("not found")
	ErrAlreadyExists            = errors.New("already exists")
	ErrMultiParentsNotSupported = errors.New("multi parents not supported")
	ErrNotReadable              = errors.New("not readable")
	ErrNotRemovable             = errors.New("not removable")
)

type wrapError struct {
	underlying error
	msg        string
	cause      error
}

var _ error = (*wrapError)(nil)

func NewAPIError(msg string, cause error) error {
	return &wrapError{
		underlying: ErrAPIError,
		msg:        msg,
		cause:      cause,
	}
}

func NewIOError(msg string, cause error) error {
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
