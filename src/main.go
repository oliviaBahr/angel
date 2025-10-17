package main

import (
	"angel/src/cmd"
	"angel/src/core"

	"github.com/alecthomas/kong"
)

const VERSION = "0.1.0"

var cli struct {
	Start     cmd.StartCmd     `cmd:"" help:"Start a service."`
	Stop      cmd.StopCmd      `cmd:"" help:"Stop a service."`
	Restart   cmd.RestartCmd   `cmd:"" help:"Restart a service."`
	Status    cmd.StatusCmd    `cmd:"" help:"Show service status."`
	List      cmd.ListCmd      `cmd:"" aliases:"ls" help:"List services."`
	Install   cmd.InstallCmd   `cmd:"" help:"Install a service."`
	Uninstall cmd.UninstallCmd `cmd:"" help:"Uninstall a service."`
	Show      cmd.ShowCmd      `cmd:"" help:"Show service daemon."`
	Edit      cmd.EditCmd      `cmd:"" help:"Edit service daemon."`
	Version   cmd.VersionCmd   `cmd:"" help:"Show version."`
}

func main() {
	// Initialize daemons at startup
	angel := core.LoadDaemons()

	ctx := kong.Parse(&cli,
		kong.Name("angel"),
		kong.Description("macOS launchd service manager"),
		kong.UsageOnError(),
		kong.Help(core.CustomHelpPrinter),
	)

	err := ctx.Run(angel, ctx)
	ctx.FatalIfErrorf(err)
}
