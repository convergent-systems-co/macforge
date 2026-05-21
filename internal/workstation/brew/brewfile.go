// Package brew provides pure-logic primitives for parsing, rewriting,
// and diffing Brewfiles. It contains no os/exec calls: the input is
// always an io.Reader and the output is always an io.Writer.
//
// Sub-issues #15 (update local-to-remote) and #16 (update remote-to-
// local) will compose these primitives with the live `brew bundle dump`
// output to drive the two-way sync. Keeping the parser/writer/diff
// engine free of side effects makes them table-testable and lets the
// exec wrappers focus on subprocess plumbing.
package brew

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// EntryKind discriminates the six recognized row types in a Brewfile.
// Real Brewfiles also carry comments and blank lines; those are dropped
// on parse and never round-tripped (see the package doc).
type EntryKind int

const (
	// KindTap is a `tap "owner/repo"` row, optionally with a URL.
	KindTap EntryKind = iota
	// KindBrew is a `brew "formula"` row, optionally with build args.
	KindBrew
	// KindCask is a `cask "name"` row.
	KindCask
	// KindMas is a `mas "App Name", id: 123456` Mac App Store row.
	KindMas
)

// String returns the lowercase keyword used in the Brewfile source.
// It is the inverse of the parser's leading-keyword match.
func (k EntryKind) String() string {
	switch k {
	case KindTap:
		return "tap"
	case KindBrew:
		return "brew"
	case KindCask:
		return "cask"
	case KindMas:
		return "mas"
	default:
		return fmt.Sprintf("EntryKind(%d)", int(k))
	}
}

// Entry is one parseable line in a Brewfile.
//
// Name is the canonical first-argument string ("owner/repo", "ripgrep",
// "Xcode", etc.) without its surrounding quotes. Extras is the verbatim
// trailing text after the closing quote of Name — including the leading
// comma and space — so callers can preserve `args: [...]`, tap URLs,
// and `id: NNN` on round-trip without re-parsing those payloads.
type Entry struct {
	Kind   EntryKind
	Name   string
	Extras string
}

// Brewfile is an ordered list of Entry values. The order matches the
// order of recognized rows in the source. Comments and blank lines are
// not preserved.
type Brewfile struct {
	Entries []Entry
}

// keywordToKind maps the leading word of a Brewfile row to its kind.
// Kept as a small table so adding a future row type (e.g. `vscode`) is
// a one-line change.
var keywordToKind = map[string]EntryKind{
	"tap":  KindTap,
	"brew": KindBrew,
	"cask": KindCask,
	"mas":  KindMas,
}

// Parse reads a Brewfile from r and returns its Entry list.
//
// Comments (`# ...`) and blank lines are skipped. Each remaining line
// MUST begin with one of the recognized keywords (tap / brew / cask /
// mas) followed by a double-quoted name; anything else is a hard error
// that names the offending line number and content. Single-quoted names
// are rejected — real `brew bundle dump` output uses doubles, and we
// keep the surface small.
func Parse(r io.Reader) (Brewfile, error) {
	var b Brewfile
	scanner := bufio.NewScanner(r)
	// Default Scanner buffer is 64KB per line; Brewfiles never approach
	// that, but raise it to 1MB defensively in case an absurd args list
	// shows up.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineno := 0
	for scanner.Scan() {
		lineno++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		entry, err := parseLine(trimmed, lineno)
		if err != nil {
			return Brewfile{}, err
		}
		b.Entries = append(b.Entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return Brewfile{}, fmt.Errorf("brewfile: scan: %w", err)
	}
	return b, nil
}

// parseLine recognizes a single non-empty, non-comment row.
//
// The shape we accept is:
//
//	<keyword> "<name>"[<extras>]
//
// where <keyword> is one of tap/brew/cask/mas, <name> is a double-quoted
// string with no embedded double quotes, and <extras> is whatever
// follows the closing quote — typically `, args: [...]`, `, id: NNN`,
// or `, "https://..."`. Extras is captured verbatim (no trimming) so a
// Write round-trip is byte-stable.
func parseLine(line string, lineno int) (Entry, error) {
	// Split off the leading keyword. The keyword is followed by at
	// least one space before the opening quote in every shape we accept.
	space := strings.IndexByte(line, ' ')
	if space < 0 {
		return Entry{}, unrecognized(lineno, line)
	}
	keyword := line[:space]
	kind, ok := keywordToKind[keyword]
	if !ok {
		return Entry{}, unrecognized(lineno, line)
	}

	rest := strings.TrimLeft(line[space+1:], " \t")
	if rest == "" || rest[0] != '"' {
		return Entry{}, unrecognized(lineno, line)
	}
	// Find the closing quote of the name. We do not support escaped
	// quotes inside the name: real Brewfile entries never contain them.
	closing := strings.IndexByte(rest[1:], '"')
	if closing < 0 {
		return Entry{}, unrecognized(lineno, line)
	}
	name := rest[1 : 1+closing]
	extras := rest[1+closing+1:]

	return Entry{
		Kind:   kind,
		Name:   name,
		Extras: extras,
	}, nil
}

func unrecognized(lineno int, line string) error {
	return fmt.Errorf("brewfile: line %d: unrecognized entry: %q", lineno, line)
}

// Write renders b to w. Each Entry produces one canonical line:
//
//	tap "<name>"<extras>
//	brew "<name>"<extras>
//	cask "<name>"<extras>
//	mas "<name>"<extras>
//
// A trailing newline follows every entry. When the input to Parse was
// already canonical (one entry per line, double-quoted names, no
// leading whitespace, LF line endings), Write produces a byte-identical
// rendering modulo dropped comments and blank lines.
func (b Brewfile) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)
	for _, e := range b.Entries {
		if _, err := fmt.Fprintf(bw, "%s \"%s\"%s\n", e.Kind.String(), e.Name, e.Extras); err != nil {
			return fmt.Errorf("brewfile: write: %w", err)
		}
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("brewfile: flush: %w", err)
	}
	return nil
}
