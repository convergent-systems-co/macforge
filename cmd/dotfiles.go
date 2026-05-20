package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func dotfilesCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:  "dotfiles",
		Usage: "Dotfiles operations",
		Commands: []*cli.Command{
			{
				Name:   "apply",
				Usage:  "Copy repo dotfiles into $HOME",
				Action: notImplemented("dotfiles apply", 17),
			},
		},
	}
}
