package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func statusCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:   "status",
		Usage:  "Read-only summary of drift between this Mac and the repo",
		Action: notImplemented("status", 11),
	}
}
