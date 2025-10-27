package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func UsageStringFunc(cmd *cobra.Command) func(*cobra.Command) error {
	return func(cmd *cobra.Command) error {
		shortFlags := []string{}
		onlyLongFlags := 0
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Shorthand != "" {
				shortFlags = append(shortFlags, flag.Shorthand)
			} else {
				onlyLongFlags++
			}
		})

		// Extract argument names from Use field
		argNames := parseArgsFromUse(cmd.Use)

		usageStr := fmt.Sprintf("angel %s", cmd.Name())
		for _, arg := range argNames {
			usageStr += fmt.Sprintf(" <%s>", strings.ToUpper(arg))
		}
		if len(shortFlags) > 0 {
			usageStr += fmt.Sprintf(" [-%s]", strings.Join(shortFlags, ""))
		}
		if onlyLongFlags > 0 {
			usageStr += " [--FLAGS]"
		}

		t := table.New().Border(lipgloss.HiddenBorder())
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			longName := "--" + flag.Name
			shortName := ""
			if flag.Shorthand != "" {
				shortName = "-" + flag.Shorthand
			}
			t.Row(shortName, longName, flag.Usage)
		})

		fmt.Println("Usage: " + usageStr)
		fmt.Println("\nAliases: " + strings.Join(cmd.Aliases, ", "))
		fmt.Print("\nFlags:")
		fmt.Println(HelpBodyStyle.Render(t.String()))
		return nil
	}
}

func parseArgsFromUse(use string) []string {
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(use, -1)

	args := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			args = append(args, match[1])
		}
	}

	return args
}
