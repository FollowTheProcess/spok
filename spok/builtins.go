package spok

import "path/filepath"

// builtin is a spok built in function.
type builtin func(...string) (string, error)

// package scoped map mapping the names of the builtins to their underlying function.
var builtins = map[string]builtin{
	"join": join,
}

// join joins up filepath parts with an OS specific separator.
func join(parts ...string) (string, error) {
	return filepath.Join(parts...), nil
}
