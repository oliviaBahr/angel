package cmd

import (
	"fmt"
	"os"

	"angel/src/core"

	"github.com/spf13/cobra"
)

func NewShowCmd(angel *core.Angel) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [NAME]",
		Short: "Show service daemon.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_ = args

			return angel.Daemons.WithMatch(name, false, func(daemon core.Daemon) error {
				content, err := os.ReadFile(daemon.SourcePath)
				if err != nil {
					return err
				}
				fmt.Print(string(content))
				return nil
			})
		},
	}

	cmd.Flags().StringP("format", "f", "pretty", "Format to show. (xml|json|pretty)")

	return cmd
}
