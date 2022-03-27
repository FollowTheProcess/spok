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
	Show      string // The --show flag
	Spokfile  string // The --spokfile flag
	Variables bool   // The --variables flag
	Fmt       bool   // The --fmt flag
	Init      bool   // The --init flag
	Clean     bool   // The --clean flag
	Check     bool   // The --check flag
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags.
func (a *App) Run(tasks []string) error {
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
	case a.Options.Check:
		fmt.Fprintln(a.Out, "Check spokfile for syntax errors")
	default:
		fmt.Fprintf(a.Out, "Running tasks: %v\n", tasks)
	}

	return nil
}
