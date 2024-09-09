package main

import (
	"os"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/spok/cli/cmd"
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
