package cmd

import (
	"context"

	"github.com/polliard/macheim/internal/brew"
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func brewCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "brew",
		Usage: "Homebrew operations",
		Commands: []*cli.Command{
			{
				Name:  "install",
				Usage: "Install Homebrew itself",
				Action: func(_ context.Context, _ *cli.Command) error {
					return brew.Install(rt)
				},
			},
			{
				Name:   "bundle",
				Usage:  "Apply the repo Brewfile",
				Action: notImplemented("brew bundle", 13),
			},
		},
	}
}
