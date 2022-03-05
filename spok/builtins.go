package spok

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// builtin is a spok built in function.
type builtin func(...string) (string, error)

// package scoped map mapping the names of the builtins to their underlying function.
var builtins = map[string]builtin{
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
	if len(command) == 0 {
		return "", errors.New("exec requires at least 1 argument")
	}
	var cmd *exec.Cmd
	bin := command[0]
	if len(command) > 1 {
		args := command[1:]
		cmd = exec.Command(bin, args...)
	} else {
		cmd = exec.Command(bin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %v exited with a non-zero exit code, stderr: %s", command, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
