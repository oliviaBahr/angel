package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const VERSION = "0.1.0"

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("angel version %s\n", VERSION)
			return nil
		},
	}
}
