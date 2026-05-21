package cmd

import (
	"context"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/doctor"
	"github.com/urfave/cli/v3"
)

func doctorCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Sanity-check the macheim environment",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return doctor.Run(rt, cmd.Root().Writer)
		},
	}
}
