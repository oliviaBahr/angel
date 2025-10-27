package cmd

import (
	"fmt"
	"os"

	"angel/src/core"
	"angel/src/core/config"

	"github.com/alecthomas/kong"
)

type ShowCmd struct {
	Name   string `arg:"" help:"Service name to show."`
	Format string `short:"f" help:"Format to show." enum:"xml,json,pretty" default:"pretty" placeholder:""`
}

func (s *ShowCmd) Run(a *core.Angel, config *config.Config, ctx *kong.Context) error {
	return a.Daemons.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		content, err := os.ReadFile(daemon.SourcePath)
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	})
}
