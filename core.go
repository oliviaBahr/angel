package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"howett.net/plist"
)

// Daemon represents a service daemon with required and optional fields
type Daemon struct {
	// Required
	Name   string `json:"name"`
	Path   string `json:"path"`
	Domain string `json:"domain"`

	// Optional - plist fields
	Program                string            `plist:"Program,omitempty" json:"program,omitempty"`                               // Path to executable
	ProgramArguments       []string          `plist:"ProgramArguments,omitempty" json:"programarguments,omitempty"`             // Command line arguments
	RunAtLoad              bool              `plist:"RunAtLoad,omitempty" json:"runatload,omitempty"`                           // Start when loaded
	KeepAlive              bool              `plist:"KeepAlive,omitempty" json:"keepalive,omitempty"`                           // Restart if exits
	WorkingDirectory       string            `plist:"WorkingDirectory,omitempty" json:"workingdirectory,omitempty"`             // Working directory
	StandardOutPath        string            `plist:"StandardOutPath,omitempty" json:"standardoutpath,omitempty"`               // Stdout log file
	StandardErrorPath      string            `plist:"StandardErrorPath,omitempty" json:"standarderrorpath,omitempty"`           // Stderr log file
	EnvironmentVariables   map[string]string `plist:"EnvironmentVariables,omitempty" json:"environmentvariables,omitempty"`     // Environment vars
	StartInterval          int               `plist:"StartInterval,omitempty" json:"startinterval,omitempty"`                   // Restart interval (seconds)
	StartOnMount           bool              `plist:"StartOnMount,omitempty" json:"startonmount,omitempty"`                     // Start when filesystem mounts
	ThrottleInterval       int               `plist:"ThrottleInterval,omitempty" json:"throttleinterval,omitempty"`             // Throttle restart attempts
	ProcessType            string            `plist:"ProcessType,omitempty" json:"processtype,omitempty"`                       // Process type (Background, Standard, etc.)
	SessionCreate          bool              `plist:"SessionCreate,omitempty" json:"sessioncreate,omitempty"`                   // Create session
	LaunchOnlyOnce         bool              `plist:"LaunchOnlyOnce,omitempty" json:"launchonlyonce,omitempty"`                 // Run only once
	LimitLoadToSessionType string            `plist:"LimitLoadToSessionType,omitempty" json:"limitloadtosessiontype,omitempty"` // Session type limit
}

// Angel manages all daemon information
type Angel struct {
	user_dirs   []string
	system_dirs []string
	custom_dirs []string
	all_dirs    []string
	uid         string
	isRoot      bool
	daemons     map[string]Daemon
}

func LoadDaemons() *Angel {
	a := &Angel{
		user_dirs:   []string{filepath.Join(os.Getenv("HOME"), "Library/LaunchAgents"), "/Library/LaunchAgents"},
		system_dirs: []string{"/Library/LaunchDaemons", "/System/Library/LaunchDaemons"},
		custom_dirs: strings.Split(os.Getenv("ANGEL_DIRS"), ":"),
		all_dirs:    []string{},
		uid:         strconv.Itoa(os.Geteuid()),
		isRoot:      os.Geteuid() == 0,
		daemons:     make(map[string]Daemon),
	}
	a.all_dirs = append(a.all_dirs, a.user_dirs...)
	a.all_dirs = append(a.all_dirs, a.system_dirs...)
	a.all_dirs = append(a.all_dirs, a.custom_dirs...)
	for _, dir := range a.all_dirs {
		matches, err := filepath.Glob(filepath.Join(dir, "*.plist"))
		if err != nil {
			continue
		}
		seshToDomain := map[string]string{
			"Aqua":        "gui" + "/" + a.uid,
			"Background":  "user" + "/" + a.uid,
			"LoginWindow": "user" + "/" + a.uid,
			"System":      "system",
		}

		for _, filename := range matches {
			// Parse the plist file to extract all properties
			var daemon Daemon
			daemon.Name = strings.TrimSuffix(filepath.Base(filename), ".plist")
			daemon.Path = filename

			content, err := os.ReadFile(filename)
			if err != nil {
				// This is a non-fatal error during daemon loading, continue processing
				continue
			}

			decoder := plist.NewDecoder(strings.NewReader(string(content)))
			decoder.Decode(&daemon)

			daemon.Domain = seshToDomain[daemon.LimitLoadToSessionType]
			a.daemons[daemon.Name] = daemon
		}
	}
	return a
}

type StartCmd struct {
	Name string `arg:"" help:"Service name to start."`
}

func (s *StartCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, s.Name, false, ctx, func(daemon Daemon) error {
		// bootstrap could fail if the service is already loaded. keep going
		exec.Command("launchctl", "bootstrap", daemon.Domain, daemon.Path).Output()

		// kickstart
		output, err := exec.Command("launchctl", "kickstart", daemon.Domain+"/"+daemon.Name).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("started %s\n", daemon.Name)
		return nil
	})
}

type StopCmd struct {
	Name string `arg:"" help:"Service name to stop."`
}

func (s *StopCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, s.Name, false, ctx, func(daemon Daemon) error {
		// Use bootout to remove the service from the domain
		output, err := exec.Command("launchctl", "bootout", daemon.Domain, daemon.Path).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("stopped %s\n", daemon.Name)
		return nil
	})
}

