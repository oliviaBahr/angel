package launchctl

import (
	"angel/src/types"
	"os/exec"
)

func Bootstrap(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("bootstrap", daemon.DomainStr, daemon.SourcePath)
}

func Bootout(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("bootout", daemon.DomainStr, daemon.SourcePath)
}

func Enable(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("enable", serviceTarget(daemon))
}

func Disable(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("disable", serviceTarget(daemon))
}

func Kickstart(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("kickstart", serviceTarget(daemon))
}

func KickstartKill(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("kickstart", "-k", serviceTarget(daemon))
}

func Kill(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("kill", serviceTarget(daemon))
}

func Print(daemon types.Daemon) (output []byte, error error) {
	return launchctlExec("print", serviceTarget(daemon))
}

func List() (output []byte, error error) {
	return launchctlExec("list")
}

// Helpers

func launchctlExec(args ...string) (output []byte, error error) {
	return exec.Command("launchctl", args...).Output()
}

func serviceTarget(daemon types.Daemon) string {
	return daemon.DomainStr + "/" + daemon.Name
}
