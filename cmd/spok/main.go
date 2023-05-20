package main

import (
	"os"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/spok/cli/cmd"
)

func main() {
	rootCmd := cmd.BuildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		msg.Error("%s", err)
		os.Exit(1)
	}
}
