package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func zshCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:  "zsh",
		Usage: "Zsh shell setup operations",
		Commands: []*cli.Command{
			{
				Name:   "setup",
				Usage:  "Set up zsh as the macheim-managed shell",
				Action: notImplemented("zsh setup", 20),
			},
		},
	}
}
