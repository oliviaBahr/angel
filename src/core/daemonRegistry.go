package core

import (
	"fmt"
	fp "path/filepath"
	"regexp"

	"github.com/alecthomas/kong"
)

type DaemonRegistry struct {
	Map map[string]Daemon
}

func NewDaemonRegistry(plistDirs []PlistDir) *DaemonRegistry {
	daemonRegistry := &DaemonRegistry{
		Map: make(map[string]Daemon),
	}
	for _, plistDir := range plistDirs {
		matches, err := fp.Glob(plistDir.Path + "/*.plist")
		if err != nil || matches == nil {
			continue
		}
		for _, filename := range matches {
			daemon := NewDaemon(filename, plistDir.Domain, plistDir.ForUseBy)
			daemonRegistry.Map[daemon.Name] = daemon
		}
	}
	return daemonRegistry
}

func (r *DaemonRegistry) WithMatch(query string, exact bool, ctx *kong.Context, execFn func(Daemon) error) error {
	matches := r.findMatches(query, exact, ctx)
	return execFn(matches[0])
}

func (r *DaemonRegistry) WithMatches(query string, exact bool, ctx *kong.Context, execFn func(Daemon) error) error {
	matches := r.findMatches(query, exact, ctx)
	for _, daemon := range matches {
		if err := execFn(daemon); err != nil {
			return err
		}
	}
	return nil
}

func (r *DaemonRegistry) findMatches(query string, exact bool, ctx *kong.Context) []Daemon {
	if exact {
		query = fmt.Sprintf("^%s$", query)
	}
	pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(query))
	var matches []Daemon

	for _, daemon := range r.Map {
		if pattern.MatchString(daemon.Name) {
			matches = append(matches, daemon)
		}
	}

	if len(matches) == 0 {
		if query != "" {
			ctx.Errorf("No daemon found matching '%s'", query)
		}
		ctx.Exit(0)
	}
	return matches
}
