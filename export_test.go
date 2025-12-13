package drivefs

// This file provides test helpers that expose internal package constructs
// to the external test package (drivefs_test).

// NewDriveError constructs a drive error using the internal constructor.
// This is exported for testing purposes only.
func NewDriveError(msg string, cause error) error {
	return newDriveError(msg, cause)
}

// NewIOError constructs an I/O error using the internal constructor.
// This is exported for testing purposes only.
func NewIOError(msg string, cause error) error {
	return newIOError(msg, cause)
}
