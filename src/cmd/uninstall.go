package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type UninstallCmd struct {
	Name string `arg:"" help:"Service name to uninstall."`
}

func (u *UninstallCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(u.Name, false, ctx, func(daemon core.Daemon) error {
		// First bootout the service
		exec.Command("launchctl", "bootout", daemon.Domain, daemon.SourcePath).Output()
		// Bootout failure is not critical, continue

		// Then remove the file
		if _, err := os.Stat(daemon.SourcePath); err == nil {
			if err := os.Remove(daemon.SourcePath); err != nil {
				return err
			}
			fmt.Printf("uninstalled %s\n", daemon.Name)
		}
		return nil
	})
}
