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
	Flags   *Flags
	Version string
	Commit  string
}

// Flags holds all the flag options for spok
type Flags struct {
	Show      string
	Version   bool
	Variables bool
}

func (a *App) Run(args []string) error {
	if a.Flags.Version {
		return a.showVersion()
	}
	fmt.Fprintf(a.Out, "spok called with args: %v\n", args)
	fmt.Fprintf(a.Out, "Flags: %#v\n", a.Flags)
	return nil
}

func (a *App) showVersion() error {
	fmt.Println("Version:", a.Version)
	fmt.Println("Commit:", a.Commit)
	return nil
}
