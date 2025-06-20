package main

import (
	"os"

	"go.followtheprocess.codes/msg"
	"go.followtheprocess.codes/spok/cli/cmd"
)

func main() {
	if err := run(); err != nil {
		msg.Error("%s", err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd, err := cmd.BuildRootCmd()
	if err != nil {
		return err
	}
	return rootCmd.Execute()
}
