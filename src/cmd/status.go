package cmd

import (
	"fmt"

	"angel/src/core"

	"github.com/alecthomas/kong"
)

type StatusCmd struct {
	Name string `arg:"" help:"Name of the services."`
}

func (s *StatusCmd) Run(angel *core.Angel, ctx *kong.Context) error {
	fmt.Println("Not implemented")
	return nil
}
