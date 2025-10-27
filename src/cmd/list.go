package cmd

import (
	"angel/src/core"
	"angel/src/types"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func NewListCmd(angel *core.Angel) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [PATTERN]",
		Aliases: []string{"ls"},
		Short:   "List services.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := ""
			if len(args) > 0 {
				pattern = args[0]
			}
			exact, _ := cmd.Flags().GetBool("exact")

			matchingDaemons, err := angel.Daemons.GetMatches(pattern, exact)
			if err != nil {
				return err
			}
			daemonsByDomain := core.SortDaemonsByDomain(matchingDaemons)
			for domain, daemons := range daemonsByDomain {
				t := table.New()
				showDynamic, _ := cmd.Flags().GetBool("dynamic")
				if domain == types.DomainUnknown && !showDynamic {
					continue
				}
				for _, daemon := range daemons {
					if daemon.ForUseBy == types.ForApple {
						showApple, _ := cmd.Flags().GetBool("apple")
						if !showApple {
							continue
						}
					}
					t.Row(setStatusIconColor(daemon), daemon.Name, daemon.SourcePath)
				}
				fmt.Println(domain)
				fmt.Println(lipgloss.NewStyle().Padding(0, 2).Render(t.String()))
			}

			return nil
		},
	}

	cmd.Flags().BoolP("exact", "e", false, "Exact match.")
	cmd.Flags().BoolP("apple", "a", false, "Show Apple daemons.")
	cmd.Flags().BoolP("dynamic", "d", false, "Show dynamically loaded daemons (daemons without a plist file).")

	return cmd
}

func setStatusIconColor(daemon types.Daemon) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	if daemon.LastExitCode != nil && *daemon.LastExitCode == 0 {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	}
	return style.Render("●")
}
