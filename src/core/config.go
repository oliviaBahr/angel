package core

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"
)

type directory struct {
	Path   string `yaml:"path"`
	Domain Domain `yaml:"domain"`
}

type Config struct {
	Dirs []directory `yaml:"dirs"`
}

func LoadConfig(ctx *kong.Context) *Config {
	config := &Config{
		Dirs: []directory{},
	}

	configPath := getConfigPath()
	if configPath == nil {
		return config // Return empty config if none found
	}

	parseConfig(*configPath, config, ctx)

	// expand ~ in config paths
	for _, dir := range config.Dirs {
		dir.Path = os.ExpandEnv(dir.Path)
	}

	return config
}

func parseConfig(path string, config *Config, ctx *kong.Context) {
	data, err := os.ReadFile(path)
	if err != nil {
		ctx.Errorf("failed to read config file: %s", err)
		ctx.Exit(1)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		ctx.Errorf("invalid config file: %s", err)
		ctx.Exit(1)
	}
}

func getConfigPath() *string {
	var path string
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".angelrc")); err == nil {
		path = filepath.Join(os.Getenv("HOME"), ".angelrc")
	}
	if _, err := os.Stat(filepath.Join(getXDGConfigHome(), "angel", ".angelrc")); err == nil {
		path = filepath.Join(getXDGConfigHome(), "angel", ".angelrc")
	}
	return &path
}

func getXDGConfigHome() string {
	confDir := os.Getenv("XDG_CONFIG_HOME")
	if confDir == "" {
		confDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return confDir
}
