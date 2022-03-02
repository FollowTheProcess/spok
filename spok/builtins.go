package spok

import "path/filepath"

// Join joins any number of path elements into a single path, separating
// the elements with the OS specific separator.
func Join(parts ...string) string {
	return filepath.Join(parts...)
}
