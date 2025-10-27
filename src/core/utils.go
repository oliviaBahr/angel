package core

import (
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
