// Package builtins implements the built in functions supported by spok,
// it also exports functions which other packages may use to retrieve and call a builtin
// function by name.
package builtins

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

// Builtin is a spok built in function.
type Builtin func(...string) (string, error)

// read-only package scoped map mapping the names of the builtins to their underlying function
// client packages access this through the Get function below.
var builtins = map[string]Builtin{
	"join": join,
	"exec": execute,
}

// join joins up filepath parts with an OS specific separator.
func join(parts ...string) (string, error) {
	return filepath.Join(parts...), nil
}

// execute executes an external command and returns the stdout to the caller
// leading and trailing whitespace will be trimmed prior to returning, if the command
// returns a non-zero exit code, this will be reported as an error and the stderr of the
// underlying command will be included in the error message.
func execute(command ...string) (string, error) {
	if len(command) != 1 {
		return "", errors.New("exec takes the shell command as a single string argument")
	}
	com, err := shlex.Split(command[0])
	if err != nil {
		return "", fmt.Errorf("could not split command into parts: %w", err)
	}
	var cmd *exec.Cmd
	bin := com[0]
	if len(com) > 1 {
		args := com[1:]
		cmd = exec.Command(bin, args...)
	} else {
		cmd = exec.Command(bin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %v exited with a non-zero exit code.\nstdout: %s\nstderr: %s", command, stdout.String(), stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Get looks up a builtin function by name, it returns the Builtin and a bool
// indicating whether or not it was found in the same way that item, ok is used
// for maps.
func Get(name string) (Builtin, bool) {
	fn, ok := builtins[name]
	if !ok {
		return nil, false
	}
	return fn, true
}
