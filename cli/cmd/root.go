// Package cmd implements the spok CLI
package cmd

import (
	"fmt"
	"os"

	"github.com/FollowTheProcess/spok/cli/app"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version     = "dev"                                // spok version, set at compile time by ldflags
	commit      = ""                                   // spok version's commit hash, set at compile time by ldflags
	headerStyle = color.New(color.FgWhite, color.Bold) // Setting header style to use in usage message (usage.go)
)

// BuildRootCmd builds and returns the root spok CLI command
func BuildRootCmd() *cobra.Command {
	options := app.Options{}
	spok := &app.App{
		Out:     os.Stdout,
		Options: options,
		Version: version,
		Commit:  commit,
	}

	rootCmd := &cobra.Command{
		Use:           "spok [tasks]...",
		Version:       version,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "It's a build system Jim, but not as we know it!",
		Long: heredoc.Doc(`
		
		It's a build system Jim, but not as we know it!

		Spok is a lightweight build system and command runner, inspired by things like
		make, just etc.

		However, spok offers a number of additional features such as:

		- Cleaner, more developer-friendly syntax
		- Full cross compatibility
		- Incremental runs based on file hashing and sum checks
		- Parallel execution by default
		`),
		Example: heredoc.Doc(`

		# Spok prints all tasks by default
		$ spok

		# Run tasks named 'test' and 'lint'
		$ spok task lint

		# Show all defined variables in the 'spokfile'
		$ spok --variables

		# Show a task's source code
		$ spok --show <task>

		# Format the spokfile
		$ spok --fmt
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return spok.Run(args)
		},
	}

	flags := rootCmd.Flags()
	flags.BoolVar(&options.Variables, "variables", false, "Show all defined variables in spokfile.")
	flags.StringVar(&options.Show, "show", "", "Show the source code for a task.")
	flags.BoolVar(&options.Fmt, "fmt", false, "Format the spokfile.")

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{printf "Version: %s\nCommit: %s\n"}}`, version, commit))

	return rootCmd
}
