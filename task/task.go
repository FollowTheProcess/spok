// Package task handles core spok functionality related to the processing of declared
// tasks e.g. expanding glob patterns, parsing from an ast node etc.
package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/FollowTheProcess/spok/ast"
	"github.com/FollowTheProcess/spok/iostream"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/fatih/color"
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

// Run runs a task commands in order, echoing each one to out and returning the list of results
// containing the exit status, stdout and stderr of each command.
//
// If the task has no commands, this becomes a no-op.
func (t *Task) Run(runner shell.Runner, stream iostream.IOStream, env []string) (shell.Results, error) {
	echoStyle := color.New(color.Bold)

	var results shell.Results
	for _, cmd := range t.Commands {
		echoStyle.Fprintln(stream.Stdout, cmd)
		result, err := runner.Run(cmd, stream, t.Name, env)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Result encodes the overall result of running a task which
// may involve any number of shell commands.
type Result struct {
	Task           string        `json:"task"`    // The name of the task
	CommandResults shell.Results `json:"results"` // The results of running the tasks commands
	Skipped        bool          `json:"skipped"` // Whether the task was skipped or run
}

// Ok returns whether or not the task was successful, true if
// all commands exited with 0, else false.
func (r Result) Ok() bool {
	return r.CommandResults.Ok()
}

// Results is a collection of task results.
type Results []Result

// Ok reports whether all results in the collection are ok.
func (r Results) Ok() bool {
	for _, result := range r {
		if !result.Ok() {
			return false
		}
	}
	return true
}

// JSON returns the Results as JSON.
func (r Results) JSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
