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
	"time"

	"github.com/joho/godotenv"
	"go.followtheprocess.codes/hue"
	"go.followtheprocess.codes/hue/tabwriter"
	"go.followtheprocess.codes/msg"
	"go.followtheprocess.codes/spok/cache"
	"go.followtheprocess.codes/spok/file"
	"go.followtheprocess.codes/spok/iostream"
	"go.followtheprocess.codes/spok/logger"
	"go.followtheprocess.codes/spok/parser"
	"go.followtheprocess.codes/spok/shell"
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

const (
	filePerms = 0o666 // The permissions to use when creating the spokfile
	minWidth  = 1     // The minimum width of columns in the output
	tabWidth  = 8     // The width of tabs in the output
	padding   = 2     // The padding between columns in the output
	padChar   = ' '   // The character to use for padding in the output
	flags     = 0     // Tabwriter flags
)

const (
	titleStyle = hue.Bold
	taskStyle  = hue.Cyan | hue.Bold
	descStyle  = hue.BrightBlack | hue.Italic
)

// App represents the spok program.
type App struct {
	stream  iostream.IOStream // Where spok writes output to
	Options *Options          // All the CLI options
	logger  logger.Logger     // Spok's logger, prints debug messages to stderr if --debug is used
}

// Options holds all the flag options for spok, these will be at their zero values
// if the flags were not set and the value of the flag otherwise.
type Options struct {
	Spokfile  string // The path to the spokfile (defaults to find, overridden by --spokfile)
	Variables bool   // The --vars flag
	Fmt       bool   // The --fmt flag
	Init      bool   // The --init flag
	Clean     bool   // The --clean flag
	Force     bool   // The --force flag
	Debug     bool   // The --debug flag
	Quiet     bool   // The --quiet flag
	JSON      bool   // The --json flag
	Show      bool   // The --show flag
}

// New creates and returns a new App.
func New(stream iostream.IOStream) *App {
	options := &Options{}
	spok := &App{
		stream:  stream,
		Options: options,
	}
	return spok
}

// Run is the entry point to the spok program, the only arguments spok accepts are names
// of tasks, all other logic is handled via flags.
func (a *App) Run(tasks []string) error {
	if a.Options.Init {
		return a.initialise()
	}

	if a.Options.Quiet {
		if a.Options.Debug {
			return errors.New("--debug cannot be used with --quiet")
		}
		a.setStream(iostream.Null())
	}

	// If we want task output as json, we don't want it printing
	// to stdout too
	if a.Options.JSON {
		a.setStream(iostream.Null())
	}

	if err := a.setup(); err != nil {
		return err
	}
	// Flush the logger
	defer a.logger.Sync() //nolint: errcheck

	parseStart := time.Now()
	contents, err := os.ReadFile(a.Options.Spokfile)
	if err != nil {
		return err
	}

	tree, err := parser.New(string(contents)).Parse()
	if err != nil {
		return err
	}
	a.logger.Debug("Parsed spokfile at %s in %v", a.Options.Spokfile, time.Since(parseStart))

	spokfile, err := file.New(tree, filepath.Dir(a.Options.Spokfile), a.logger)
	if err != nil {
		return err
	}

	runner := shell.NewIntegratedRunner()

	switch {
	case a.Options.Fmt:
		msg.Finfo(a.stream.Stdout, "Formatting spokfile at %q", a.Options.Spokfile)
		return os.WriteFile(a.Options.Spokfile, []byte(tree.String()), filePerms)
	case a.Options.Variables:
		return a.showVariables(spokfile)
	case a.Options.Clean:
		return a.handleClean(spokfile, runner)
	case a.Options.Show:
		return a.showTasks(spokfile)
	default:
		if len(tasks) == 0 {
			// No tasks provided, handle default actions
			return a.handleDefault(spokfile, runner)
		}

		a.logger.Debug("Running requested tasks: %v", tasks)

		return a.runTasks(spokfile, runner, tasks...)
	}
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

	log, err := logger.NewZapLogger(a.Options.Debug)
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

	// Ensure we make the spokfile path absolute in case the user
	// provided --spokfile with a relative path
	a.Options.Spokfile, err = filepath.Abs(a.Options.Spokfile)
	if err != nil {
		return err
	}

	if filepath.Base(a.Options.Spokfile) != file.NAME {
		return fmt.Errorf("invalid spokfile file name. Got %s, Expected %s", filepath.Base(a.Options.Spokfile), file.NAME)
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
		return fmt.Errorf("could not load .env file: %w", err)
	}
	a.logger.Debug("Loaded .env file at %s", dotenvPath)

	return nil
}

