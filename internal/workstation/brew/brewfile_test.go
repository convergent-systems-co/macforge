//go:build !windows

package brew

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  []Entry
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "single brew",
			input: "brew \"ripgrep\"\n",
			want: []Entry{
				{Kind: KindBrew, Name: "ripgrep", Extras: ""},
			},
		},
		{
			name: "four kinds in order",
			input: "tap \"homebrew/cask\"\n" +
				"brew \"ripgrep\"\n" +
				"cask \"firefox\"\n" +
				"mas \"Xcode\", id: 497799835\n",
			want: []Entry{
				{Kind: KindTap, Name: "homebrew/cask", Extras: ""},
				{Kind: KindBrew, Name: "ripgrep", Extras: ""},
				{Kind: KindCask, Name: "firefox", Extras: ""},
				{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
			},
		},
		{
			name:  "brew with args",
			input: "brew \"vim\", args: [\"with-python3\"]\n",
			want: []Entry{
				{Kind: KindBrew, Name: "vim", Extras: ", args: [\"with-python3\"]"},
			},
		},
		{
			name:  "mas with id",
			input: "mas \"Xcode\", id: 497799835\n",
			want: []Entry{
				{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
			},
		},
		{
			name:  "tap with URL",
			input: "tap \"homebrew/cask-fonts\", \"https://github.com/Homebrew/homebrew-cask-fonts.git\"\n",
			want: []Entry{
				{
					Kind:   KindTap,
					Name:   "homebrew/cask-fonts",
					Extras: ", \"https://github.com/Homebrew/homebrew-cask-fonts.git\"",
				},
			},
		},
		{
			name: "mixed with comments and blanks",
			input: "# top-level comment\n" +
				"\n" +
				"tap \"homebrew/cask\"\n" +
				"   \n" +
				"# section: brews\n" +
				"brew \"ripgrep\"\n" +
				"brew \"vim\", args: [\"with-python3\"]\n" +
				"\n" +
				"cask \"firefox\"\n" +
				"mas \"Xcode\", id: 497799835\n",
			want: []Entry{
				{Kind: KindTap, Name: "homebrew/cask", Extras: ""},
				{Kind: KindBrew, Name: "ripgrep", Extras: ""},
				{Kind: KindBrew, Name: "vim", Extras: ", args: [\"with-python3\"]"},
				{Kind: KindCask, Name: "firefox", Extras: ""},
				{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse: unexpected error: %v", err)
			}
			if !equalEntries(got.Entries, tc.want) {
				t.Fatalf("Parse entries mismatch:\n got: %#v\nwant: %#v", got.Entries, tc.want)
			}
		})
	}
}

func TestParseUnrecognized(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{name: "bare word", input: "ripgrep\n"},
		{name: "unknown keyword", input: "formula \"ripgrep\"\n"},
		{name: "missing quotes", input: "brew ripgrep\n"},
		{name: "single quotes", input: "brew 'ripgrep'\n"},
		{name: "unterminated quote", input: "brew \"ripgrep\n"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(strings.NewReader(tc.input))
			if err == nil {
				t.Fatalf("Parse: expected error, got nil for input %q", tc.input)
			}
			if !strings.Contains(err.Error(), "line 1") {
				t.Fatalf("Parse: error should mention line number; got %v", err)
			}
		})
	}
}

