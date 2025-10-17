package cmd

import (
	"fmt"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type StopCmd struct {
	Name string `arg:"" help:"Service name to stop."`
}

func (s *StopCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		// Use bootout to remove the service from the domain
		output, err := exec.Command("launchctl", "bootout", daemon.Domain, daemon.SourcePath).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("stopped %s\n", daemon.Name)
		return nil
	})
}
