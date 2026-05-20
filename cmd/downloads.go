package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func downloadsCommand(rt *config.Runtime) *cli.Command {
	_ = rt
	return &cli.Command{
		Name:   "downloads",
		Usage:  "Fetch optional downloads listed in the repo",
		Action: notImplemented("downloads", 20),
	}
}
