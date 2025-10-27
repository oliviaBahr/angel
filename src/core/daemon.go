package core

import (
	"fmt"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"

	"howett.net/plist"
)

type plistDir struct {
	Path     string
	Domain   Domain
	ForUseBy ForWhom
}

var plistDirs = []plistDir{
	{Path: "/System/Library/LaunchDaemons", Domain: DomainSystem, ForUseBy: ForApple},
	{Path: "/System/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForApple},
	{Path: "/Library/LaunchDaemons", Domain: DomainSystem, ForUseBy: ForThirdParty},
	{Path: "/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForThirdParty},
	{Path: userHome() + "/Library/LaunchAgents", Domain: DomainUser, ForUseBy: ForUser},
	{Path: userHome() + "/.config/angel/user", Domain: DomainUser, ForUseBy: ForAngel},
	{Path: userHome() + "/.config/angel/system", Domain: DomainSystem, ForUseBy: ForAngel},
	{Path: userHome() + "/.config/angel/gui", Domain: DomainGui, ForUseBy: ForAngel},
}

type Daemon struct {
	Name       string
	SourcePath string
	Domain     string
	ForUseBy   ForWhom
	Plist      *Plist
}

type Plist struct {
	Program                string            `plist:"Program,omitempty"`                // Path to executable
	ProgramArguments       []string          `plist:"ProgramArguments,omitempty"`       // Command line arguments
	RunAtLoad              bool              `plist:"RunAtLoad,omitempty"`              // Start when loaded
	KeepAlive              bool              `plist:"KeepAlive,omitempty"`              // Restart if exits
	WorkingDirectory       string            `plist:"WorkingDirectory,omitempty"`       // Working directory
	StandardOutPath        string            `plist:"StandardOutPath,omitempty"`        // Stdout log file
	StandardErrorPath      string            `plist:"StandardErrorPath,omitempty"`      // Stderr log file
	EnvironmentVariables   map[string]string `plist:"EnvironmentVariables,omitempty"`   // Environment vars
	StartInterval          int               `plist:"StartInterval,omitempty"`          // Restart interval (seconds)
	StartOnMount           bool              `plist:"StartOnMount,omitempty"`           // Start when filesystem mounts
	ThrottleInterval       int               `plist:"ThrottleInterval,omitempty"`       // Throttle restart attempts
	ProcessType            string            `plist:"ProcessType,omitempty"`            // Process type (Background, Standard, etc.)
	SessionCreate          bool              `plist:"SessionCreate,omitempty"`          // Create session
	LaunchOnlyOnce         bool              `plist:"LaunchOnlyOnce,omitempty"`         // Run only once
	LimitLoadToSessionType string            `plist:"LimitLoadToSessionType,omitempty"` // Session type limit
}

func NewDaemon(filename string, dirDomain Domain, forUseBy ForWhom) Daemon {
	plistData := &Plist{}
	content, _ := os.ReadFile(filename)
	_, _ = plist.Unmarshal(content, plistData)
	domain := domainFromSessionType(plistData.LimitLoadToSessionType)
	if domain == DomainUnknown {
		domain = dirDomain
	}

	return Daemon{
		Name:       strings.TrimSuffix(fp.Base(filename), ".plist"),
		SourcePath: filename,
		Plist:      plistData,
		Domain:     domainStr(os.Geteuid(), domain),
		ForUseBy:   forUseBy,
	}
}

func domainFromSessionType(sessionType string) Domain {
	switch sessionType {
	case "Aqua":
		return DomainGui
	case "Background":
		return DomainUser
	case "LoginWindow":
		return DomainUser
	case "System":
		return DomainSystem
	default:
		return DomainUnknown
	}
}

func domainStr(uid int, domain Domain) string {
	switch domain {
	case DomainSystem:
		return "system"
	case DomainUser:
		return "user/" + strconv.Itoa(uid)
	case DomainGui:
		return "gui/" + strconv.Itoa(uid)
	default:
		return "Unknown"
	}
}

type DaemonRegistry struct {
	Map map[string]Daemon
}

func NewDaemonRegistry() *DaemonRegistry {
	// Add user-defined plist directories
	for _, cfgDir := range LoadedConfig.Directories {
		plistDirs = append(plistDirs, plistDir{
			Path:     cfgDir.Path,
			Domain:   cfgDir.Domain,
			ForUseBy: ForUser,
		})
	}

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

func (r *DaemonRegistry) WithMatch(query string, exact bool, execFn func(Daemon) error) error {
	matches, err := r.findMatches(query, exact)
	if err != nil {
		return err
	}
	return execFn(matches[0])
}

func (r *DaemonRegistry) WithMatches(query string, exact bool, execFn func(Daemon) error) error {
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

func (r *DaemonRegistry) findMatches(query string, exact bool) ([]Daemon, error) {
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
			return nil, fmt.Errorf("no daemon found matching '%s'", query)
		}
		return []Daemon{}, nil
	}
	return matches, nil
}
