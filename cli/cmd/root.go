// Package cmd implements the spok CLI
package cmd

import (
	"fmt"

	"github.com/FollowTheProcess/cli"
	"github.com/FollowTheProcess/spok/cli/app"
	"github.com/FollowTheProcess/spok/iostream"
)

const (
	short = "It's a build system Jim, but not as we know it!"
	long  = `
Spok is a lightweight build system and command runner, inspired by things like
make, just etc.

However, spok offers a number of additional features such as:

- Cleaner, more developer-friendly syntax
- Full cross compatibility
- No dependency on any form of shell
- Load .env files by default
- Incremental runs based on file hashing and sum checks
`
)

var (
	version   = "dev" // spok version, set at compile time by ldflags
	commit    = ""    // spok version's commit hash, set at compile time by ldflags
	buildDate = ""    // spok build date, set at compile time by ldflags
)

// BuildRootCmd builds and returns the root spok CLI command.
func BuildRootCmd() (*cli.Command, error) {
	spok := app.New(iostream.OS())

	root, err := cli.New(
		"spok",
		cli.Short(short),
		cli.Long(long),
		cli.Example("Spok prints all tasks by default", "spok"),
		cli.Example("Run tasks named 'test' and 'lint'", "spok test lint"),
		cli.Example("Show all defined variables in the spokfile", "spok --vars"),
		cli.Example("Format the spokfile", "spok --fmt"),
		cli.Version(version),
		cli.Commit(commit),
		cli.BuildDate(buildDate),
		cli.Flag(&spok.Options.Variables, "vars", cli.NoShortHand, false, "Show all defined variables in the spokfile"),
		cli.Flag(&spok.Options.Fmt, "fmt", cli.NoShortHand, false, "Format the spokfile"),
		cli.Flag(&spok.Options.Spokfile, "spokfile", cli.NoShortHand, "", "The path to the spokfile (defaults to '$CWD/spokfile')"),
		cli.Flag(&spok.Options.Init, "init", cli.NoShortHand, false, "Initialise a new spokfile in $CWD"),
		cli.Flag(&spok.Options.Clean, "clean", 'c', false, "Remove all build artifacts"),
		cli.Flag(&spok.Options.Force, "force", 'f', false, "Bypass file hash checks and force requested tasks to run"),
		cli.Flag(&spok.Options.Debug, "debug", cli.NoShortHand, false, "Show debug info"),
		cli.Flag(&spok.Options.Quiet, "quiet", 'q', false, "Silence all CLI output."),
		cli.Flag(&spok.Options.JSON, "json", 'j', false, "Output task results as JSON"),
		cli.Flag(&spok.Options.Show, "show", 's', false, "Show all tasks defined in the spokfile"),
		cli.Allow(cli.AnyArgs()),
		cli.Run(func(cmd *cli.Command, args []string) error {
			return spok.Run(args)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("could not build spok CLI command: %w", err)
	}

	return root, nil
}
