package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func macosCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:  "macos",
		Usage: "macOS system operations",
		Commands: []*cli.Command{
			{
				Name:   "defaults",
				Usage:  "Apply the repo macOS defaults manifest",
				Action: notImplemented("macos defaults", 20),
			},
		},
	}
}
