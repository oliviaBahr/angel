package cmd

import (
	"fmt"

	"angel/src/core"
	"angel/src/core/launchctl"

	"github.com/alecthomas/kong"
)

type StopCmd struct {
	Name string `arg:"" help:"Service name to stop."`
}

func (s *StopCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		output, err := launchctl.Kill(daemon)
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("stopped %s\n", daemon.Name)
		return nil
	})
}
