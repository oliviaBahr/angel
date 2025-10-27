package cmd

import (
	"fmt"
	"reflect"

	"angel/src/cmd/launchctl"
	"angel/src/core"
	"angel/src/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

// the pride of this package; the shame of launchctl
func NewStatusCmd(angel *core.Angel) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [NAME]",
		Short: "Show service status.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return angel.Daemons.WithMatch(name, false, func(daemon types.Daemon) error {
				printOutput, err := launchctl.Print(daemon)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}
				deamonInfo, err := launchctl.ParseLaunchctlPrint(printOutput)
				if err != nil {
					return fmt.Errorf("failed to get status: %w", err)
				}

				fmt.Println(core.BoldStyle.Render(daemon.Name) + " " + getStatusStr(deamonInfo.Get("state")))

				t := table.New().Border(lipgloss.HiddenBorder()).
					Row("Domain:", daemon.DomainStr).
					Row("Active:", deamonInfo.Get("active count")).
					Row("Source:", deamonInfo.Get("path")).
					Row("Type:", deamonInfo.Get("type"))

				fmt.Print(t)

				if long, _ := cmd.Flags().GetBool("long"); long {
					t := table.New().Border(lipgloss.HiddenBorder())
					if daemon.Plist != nil {
						v := reflect.ValueOf(daemon.Plist).Elem()

						for i := 0; i < v.NumField(); i++ {
							field := v.Field(i)

							// Skip zero values
							if !field.IsValid() || field.IsZero() {
								continue
							}

							fieldName := v.Type().Field(i).Name
							fieldValue := fmt.Sprintf("%v", field.Interface())
							t.Row(fieldName+":", fieldValue)
						}
					}

					// Add all launchctl data to table
					if dataMap := deamonInfo.GetAll(); dataMap != nil {
						for key, value := range dataMap {
							valueStr := fmt.Sprintf("%v", value)
							t.Row(key+":", valueStr)
						}
					}
					fmt.Print(t)
				}
				fmt.Print("\n")
				return nil
			})
		},
	}
	cmd.Flags().BoolP("long", "l", false, "Show long output.")
	return cmd
}

func getStatusStr(status string) string {
	dot := "●"
	switch status {
	case "running":
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("●")
	case "stopped":
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("●")
	case "launched":
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("●")
	case "exited":
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("●")
	default:
		dot = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("●")
	}
	return dot + " " + status
}
