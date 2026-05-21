package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func bootstrapCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:   "bootstrap",
		Usage:  "Run everything end-to-end on a fresh Mac",
		Action: notImplemented("bootstrap", 19),
	}
}
