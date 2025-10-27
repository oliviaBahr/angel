package core

import (
	"fmt"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"

	"angel/src/cmd/launchctl"
	"angel/src/types"

	"howett.net/plist"
)

type plistDir struct {
	Path     string
	Domain   types.Domain
	ForUseBy types.ForWhom
}

var plistDirs = []plistDir{
	{Path: "/System/Library/LaunchDaemons", Domain: types.DomainSystem, ForUseBy: types.ForApple},
	{Path: "/System/Library/LaunchAgents", Domain: types.DomainUser, ForUseBy: types.ForApple},
	{Path: "/System/Library/LaunchAngels", Domain: types.DomainGui, ForUseBy: types.ForApple},
	{Path: "/Library/LaunchDaemons", Domain: types.DomainSystem, ForUseBy: types.ForThirdParty},
	{Path: "/Library/LaunchAgents", Domain: types.DomainUser, ForUseBy: types.ForThirdParty},
	{Path: userHome() + "/Library/LaunchAgents", Domain: types.DomainUser, ForUseBy: types.ForUser},
	{Path: userHome() + "/.config/angel/user", Domain: types.DomainUser, ForUseBy: types.ForAngel},
	{Path: userHome() + "/.config/angel/system", Domain: types.DomainSystem, ForUseBy: types.ForAngel},
	{Path: userHome() + "/.config/angel/gui", Domain: types.DomainGui, ForUseBy: types.ForAngel},
}

func NewDaemon(filenameOrDaemonName string, dirDomain types.Domain, forUseBy types.ForWhom, pid *int, lastExitCode *int) types.Daemon {
	plistData := &types.Plist{}
	name := ""
	sourcePath := "unknown"
	domain := dirDomain
	if strings.Contains(filenameOrDaemonName, "/") {
		filename := filenameOrDaemonName
		content, _ := os.ReadFile(filename)
		_, _ = plist.Unmarshal(content, plistData)
		domain = domainFromSessionType(plistData.LimitLoadToSessionType)
		if domain == types.DomainUnknown {
			domain = dirDomain
		}
		// Use Label from plist if it exists, otherwise fall back to filename
		if plistData.Label != "" {
			name = plistData.Label
		} else {
			name = strings.TrimSuffix(fp.Base(filename), ".plist")
		}
		sourcePath = filename
	} else {
		name = filenameOrDaemonName
	}

	return types.Daemon{
		Name:         name,
		SourcePath:   sourcePath,
		Plist:        plistData,
		Domain:       domain,
		DomainStr:    domainStr(os.Geteuid(), domain),
		ForUseBy:     forUseBy,
		PID:          pid,
		LastExitCode: lastExitCode,
	}
}

func domainFromSessionType(sessionType string) types.Domain {
	switch sessionType {
	case "Aqua":
		return types.DomainGui
	case "Background":
		return types.DomainUser
	case "LoginWindow":
		return types.DomainUser
	case "System":
		return types.DomainSystem
	default:
		return types.DomainUnknown
	}
}

func domainStr(uid int, domain types.Domain) string {
	switch domain {
	case types.DomainSystem:
		return "system"
	case types.DomainUser:
		return "user/" + strconv.Itoa(uid)
	case types.DomainGui:
		return "gui/" + strconv.Itoa(uid)
	default:
		return "Unknown"
	}
}

type DaemonRegistry struct {
	Map map[string]types.Daemon
}

func NewDaemonRegistry() *DaemonRegistry {
	// Add user-defined plist directories
	for _, cfgDir := range LoadedConfig.Directories {
		plistDirs = append(plistDirs, plistDir{
			Path:     cfgDir.Path,
			Domain:   cfgDir.Domain,
			ForUseBy: types.ForUser,
		})
	}

	daemonRegistry := &DaemonRegistry{
		Map: make(map[string]types.Daemon),
	}
	for _, plistDir := range plistDirs {
		matches, err := fp.Glob(plistDir.Path + "/*.plist")
		if err != nil || matches == nil {
			continue
		}
		for _, filename := range matches {
			daemon := NewDaemon(filename, plistDir.Domain, plistDir.ForUseBy, nil, nil)
			daemonRegistry.Map[daemon.Name] = daemon
		}
	}
	// add running daemons from anywhere else
	listOutput, err := launchctl.List()
	if err != nil {
		return daemonRegistry
	}
	for _, line := range strings.Split(string(listOutput), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		pid, _ := strconv.Atoi(parts[0])
		lastExitCode, _ := strconv.Atoi(parts[1])
		name := parts[2]
		daemon, exists := daemonRegistry.Map[name]
		if !exists {
			daemon = NewDaemon(name, types.DomainUnknown, types.ForThirdParty, &pid, &lastExitCode)
			daemonRegistry.Map[name] = daemon
		} else {
			// add pid and last exit code to existing daemon
			daemon.PID = &pid
			daemon.LastExitCode = &lastExitCode
			daemonRegistry.Map[name] = daemon
		}
	}
	return daemonRegistry
}

func (r *DaemonRegistry) WithMatch(query string, exact bool, execFn func(types.Daemon) error) error {
	matches, err := r.findMatches(query, exact)
	if err != nil {
		return err
	}
	return execFn(matches[0])
}

func (r *DaemonRegistry) WithMatches(query string, exact bool, execFn func(types.Daemon) error) error {
	matches, err := r.findMatches(query, exact)
	if err != nil {
		return err
	}
	for _, daemon := range matches {
		if err := execFn(daemon); err != nil {
			return err
		}
	}
	return nil
}

func (r *DaemonRegistry) GetMatches(query string, exact bool) ([]types.Daemon, error) {
	return r.findMatches(query, exact)
}

func (r *DaemonRegistry) findMatches(query string, exact bool) ([]types.Daemon, error) {
	if exact {
		query = fmt.Sprintf("^%s$", query)
	}
	pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(query))
	var matches []types.Daemon

	for _, daemon := range r.Map {
		if pattern.MatchString(daemon.Name) {
			matches = append(matches, daemon)
		}
	}

	if len(matches) == 0 {
		if query != "" {
			return nil, fmt.Errorf("no daemon found matching '%s'", query)
		}
		return []types.Daemon{}, nil
	}
	return matches, nil
}
