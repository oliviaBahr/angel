package core

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Style functions for consistent formatting
var (
	green    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render      // light green
	magenta  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F37878")).Render // light magenta
	dimmed   = lipgloss.NewStyle().Faint(true).Render
	dimunder = lipgloss.NewStyle().Faint(true).Underline(true).Render
	bold     = lipgloss.NewStyle().Bold(true).Render
	indented = lipgloss.NewStyle().Padding(0, 2).Render
)

// customHelpPrinter provides custom help formatting that hides default values from flag names
func CustomHelpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	fmt.Println()
	selected := ctx.Selected()
	if selected == nil {
		return printRootHelp(ctx)
	}
	return printCommandHelp(selected)
}

func printRootHelp(ctx *kong.Context) error {
	fmt.Println(ctx.Model.Help)

	fmt.Println(green("\nUsage"))
	fmt.Println(indented(ctx.Model.Name + magenta(" <command>") + shortFlags(ctx.Model.Flags)))

	fmt.Println(green("\nCommands"))
	fmt.Println(indented(MakeTable(ctx.Model.Leaves(false), true).String()))

	if len(ctx.Model.AllFlags(false)) > 0 {
		fmt.Println(green("\nFlags"))
		fmt.Println(indented(MakeTable(ctx.Model.AllFlags(false), true).String()))
	}

	fmt.Println()
	return nil
}

func printCommandHelp(selected *kong.Node) error {
	fmt.Println(selected.Help)

	fmt.Println(green("\nUsage"))
	fmt.Println(indented(CommandUsageStr(selected)))

	fmt.Println(green("\nArguments"))
	fmt.Println(indented(MakeTable(selected.Positional, false).String()))

	if len(selected.Flags) > 0 {
		fmt.Println(green("\nFlags"))
		fmt.Println(indented(MakeTable(selected.AllFlags(false), false).String()))
	}

	fmt.Println()
	return nil
}

func makeTable(rows [][]string) *table.Table {
	return table.New().
		Rows(rows...).
		BorderTop(false).
		BorderBottom(false).
		BorderRight(false).
		BorderLeft(false).
		BorderColumn(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 1 {
				return lipgloss.NewStyle().Padding(0, 2)
			}
			return lipgloss.NewStyle()
		})
}

func MakeTable(data interface{}, showHelp bool) *table.Table {
	var rows [][]string

	switch v := data.(type) {
	case []*kong.Positional:
		for _, arg := range v {
			rows = append(rows, []string{magenta(arg.Summary()), arg.Help})
		}
	case []*kong.Node:
		for _, cmd := range v {
			rows = append(rows, []string{CommandUsageStr(cmd), cmd.Help})
		}
	case [][]*kong.Flag:
		for _, flagGroup := range v {
			for _, flag := range flagGroup {
				if flag.Name == "help" && !showHelp {
					continue
				}
				flagName := fmt.Sprintf("-%c --%s", flag.Short, flag.Name)
				rows = append(rows, []string{flagName, FormatHelpStr(flag.Value)})
			}
		}
	}

	return makeTable(rows)
}

func CommandUsageStr(cmd *kong.Node) string {
	var sb strings.Builder

	if cmd.Name != "" {
		sb.WriteString(bold(cmd.Name))
	}
	if aliasesStr := strings.Join(cmd.Aliases, ","); aliasesStr != "" {
		sb.WriteString(dimmed(fmt.Sprintf(" (%s)", aliasesStr)))
	}
	for _, arg := range cmd.Positional {
		if arg.Required {
			sb.WriteString(magenta(fmt.Sprintf(" <%s>", arg.Name)))
		} else {
			sb.WriteString(dimmed(magenta(fmt.Sprintf(" <%s>", arg.Name))))
		}
	}
	sb.WriteString(shortFlags(cmd.Flags))
	return sb.String()
}

func shortFlags(flags []*kong.Flag) string {
	if len(flags) > 0 {
		var flagstr strings.Builder
		for _, flag := range flags {
			if flag.Short != 0 {
				flagstr.WriteString(string(flag.Short))
			}
		}
		if flagstr.Len() > 0 {
			return fmt.Sprintf(" [-%s]", flagstr.String())
		}
	}
	return ""
}

// add possible enum values to the help string
func FormatHelpStr(value *kong.Value) string {
	if value.Enum != "" {
		enumValues := strings.Split(value.Enum, ",")

		var formattedValues []string
		for _, enumValue := range enumValues {
			if enumValue == value.Default {
				formattedValues = append(formattedValues, dimunder(enumValue))
			} else {
				formattedValues = append(formattedValues, enumValue)
			}
		}

		enumText := strings.Join(formattedValues, "|")
		// dimming the whole thing at once doesn't work idk why
		value.Help += dimmed(" [") + dimmed(enumText) + dimmed("]")
	}
	return value.Help
}
