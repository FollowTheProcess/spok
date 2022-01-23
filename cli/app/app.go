// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"
)

var (
	version = "dev" // spok version, set at compile time by ldflags
	commit  = ""    // spok version's commit hash, set at compile time by ldflags
)

// App represents the spok program
type App struct {
	Out   io.Writer
	Flags *Flags
}

// Flags holds all the flag options for spok
type Flags struct {
	Show      string
	Version   bool
	Variables bool
}

// New creates a new default App configured to talk to os.Stdout
func New(flags *Flags) *App {
	return &App{Out: os.Stdout, Flags: flags}
}

func (a *App) Run(args []string) error {
	fmt.Fprintf(a.Out, "spok called with args: %v\n", args)
	fmt.Fprintf(a.Out, "Flags: %#v\n", a.Flags)
	return nil
}
