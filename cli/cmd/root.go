// Package cmd implements the spok CLI
package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

var (
	version = "dev" // spok version, set at compile time by ldflags
	commit  = ""    // spok version's commit hash, set at compile time by ldflags
)

func BuildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "spok <command> [flags]",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "It's a build system Jim, but not as we know it!",
		Long: heredoc.Doc(`
		
		Longer description of your CLI.
		`),
		Example: heredoc.Doc(`

		$ spok hello

		$ spok version

		$ spok --help
		`),
	}

	// Attach child commands
	rootCmd.AddCommand(
		buildVersionCmd(),
		buildHelloCommand(),
	)

	return rootCmd
}
