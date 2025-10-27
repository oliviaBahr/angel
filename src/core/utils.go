package core

import (
	"angel/src/types"
	"os"

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

func SortDaemonsByDomain(daemons []types.Daemon) map[types.Domain][]types.Daemon {
	daemonsByDomain := map[types.Domain][]types.Daemon{
		types.DomainSystem:  {},
		types.DomainUser:    {},
		types.DomainGui:     {},
		types.DomainUnknown: {},
	}
	for _, daemon := range daemons {
		daemonsByDomain[daemon.Domain] = append(daemonsByDomain[daemon.Domain], daemon)
	}
	return daemonsByDomain
}
