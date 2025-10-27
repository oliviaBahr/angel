package core

import (
	"angel/src/types"
	"os"
	fp "path/filepath"

	"github.com/charmbracelet/log"
)

func userHome() (dir string) {
	dir = os.Getenv("HOME")
	if dir == "/var/root" {
		log.Error("running as root user not supported. Use sudo to run as root.")
		os.Exit(1)
	}
	return dir
}

func SortDaemonsByDomain(daemons []types.Daemon) map[string][]types.Daemon {
	daemonsByDomain := map[string][]types.Daemon{
		"system":   {},
		"user/501": {},
		"gui/501":  {},
		"unknown":  {},
	}
	for _, daemon := range daemons {
		daemonsByDomain[daemon.DomainStr] = append(daemonsByDomain[daemon.DomainStr], daemon)
	}
	return daemonsByDomain
}

func SortDaemonsByParentDir(daemons []types.Daemon) map[string][]types.Daemon {
	daemonsByParentDir := map[string][]types.Daemon{}
	for _, daemon := range daemons {
		daemonsByParentDir[fp.Dir(daemon.SourcePath)] = append(daemonsByParentDir[fp.Dir(daemon.SourcePath)], daemon)
	}
	return daemonsByParentDir
}
