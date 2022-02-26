// Package spok implements core spok types and behaviour e.g. tasks, file.
package spok

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/bmatcuk/doublestar/v4"
)

// Canonical spokfile filename.
const spokfile = "spokfile"

// errNoSpokfile is what happens when spok can't find a spokfile.
var errNoSpokfile = errors.New("no spokfile found")

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

// expandGlob expands out the glob pattern from root and returns all the matches.
func expandGlob(root, pattern string) ([]string, error) {
	matches, err := doublestar.Glob(os.DirFS(root), pattern)
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern '%s': %w", pattern, err)
	}
	return matches, nil
}

// File represents a concrete spokfile.
type File struct {
	Path  string            // The absolute path to the spokfile
	Vars  map[string]string // Global variables in IDENT: value form
	Tasks []Task            // Defined tasks
}

// Task represents a spok Task.
type Task struct {
	Doc              string   // The task docstring
	Name             string   // Task name
	TaskDependencies []string // Other tasks this task depends on (by name)
	FileDependencies []string // Filepaths this task depends on (globs expanded)
	Commands         []string // Shell commands to run
}

// newTask parses a task AST node into a concrete task.
func newTask(t ast.Task, root string) (Task, error) {
	var fileDeps []string
	var taskDeps []string
	var commands []string

	for _, dep := range t.Dependencies {
		switch {
		case dep.Type() == ast.NodeString:
			// String means file dependency
			if strings.Contains(dep.String(), "*") {
				matches, err := expandGlob(root, dep.String())
				if err != nil {
					return Task{}, err
				}
				fileDeps = append(fileDeps, matches...)
			} else {
				fileDeps = append(fileDeps, dep.String())
			}
		case dep.Type() == ast.NodeIdent:
			// Ident means it depends on another task
			taskDeps = append(taskDeps, dep.String())
		default:
			return Task{}, fmt.Errorf("unknown dependency: %s", dep)
		}
	}

	for _, cmd := range t.Commands {
		commands = append(commands, cmd.Command)
	}
	task := Task{
		Doc:              strings.TrimSpace(t.Docstring.Text),
		Name:             t.Name.Name,
		TaskDependencies: taskDeps,
		FileDependencies: fileDeps,
		Commands:         commands,
	}
	return task, nil
}
