// Package task implements core spok functionality to do with it's tasks.
package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/bmatcuk/doublestar/v4"
)

// expandGlob expands out the glob pattern from root and returns all the matches,
// the matches are made absolute before returning, root should be absolute.
func expandGlob(root, pattern string) ([]string, error) {
	matches, err := doublestar.Glob(os.DirFS(root), pattern)
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern '%s': %w", filepath.Join(root, pattern), err)
	}

	var absMatches []string
	for _, match := range matches {
		joined := filepath.Join(root, match)
		abs, err := filepath.Abs(joined)
		if err != nil {
			return nil, fmt.Errorf("could not resolve path '%s' to absolute: %w", joined, err)
		}
		absMatches = append(absMatches, abs)
	}

	return absMatches, nil
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
				// We have a glob pattern
				matches, err := expandGlob(root, dep.String())
				if err != nil {
					return Task{}, err
				}
				fileDeps = append(fileDeps, matches...)
			} else {
				// We have something like "file.go"
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
