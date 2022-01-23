package cmd

import (
	"fmt"

	"github.com/FollowTheProcess/spok/cli/app"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

func buildHelloCommand() *cobra.Command {
	app := app.New()

	listCmd := &cobra.Command{
		Use:   "hello",
		Args:  cobra.NoArgs,
		Short: "Say hello",
		Long: heredoc.Doc(`
		
		Say hello.

		The hello command is a silly command that simply
		prints a hello welcome message
		`),
		Example: heredoc.Doc(`
		
		$ spok hello
		`),

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Hello(); err != nil {
				return fmt.Errorf("error saying hello: %w", err)
			}
			return nil
		},
	}

	return listCmd
}
