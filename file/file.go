// Package file implements the core functionality to do with the spokfile.
package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/FollowTheProcess/spok/task"
)

// Canonical spokfile filename.
const spokfile = "spokfile"

// errNoSpokfile is what happens when spok can't find a spokfile.
var errNoSpokfile = errors.New("no spokfile found")

// File represents a concrete spokfile.
type File struct {
	Path  string            // The absolute path to the spokfile
	Vars  map[string]string // Global variables in IDENT: value form
	Tasks []task.Task       // Defined tasks
}

// find climbs the file tree from 'start' to 'stop' looking for a spokfile,
// if it hits 'stop' before finding one, an ErrNoSpokfile will be returned
// If a spokfile is found, it's absolute path will be returned.
func find(start, stop string) (string, error) {
	for {
		entries, err := os.ReadDir(start)
		if err != nil {
			return "", fmt.Errorf("could not read directory '%s': %w", start, err)
		}

		for _, e := range entries {
			if !e.IsDir() && e.Name() == spokfile {
				// We've found it
				abs, err := filepath.Abs(filepath.Join(start, e.Name()))
				if err != nil {
					return "", fmt.Errorf("could not resolve '%s': %w", e.Name(), err)
				}
				return abs, nil
			} else if start == stop {
				return "", errNoSpokfile
			}
		}
		start = filepath.Dir(start)
	}
}
