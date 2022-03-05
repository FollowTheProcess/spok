package spok

import "path/filepath"

type builtin func(...string) (string, error)

var builtins = map[string]builtin{
	"join": func(parts ...string) (string, error) { return filepath.Join(parts...), nil },
}
