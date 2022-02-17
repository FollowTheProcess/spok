// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"

	"github.com/FollowTheProcess/spok/parser"
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
		spokfile, err := os.ReadFile("/Users/tomfleet/Development/spok/spokfile")
		if err != nil {
			return err
		}
		_, err = parser.New(string(spokfile)).Parse()
		if err != nil {
			return err
		}
	}

	return nil
}
