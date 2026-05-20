package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func doctorCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:   "doctor",
		Usage:  "Sanity-check the macheim environment",
		Action: notImplemented("doctor", 10),
	}
}
