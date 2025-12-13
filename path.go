package drivefs

// Path represents an absolute path in Google Drive.
// Paths must start with '/' and use forward slashes as separators (e.g., "/folder/subfolder/file").
// Relative path components like "." and ".." are not allowed.
type Path string
