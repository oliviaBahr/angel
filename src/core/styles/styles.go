package styles

import "github.com/charmbracelet/lipgloss"

var (
	greenStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))      // light green
	magentaStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F37878")) // light magenta
	dimmedStyle   = lipgloss.NewStyle().Faint(true)
	dimunderStyle = lipgloss.NewStyle().Faint(true).Underline(true)
	boldStyle     = lipgloss.NewStyle().Bold(true)
	indentedStyle = lipgloss.NewStyle().Padding(0, 2)
	underStyle    = lipgloss.NewStyle().Underline(true)
)

func Green(s string) string {
	return greenStyle.Render(s)
}

func Magenta(s string) string {
	return magentaStyle.Render(s)
}

func Dimmed(s string) string {
	return dimmedStyle.Render(s)
}

func Dimunder(s string) string {
	return dimunderStyle.Render(s)
}

func Bold(s string) string {
	return boldStyle.Render(s)
}

func Indented(s string) string {
	return indentedStyle.Render(s)
}

func Under(s string) string {
	return underStyle.Render(s)
}