type RestartCmd struct {
	Name string `arg:"" help:"Service name to restart."`
}

func (r *RestartCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, r.Name, false, ctx, func(daemon Daemon) error {
		output, err := exec.Command("launchctl", "kickstart", "-k", daemon.Domain+"/"+daemon.Name).Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		fmt.Printf("restarted %s\n", daemon.Name)
		return nil
	})
}

type StatusCmd struct {
	Name string `arg:"" help:"Name of the services."`
}

func (s *StatusCmd) Run(a *Angel, ctx *kong.Context) error {
	if s.Name != "" {
		// Use shell for piping to grep
		cmd := fmt.Sprintf("launchctl list | grep -i \"%s\"", patternForGrep(s.Name, false))
		execCmd := exec.Command("sh", "-c", cmd)
		output, err := execCmd.Output()
		if err != nil {
			return err
		}
		fmt.Print(string(output))
		return nil
	}

	// No pattern, just run launchctl list directly
	output, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return err
	}
	fmt.Print(string(output))
	return nil
}

type ListCmd struct {
	Pattern  string `arg:"" optional:"" help:"Pattern to filter services."`
	Exact    bool   `short:"e" help:"Exact match."`
	Long     bool   `short:"l" help:"Show long format."`
	TestFlag bool   `help:"Test flag without short version."`
}

func (l *ListCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, l.Pattern, false, ctx, func(matches []Daemon) error {
		if l.Long {
			var paths []string
			for _, daemon := range matches {
				paths = append(paths, daemon.Path)
			}
			sort.Strings(paths)
			for _, path := range paths {
				fmt.Println(path)
			}
		} else {
			var names []string
			for _, daemon := range matches {
				names = append(names, daemon.Name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Println(name)
			}
		}
		return nil
	})
}

type InstallCmd struct {
	File    string `arg:"" help:"Daemon file to install."`
	Symlink bool   `short:"s" help:"Create symlink instead of copying."`
}

func (i *InstallCmd) Run(a *Angel, ctx *kong.Context) error {
	for _, dir := range a.all_dirs {
		if _, err := os.Stat(dir); err == nil {
			targetPath := filepath.Join(dir, filepath.Base(i.File))

			if i.Symlink {
				if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
					return err
				}
				if err := os.Symlink(i.File, targetPath); err != nil {
					return err
				}
			} else {
				// Copy file
				input, err := os.ReadFile(i.File)
				if err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, input, 0644); err != nil {
					return err
				}
			}

			// Bootstrap the service after installation
			daemon, exists := a.daemons[i.File]
			if !exists {
				return fmt.Errorf("daemon not found")
			}
			output, err := exec.Command("launchctl", "bootstrap", daemon.Domain, targetPath).Output()
			if err != nil {
				return err
			}
			fmt.Print(string(output))

			fmt.Printf("%s installed to %s\n", i.File, dir)
			return nil
		}
	}

	return fmt.Errorf("no suitable directory found for installation")
}

type UninstallCmd struct {
	Name string `arg:"" help:"Service name to uninstall."`
}

func (u *UninstallCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, u.Name, false, ctx, func(daemon Daemon) error {
		// First bootout the service
		exec.Command("launchctl", "bootout", daemon.Domain, daemon.Path).Output()
		// Bootout failure is not critical, continue

		// Then remove the file
		if _, err := os.Stat(daemon.Path); err == nil {
			if err := os.Remove(daemon.Path); err != nil {
				return err
			}
			fmt.Printf("uninstalled %s\n", daemon.Name)
		}
		return nil
	})
}

type ShowCmd struct {
	Name   string `arg:"" help:"Service name to show."`
	Format string `short:"f" help:"Format to show." enum:"xml,json,pretty" default:"pretty" placeholder:""`
}

func (s *ShowCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, s.Name, false, ctx, func(daemon Daemon) error {
		content, err := os.ReadFile(daemon.Path)
		if err != nil {
			return err
		}
		fmt.Print(string(content))
		return nil
	})
}

type EditCmd struct {
	Name string `arg:"" help:"Service name to edit."`
}

func (e *EditCmd) Run(a *Angel, ctx *kong.Context) error {
	return withMatch(a, e.Name, false, ctx, func(daemon Daemon) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			return fmt.Errorf("EDITOR environment variable is not set")
		}

		cmd := exec.Command(editor, daemon.Path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	})
}

type VersionCmd struct{}

func (v *VersionCmd) Run(a *Angel, ctx *kong.Context) error {
	fmt.Printf("angel version %s\n", VERSION)
	return nil
}

// Helper functions
func patternForGrep(s string, exact bool) string {
	if exact {
		return fmt.Sprintf("^%s$", s)
	}
	return s
}

func withMatch(daemons *Angel, query string, exact bool, ctx *kong.Context, fn interface{}) error {
	pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(patternForGrep(query, exact)))
	var matches []Daemon

	for agent, daemon := range daemons.daemons {
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
			printDaemons(matches)
			os.Exit(0)
		}
		return f(matches[0])
	case func([]Daemon) error:
		return f(matches)
	default:
		return fmt.Errorf("callback must be func(Daemon) error or func([]Daemon) error")
	}
}

func printDaemons(daemons []Daemon) {
	for _, daemon := range daemons {
		fmt.Println(daemon.Name)
	}
}
