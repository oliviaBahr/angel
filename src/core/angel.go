package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"howett.net/plist"
)

// Angel manages all daemon information
type Angel struct {
	UserDirs    []string
	SystemDirs  []string
	GuiDirs     []string
	CustomPaths []CustomPath
	AllDirs     []string
	AllPaths    []string
	Uid         string
	IsRoot      bool
	Daemons     map[string]Daemon
}

func LoadDaemons() *Angel {
	// Load configuration from YAML file
	cfg := LoadConfig()

	a := &Angel{
		UserDirs:    []string{filepath.Join(os.Getenv("HOME"), "Library/LaunchAgents"), "/Library/LaunchAgents"},
		SystemDirs:  []string{"/Library/LaunchDaemons", "/System/Library/LaunchDaemons"},
		GuiDirs:     []string{},
		CustomPaths: []CustomPath{},
		AllDirs:     []string{},
		AllPaths:    []string{},
		Uid:         strconv.Itoa(os.Geteuid()),
		IsRoot:      os.Geteuid() == 0,
		Daemons:     make(map[string]Daemon),
	}

	// Load custom paths from config
	a.CustomPaths = cfg.Paths

	// Add all directories to AllDirs
	a.AllDirs = append(a.AllDirs, a.UserDirs...)
	a.AllDirs = append(a.AllDirs, a.SystemDirs...)
	a.AllDirs = append(a.AllDirs, a.GuiDirs...)

	// Process custom paths from config
	for _, customPath := range a.CustomPaths {
		path := customPath.Path
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				a.AllDirs = append(a.AllDirs, path)
			} else {
				a.AllPaths = append(a.AllPaths, path)
			}
		}
	}
	// Process directories (find all .plist files in each directory)
	for _, dir := range a.AllDirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*.plist"))
		if err != nil {
			continue
		}

		// Determine domain based on directory location
		var domain string
		if contains(a.UserDirs, dir) {
			domain = "user" + "/" + a.Uid
		} else if contains(a.SystemDirs, dir) {
			domain = "system"
		} else if contains(a.GuiDirs, dir) {
			domain = "gui" + "/" + a.Uid
		} else {
			// Check if this directory is from a custom path
			foundCustomDomain := false
			for _, customPath := range a.CustomPaths {
				if customPath.Path == dir {
					domain = string(customPath.Domain)
					foundCustomDomain = true
					break
				}
			}
			if !foundCustomDomain {
				// Fallback to default domain mapping
				seshToDomain := map[string]string{
					"Aqua":        "gui" + "/" + a.Uid,
					"Background":  "user" + "/" + a.Uid,
					"LoginWindow": "user" + "/" + a.Uid,
					"System":      "system",
				}
				domain = seshToDomain["Background"] // Default fallback
			}
		}

		for _, filename := range matches {
			// Parse the plist file to extract all properties
			var daemon Daemon
			daemon.Name = strings.TrimSuffix(filepath.Base(filename), ".plist")
			daemon.SourcePath = filename

			content, err := os.ReadFile(filename)
			if err != nil {
				// This is a non-fatal error during daemon loading, continue processing
				continue
			}

			decoder := plist.NewDecoder(strings.NewReader(string(content)))
			decoder.Decode(&daemon)

			// Use directory-based domain, but allow plist to override if it has LimitLoadToSessionType
			if daemon.LimitLoadToSessionType != "" {
				seshToDomain := map[string]string{
					"Aqua":        "gui" + "/" + a.Uid,
					"Background":  "user" + "/" + a.Uid,
					"LoginWindow": "user" + "/" + a.Uid,
					"System":      "system",
				}
				if mappedDomain, exists := seshToDomain[daemon.LimitLoadToSessionType]; exists {
					domain = mappedDomain
				}
			}

			daemon.Domain = domain
			a.Daemons[daemon.Name] = daemon
		}
	}

	// Process individual files
	for _, filePath := range a.AllPaths {
		// Only process .plist files
		if filepath.Ext(filePath) != ".plist" {
			continue
		}

		// Determine domain based on file location
		var domain string
		// Check if this file is from a custom path
		foundCustomDomain := false
		for _, customPath := range a.CustomPaths {
			if customPath.Path == filePath {
				domain = string(customPath.Domain)
				foundCustomDomain = true
				break
			}
		}
		if !foundCustomDomain {
			// For individual files, we need to determine domain based on which directory they came from
			// Since files are processed from AllPaths, we'll use a fallback approach
			seshToDomain := map[string]string{
				"Aqua":        "gui" + "/" + a.Uid,
				"Background":  "user" + "/" + a.Uid,
				"LoginWindow": "user" + "/" + a.Uid,
				"System":      "system",
			}
			domain = seshToDomain["Background"] // Default fallback
		}

		// Parse the plist file to extract all properties
		var daemon Daemon
		daemon.Name = strings.TrimSuffix(filepath.Base(filePath), ".plist")
		daemon.SourcePath = filePath

		content, err := os.ReadFile(filePath)
		if err != nil {
			// This is a non-fatal error during daemon loading, continue processing
			continue
		}

		decoder := plist.NewDecoder(strings.NewReader(string(content)))
		decoder.Decode(&daemon)

		// Use directory-based domain, but allow plist to override if it has LimitLoadToSessionType
		if daemon.LimitLoadToSessionType != "" {
			seshToDomain := map[string]string{
				"Aqua":        "gui" + "/" + a.Uid,
				"Background":  "user" + "/" + a.Uid,
				"LoginWindow": "user" + "/" + a.Uid,
				"System":      "system",
			}
			if mappedDomain, exists := seshToDomain[daemon.LimitLoadToSessionType]; exists {
				domain = mappedDomain
			}
		}

		daemon.Domain = domain
		a.Daemons[daemon.Name] = daemon
	}

	return a
}

// Helper function to check if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// PatternForGrep returns a pattern for grep matching
func (a *Angel) PatternForGrep(s string, exact bool) string {
	if exact {
		return fmt.Sprintf("^%s$", s)
	}
	return s
}

// WithMatch executes a callback function with daemons matching the query
func (a *Angel) WithMatch(query string, exact bool, ctx *kong.Context, fn interface{}) error {
	pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(a.PatternForGrep(query, exact)))
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
		os.Exit(0)
	}

	switch f := fn.(type) {
	case func(Daemon) error:
		if len(matches) > 1 {
			ctx.Errorf("Multiple daemons found matching '%s':", query)
			a.PrintDaemons(matches)
			os.Exit(0)
		}
		return f(matches[0])
	case func([]Daemon) error:
		return f(matches)
	default:
		return fmt.Errorf("callback must be func(Daemon) error or func([]Daemon) error")
	}
}

// PrintDaemons prints the names of the given daemons
func (a *Angel) PrintDaemons(daemons []Daemon) {
	for _, daemon := range daemons {
		fmt.Println(daemon.Name)
	}
}
