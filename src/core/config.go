package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
)

type Config struct {
	Directories []struct {
		Path   string
		Domain Domain
	}
}

func setupViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(".angelrc")
	v.SetConfigType("yaml")
	v.AddConfigPath(os.Getenv("HOME"))
	v.AddConfigPath(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "angel"))
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "angel"))
	return v
}

func LoadConfig(ctx *kong.Context) *Config {
	v := setupViper()
	config := &Config{}

	if os.Getenv("HOME") == "" {
		return config
	}

	_ = v.ReadInConfig()
	if err := v.Unmarshal(config); err != nil {
		ctx.Errorf("failed to read config file: %s", err)
		fmt.Println(err)
		ctx.Exit(1)
		os.Exit(1)
	}

	// expand ~ in config paths
	for i, dir := range config.Directories {
		config.Directories[i].Path = strings.Replace(dir.Path, "~", userHome(), 1)
	}

	return config
}
