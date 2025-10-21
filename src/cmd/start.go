package cmd

import (
	"fmt"

	"angel/src/core"
	"angel/src/core/launchctl"

	"github.com/alecthomas/kong"
)

type StartCmd struct {
	Name string `arg:"" help:"Service name to start."`
}

func (s *StartCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		output, err := launchctl.KickstartKill(daemon)
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("started %s\n", daemon.Name)
		return nil
	})
}
