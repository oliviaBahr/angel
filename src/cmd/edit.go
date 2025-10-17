package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type EditCmd struct {
	Name string `arg:"" help:"Service name to edit."`
}

func (e *EditCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	return angel.WithMatch(e.Name, false, ctx, func(daemon core.Daemon) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return fmt.Errorf("EDITOR environment variable is not set")
		}

		cmd := exec.Command(editor, daemon.SourcePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	})
}
