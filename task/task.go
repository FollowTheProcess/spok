// Package task handles core spok functionality related to the processing of declared
// tasks e.g. expanding glob patterns, parsing from an ast node etc.
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
	Doc               string   // The task docstring
	Name              string   // Task name
	NamedDependencies []string // Other tasks or idents this task depends on (by name)
	FileDependencies  []string // Filepaths this task depends on (globs expanded)
	Commands          []string // Shell commands to run
	NamedOutputs      []string // Other outputs by ident
	FileOutputs       []string // Filepaths this task outputs
}

// New parses a task AST node into a concrete task,
// root is the absolute path of the directory to use as the root for
// glob expansion, typically the path to the spokfile.
func New(t ast.Task, root string) (Task, error) {
	var fileDeps []string
	var nameDeps []string
	var commands []string
	var fileOutputs []string
	var nameOutputs []string

	for _, dep := range t.Dependencies {
		switch {
		case dep.Type() == ast.NodeString:
			// String means file dependency
			str, ok := dep.(ast.String)
			if !ok {
				return Task{}, fmt.Errorf("Task dependency had type ast.NodeString but could not be converted to an ast.String: %s", dep)
			}
			if strings.Contains(str.Text, "*") {
				// We have a glob pattern
				matches, err := expandGlob(root, str.Text)
				if err != nil {
					return Task{}, err
				}
				fileDeps = append(fileDeps, matches...)
			} else {
				// We have something like "file.go"
				fileDeps = append(fileDeps, str.Text)
			}
		case dep.Type() == ast.NodeIdent:
			// Ident means it depends on another task
			nameDeps = append(nameDeps, dep.String())
		default:
			return Task{}, fmt.Errorf("unknown dependency: %s", dep)
		}
	}

	for _, cmd := range t.Commands {
		commands = append(commands, cmd.Command)
	}

	for _, out := range t.Outputs {
		switch {
		case out.Type() == ast.NodeString:
			// String means file
			str, ok := out.(ast.String)
			if !ok {
				return Task{}, fmt.Errorf("Task output had type ast.NodeString but could not be converted to an ast.String: %s", out)
			}
			if strings.Contains(str.Text, "*") {
				// We have a glob pattern
				matches, err := expandGlob(root, str.Text)
				if err != nil {
					return Task{}, err
				}
				fileOutputs = append(fileOutputs, matches...)
			} else {
				// We have something like "file.go"
				fileOutputs = append(fileOutputs, str.Text)
			}
		case out.Type() == ast.NodeIdent:
			// Ident means it outputs something named by global scope
			nameOutputs = append(nameOutputs, out.String())
		default:
			return Task{}, fmt.Errorf("unknown dependency: %s", out)
		}
	}

	task := Task{
		Doc:               strings.TrimSpace(t.Docstring.Text),
		Name:              t.Name.Name,
		NamedDependencies: nameDeps,
		FileDependencies:  fileDeps,
		Commands:          commands,
		NamedOutputs:      nameOutputs,
		FileOutputs:       fileOutputs,
	}
	return task, nil
}
