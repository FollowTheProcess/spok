// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/spok/cache"
	"github.com/FollowTheProcess/spok/file"
	"github.com/FollowTheProcess/spok/iostream"
	"github.com/FollowTheProcess/spok/logger"
	"github.com/FollowTheProcess/spok/parser"
	"github.com/FollowTheProcess/spok/shell"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/juju/ansiterm/tabwriter"
)

const demoSpokfile string = `# This is a spokfile example

VERSION := "0.3.0"

# Run the unit tests
task test("**/*.go") {
    go test ./...
}

# Which version am I
task version() {
    echo {{.VERSION}}
}
`

const gitIgnoreText string = `
# Ignore the spok cache directory
.spok/
`

// App represents the spok program.
type App struct {
	stream  iostream.IOStream // Where spok writes output to
	Options *Options          // All the CLI options
	logger  logger.Logger     // Spok's logger, prints debug messages to stderr if --verbose is used
	printer msg.Printer       // Spok's printer, prints user messages to stdout
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
	Verbose   bool   // The --verbose flag
	Quiet     bool   // The --quiet flag
}

// New creates and returns a new App.
func New(stream iostream.IOStream) *App {
	options := &Options{}
	printer := msg.Default()
	printer.Stdout = stream.Stdout
	printer.Stderr = stream.Stderr
	spok := &App{
		stream:  stream,
		Options: options,
		printer: printer,
	}
	return spok
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags.
func (a *App) Run(tasks []string) error {
	if a.Options.Init {
		return initialise()
	}
	if err := a.setup(); err != nil {
		return err
	}
	// Flush the logger
	defer a.logger.Sync() // nolint: errcheck

	a.logger.Debug("Parsing spokfile at %s", a.Options.Spokfile)
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
	case a.Options.Clean:
		return a.clean(spokfile)
	default:
		if len(tasks) == 0 {
			// No tasks provided, show defined tasks and exit
			return a.showTasks(spokfile)
		}

		if a.Options.Quiet {
			a.stream = iostream.Null()
		}

		a.logger.Debug("Running requested tasks: %v", tasks)

		results, err := spokfile.Run(a.logger, a.stream, runner, a.Options.Force, tasks...)
		if err != nil {
			return err
		}

		for _, result := range results {
			if !result.Ok() {
				a.logger.Debug("Command in task %q exited with non-zero status", result.Task)
				for _, cmd := range result.CommandResults {
					if !cmd.Ok() {
						// We've found the one
						return fmt.Errorf("Command %q in task %q exited with status %d\nStdout:\n-------\n %s\nStderr:\n-------\n %s", cmd.Cmd, result.Task, cmd.Status, cmd.Stdout, cmd.Stderr)
					}
				}
			}
			if result.Skipped {
				skipStyle := color.New(color.FgYellow, color.Bold)
				skipStyle.Fprintf(a.stream.Stdout, "- Task %q skipped as none of its dependencies have changed\n", result.Task)
			} else {
				a.printer.Goodf("Task %q completed successfully", result.Task)
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

	log, err := logger.NewZapLogger(a.Options.Verbose)
	if err != nil {
		return err
	}
	a.logger = log

	if a.Options.Spokfile == "" {
		// The --spokfile flag has not been set, find the default
		// findErr to avoid shadowing Getwd err
		spokfilePath, findErr := file.Find(a.logger, cwd, home)
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

	a.logger.Debug("Found spokfile at %s", a.Options.Spokfile)

	// Auto load .env file (if present) to be present in os.Environ
	a.logger.Debug("Looking for .env file")
	dotenvPath := filepath.Join(filepath.Dir(a.Options.Spokfile), ".env")

	if !exists(dotenvPath) {
		a.logger.Debug("No .env file found")
		return nil
	}

	if err := godotenv.Load(dotenvPath); err != nil {
		return fmt.Errorf("Could not load .env file: %w", err)
	}
	a.logger.Debug("Loaded .env file at %s", dotenvPath)

	return nil
}

// Initialise writes the demo spokfile to the cwd.
func initialise() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, "spokfile")
	gitIgnorePath := filepath.Join(cwd, ".gitignore")

	if exists(path) {
		return fmt.Errorf("spokfile already exists at %s", path)
	}
	if err = os.WriteFile(path, []byte(demoSpokfile), 0666); err != nil {
		return err
	}

	file, err := os.OpenFile(gitIgnorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(gitIgnoreText)
	if err != nil {
		return err
	}

	return nil
}

// show Tasks shows a pretty representation of the defined tasks and their
// docstrings in alphabetical order.
func (a *App) showTasks(spokfile *file.SpokFile) error {
	writer := tabwriter.NewWriter(a.stream.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)
	taskStyle := color.New(color.FgHiCyan, color.Bold)
	descStyle := color.New(color.FgHiBlack, color.Italic)

	// sort.Sort(task.ByName(spokfile.Tasks))
	fmt.Fprintf(a.stream.Stdout, "Tasks defined in %s:\n", spokfile.Path)
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
	writer := tabwriter.NewWriter(a.stream.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	titleStyle := color.New(color.FgHiWhite, color.Bold)

	fmt.Fprintf(a.stream.Stdout, "Variables defined in %s:\n", spokfile.Path)
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

// clean removes all declared outputs in the spokfile.
func (a *App) clean(spokfile *file.SpokFile) error {
	var toRemove []string
	for _, task := range spokfile.Tasks {
		// Gather up all the declared file outputs
		for _, fileOutput := range task.FileOutputs {
			resolved, err := filepath.Abs(fileOutput)
			if err != nil {
				return err
			}
			_, err = os.Stat(resolved)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					// If it doesn't exist we can ignore the error
					return err
				}
			}
			toRemove = append(toRemove, resolved)
		}

		// Same with all named outputs
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
				if !errors.Is(err, fs.ErrNotExist) {
					// If it doesn't exist we can ignore the error
					return err
				}
			}
			toRemove = append(toRemove, resolved)
		}
	}

	// Finally, add spok's own cache to the clean list
	cachePath := filepath.Join(spokfile.Dir, cache.Dir)
	toRemove = append(toRemove, cachePath)

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

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
