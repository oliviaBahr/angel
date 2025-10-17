package cmd

import (
	"fmt"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type RestartCmd struct {
	Name string `arg:"" help:"Service name to restart."`
}

func (r *RestartCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(r.Name, false, ctx, func(daemon core.Daemon) error {
		output, err := exec.Command("launchctl", "kickstart", "-k", daemon.Domain+"/"+daemon.Name).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("restarted %s\n", daemon.Name)
		return nil
	})
}
