package core

import (
	"os"

	"angel/src/core/config"
	"angel/src/core/constants"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/log"
)

type Angel struct {
	Daemons *DaemonRegistry
	IsRoot  bool
}

func LoadAngel(ctx *kong.Context) *Angel {
	config.LoadConfig(ctx)
	plistDirs := loadPlistDirectories(config.LoadedConfig)
	daemonRegistry := NewDaemonRegistry(plistDirs)

	return &Angel{
		Daemons: daemonRegistry,
		IsRoot:  os.Geteuid() == 0,
	}
}

func loadPlistDirectories(cfg *config.Config) []PlistDir {
	home := userHome()
	dirs := []PlistDir{
		{Path: "/System/Library/LaunchDaemons", Domain: constants.DomainSystem, ForUseBy: constants.ForApple},
		{Path: "/System/Library/LaunchAgents", Domain: constants.DomainUser, ForUseBy: constants.ForApple},
		{Path: "/Library/LaunchDaemons", Domain: constants.DomainSystem, ForUseBy: constants.ForThirdParty},
		{Path: "/Library/LaunchAgents", Domain: constants.DomainUser, ForUseBy: constants.ForThirdParty},
		{Path: home + "/Library/LaunchAgents", Domain: constants.DomainUser, ForUseBy: constants.ForUser},
		{Path: home + "/.config/angel/user", Domain: constants.DomainUser, ForUseBy: constants.ForAngel},
		{Path: home + "/.config/angel/system", Domain: constants.DomainSystem, ForUseBy: constants.ForAngel},
		{Path: home + "/.config/angel/gui", Domain: constants.DomainGui, ForUseBy: constants.ForAngel},
	}

	// Add user-defined directories from config
	for _, cfgDir := range cfg.Directories {
		dirs = append(dirs, PlistDir{
			Path:     cfgDir.Path,
			Domain:   cfgDir.Domain,
			ForUseBy: constants.ForAngel,
		})
	}

	return dirs
}

func userHome() (dir string) {
	dir = os.Getenv("HOME")
	if dir == "/var/root" {
		log.Error("running as root user not supported")
		os.Exit(1)
	}
	return dir
}
