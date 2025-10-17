package core

// Daemon represents a service daemon with required and optional fields
type Daemon struct {
	// Required
	Name       string
	SourcePath string
	Domain     string

	// Optional - plist fields
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
