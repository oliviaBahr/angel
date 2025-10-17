package cmd

import (
	"fmt"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	fmt.Printf("angel version %s\n", "0.1.0")
	return nil
}
