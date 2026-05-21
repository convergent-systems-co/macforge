package brew

import (
	"errors"

	"github.com/polliard/macheim/internal/config"
)

// UpdateLocalToRemote dumps the live brew state, diffs against the repo
// Brewfile, prompts for confirmation, and rewrites <repo>/Brewfile on
// confirm. Refuses when no repo is configured (embed-fallback mode).
//
// This placeholder lets the cmd-level dispatch compile while the full
// implementation lands in a follow-up commit on this branch.
func UpdateLocalToRemote(_ *config.Runtime) error {
	return errors.New("brew local-to-remote: not yet implemented")
}
