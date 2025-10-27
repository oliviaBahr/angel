package cmd

import (
	"angel/src/core"
	"angel/src/core/config"
	"angel/src/core/constants"
	"fmt"
	fp "path/filepath"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss/table"
)

type ListCmd struct {
	Pattern  string `arg:"" optional:"" help:"Pattern to filter services."`
	Exact    bool   `short:"e" help:"Exact match."`
	TestFlag bool   `help:"Test flag without short version."`
}

func (l *ListCmd) Run(a *core.Angel, config *config.Config, ctx *kong.Context) error {
	t := table.New()
	err := a.Daemons.WithMatches(l.Pattern, false, ctx, func(daemon core.Daemon) error {
		if daemon.ForUseBy != constants.ForApple {
			srcDir := fp.Dir(daemon.SourcePath)
			t.Row(daemon.Name, srcDir)
		}
		return nil
	})
	fmt.Println(t.String())
	return err
}
