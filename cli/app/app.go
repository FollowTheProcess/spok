// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/spok/cache"
	"github.com/FollowTheProcess/spok/file"
	"github.com/FollowTheProcess/spok/parser"
	"github.com/fatih/color"
	"github.com/juju/ansiterm/tabwriter"
	"go.uber.org/zap"
)

// App represents the spok program.
type App struct {
	out     io.Writer          // Where to write to
	Options *Options           // All the CLI options
	printer *msg.Printer       // Spok's printer
	logger  *zap.SugaredLogger // The spok logger
}

// Options holds all the flag options for spok, these will be at their zero values
// if the flags were not set and the value of the flag otherwise.
type Options struct {
	Show      string // The --show flag
	Spokfile  string // The path to the spokfile (defaults to find, overridden by --spokfile)
	Variables bool   // The --variables flag
	Fmt       bool   // The --fmt flag
	Init      bool   // The --init flag
	Clean     bool   // The --clean flag
	Force     bool   // The --force flag
	Sync      bool   // The --sync flag
	Verbose   bool   // The --verbose flag
	Quiet     bool   // The --quiet flag
}

// New creates and returns a new App.
func New(out io.Writer) *App {
	options := &Options{}
	printer := msg.Default()
	spok := &App{
		out:     out,
		Options: options,
		printer: printer,
	}
	return spok
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags.
func (a *App) Run(tasks []string) error {
	if err := a.setup(); err != nil {
		return err
	}
	// Flush the logger
	defer a.logger.Sync() // nolint: errcheck

	a.logger.Debugf("Parsing spokfile at %s", a.Options.Spokfile)
	spokfile, err := a.parse(a.Options.Spokfile)
	if err != nil {
		return err
	}

	switch {
	case a.Options.Fmt:
		fmt.Fprintln(a.out, "Format spokfile")
	case a.Options.Variables:
		return a.showVariables(spokfile)
	case a.Options.Show != "":
		fmt.Fprintf(a.out, "Show source code for task: %s\n", a.Options.Show)
	case a.Options.Clean:
		return a.cleanOutputs(spokfile)
	case a.Options.Init:
		fmt.Fprintf(a.out, "Create a new spokfile at %s\n", a.Options.Spokfile)
	default:
		switch len(tasks) {
		case 0:
			// No tasks provided, show defined tasks and exit
			return a.showTasks(spokfile)
		default:
			a.logger.Debugf("Running requested tasks: %v", tasks)
			results, err := spokfile.Run(a.Options.Sync, a.Options.Force, tasks...)
			if err != nil {
				return err
			}

			for _, result := range results {
				if !result.Ok() {
					// TODO: Drill down into which one failed and show nice error
					a.logger.Debugf("Command in task %q exited with non-zero status", result.Task)
					for _, cmd := range result.CommandResults {
						if !cmd.Ok() {
							// We've found the one
							return fmt.Errorf("Command %q in task %q exited with status %d\nStdout: %s\nStderr: %s", cmd.Cmd, result.Task, cmd.Status, cmd.Stdout, cmd.Stderr)
						}
					}

					return fmt.Errorf("Task %q failed", result.Task)
				}
				for _, cmd := range result.CommandResults {
					fmt.Fprint(a.out, cmd.Stdout)
				}
			}
		}
	}

	return nil
}

// parse fully parses a spokfile located at path and returns the
// concrete Spokfile object along with any errors.
func (a *App) parse(path string) (*file.SpokFile, error) {
	contents, err := os.ReadFile(a.Options.Spokfile)
	if err != nil {
		return nil, err
	}

	tree, err := parser.New(string(contents)).Parse()
	if err != nil {
		return nil, err
	}

	spokfile, err := file.New(tree, filepath.Dir(a.Options.Spokfile))
	if err != nil {
		return spokfile, err
	}

	return spokfile, nil
}

// setup performs one time initialise actions like finding the cwd and $HOME
// and setting the path to the spokfile.
func (a *App) setup() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if a.Options.Spokfile == "" {
		// The --spokfile flag has not been set, find the default
		// findErr to avoid shadowing Getwd err
		spokfilePath, findErr := file.Find(cwd, home)
		if findErr != nil {
			return findErr
		}
		a.Options.Spokfile = spokfilePath
	}

	// Ensure we make the spokfile path absolute incase the user
	// provided --spokfile with a relative path
	a.Options.Spokfile, err = filepath.Abs(a.Options.Spokfile)
	if err != nil {
		return err
	}

	if filepath.Base(a.Options.Spokfile) != file.Name {
		return fmt.Errorf("Invalid spokfile file name. Got %s, Expected %s", filepath.Base(a.Options.Spokfile), file.Name)
	}

	// Set up the logger
	level := zap.InfoLevel
	if a.Options.Verbose {
		level = zap.DebugLevel
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	cfg.EncoderConfig.TimeKey = ""
	logger, err := cfg.Build(zap.IncreaseLevel(level))
	if err != nil {
		return err
	}
	sugar := logger.Sugar()
	a.logger = sugar

	// Initialise the cache
	if err := cache.Init(filepath.Dir(a.Options.Spokfile)); err != nil {
		return err
	}

	return nil
}

// show Tasks shows a pretty representation of the defined tasks and their
// docstrings in alphabetical order.
func (a *App) showTasks(spokfile *file.SpokFile) error {
	writer := tabwriter.NewWriter(a.out, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)
	taskStyle := color.New(color.FgHiCyan, color.Bold)
	descStyle := color.New(color.FgHiBlack, color.Italic)

	// sort.Sort(task.ByName(spokfile.Tasks))
	fmt.Fprintf(a.out, "Tasks defined in %s:\n", spokfile.Path)
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

// showVariables shows all the defined spokfile variables and their set values.
func (a *App) showVariables(spokfile *file.SpokFile) error {
	writer := tabwriter.NewWriter(a.out, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)

	fmt.Fprintf(a.out, "Variables defined in %s:\n", spokfile.Path)
	titleStyle.Fprintln(writer, "Name\tValue")

	names := make([]string, 0, len(spokfile.Vars))
	for n := range spokfile.Vars {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		line := fmt.Sprintf("%s\t%s\n", name, spokfile.Vars[name])
		fmt.Fprint(writer, line)
	}
	return writer.Flush()
}

// cleanOutputs removes all declared outputs in the spokfile.
func (a *App) cleanOutputs(spokfile *file.SpokFile) error {
	for _, task := range spokfile.Tasks {
		for _, fileOutput := range task.FileOutputs {
			resolved, err := filepath.Abs(fileOutput)
			if err != nil {
				return err
			}
			err = os.RemoveAll(resolved)
			if err != nil {
				return fmt.Errorf("Could not remove %s: %v", resolved, err)
			}
			msg.Goodf("Removed %s", resolved)
		}

		for _, namedOutput := range task.NamedOutputs {
			// NamedOutputs are just idents that point to filepaths
			actual, ok := spokfile.Vars[namedOutput]
			if !ok {
				return fmt.Errorf("Named output %s is not defined", namedOutput)
			}
			resolved, err := filepath.Abs(actual)
			if err != nil {
				return err
			}
			err = os.RemoveAll(resolved)
			if err != nil {
				return fmt.Errorf("Could not remove %s: %v", resolved, err)
			}
			msg.Goodf("Removed %s", resolved)
		}
	}
	return nil
}
