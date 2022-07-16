// Package task handles core spok functionality related to the processing of declared
// tasks e.g. expanding glob patterns, parsing from an ast node etc.
package task

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/bmatcuk/doublestar/v4"
)

// Task represents a spok Task.
type Task struct {
	Doc                  string   // The task docstring
	Name                 string   // Task name
	Root                 string   // The root dir (typically the spokfile dir)
	TaskDependencies     []string // Other tasks or idents this task depends on (by name)
	FileDependencies     []string // Filepaths this task depends on
	GlobDependencies     []string // Filepath dependencies that are specified as glob patterns
	Commands             []string // Shell commands to run
	NamedOutputs         []string // Other outputs by ident
	FileOutputs          []string // Filepaths this task outputs
	GlobOutputs          []string // Filepaths this task outputs that are specified as glob patterns
	ExpandedDependencies []string // Combination of FileDependencies and the expansion of every GlobDependency, only populated on run
}

// Run runs a task commands in order, returning the list of results containing
// the exit status, stdout and stderr of each command.
func (t *Task) Run() ([]shell.Result, error) {
	if len(t.Commands) == 0 {
		return nil, fmt.Errorf("Task %q has no commands", t.Name)
	}

	// Expand glob patterns on the task
	if err := t.expandGlobs(); err != nil {
		return nil, err
	}

	// TODO: Hash everything in expandedDependencies and compare against
	// cached hash sum to see whether or not we should run

	var results []shell.Result
	for _, cmd := range t.Commands {
		result, err := shell.Run(cmd, t.Name, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// expandGlobs looks at all the glob patterns defined as dependencies on the task,
// expands them all and populates the expandedDependencies field on the struct.
func (t *Task) expandGlobs() error {
	// If the expandedDependencies is already populated, do nothing
	if len(t.ExpandedDependencies) != 0 {
		return nil
	}
	for _, pattern := range t.GlobDependencies {
		matches, err := expandGlob(t.Root, pattern)
		if err != nil {
			return err
		}
		t.ExpandedDependencies = append(t.ExpandedDependencies, matches...)
	}

	// Add in the normal file dependencies too
	t.ExpandedDependencies = append(t.ExpandedDependencies, t.FileDependencies...)
	return nil
}

// New parses a task AST node into a concrete task,
// root is the absolute path of the directory to use as the root for
// glob expansion, typically the path to the spokfile.
func New(t ast.Task, root string, vars map[string]string) (*Task, error) {
	var (
		fileDeps     []string
		globDeps     []string
		taskDeps     []string
		commands     []string
		fileOutputs  []string
		globOutputs  []string
		namedOutputs []string
	)

	for _, dep := range t.Dependencies {
		switch {
		case dep.Type() == ast.NodeString:
			// String means file dependency, in which case Literal is the go representation of the string
			if strings.Contains(dep.Literal(), "*") {
				// We have a glob pattern
				globDeps = append(globDeps, dep.Literal())
			} else {
				// We have something like "file.go"
				fileDeps = append(fileDeps, filepath.Join(root, dep.Literal()))
			}
		case dep.Type() == ast.NodeIdent:
			// Ident means it depends on another task
			taskDeps = append(taskDeps, dep.Literal())
		default:
			return nil, fmt.Errorf("unknown dependency: %s", dep)
		}
	}

	for _, cmd := range t.Commands {
		expanded, err := expandVars(cmd.Command, vars)
		if err != nil {
			return nil, err
		}
		commands = append(commands, expanded)
	}

	for _, out := range t.Outputs {
		switch {
		case out.Type() == ast.NodeString:
			// String means file
			if strings.Contains(out.Literal(), "*") {
				// We have a glob pattern
				globOutputs = append(globOutputs, out.Literal())
			} else {
				// We have something like "file.go"
				fileOutputs = append(fileOutputs, filepath.Join(root, out.Literal()))
			}
		case out.Type() == ast.NodeIdent:
			// Ident means it outputs something named by global scope
			namedOutputs = append(namedOutputs, out.Literal())
		default:
			return nil, fmt.Errorf("unknown dependency: %s", out)
		}
	}

	task := &Task{
		Doc:              strings.TrimSpace(t.Docstring.Text),
		Name:             t.Name.Name,
		TaskDependencies: taskDeps,
		FileDependencies: fileDeps,
		GlobDependencies: globDeps,
		Commands:         commands,
		NamedOutputs:     namedOutputs,
		FileOutputs:      fileOutputs,
		GlobOutputs:      globOutputs,
	}
	return task, nil
}

// expandVars performs a find and replace on any templated variables in
// a command, using the provided variables map.
func expandVars(command string, vars map[string]string) (string, error) {
	tmp := template.New("tmp")
	parsed, err := tmp.Parse(command)
	if err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	if err := parsed.Execute(out, vars); err != nil {
		return "", err
	}

	return out.String(), nil
}

// expandGlob expands out the glob pattern from root and returns all the matches,
// the matches are made absolute before returning, root should be absolute.
func expandGlob(root, pattern string) ([]string, error) {
	matches, err := doublestar.FilepathGlob(filepath.Join(root, pattern))
	if err != nil {
		return nil, fmt.Errorf("could not expand glob pattern '%s': %w", filepath.Join(root, pattern), err)
	}
	absMatches := make([]string, 0, len(matches))
	for _, match := range matches {
		abs, err := filepath.Abs(match)
		if err != nil {
			return nil, fmt.Errorf("could not resolve path '%s' to absolute: %w", match, err)
		}
		absMatches = append(absMatches, abs)
	}

	return absMatches, nil
}
