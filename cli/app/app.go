// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"

	"github.com/FollowTheProcess/spok/pkg/hello"
)

// App represents the spok program
type App struct {
	Out io.Writer
}

// New creates a new default App configured to talk to os.Stdout
func New() *App {
	return &App{Out: os.Stdout}
}

// Hello is the handler for the spok hello command
func (a *App) Hello() error {

	message := hello.Say("A Thing")

	fmt.Fprintln(a.Out, message)
	return nil
}
