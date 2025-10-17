package cmd

import (
	"fmt"
	"sort"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type ListCmd struct {
	Pattern  string `arg:"" optional:"" help:"Pattern to filter services."`
	Exact    bool   `short:"e" help:"Exact match."`
	Long     bool   `short:"l" help:"Show long format."`
	TestFlag bool   `help:"Test flag without short version."`
}

func (l *ListCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(l.Pattern, false, ctx, func(matches []core.Daemon) error {
		if l.Long {
			var paths []string
			for _, daemon := range matches {
				paths = append(paths, daemon.SourcePath)
			}
			sort.Strings(paths)
			for _, path := range paths {
				fmt.Println(path)
			}
		} else {
			var names []string
			for _, daemon := range matches {
				names = append(names, daemon.Name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Println(name)
			}
		}
		return nil
	})
}
