package brew

// DiffResult is the three-way partition produced by Diff.
//
// Added are entries present in the local Brewfile but missing from the
// remote one — i.e. things the live Mac has that the repo does not.
// Removed are entries present in remote but missing from local — things
// the repo records that the live Mac has dropped. Unchanged are entries
// present in both; their Extras come from the local copy so a remote
// rewrite picks up any local change to `args: [...]` or a tap URL.
type DiffResult struct {
	Added     []Entry // in local, not in remote (repo would gain these)
	Removed   []Entry // in remote, not in local (repo would lose these)
	Unchanged []Entry // in both — local copy wins for Extras
}

// entryKey is the identity used by Diff. Two entries are "the same
// entry" when their Kind and Name match; differing Extras (build args,
// tap URLs, mas ids) do not split them into separate entries. The
// local copy's Extras then wins in the Unchanged slice, which is what
// `update local-to-remote` (#15) needs: when the user has changed
// `brew "vim", args: [...]` on the Mac, the repo should adopt that.
type entryKey struct {
	Kind EntryKind
	Name string
}

func (e Entry) key() entryKey {
	return entryKey{Kind: e.Kind, Name: e.Name}
}

// Diff compares the local Brewfile (e.g. the output of `brew bundle
// dump --describe --file=-`) against the remote Brewfile (the file on
// disk in the repo) and reports what would change if we synced
// repo -> local.
//
// Order preservation:
//   - Added items appear in local's order.
//   - Removed items appear in remote's order.
//   - Unchanged items appear in local's order.
//
// Identity is (Kind, Name) only. An Extras difference on the same key
// does not produce an Added/Removed pair; it surfaces in Unchanged via
// the local copy's Extras. Callers that care about Extras drift can
// detect it by zipping Unchanged with a remote lookup.
func Diff(local, remote Brewfile) DiffResult {
	remoteByKey := make(map[entryKey]Entry, len(remote.Entries))
	for _, e := range remote.Entries {
		// First occurrence wins; duplicate keys in a Brewfile would be
		// a user-level defect, and silently keeping the earliest match
		// produces deterministic diffs.
		if _, exists := remoteByKey[e.key()]; !exists {
			remoteByKey[e.key()] = e
		}
	}
	localByKey := make(map[entryKey]struct{}, len(local.Entries))
	for _, e := range local.Entries {
		localByKey[e.key()] = struct{}{}
	}

	result := DiffResult{}
	for _, e := range local.Entries {
		if _, ok := remoteByKey[e.key()]; ok {
			result.Unchanged = append(result.Unchanged, e)
		} else {
			result.Added = append(result.Added, e)
		}
	}
	for _, e := range remote.Entries {
		if _, ok := localByKey[e.key()]; !ok {
			result.Removed = append(result.Removed, e)
		}
	}
	return result
}