// Initialise writes the demo spokfile to the cwd.
func (a *App) initialise() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, "spokfile")
	gitIgnorePath := filepath.Join(cwd, ".gitignore")

	if exists(path) {
		return fmt.Errorf("spokfile already exists at %s", path)
	}
	if err = os.WriteFile(path, []byte(demoSpokfile), filePerms); err != nil {
		return err
	}

	file, err := os.OpenFile(gitIgnorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerms)
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

// runTasks is a helper that runs the request spokfile tasks.
func (a *App) runTasks(spokfile *file.SpokFile, runner shell.Runner, tasks ...string) error {
	results, err := spokfile.Run(a.stream, runner, a.Options.Force, tasks...)
	if err != nil {
		return err
	}

	for _, result := range results {
		if !result.Ok() {
			for _, cmd := range result.CommandResults {
				if !cmd.Ok() {
					// We've found the one
					return fmt.Errorf("command %q in task %q exited with status %d", cmd.Cmd, result.Task, cmd.Status)
				}
			}
		}
		if result.Skipped {
			msg.Fwarn(a.stream.Stdout, "Task %q skipped as none of it's dependencies have changed", result.Task)
		} else {
			msg.Fsuccess(a.stream.Stdout, "Task %q completed successfully", result.Task)
		}
	}

	if a.Options.JSON {
		text, err := results.JSON()
		if err != nil {
			return err
		}
		fmt.Println(text)
	}

	return nil
}

// show Tasks shows a pretty representation of the defined tasks and their
// docstrings in alphabetical order.
func (a *App) showTasks(spokfile *file.SpokFile) error {
	writer := tabwriter.NewWriter(a.stream.Stdout, 0, tabWidth, 1, '\t', tabwriter.AlignRight)

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
	writer := tabwriter.NewWriter(a.stream.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.AlignRight)

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

// handleClean removes all declared outputs in the spokfile, either by using spok's own
// cleaning of all declared outputs and it's own cache, or by a custom task written
// by the user.
func (a *App) handleClean(spokfile *file.SpokFile, runner shell.Runner) error {
	if spokfile.HasTask("clean") {
		return a.runTasks(spokfile, runner, "clean")
	}
	return a.clean(spokfile)
}

// handleDefault implements the default actions for spok, this defaults to
// showing all the defined tasks but if the user has a task named "default"
// this will be run instead.
func (a *App) handleDefault(spokfile *file.SpokFile, runner shell.Runner) error {
	if spokfile.HasTask("default") {
		return a.runTasks(spokfile, runner, "default")
	}
	return a.showTasks(spokfile)
}

// clean is the default implementation of --clean if the user has
// not defined a clean task in the spokfile itself.
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
				return fmt.Errorf("named output %s is not defined", namedOutput)
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
		msg.Fsuccess(a.stream.Stdout, "Nothing to remove")
		return nil
	}

	for _, file := range toRemove {
		err := os.RemoveAll(file)
		if err != nil {
			return fmt.Errorf("could not remove %s: %w", file, err)
		}
		fmt.Fprintf(a.stream.Stdout, "Removed %s\n", file)
	}
	msg.Fsuccess(a.stream.Stdout, "Done")
	return nil
}

// setStream reassigns all the app's IO streams to match the one passed in.
func (a *App) setStream(stream iostream.IOStream) {
	a.stream = stream
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
