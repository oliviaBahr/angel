package cmd

import (
	"fmt"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type StartCmd struct {
	Name string `arg:"" help:"Service name to start."`
}

func (s *StartCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		// bootstrap could fail if the service is already loaded. keep going
		exec.Command("launchctl", "bootstrap", daemon.Domain, daemon.SourcePath).Output()

		// kickstart
		output, err := exec.Command("launchctl", "kickstart", daemon.Domain+"/"+daemon.Name).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("started %s\n", daemon.Name)
		return nil
	})
}
