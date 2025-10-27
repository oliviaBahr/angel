package cmd

import (
	"fmt"

	"angel/src/core"
	"angel/src/core/config"
	"angel/src/core/launchctl"
	"angel/src/core/styles"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss/table"
)

type StatusCmd struct {
	Name string `arg:"" help:"Name of the services."`
}

func (s *StatusCmd) Run(a *core.Angel, config *config.Config, ctx *kong.Context) error {
	return a.Daemons.WithMatch(s.Name, false, ctx, func(daemon core.Daemon) error {
		// Get raw launchctl output
		deamonInfo, err := launchctl.Status(daemon)
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		t := table.New().StyleFunc(styles.TableStyleFunc).
			Row("Name", daemon.Name).
			Row("State", deamonInfo.Get("state")).
			Row("Domain", daemon.Domain).
			Row("Active Count", deamonInfo.Get("active count")).
			Row("Source Path", deamonInfo.Get("path")).
			Row("Type", deamonInfo.Get("type"))

		fmt.Println(t.String())
		return nil
	})
}