func TestParseUnrecognizedLineNumber(t *testing.T) {
	t.Parallel()

	// Three lines of valid content, then a bad one on line 5 (blank +
	// comment do not advance the count past their position).
	input := "tap \"homebrew/cask\"\n" +
		"# a comment\n" +
		"brew \"ripgrep\"\n" +
		"\n" +
		"nonsense here\n"

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatalf("Parse: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "line 5") {
		t.Fatalf("Parse: expected error to reference line 5, got %v", err)
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	// Each input here is already canonical: one entry per line, double-
	// quoted names, no leading whitespace, no comments, no blank lines.
	// Write(Parse(input)) must reproduce input byte-for-byte.
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "single brew",
			input: "brew \"ripgrep\"\n",
		},
		{
			name: "four kinds",
			input: "tap \"homebrew/cask\"\n" +
				"brew \"ripgrep\"\n" +
				"cask \"firefox\"\n" +
				"mas \"Xcode\", id: 497799835\n",
		},
		{
			name:  "brew with args",
			input: "brew \"vim\", args: [\"with-python3\"]\n",
		},
		{
			name:  "tap with URL",
			input: "tap \"homebrew/cask-fonts\", \"https://github.com/Homebrew/homebrew-cask-fonts.git\"\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			var buf bytes.Buffer
			if err := parsed.Write(&buf); err != nil {
				t.Fatalf("Write: %v", err)
			}
			if buf.String() != tc.input {
				t.Fatalf("round trip mismatch:\n got: %q\nwant: %q", buf.String(), tc.input)
			}
		})
	}
}

func TestParseCanonicalFixture(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "canonical.Brewfile"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	bf, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []Entry{
		{Kind: KindTap, Name: "homebrew/cask"},
		{Kind: KindTap, Name: "homebrew/cask-fonts", Extras: ", \"https://github.com/Homebrew/homebrew-cask-fonts.git\""},
		{Kind: KindBrew, Name: "git"},
		{Kind: KindBrew, Name: "ripgrep"},
		{Kind: KindBrew, Name: "vim", Extras: ", args: [\"with-python3\"]"},
		{Kind: KindCask, Name: "firefox"},
		{Kind: KindCask, Name: "iterm2"},
		{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
	}
	if !equalEntries(bf.Entries, want) {
		t.Fatalf("canonical fixture entries mismatch:\n got: %#v\nwant: %#v", bf.Entries, want)
	}

	// Round-trip the canonical fixture: byte-for-byte identical.
	var buf bytes.Buffer
	if err := bf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != string(data) {
		t.Fatalf("canonical round trip mismatch:\n got: %q\nwant: %q", buf.String(), string(data))
	}
}

func TestParseMixedFixture(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "mixed.Brewfile"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	bf, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Mixed fixture has comments and blank lines: the parsed list must
	// match the canonical fixture's entry set, in order.
	want := []Entry{
		{Kind: KindTap, Name: "homebrew/cask"},
		{Kind: KindTap, Name: "homebrew/cask-fonts", Extras: ", \"https://github.com/Homebrew/homebrew-cask-fonts.git\""},
		{Kind: KindBrew, Name: "git"},
		{Kind: KindBrew, Name: "ripgrep"},
		{Kind: KindBrew, Name: "vim", Extras: ", args: [\"with-python3\"]"},
		{Kind: KindCask, Name: "firefox"},
		{Kind: KindCask, Name: "iterm2"},
		{Kind: KindMas, Name: "Xcode", Extras: ", id: 497799835"},
	}
	if !equalEntries(bf.Entries, want) {
		t.Fatalf("mixed fixture entries mismatch:\n got: %#v\nwant: %#v", bf.Entries, want)
	}

	// Mixed -> Write produces the canonical form (comments/blanks
	// dropped), so it must equal the canonical fixture byte-for-byte.
	canonical, err := os.ReadFile(filepath.Join("testdata", "canonical.Brewfile"))
	if err != nil {
		t.Fatalf("read canonical fixture: %v", err)
	}
	var buf bytes.Buffer
	if err := bf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.String() != string(canonical) {
		t.Fatalf("mixed -> canonical write mismatch:\n got: %q\nwant: %q", buf.String(), string(canonical))
	}
}

func TestEntryKindString(t *testing.T) {
	t.Parallel()

	cases := map[EntryKind]string{
		KindTap:  "tap",
		KindBrew: "brew",
		KindCask: "cask",
		KindMas:  "mas",
	}
	for k, want := range cases {
		if got := k.String(); got != want {
			t.Errorf("EntryKind(%d).String() = %q, want %q", int(k), got, want)
		}
	}
}

func equalEntries(a, b []Entry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}