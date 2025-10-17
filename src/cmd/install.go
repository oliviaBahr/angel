package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type InstallCmd struct {
	File    string `arg:"" help:"Daemon file to install."`
	Symlink bool   `short:"s" help:"Create symlink instead of copying."`
}

func (i *InstallCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	for _, dir := range angel.AllDirs {
		if _, err := os.Stat(dir); err == nil {
			targetPath := filepath.Join(dir, filepath.Base(i.File))

			if i.Symlink {
				if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
					return err
				}
				if err := os.Symlink(i.File, targetPath); err != nil {
					return err
				}
			} else {
				// Copy file
				input, err := os.ReadFile(i.File)
				if err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, input, 0644); err != nil {
					return err
				}
			}

			// Bootstrap the service after installation
			daemon, exists := angel.Daemons[i.File]
			if !exists {
				return fmt.Errorf("daemon not found")
			}
			output, err := exec.Command("launchctl", "bootstrap", daemon.Domain, targetPath).Output()
			if err != nil {
				return err
			}
			fmt.Print(string(output))

			fmt.Printf("%s installed to %s\n", i.File, dir)
			return nil
		}
	}

	return fmt.Errorf("no suitable directory found for installation")
}
