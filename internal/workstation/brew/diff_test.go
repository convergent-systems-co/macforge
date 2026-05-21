//go:build !windows

package brew

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	brew := func(name string, extras ...string) Entry {
		e := Entry{Kind: KindBrew, Name: name}
		if len(extras) > 0 {
			e.Extras = extras[0]
		}
		return e
	}
	cask := func(name string) Entry { return Entry{Kind: KindCask, Name: name} }

	cases := []struct {
		name        string
		local       []Entry
		remote      []Entry
		wantAdded   []Entry
		wantRemoved []Entry
		wantUnchang []Entry
	}{
		{
			name:        "empty + empty",
			local:       nil,
			remote:      nil,
			wantAdded:   nil,
			wantRemoved: nil,
			wantUnchang: nil,
		},
		{
			name:        "identical single entry",
			local:       []Entry{brew("foo")},
			remote:      []Entry{brew("foo")},
			wantAdded:   nil,
			wantRemoved: nil,
			wantUnchang: []Entry{brew("foo")},
		},
		{
			name:        "local has an extra",
			local:       []Entry{brew("foo"), brew("bar")},
			remote:      []Entry{brew("foo")},
			wantAdded:   []Entry{brew("bar")},
			wantRemoved: nil,
			wantUnchang: []Entry{brew("foo")},
		},
		{
			name:        "remote has an extra",
			local:       []Entry{brew("foo")},
			remote:      []Entry{brew("foo"), brew("bar")},
			wantAdded:   nil,
			wantRemoved: []Entry{brew("bar")},
			wantUnchang: []Entry{brew("foo")},
		},
		{
			name:        "same name different kind",
			local:       []Entry{brew("foo")},
			remote:      []Entry{cask("foo")},
			wantAdded:   []Entry{brew("foo")},
			wantRemoved: []Entry{cask("foo")},
			wantUnchang: nil,
		},
		{
			name:        "extras differ on same entry",
			local:       []Entry{brew("vim", ", args: [\"with-python3\"]")},
			remote:      []Entry{brew("vim")},
			wantAdded:   nil,
			wantRemoved: nil,
			// Unchanged surfaces the local copy (Extras preserved).
			wantUnchang: []Entry{brew("vim", ", args: [\"with-python3\"]")},
		},
		{
			name: "added preserves local order; removed preserves remote order",
			local: []Entry{
				brew("foo"),
				brew("alpha"),
				brew("zeta"),
			},
			remote: []Entry{
				brew("foo"),
				brew("omega"),
				brew("beta"),
			},
			// Added in local-order: alpha, zeta.
			wantAdded: []Entry{brew("alpha"), brew("zeta")},
			// Removed in remote-order: omega, beta.
			wantRemoved: []Entry{brew("omega"), brew("beta")},
			wantUnchang: []Entry{brew("foo")},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Diff(Brewfile{Entries: tc.local}, Brewfile{Entries: tc.remote})
			if !equalEntries(got.Added, tc.wantAdded) {
				t.Errorf("Added mismatch:\n got: %#v\nwant: %#v", got.Added, tc.wantAdded)
			}
			if !equalEntries(got.Removed, tc.wantRemoved) {
				t.Errorf("Removed mismatch:\n got: %#v\nwant: %#v", got.Removed, tc.wantRemoved)
			}
			if !equalEntries(got.Unchanged, tc.wantUnchang) {
				t.Errorf("Unchanged mismatch:\n got: %#v\nwant: %#v", got.Unchanged, tc.wantUnchang)
			}
		})
	}
}

func TestDiffFixture(t *testing.T) {
	t.Parallel()

	localData, err := os.ReadFile(filepath.Join("testdata", "local.Brewfile"))
	if err != nil {
		t.Fatalf("read local fixture: %v", err)
	}
	remoteData, err := os.ReadFile(filepath.Join("testdata", "remote.Brewfile"))
	if err != nil {
		t.Fatalf("read remote fixture: %v", err)
	}

	local, err := Parse(bytes.NewReader(localData))
	if err != nil {
		t.Fatalf("Parse local: %v", err)
	}
	remote, err := Parse(bytes.NewReader(remoteData))
	if err != nil {
		t.Fatalf("Parse remote: %v", err)
	}

	got := Diff(local, remote)

	// local adds: cask "slack"
	// remote loses: tap homebrew/cask-fonts (with URL), cask iterm2
	// shared (extras from local): tap homebrew/cask, brew git, brew
	//   ripgrep, brew vim (with args: from local), cask firefox, mas
	//   Xcode
	wantAdded := []Entry{
		{Kind: KindCask, Name: "slack"},
	}
	wantRemoved := []Entry{
		{Kind: KindTap, Name: "homebrew/cask-fonts", Extras: ", \"https://github.com/Homebrew/homebrew-cask-fonts.git\""},
		{Kind: KindCask, Name: "iterm2"},
	}
	wantUnchanged := []Entry{
		{Kind: KindTap, Name: "homebrew/cask"},
		{Kind: KindBrew, Name: "git"},
		{Kind: KindBrew, Name: "ripgrep"},
		{Kind: KindBrew, Name: "vim", Extras: ", args: [\"with-python3\"]"},
		{Kind: KindCask, Name: "firefox"},
		{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
	}

	if !equalEntries(got.Added, wantAdded) {
		t.Errorf("fixture Added mismatch:\n got: %#v\nwant: %#v", got.Added, wantAdded)
	}
	if !equalEntries(got.Removed, wantRemoved) {
		t.Errorf("fixture Removed mismatch:\n got: %#v\nwant: %#v", got.Removed, wantRemoved)
	}
	if !equalEntries(got.Unchanged, wantUnchanged) {
		t.Errorf("fixture Unchanged mismatch:\n got: %#v\nwant: %#v", got.Unchanged, wantUnchanged)
	}
}