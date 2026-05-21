package dotfiles

import (
	"errors"

	"github.com/polliard/macheim/internal/config"
)

// UpdateLocalToRemote walks $HOME against <repo>/dotfiles/, classifies
// each file ($HOME-side vs repo-side), prompts the user, and copies
// changed/new $HOME files back into the repo. Backs up overwritten
// repo files to <repo>/.macheim-repo-backups/<ISO-8601>/.
//
// This placeholder lets the cmd-level dispatch compile while the full
// implementation lands in a follow-up commit on this branch.
func UpdateLocalToRemote(_ *config.Runtime) error {
	return errors.New("dotfiles local-to-remote: not yet implemented")
}
