package cmd

import (
	"fmt"
	"os"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type ShowCmd struct {
	Name   string `arg:"" help:"Service name to show."`
	Format string `short:"f" help:"Format to show." enum:"xml,json,pretty" default:"pretty" placeholder:""`
}

func (s *ShowCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		content, err := os.ReadFile(daemon.SourcePath)
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	})
}
