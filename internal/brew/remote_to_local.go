package brew

import (
	"errors"

	"github.com/polliard/macheim/internal/config"
)

// UpdateRemoteToLocal optionally pulls the repo (skipped under noPull),
// refuses if the repo is dirty, then runs brew bundle. With prune,
// passes --cleanup to brew bundle to remove formulae no longer listed.
//
// This placeholder lets the cmd-level dispatch compile while the full
// implementation lands in a follow-up commit on this branch.
func UpdateRemoteToLocal(_ *config.Runtime, _, _ bool) error {
	return errors.New("brew remote-to-local: not yet implemented")
}
