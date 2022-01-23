// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
)

// App represents the spok program
type App struct {
	Out     io.Writer
	Options *Options
	Version string
	Commit  string
}

// Options holds all the flag options for spok
type Options struct {
	Show      string
	Variables bool
	Fmt       bool
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags
func (a *App) Run(tasks []string) error {
	fmt.Fprintf(a.Out, "spok called with args: %v\n", tasks)
	fmt.Fprintf(a.Out, "Flags: %#v\n", a.Options)
	return nil
}
