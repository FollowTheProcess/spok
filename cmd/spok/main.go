package main

import (
	"fmt"
	"os"

	"github.com/FollowTheProcess/spok/cli/cmd"
	"github.com/fatih/color"
)

func main() {
	rootCmd := cmd.BuildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgHiRed, color.Bold).Fprint(os.Stderr, "✘ Error: ")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
