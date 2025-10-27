package cmd

import (
	"angel/src/core"
	"angel/src/types"
	"fmt"
	"slices"
	"strconv"

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
			sortBy := cmd.Flags().Lookup("sort").Value.String()
			sortFn := core.SortDaemonsByParentDir
			if sortBy == "domain" {
				sortFn = core.SortDaemonsByDomain
			}
			sortedDaemons := sortFn(matchingDaemons)

			// Get keys and sort them alphabetically
			keys := make([]string, 0, len(sortedDaemons))
			for key := range sortedDaemons {
				keys = append(keys, key)
			}
			slices.Sort(keys)

			for _, key := range keys {
				daemons := sortedDaemons[key]
				rows := [][]string{}
				for _, daemon := range daemons {
					if daemon.ForUseBy == types.ForApple {
						showApple, _ := cmd.Flags().GetBool("apple")
						if !showApple {
							continue
						}
					}
					showDynamic, _ := cmd.Flags().GetBool("dynamic")
					if daemon.Domain == types.DomainUnknown && !showDynamic {
						continue
					}
					rows = append(rows, []string{getExitCode(daemon), daemon.Name, daemon.SourcePath})
				}
				if len(rows) > 0 {
					t := table.New().Border(lipgloss.HiddenBorder()).Rows(rows...)
					fmt.Print(lipgloss.NewStyle().Underline(true).Bold(true).Render(key))
					fmt.Println(t.String())
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP("exact", "e", false, "Exact match.")
	cmd.Flags().BoolP("apple", "a", false, "Show Apple daemons.")
	cmd.Flags().BoolP("dynamic", "d", false, "Show dynamically loaded daemons (daemons without a plist file).")
	cmd.Flags().FuncP("sort", "s", "Sort by. (parent|domain)", func(value string) error {
		if !slices.Contains([]string{"parent", "domain"}, value) {
			return fmt.Errorf("invalid sort value: %s", value)
		}
		return nil
	})
	return cmd
}

func getExitCode(daemon types.Daemon) string {
	if daemon.LastExitCode != nil {
		return strconv.Itoa(*daemon.LastExitCode)
	}
	return "-"
}
