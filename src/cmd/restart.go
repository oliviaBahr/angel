package cmd

import (
	"fmt"

	"angel/src/cmd/launchctl"
	"angel/src/core"

	"github.com/spf13/cobra"
)

func NewRestartCmd(angel *core.Angel) *cobra.Command {
	return &cobra.Command{
		Use:   "restart [NAME]",
		Short: "Restart a service.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return angel.Daemons.WithMatch(name, false, func(daemon core.Daemon) error {
				output, err := launchctl.KickstartKill(daemon)
				if err != nil {
					return err
				}
				fmt.Print(string(output))
				fmt.Printf("restarted %s\n", daemon.Name)
				return nil
			})
		},
	}
}
