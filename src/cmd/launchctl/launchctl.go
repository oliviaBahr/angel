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
	return launchctl("enable", serviceTarget(daemon))
}

func Disable(daemon core.Daemon) (output []byte, error error) {
	return launchctl("disable", serviceTarget(daemon))
}

func Kickstart(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kickstart", serviceTarget(daemon))
}

func KickstartKill(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kickstart", "-k", serviceTarget(daemon))
}

func Kill(daemon core.Daemon) (output []byte, error error) {
	return launchctl("kill", serviceTarget(daemon))
}

func Print(daemon core.Daemon) (output []byte, error error) {
	return launchctl("print", serviceTarget(daemon))
}

// Helpers

func launchctl(args ...string) (output []byte, error error) {
	return exec.Command("launchctl", args...).Output()
}

func serviceTarget(daemon core.Daemon) string {
	return daemon.Domain + "/" + daemon.Name
}
