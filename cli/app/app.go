// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/FollowTheProcess/spok/file"
	"github.com/FollowTheProcess/spok/parser"
	"github.com/fatih/color"
	"github.com/juju/ansiterm/tabwriter"
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	spokfilePath, err := file.Find(cwd, home)
	if err != nil {
		return err
	}

	contents, err := os.ReadFile(spokfilePath)
	if err != nil {
		return err
	}

	tree, err := parser.New(string(contents)).Parse()
	if err != nil {
		return err
	}

	spokfile, err := file.New(tree, filepath.Dir(spokfilePath))
	if err != nil {
		return err
	}

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
		switch len(tasks) {
		case 0:
			// No tasks provided, show defined tasks and exit
			return a.showTasks(spokfile)
		default:
			fmt.Fprintf(a.Out, "Running tasks: %v\n", tasks)
		}
	}

	return nil
}

// show Tasks shows a pretty representation of the defined tasks and their
// docstrings in alphabetical order.
func (a *App) showTasks(spokfile file.SpokFile) error {
	writer := tabwriter.NewWriter(a.Out, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)
	taskStyle := color.New(color.FgHiCyan, color.Bold)
	descStyle := color.New(color.FgHiBlack, color.Italic)

	// sort.Sort(task.ByName(spokfile.Tasks))
	fmt.Fprintf(a.Out, "Tasks defined in %s:\n", spokfile.Path)
	titleStyle.Fprintln(writer, "Name\tDescription")

	names := make([]string, 0, len(spokfile.Tasks))
	for n := range spokfile.Tasks {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		line := fmt.Sprintf("%s\t%s\n", taskStyle.Sprint(name), descStyle.Sprint(spokfile.Tasks[name].Doc))
		fmt.Fprint(writer, line)
	}

	return writer.Flush()
}
