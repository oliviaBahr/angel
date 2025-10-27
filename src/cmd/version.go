package cmd

import (
	"fmt"

	"angel/src/core"
	"angel/src/core/config"

	"github.com/alecthomas/kong"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(a *core.Angel, config *config.Config, ctx *kong.Context) error {
	fmt.Printf("angel version %s\n", "0.1.0")
	return nil
}
