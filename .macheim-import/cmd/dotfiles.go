package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/dotfiles"
	"github.com/urfave/cli/v3"
)

func dotfilesCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "dotfiles",
		Usage: "Dotfiles operations",
		Commands: []*cli.Command{
			{
				Name:  "apply",
				Usage: "Copy repo dotfiles into $HOME",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					repoPath, _, err := rt.ResolveRepoPath()
					if err != nil {
						return err
					}
					if repoPath == "" {
						return errors.New("dotfiles apply: no repo configured; set --repo, MACHEIM_REPO, or ~/.config/macheim/config.yaml")
					}
					homePath, err := os.UserHomeDir()
					if err != nil {
						return err
					}
					result, err := dotfiles.Apply(rt, repoPath, homePath)
					if err != nil {
						return err
					}
					_, _ = fmt.Fprintf(cmd.Root().Writer, "copied: %d, skipped: %d, backed up: %d", len(result.Copied), len(result.Skipped), len(result.BackedUp))
					if result.BackupDir != "" {
						_, _ = fmt.Fprintf(cmd.Root().Writer, " (backups at %s)", result.BackupDir)
					}
					_, _ = fmt.Fprintln(cmd.Root().Writer)
					return nil
				},
			},
		},
	}
}
