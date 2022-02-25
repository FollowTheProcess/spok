// Package spok implements core spok types and behaviour e.g. tasks, file.
package spok

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Canonical spokfile filename.
const spokfile = "spokfile"

// ErrNoSpokfile is what happens when spok can't find a spokfile.
var ErrNoSpokfile = errors.New("no spokfile found")

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
				return "", ErrNoSpokfile
			}
		}
		start = filepath.Dir(start)
	}
}

// File represents a concrete spokfile.
type File struct {
	Vars  map[string]string // Global variables in IDENT: value form
	Tasks []Task            // Defined tasks
}

// Task represents a spok task.
type Task struct {
	Doc              string   // The task docstring
	Name             string   // Task name
	TaskDependencies []Task   // Other tasks this task depends on
	FileDependencies []string // Filepaths this task depends on (globs expanded)
	Commands         []string // Shell commands to run
}
