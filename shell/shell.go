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
	"fmt"
	"os"
	"strings"
	"time"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// The default timeout after which shell commands will be aborted.
const timeout = 15 * time.Second

// Result holds the result of running a shell command.
type Result struct {
	Stdout string // The stdout of the command
	Stderr string // The stderr of the command
	Status int    // The exit status of the command
}

// Run runs the shell command cmd belonging to task with
// environment variables set, if env is empty or nil, os.Environ is used.
func Run(cmd, task string, env []string) (Result, error) {
	prog, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return Result{}, fmt.Errorf("Command %q in task %q not valid shell syntax: %w", cmd, task, err)
	}

	if len(env) == 0 {
		env = os.Environ()
	}

	var result Result
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runner, err := interp.New(
		interp.Params("-e"),
		interp.Env(expand.ListEnviron(env...)),
		interp.ExecHandler(interp.DefaultExecHandler(timeout)),
		interp.OpenHandler(interp.DefaultOpenHandler()),
		interp.StdIO(nil, stdout, stderr),
		interp.Dir(""),
	)
	if err != nil {
		return Result{}, err
	}

	// TODO: Context?
	err = runner.Run(context.Background(), prog)
	if err != nil {
		// If ok then the error is an exit status, if not it's some other error
		status, ok := interp.IsExitStatus(err)
		if !ok {
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
