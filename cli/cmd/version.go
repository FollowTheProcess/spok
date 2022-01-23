package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func buildVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: "Show spok's version info",
		Run: func(cmd *cobra.Command, args []string) {
			ver := color.CyanString("spok version")
			sha := color.CyanString("commit")

			fmt.Printf("%s: %s\n", ver, version)
			fmt.Printf("%s: %s\n", sha, commit)
		},
	}

	return versionCmd
}
