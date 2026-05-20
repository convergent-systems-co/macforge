package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func updateCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:  "update",
		Usage: "Reconcile state between this Mac and the repo",
		Commands: []*cli.Command{
			{
				Name:   "local-to-remote",
				Usage:  "Update the repo to match local state",
				Action: notImplemented("update local-to-remote", 15),
			},
			{
				Name:   "remote-to-local",
				Usage:  "Update this Mac to match the repo",
				Action: notImplemented("update remote-to-local", 16),
			},
		},
	}
}
