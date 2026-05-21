package cmd

import (
	"context"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/status"
	"github.com/urfave/cli/v3"
)

func statusCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Read-only summary of drift between this Mac and the repo",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return status.Run(rt, cmd.Root().Writer)
		},
	}
}
