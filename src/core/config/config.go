package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"angel/src/core/constants"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var LoadedConfig *Config

// ColorHookFunc returns a decode hook function for converting strings to Color
func ColorHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if t != reflect.TypeOf(lipgloss.Color("")) {
			return data, nil
		}

		// Handle different input types
		switch v := data.(type) {
		case string:
			return lipgloss.Color(v), nil
		case lipgloss.Color:
			// Already a lipgloss.Color, return as-is
			return v, nil
		default:
			// For other types (like termenv.RGBColor), convert to string first
			return lipgloss.Color(fmt.Sprintf("%v", v)), nil
		}
	}
}

type Config struct {
	Directories []struct {
		Path   string
		Domain constants.Domain
	}

	Colors struct {
		Background    lipgloss.Color
		Foreground    lipgloss.Color
		Title         lipgloss.Color
		SectionHeader lipgloss.Color
		Command       lipgloss.Color
		Argument      lipgloss.Color
		Flag          lipgloss.Color
		TableBorder   lipgloss.Color
	}
}

func setupViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(".angelrc")
	v.SetConfigType("yaml")
	v.AddConfigPath(os.Getenv("HOME"))
	v.AddConfigPath(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "angel"))
	v.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "angel"))
	defaultRenderer := lipgloss.DefaultRenderer()
	v.SetDefault("colors.background", defaultRenderer.Output().BackgroundColor())
	v.SetDefault("colors.foreground", defaultRenderer.Output().ForegroundColor())
	v.SetDefault("colors.title", defaultRenderer.Output().ForegroundColor())
	v.SetDefault("colors.sectionHeader", lipgloss.Color("5"))
	v.SetDefault("colors.command", defaultRenderer.Output().ForegroundColor())
	v.SetDefault("colors.argument", lipgloss.Color("9"))
	v.SetDefault("colors.flag", lipgloss.Color("13"))
	v.SetDefault("colors.tableBorder", lipgloss.Color("5"))
	return v
}

func LoadConfig(ctx *kong.Context) {
	v := setupViper()
	config := &Config{}

	home := os.Getenv("HOME")
	if home == "" || home == "/var/root" {
		return
	}

	_ = v.ReadInConfig()
	if err := v.Unmarshal(config, viper.DecodeHook(ColorHookFunc())); err != nil {
		ctx.Errorf("failed to read config file: %s", err)
		fmt.Println(err)
		ctx.Exit(1)
		os.Exit(1)
	}

	// expand ~ in config paths
	for i, dir := range config.Directories {
		config.Directories[i].Path = strings.Replace(dir.Path, "~", home, 1)
	}

	LoadedConfig = config
}
