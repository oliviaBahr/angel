package main

import (
	"fmt"
	"os"

	"angel/src/cmd"
	"angel/src/core"

	"github.com/spf13/cobra"
)

const VERSION = "0.1.0"

func main() {
	// Load Angel instance before any command runs
	angel, err := core.LoadAngel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading angel: %v\n", err)
		os.Exit(1)
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:     "angel",
		Short:   "macOS launchd service manager",
		Version: VERSION,
	}

	// Add commands with angel passed via closure
	rootCmd.AddCommand(cmd.NewStartCmd(angel))
	rootCmd.AddCommand(cmd.NewStopCmd(angel))
	rootCmd.AddCommand(cmd.NewRestartCmd(angel))
	rootCmd.AddCommand(cmd.NewStatusCmd(angel))
	rootCmd.AddCommand(cmd.NewListCmd(angel))
	rootCmd.AddCommand(cmd.NewShowCmd(angel))
	rootCmd.AddCommand(cmd.NewVersionCmd())

	for _, cmd := range rootCmd.Commands() {
		cmd.SetUsageFunc(core.UsageStringFunc(cmd))
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
