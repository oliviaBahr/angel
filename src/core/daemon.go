package core

import (
	"os"
	fp "path/filepath"
	"strconv"
	"strings"

	"angel/src/core/constants"

	"howett.net/plist"
)

type PlistDir struct {
	Path     string
	Domain   constants.Domain
	ForUseBy constants.ForWhom
}

type Daemon struct {
	Name       string
	SourcePath string
	Domain     string
	ForUseBy   constants.ForWhom
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

func NewDaemon(filename string, dirDomain constants.Domain, forUseBy constants.ForWhom) Daemon {
	plistData := &Plist{}
	content, _ := os.ReadFile(filename)
	_, _ = plist.Unmarshal(content, plistData)
	domain := domainFromSessionType(plistData.LimitLoadToSessionType)
	if domain == constants.DomainUnknown {
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

func domainFromSessionType(sessionType string) constants.Domain {
	switch sessionType {
	case "Aqua":
		return constants.DomainGui
	case "Background":
		return constants.DomainUser
	case "LoginWindow":
		return constants.DomainUser
	case "System":
		return constants.DomainSystem
	default:
		return constants.DomainUnknown
	}
}

func domainStr(uid int, domain constants.Domain) string {
	switch domain {
	case constants.DomainSystem:
		return "system"
	case constants.DomainUser:
		return "user/" + strconv.Itoa(uid)
	case constants.DomainGui:
		return "gui/" + strconv.Itoa(uid)
	default:
		return "Unknown"
	}
}
