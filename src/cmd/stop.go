package cmd

import (
	"fmt"

	"angel/src/cmd/launchctl"
	"angel/src/core"

	"github.com/spf13/cobra"
)

func NewStopCmd(angel *core.Angel) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [NAME]",
		Short: "Stop a service.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return angel.Daemons.WithMatch(name, false, func(daemon core.Daemon) error {
				output, err := launchctl.Kill(daemon)
				if err != nil {
					return err
				}
				fmt.Print(string(output))
				fmt.Printf("stopped %s\n", daemon.Name)
				return nil
			})
		},
	}
}
