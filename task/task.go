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
	"github.com/FollowTheProcess/spok/hash"
	"github.com/FollowTheProcess/spok/shell"
)

// Task represents a spok Task.
type Task struct {
	Doc              string   // The task docstring
	Name             string   // Task name
	TaskDependencies []string // Other tasks or idents this task depends on (by name)
	FileDependencies []string // Filepaths this task depends on
	GlobDependencies []string // Filepath dependencies that are specified as glob patterns
	Commands         []string // Shell commands to run
	NamedOutputs     []string // Other outputs by ident
	FileOutputs      []string // Filepaths this task outputs
	GlobOutputs      []string // Filepaths this task outputs that are specified as glob patterns
}

// Run runs a task commands in order, returning the list of results containing
// the exit status, stdout and stderr of each command.
func (t *Task) Run() ([]shell.Result, error) {
	if len(t.Commands) == 0 {
		return nil, fmt.Errorf("Task %q has no commands", t.Name)
	}

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

// ShouldRun hashes the tasks expanded dependencies and compares against a
// previously cached value to determine whether or not the task should be run
// (i.e. if any dependency has changed). ShouldRun must be called after the glob
// patterns have been expanded, to do otherwise will return an error.
func (t *Task) ShouldRun(files []string, hasher hash.Hasher, cached string) (bool, error) {
	digest, err := hasher.Hash(files)
	if err != nil {
		return false, err
	}

	return digest != cached, nil
}

// New parses a task AST node into a concrete task,
// root is the absolute path of the directory to use as the root for
// glob expansion, typically the path to the spokfile.
func New(t ast.Task, root string, vars map[string]string) (Task, error) {
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
			return Task{}, fmt.Errorf("unknown dependency: %s", dep)
		}
	}

	for _, cmd := range t.Commands {
		expanded, err := expandVars(cmd.Command, vars)
		if err != nil {
			return Task{}, err
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
			return Task{}, fmt.Errorf("unknown dependency: %s", out)
		}
	}

	task := Task{
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
