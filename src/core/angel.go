package core

import (
	"fmt"
	"os"
	fp "path/filepath"
	"reflect"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/log"
)

type Angel struct {
	Daemons map[string]Daemon
	IsRoot  bool
}

func LoadAngel(ctx *kong.Context) *Angel {
	cfg := LoadConfig(ctx)
	daemons := loadDaemons(cfg)

	return &Angel{
		Daemons: daemons,
		IsRoot:  os.Geteuid() == 0,
	}
}

func loadDaemons(cfg *Config) map[string]Daemon {
	daemons := make(map[string]Daemon)
	plistDirs := loadPlistDirectories(cfg)

	for _, plistDir := range plistDirs {
		matches, err := fp.Glob(plistDir.Path + "/*.plist")
		if err != nil || matches == nil {
			continue
		}

		for _, filename := range matches {
			daemon := NewDaemon(filename, plistDir.Domain, plistDir.ForUseBy)
			daemons[daemon.Name] = daemon
		}
	}

	return daemons
}

func loadPlistDirectories(cfg *Config) []PlistDir {
	home := userHome()
	dirs := []PlistDir{
		{Path: "/System/Library/LaunchDaemons", Domain: DomainSystem, ForUseBy: ForApple},
		{Path: "/System/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForApple},
		{Path: "/Library/LaunchDaemons", Domain: DomainSystem, ForUseBy: ForThirdParty},
		{Path: "/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForThirdParty},
		{Path: home + "/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForUser},
		{Path: home + "/.config/angel/user", Domain: DomainUser, ForUseBy: ForAngel},
		{Path: home + "/.config/angel/system", Domain: DomainSystem, ForUseBy: ForAngel},
		{Path: home + "/.config/angel/gui", Domain: DomainGui, ForUseBy: ForAngel},
	}

	// Add user-defined directories from config
	for _, cfgDir := range cfg.Directories {
		dirs = append(dirs, PlistDir{
			Path:     cfgDir.Path,
			Domain:   cfgDir.Domain,
			ForUseBy: ForAngel,
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

// WithMatch executes a callback function with daemons matching the query
func (a *Angel) WithMatch(query string, exact bool, ctx *kong.Context, execFn interface{}) error {
	if exact {
		query = fmt.Sprintf("^%s$", query)
	}
	pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(query))
	var matches []Daemon

	for agent, daemon := range a.Daemons {
		if pattern.MatchString(agent) {
			matches = append(matches, daemon)
		}
	}

	if len(matches) == 0 {
		if query != "" {
			ctx.Errorf("No daemon found matching '%s'", query)
		}
		ctx.Exit(0)
	}

	inputType := reflect.TypeOf(execFn).In(0)
	switch inputType {
	case reflect.TypeOf([]Daemon{}):
		return execFn.(func([]Daemon) error)(matches)
	case reflect.TypeOf(Daemon{}):
		return execFn.(func(Daemon) error)(matches[0])
	default:
		panic("this should never happen. callback must be func(Daemon) error or func([]Daemon)")
	}
}
