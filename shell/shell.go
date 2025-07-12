// Package shell implements spok's command running functionality
//
// We use https://github.com/mvdan/sh so spok is entirely self contained and
// does not need an external shell at all to run.
//
// This implementation is based on a similar one in https://github.com/go-task/task
// at internal/execext/exec.go.
package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.followtheprocess.codes/spok/iostream"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// The default timeout after which shell commands will be aborted.
const timeout = 15 * time.Second

// Runner is an interface representing something capable of running shell commands
// and returning Results.
type Runner interface {
	// Run runs the shell command belonging to task with environment variables set.
	Run(cmd string, stream iostream.IOStream, task string, env []string) (Result, error)
}

// Result holds the result of running a shell command.
type Result struct {
	Cmd    string `json:"cmd"`    // The command that was run
	Stdout string `json:"stdout"` // The stdout of the command
	Stderr string `json:"stderr"` // The stderr of the command
	Status int    `json:"status"` // The exit status of the command
}

// Ok returns whether the result was successful or not.
func (r Result) Ok() bool {
	return r.Status == 0
}

// Results is a collection of shell results.
type Results []Result

// Ok reports whether all results in the collection were ok.
func (r Results) Ok() bool {
	for _, result := range r {
		if !result.Ok() {
			return false
		}
	}
	return true
}

// IntegratedRunner implements Runner by using a 100% go implementation
// of a shell interpreter, this is the most cross-compatible version of a shell
// runner possible as it does not depend on any external shell.
type IntegratedRunner struct {
	parser *syntax.Parser
}

// NewIntegratedRunner returns a shell runner with no external dependency.
func NewIntegratedRunner() IntegratedRunner {
	return IntegratedRunner{
		parser: syntax.NewParser(),
	}
}

// Run implements Runner for an IntegratedRunner, using a 100% go implementation of a shell interpreter.
//
// Command stdout and stderr will be collected into the returned Result and optionally also printed to
// the writers in the IOStream, this allows output to be captured or discarded easily.
func (i IntegratedRunner) Run(cmd string, stream iostream.IOStream, task string, env []string) (Result, error) {
	prog, err := i.parser.Parse(strings.NewReader(cmd), "")
	if err != nil {
		return Result{}, fmt.Errorf("command %q in task %q not valid shell syntax: %w", cmd, task, err)
	}

	// os.Environ() is added to env so that if nothing is passed, the
	// process environment is used, but if we do pass env vars these
	// are added as well as all the normal process env vars
	env = append(env, os.Environ()...)

	var result Result
	result.Cmd = cmd
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// Multi write to the stream as well as capture in above buffer
	stdoutMultiWriter := io.MultiWriter(stdout, stream.Stdout)
	stderrMultiWriter := io.MultiWriter(stderr, stream.Stderr)

	execHandler := func(interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return interp.DefaultExecHandler(timeout)
	}

	runner, err := interp.New(
		interp.Params("-e"),
		interp.Env(expand.ListEnviron(env...)),
		interp.ExecHandlers(execHandler),
		interp.OpenHandler(interp.DefaultOpenHandler()),
		interp.StdIO(nil, stdoutMultiWriter, stderrMultiWriter),
		interp.Dir(""),
	)
	if err != nil {
		return Result{}, err
	}

	err = runner.Run(context.Background(), prog)
	if err != nil {
		var status interp.ExitStatus
		if !errors.As(err, &status) {
			// Not an exit status but some other error, bail out
			return Result{}, err
		}

		// Exit status, set it on the result
		result.Status = int(status)
	}

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	return result, nil
}
