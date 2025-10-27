package cmd

import (
	"fmt"

	"angel/src/cmd/launchctl"
	"angel/src/core"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

// the pride of this package; the shame of launchctl
func NewStatusCmd(angel *core.Angel) *cobra.Command {
	return &cobra.Command{
		Use:   "status [NAME]",
		Short: "Show service status.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return angel.Daemons.WithMatch(name, false, func(daemon core.Daemon) error {
				printOutput, err := launchctl.Print(daemon)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}
				deamonInfo, err := launchctl.ParseLaunchctlPrint(printOutput)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}

				t := table.New().
					Row("Name", daemon.Name).
					Row("State", deamonInfo.Get("state")).
					Row("Domain", daemon.Domain).
					Row("Active Count", deamonInfo.Get("active count")).
					Row("Source Path", deamonInfo.Get("path")).
					Row("Type", deamonInfo.Get("type"))

				fmt.Println(t.String())
				return nil
			})
		},
	}
}
