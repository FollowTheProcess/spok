// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
)

// App represents the spok program.
type App struct {
	Out     io.Writer
	Options *Options
}

// Options holds all the flag options for spok, these will be at their zero values
// if the flags were not set and the value of the flag otherwise.
type Options struct {
	Show      string
	Spokfile  string
	Variables bool
	Fmt       bool
	Init      bool
	Clean     bool
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags.
func (a *App) Run(tasks []string) error {
	fmt.Fprintln(a.Out, "tasks:", tasks)
	fmt.Fprintf(a.Out, "flags: %#v\n", a.Options)

	switch {
	case a.Options.Fmt:
		fmt.Fprintln(a.Out, "Format spokfile")
	case a.Options.Variables:
		fmt.Fprintln(a.Out, "Show defined variables")
	case a.Options.Show != "":
		fmt.Fprintf(a.Out, "Show source code for task: %s\n", a.Options.Show)
	case a.Options.Spokfile != "":
		fmt.Fprintf(a.Out, "Using spokfile at: %s\n", a.Options.Spokfile)
	case a.Options.Clean:
		fmt.Fprintln(a.Out, "Clean built artifacts")
	default:
		fmt.Fprintf(a.Out, "No flags, run the following tasks: %v\n", tasks)
	}

	return nil
}
