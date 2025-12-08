package drivefs

// This file is part of the package tests (package drivefs) and provides
// helpers that allow tests in the external package to access internal
// package constructs. Helpers are exported so `drivefs_test` can call them
// via the module import path.

// NewDriveError constructs a drive-wrapped error using package-internal constructor.
func NewDriveError(msg string, cause error) error {
	return newDriveError(msg, cause)
}

// NewIOError constructs an io-wrapped error using package-internal constructor.
func NewIOError(msg string, cause error) error {
	return newIOError(msg, cause)
}
