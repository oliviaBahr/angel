package cmd

import (
	"fmt"
	"os/exec"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type StatusCmd struct {
	Name string `arg:"" help:"Name of the services."`
}

func (s *StatusCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	if s.Name != "" {
		// Use shell for piping to grep
		cmd := fmt.Sprintf("launchctl list | grep -i \"%s\"", angel.PatternForGrep(s.Name, false))
		execCmd := exec.Command("sh", "-c", cmd)
		output, err := execCmd.Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		return nil
	}

	// No pattern, just run launchctl list directly
	output, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return err
	}
	fmt.Print(string(output))
	return nil
}
