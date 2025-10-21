package launchctl

import (
	"angel/src/core"
	"os/exec"
)

func Bootstrap(daemon core.Daemon) (output []byte, error error) {
	return launchctl("bootstrap", daemon.Domain, daemon.SourcePath)
}

func Bootout(daemon core.Daemon) (output []byte, error error) {
	return launchctl("bootout", daemon.Domain, daemon.SourcePath)
}

func Enable(daemon core.Daemon) (output []byte, error error) {
	return launchctl("enable", slashJoin(daemon))
}

func Disable(daemon core.Daemon) (output []byte, error error) {
	return launchctl("disable", slashJoin(daemon))
}

func Kickstart(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kickstart", slashJoin(daemon))
}

func KickstartKill(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kickstart", "-k", slashJoin(daemon))
}

func Kill(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kill", slashJoin(daemon))
}

// Helpers

func launchctl(args ...string) (output []byte, error error) {
	return exec.Command("launchctl", args...).Output()
}

func slashJoin(daemon core.Daemon) string {
	return daemon.Domain + "/" + daemon.Name
}
