package cmd

import (
	"angel/src/core"
	"angel/src/core/styles"
	"fmt"
	fp "path/filepath"

	"github.com/alecthomas/kong"
)

type ListCmd struct {
	Pattern  string `arg:"" optional:"" help:"Pattern to filter services."`
	Exact    bool   `short:"e" help:"Exact match."`
	TestFlag bool   `help:"Test flag without short version."`
}

func (l *ListCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(l.Pattern, false, ctx, func(matches []core.Daemon) error {
		daemons := make(map[string][]string)
		for _, daemon := range matches {
			if daemon.ForUseBy != core.ForApple {
				srcDir := fp.Dir(daemon.SourcePath)
				daemons[srcDir] = append(daemons[srcDir], daemon.Name)
			}
		}
		for srcDir, names := range daemons {
			fmt.Println("\n" + styles.Under(srcDir))
			for _, name := range names {
				fmt.Println("  -", name)
			}
		}
		return nil
	})
}
