package cmd

import (
	fp "path/filepath"

	"angel/src/core"

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

			t := table.New()
			return angel.Daemons.WithMatches(pattern, exact, func(daemon core.Daemon) error {
				if daemon.ForUseBy != core.ForApple {
					srcDir := fp.Dir(daemon.SourcePath)
					t.Row(daemon.Name, srcDir)
				}
				return nil
			})
		},
	}

	cmd.Flags().BoolP("exact", "e", false, "Exact match.")

	return cmd
}
