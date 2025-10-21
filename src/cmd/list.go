package cmd

import (
	"angel/src/core"
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

func (l *ListCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	t := table.New()
	err := angel.WithMatches(l.Pattern, false, ctx, func(daemon core.Daemon) error {
		if daemon.ForUseBy != core.ForApple {
			srcDir := fp.Dir(daemon.SourcePath)
			t.Row(daemon.Name, srcDir)
		}
		return nil
	})
	fmt.Println(t.String())
	return err
}
