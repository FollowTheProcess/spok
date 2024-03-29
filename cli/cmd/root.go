// Package cmd implements the spok CLI
package cmd

import (
	"github.com/FollowTheProcess/spok/cli/app"
	"github.com/FollowTheProcess/spok/iostream"
	"github.com/spf13/cobra"
)

const (
	short = "It's a build system Jim, but not as we know it!"
	long  = `
It's a build system Jim, but not as we know it!

Spok is a lightweight build system and command runner, inspired by things like
make, just etc.

However, spok offers a number of additional features such as:

- Cleaner, more developer-friendly syntax
- Full cross compatibility
- No dependency on any form of shell
- Load .env files by default
- Incremental runs based on file hashing and sum checks
`
	example = `
# Spok prints all tasks by default
$ spok

# Run tasks named 'test' and 'lint'
$ spok test lint

# Show all defined variables in the 'spokfile'
$ spok --vars

# Format the spokfile
$ spok --fmt`
)

var (
	version   = "dev" // spok version, set at compile time by ldflags
	commit    = ""    // spok version's commit hash, set at compile time by ldflags
	buildDate = ""    // spok build date, set at compile time by ldflags
	builtBy   = ""    // spok built by, set at compile time by ldflags
)

// BuildRootCmd builds and returns the root spok CLI command.
func BuildRootCmd() *cobra.Command {
	spok := app.New(iostream.OS())

	rootCmd := &cobra.Command{
		Use:           "spok [tasks]...",
		Version:       version,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         short,
		Long:          long,
		Example:       example,
		RunE: func(cmd *cobra.Command, args []string) error {
			return spok.Run(args)
		},
	}

	// Attach the flags
	flags := rootCmd.Flags()
	flags.BoolVarP(&spok.Options.Variables, "vars", "V", false, "Show all defined variables in spokfile.")
	flags.BoolVar(&spok.Options.Fmt, "fmt", false, "Format the spokfile.")
	flags.StringVar(&spok.Options.Spokfile, "spokfile", "", "The path to the spokfile (defaults to '$CWD/spokfile').")
	flags.BoolVar(&spok.Options.Init, "init", false, "Initialise a new spokfile in $CWD.")
	flags.BoolVarP(&spok.Options.Clean, "clean", "c", false, "Remove all build artifacts.")
	flags.BoolVarP(&spok.Options.Force, "force", "f", false, "Bypass file hash checks and force running.")
	flags.BoolVarP(&spok.Options.Debug, "debug", "d", false, "Show debug logging output.")
	flags.BoolVarP(&spok.Options.Quiet, "quiet", "q", false, "Silence all CLI output.")
	flags.BoolVarP(&spok.Options.JSON, "json", "j", false, "Output task results as JSON.")
	flags.BoolVarP(&spok.Options.Show, "show", "s", false, "Show all tasks defined in the spokfile.")

	// Set our custom version and usage templates
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetVersionTemplate(versionTemplate)

	return rootCmd
}
