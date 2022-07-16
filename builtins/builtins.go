// Package builtins implements the built in functions supported by spok,
// it also exports functions which other packages may use to retrieve and call a builtin
// function by name.
package builtins

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/FollowTheProcess/spok/shell"
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
	cmd := command[0]
	result, err := shell.Run(cmd, "", nil)
	if err != nil {
		return "", err
	}
	if !result.Ok() {
		return "", fmt.Errorf("Command %q exited with a non-zero exit code.\nStdout: %s\nStderr: %s", cmd, result.Stdout, result.Stderr)
	}

	return strings.TrimSpace(result.Stdout), nil
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
