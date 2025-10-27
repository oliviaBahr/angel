package core

import (
	"os"
)

type Angel struct {
	Daemons *DaemonRegistry
	IsRoot  bool
}

func LoadAngel() (*Angel, error) {
	if err := LoadConfig(); err != nil {
		return nil, err
	}

	daemonRegistry := NewDaemonRegistry()

	return &Angel{
		Daemons: daemonRegistry,
		IsRoot:  os.Geteuid() == 0,
	}, nil
}
