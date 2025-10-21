package core

import (
	"os"
	fp "path/filepath"
	"strconv"
	"strings"

	"howett.net/plist"
)

type ForWhom int

const (
	ForUser ForWhom = iota
	ForApple
	ForThirdParty
	ForAngel
)

type PlistDir struct {
	Path     string
	Domain   Domain
	ForUseBy ForWhom
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
		Domain:     domainStr(strconv.Itoa(os.Geteuid()), domain),
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

func domainStr(uid string, domain Domain) string {
	switch domain {
	case DomainSystem:
		return "system"
	case DomainUser:
		return "user/" + uid
	case DomainGui:
		return "gui/" + uid
	default:
		return "Unknown"
	}
}
