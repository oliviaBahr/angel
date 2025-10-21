package cmd

import (
	"fmt"

	"angel/src/core"
	"angel/src/core/launchctl"

	"github.com/alecthomas/kong"
)

type RestartCmd struct {
	Name string `arg:"" help:"Service name to restart."`
}

func (r *RestartCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(r.Name, false, ctx, func(daemon core.Daemon) error {
		output, err := launchctl.KickstartKill(daemon)
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("restarted %s\n", daemon.Name)
		return nil
	})
}
