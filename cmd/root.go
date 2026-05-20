// Package cmd holds the CLI command tree built on urfave/cli/v3.
package cmd

import (
	"context"

	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

// NewRoot returns the root *cli.Command. The caller passes a *config.Runtime;
// the Before hook populates it from parsed flags. Subcommands hold the same
// pointer and read populated values from their Action handlers.
func NewRoot(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:                  "macheim",
		Usage:                 "Bootstrap and sync a Mac to a known-good state defined in a git repo.",
		Version:               rt.VersionString(),
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "Path to the macheim repo (overrides discovery)",
				Sources: cli.EnvVars("MACHEIM_REPO"),
			},
			&cli.BoolFlag{Name: "dry-run", Usage: "Print actions, change nothing"},
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Extra output"},
			&cli.BoolFlag{Name: "quiet", Aliases: []string{"q"}, Usage: "Suppress non-error output"},
			&cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "Skip confirmation prompts"},
			&cli.BoolFlag{Name: "no-color", Usage: "Disable colored output"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			rt.RepoPath = cmd.String("repo")
			rt.DryRun = cmd.Bool("dry-run")
			rt.Verbose = cmd.Bool("verbose")
			rt.Quiet = cmd.Bool("quiet")
			rt.Yes = cmd.Bool("yes")
			rt.NoColor = cmd.Bool("no-color")
			if err := rt.Validate(); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		Commands: []*cli.Command{},
	}
}
