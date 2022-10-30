// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/spok/file"
	"github.com/FollowTheProcess/spok/parser"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/juju/ansiterm/tabwriter"
	"go.uber.org/zap"
)

// App represents the spok program.
type App struct {
	stdout  io.Writer          // Where to write to
	stderr  io.Writer          // Where to write errors to
	Options *Options           // All the CLI options
	logger  *zap.SugaredLogger // Spok's logger, prints debug messages to stderr if --verbose is used
	printer msg.Printer        // Spok's printer, prints user messages to stdout
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
func New(stdout, stderr io.Writer) *App {
	options := &Options{}
	printer := msg.Default()
	printer.Stdout = stdout
	printer.Stderr = stderr
	spok := &App{
		stdout:  stdout,
		stderr:  stderr,
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
	contents, err := os.ReadFile(a.Options.Spokfile)
	if err != nil {
		return err
	}

	tree, err := parser.New(string(contents)).Parse()
	if err != nil {
		return err
	}

	spokfile, err := file.New(tree, filepath.Dir(a.Options.Spokfile))
	if err != nil {
		return err
	}

	runner := shell.NewIntegratedRunner()

	switch {
	case a.Options.Fmt:
		a.printer.Infof("Formatting spokfile at %q", a.Options.Spokfile)
		return os.WriteFile(a.Options.Spokfile, []byte(tree.String()), 0666)
	case a.Options.Variables:
		return a.showVariables(spokfile)
	case a.Options.Show != "":
		fmt.Fprintf(a.stdout, "Show source code for task: %s\n", a.Options.Show)
	case a.Options.Clean:
		return a.cleanOutputs(spokfile)
	case a.Options.Init:
		fmt.Fprintf(a.stdout, "Create a new spokfile at %s\n", a.Options.Spokfile)
	default:
		if len(tasks) == 0 {
			// No tasks provided, show defined tasks and exit
			return a.showTasks(spokfile)
		}

		a.logger.Debugf("Running requested tasks: %v", tasks)
		results, err := spokfile.Run(a.stdout, runner, a.Options.Sync, a.Options.Force, tasks...)
		if err != nil {
			return err
		}

		for _, result := range results {
			if !result.Ok() {
				a.logger.Debugf("Command in task %q exited with non-zero status", result.Task)
				for _, cmd := range result.CommandResults {
					if !cmd.Ok() {
						// We've found the one
						return fmt.Errorf("Command %q in task %q exited with status %d\nStdout:\n-------\n %s\nStderr:\n-------\n %s", cmd.Cmd, result.Task, cmd.Status, cmd.Stdout, cmd.Stderr)
					}
				}
			}
			if result.Skipped {
				a.printer.Goodf("Task %q up to date", result.Task)
			}
			for _, cmd := range result.CommandResults {
				fmt.Fprint(a.stdout, cmd.Stdout)
			}
		}
	}

	return nil
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

	if filepath.Base(a.Options.Spokfile) != file.NAME {
		return fmt.Errorf("Invalid spokfile file name. Got %s, Expected %s", filepath.Base(a.Options.Spokfile), file.NAME)
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

	// Auto load .env file (if present) to be present in os.Environ
	if err := godotenv.Load(filepath.Join(filepath.Dir(a.Options.Spokfile), ".env")); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			// A missing .env file is not an error, just don't load it in
			// If however it does exist and we can't load then report that
			return fmt.Errorf("Could not load .env file: %w", err)
		}
	}

	return nil
}

// show Tasks shows a pretty representation of the defined tasks and their
// docstrings in alphabetical order.
func (a *App) showTasks(spokfile *file.SpokFile) error {
	writer := tabwriter.NewWriter(a.stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)
	taskStyle := color.New(color.FgHiCyan, color.Bold)
	descStyle := color.New(color.FgHiBlack, color.Italic)

	// sort.Sort(task.ByName(spokfile.Tasks))
	fmt.Fprintf(a.stdout, "Tasks defined in %s:\n", spokfile.Path)
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
	writer := tabwriter.NewWriter(a.stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)

	fmt.Fprintf(a.stdout, "Variables defined in %s:\n", spokfile.Path)
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
	var toRemove []string
	for _, task := range spokfile.Tasks {
		for _, fileOutput := range task.FileOutputs {
			resolved, err := filepath.Abs(fileOutput)
			if err != nil {
				return err
			}
			_, err = os.Stat(resolved)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
			}
			toRemove = append(toRemove, resolved)
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
			_, err = os.Stat(resolved)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
			}
			toRemove = append(toRemove, resolved)
		}
	}

	if len(toRemove) == 0 {
		a.printer.Good("Nothing to remove")
		return nil
	}

	for _, file := range toRemove {
		err := os.RemoveAll(file)
		if err != nil {
			return fmt.Errorf("Could not remove %s: %w", file, err)
		}
		a.printer.Textf("Removed %s", file)
	}
	a.printer.Good("Done")
	return nil
}
