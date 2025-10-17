package core

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Domain represents the type of path domain
type Domain string

const (
	DomainUser   Domain = "user"
	DomainSystem Domain = "system"
	DomainGUI    Domain = "gui"
)

// UnmarshalYAML implements custom YAML unmarshaling with validation
func (d *Domain) UnmarshalYAML(value *yaml.Node) error {
	var domainStr string
	if err := value.Decode(&domainStr); err != nil {
		return err
	}

	switch domainStr {
	case "user", "system", "gui":
		*d = Domain(domainStr)
		return nil
	default:
		return fmt.Errorf("invalid domain '%s', must be one of: user, system, gui", domainStr)
	}
}

// CustomPath represents a path with its domain type
type CustomPath struct {
	Domain Domain `yaml:"domain"`
	Path   string `yaml:"path"`
}

// Config represents the angel configuration structure
type Config struct {
	Paths []CustomPath `yaml:"paths"`
}

// LoadConfig loads the configuration from XDG_CONFIG_HOME/angel/config.yaml or .config/angel/config.yaml
func LoadConfig() *Config {
	config := &Config{
		Paths: []CustomPath{},
	}

	// Try XDG_CONFIG_HOME first, then fallback to .config
	var configDir string
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configDir = filepath.Join(xdgConfigHome, "angel")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return config // Return empty config if we can't get home dir
		}
		configDir = filepath.Join(homeDir, ".config", "angel")
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config // Return empty config if file doesn't exist
	}

	// Read and parse the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config // Return empty config if we can't read the file
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config // Return empty config if we can't parse the YAML
	}

	return config
}
