# Design — Status (sub-epic #11)

**Tracks:** [Index #3](https://github.com/polliard/macheim/issues/3) → [Epic #25 Foundation](https://github.com/polliard/macheim/issues/25) → sub-epic [#11](https://github.com/polliard/macheim/issues/11) → leaves [#55](https://github.com/polliard/macheim/issues/55) [#56](https://github.com/polliard/macheim/issues/56) [#57](https://github.com/polliard/macheim/issues/57)
**Status:** Draft — awaiting user review
**Date:** 2026-05-21
**Author:** Thomas Polliard (with Claude Opus 4.7)

---

## 1. Objective

Implement `macheim status` as a read-only summary that tells the user what macheim sees right now — brew presence and version, repo source + last-commit + clean/dirty, and per-module drift placeholders. Always exits 0. Doctor-style rows so status and doctor share a visual language. Extract the row renderer into `internal/output/` as the second consumer requires.

## 2. Rationale

### Alternatives table — output layout

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Aligned columns with section headers | Dense; easy to read | New visual language; multi-line per-section formatting | Rejected |
| **Doctor-style rows (✓/✗/?)** | Shared visual vocabulary with doctor; row renderer reused; concise | One row per fact (commit subject + date forces a verbose-line approach) | **Chosen** |
| Flat `key=value` | Machine-parseable | Not what a human runs `status` to see | Rejected |

### Alternatives table — row renderer placement

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| **Extract `internal/output/` now** | Fulfills doctor spec §11's promise; second consumer is here; one canonical home | One refactor commit added; touches doctor (well-tested) | **Chosen** |
| Inline a tiny row helper in status | Smallest status PR | Two diverging renderers; when next consumer lands we extract anyway | Rejected |
| Both inline; extract in a follow-up PR | Single-purpose PRs | Most churn (3 PRs for 1 outcome) | Rejected |

### Alternatives table — marker semantics

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Borrow doctor's ✓/✗ for present/absent + add ? for unknown | Three states cover every cell; visually matches doctor; "?" reads naturally as "I don't know yet" | ✗ on absent reads as "broken"; readers must learn it's informational | **Chosen** |
| New scheme (●/○/?) | Avoids the "✗ = broken" misread | New vocabulary inconsistent with doctor; two visual languages | Rejected |
| Just text labels ("PRESENT"/"ABSENT"/"UNKNOWN") | Unambiguous | Loud and slow to scan; loses the visual at-a-glance affordance | Rejected |

Mitigation for the chosen path: the row's adjacent detail string distinguishes informational absence (`✗ brew  not installed`) from a failure judgment. Doctor reserves ✗ for "fix this"; status reserves ✗ for "I see nothing here." The marker glyph is the same; the surrounding language differs.

### Alternatives table — `--quiet` semantics for status

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Hide nothing under `--quiet` (status is already minimal) | Predictable | `?` placeholder rows are noise once they've been read once | Rejected |
| **Hide the three `?` drift placeholders under `--quiet`** | Operator-friendly default for "give me actionable state"; "?" rows are the only deferred / not-yet-implemented entries | Diverges from doctor's `--quiet` (which hides ✓ passes); status doesn't have a hide-passes mode | **Chosen** |
| Collapse to a one-line summary | Most compact | Loses the per-row detail; differs from doctor's behavior | Rejected |

The asymmetry is acceptable because status and doctor *are* asymmetric: doctor decides pass/fail, status reports state. The thing each command's `--quiet` hides is "noise for an operator who knows what's running" — which for doctor is pass rows, and for status is not-yet-implemented placeholders.

## 3. Scope

### Files to create

- `internal/output/output.go` — `Row(rt, w, marker, name, detail, verbose string)`, `Section(rt, w, name string)`, ANSI constants, `useColor` detection
- `internal/output/output_test.go` — table tests on emitted bytes
- `internal/status/status.go` — `Section` type, `Run(rt, w) error` orchestrator
- `internal/status/sections.go` — `brewSection`, `repoSection`, `driftSection`
- `internal/status/sections_test.go` — table tests with mocked seams
- `cmd/status_test.go` — smoke test using `cli.OsExiter` swap (mirrors `cmd/doctor_test.go`)

### Files to modify

- `internal/doctor/render.go` — replace local ANSI/useColor/row-formatting with calls into `internal/output`. The `*render` struct stays (it holds `quiet`/`verbose`/`useColor` plus the underlying writer); only the **emission helpers** delegate.
- `internal/doctor/render_test.go` — assertions stay byte-identical; only imports may change if any test was reaching into internals.
- `cmd/status.go` — replace stub body with handler that calls `status.Run(rt, cmd.Root().Writer)` and returns the error (always nil for status).

### Files NOT touched

- `cmd/root.go` — no new flags
- Anything else under `internal/`

## 4. `internal/output/output.go`

```go
// Package output provides the shared row renderer used by `macheim doctor`
// and `macheim status` (and any future read-only command that needs the
// same visual language). Intentionally has no dependency on internal/config
// or the urfave framework — pure rendering primitives.
package output

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiReset  = "\033[0m"
)

// Marker is the ✓ / ✗ / ? glyph rendered at the head of a Row.
type Marker int

const (
	MarkerOK Marker = iota
	MarkerFail
	MarkerUnknown
)

// UseColor reports whether ANSI escapes should be emitted. Returns false
// when the caller requests no color, when w is not an *os.File (e.g. a
// bytes.Buffer in tests, or an io.Pipe), or when the file is not a TTY.
func UseColor(noColor bool, w io.Writer) bool {
	if noColor {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// Row prints one labeled line:
//
//   <marker> <name>   <detail>
//
// When verbose is non-empty AND the caller's verbose flag is true, an
// indented context line follows. The colorized marker is rendered iff
// useColor. The package takes booleans rather than *config.Runtime to
// stay framework-agnostic — each consumer (doctor, status) computes
// useColor/verbose from its own state once and passes them in.
func Row(w io.Writer, useColor, verbose bool, m Marker, name, detail, probe string) {
	glyph, code := markerGlyph(m)
	if useColor {
		_, _ = fmt.Fprintf(w, "%s%s%s %s", code, glyph, ansiReset, name)
	} else {
		_, _ = fmt.Fprintf(w, "%s %s", glyph, name)
	}
	if detail != "" {
		_, _ = fmt.Fprintf(w, "  %s", detail)
	}
	_, _ = fmt.Fprintln(w)
	if verbose && probe != "" {
		_, _ = fmt.Fprintf(w, "   %s\n", probe)
	}
}

// markerGlyph maps a Marker to its (glyph, ansi-code) pair. The reset
// suffix is the caller's responsibility.
func markerGlyph(m Marker) (string, string) {
	switch m {
	case MarkerOK:
		return "✓", ansiGreen
	case MarkerFail:
		return "✗", ansiRed
	case MarkerUnknown:
		return "?", ansiYellow
	default:
		return "?", ansiReset
	}
}

// Summary prints a colored line (used for "All checks passed." style
// trailers). Empty `text` is a no-op.
func Summary(w io.Writer, useColor bool, m Marker, text string) {
	if text == "" {
		return
	}
	if !useColor {
		_, _ = fmt.Fprintln(w, text)
		return
	}
	_, code := markerGlyph(m)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", code, text, ansiReset)
}
```

### `output_test.go` shape

- `TestRow_OK_NoColor` / `TestRow_Fail_NoColor` / `TestRow_Unknown_NoColor` — exact byte output
- `TestRow_VerboseAddsIndentedLine`
- `TestRow_Color` — useColor=true; assert ANSI green/red/yellow + reset appear
- `TestSummary_Plain` / `TestSummary_Colorized`
- `TestUseColor_BytesBufferAlwaysFalse` / `TestUseColor_NoColorFlagBeatsTTY`

## 5. Doctor refactor — `internal/doctor/render.go`

The `*render` struct keeps its three behavioral fields (`quiet`/`verbose`/`useColor`) and the writer. Its `row` and `summary` methods delegate to `output`:

```go
func (r *render) row(name string, res Result) {
	if res.OK {
		if r.quiet {
			return
		}
		output.Row(r.w, r.useColor, r.verbose, output.MarkerOK, name, "", res.Probe)
		return
	}
	output.Row(r.w, r.useColor, r.verbose, output.MarkerFail, name, "", res.Probe)
	if res.Remediation != "" {
		_, _ = fmt.Fprintf(r.w, "   → %s\n", res.Remediation)
	}
}

func (r *render) summary(failed int) {
	if failed == 0 {
		output.Summary(r.w, r.useColor, output.MarkerOK, "All checks passed.")
		return
	}
	noun := "checks"
	if failed == 1 {
		noun = "check"
	}
	output.Summary(r.w, r.useColor, output.MarkerFail, fmt.Sprintf("%d %s failed.", failed, noun))
}
```

The doctor renderer has a unique extra: the `→ remediation` line on failure. That stays in doctor's render.go (it's not a generic primitive). The ANSI constants, `useColor` detection, and row+summary rendering all move to `internal/output`.

`internal/output` is decoupled from `internal/config` deliberately — it's a pure rendering package, framework-agnostic. Each consumer computes `useColor` and `verbose` from its own state once and passes them in.

## 6. `internal/status/status.go`

```go
// Package status implements `macheim status`: a read-only summary of
// what macheim sees right now. Always exits 0; sections may report
// "?" markers for not-yet-implemented checks.
package status

import (
	"io"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/output"
)

// Section is one named area of the status report.
type Section struct {
	Name string
	Run  func(rt *config.Runtime) []Row
}

// Row is one renderable line in a section's output.
type Row struct {
	Marker  output.Marker
	Name    string
	Detail  string
	Verbose string
	Hidden  bool // true → omitted under --quiet
}

// DefaultSections returns the canonical list of sections, in display order.
func DefaultSections() []Section {
	return []Section{
		brewSection(),
		repoSection(),
		driftSection(),
	}
}

// Run executes every section in order and prints the rows to w. Always
// returns nil — status is informational and never fails.
func Run(rt *config.Runtime, w io.Writer) error {
	useColor := output.UseColor(rt.NoColor, w)
	for _, s := range DefaultSections() {
		for _, row := range s.Run(rt) {
			if rt.Quiet && row.Hidden {
				continue
			}
			output.Row(w, useColor, rt.Verbose, row.Marker, row.Name, row.Detail, row.Verbose)
		}
	}
	return nil
}
```

## 7. `internal/status/sections.go`

Each section returns rows directly; no shared seam struct (the seams are scoped per section).

### brewSection

```go
func brewSection() Section {
	s := defaultBrewSeam()
	return Section{
		Name: "brew",
		Run: func(_ *config.Runtime) []Row {
			return []Row{brewRow(s)}
		},
	}
}

type brewSeam struct {
	arch    string
	stat    func(string) (os.FileInfo, error)
	version func(brewPath string) (string, error) // "/path/to/brew --version" first line
}

func brewRow(s brewSeam) Row {
	var path string
	switch s.arch {
	case "arm64":
		path = "/opt/homebrew/bin/brew"
	case "amd64":
		path = "/usr/local/bin/brew"
	default:
		return Row{Marker: output.MarkerFail, Name: "brew", Detail: fmt.Sprintf("unsupported arch %q", s.arch)}
	}
	if _, err := s.stat(path); err != nil {
		return Row{Marker: output.MarkerFail, Name: "brew", Detail: "not installed"}
	}
	v, err := s.version(path)
	if err != nil {
		return Row{Marker: output.MarkerOK, Name: "brew", Detail: fmt.Sprintf("%s (version unavailable)", path)}
	}
	return Row{Marker: output.MarkerOK, Name: "brew", Detail: fmt.Sprintf("%s (%s)", path, v)}
}

func defaultBrewSeam() brewSeam {
	return brewSeam{
		arch: runtime.GOARCH,
		stat: os.Stat,
		version: func(brewPath string) (string, error) {
			out, err := exec.Command(brewPath, "--version").Output()
			if err != nil {
				return "", err
			}
			// First line, e.g. "Homebrew 4.3.5"
			line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return line, nil
			}
			return parts[1], nil
		},
	}
}
```

We use `exec.Command` directly here, not `shell.Run`. The new `shell.Run` from sub-epic #8 takes a `*config.Runtime` and prints `[dry-run]` / verbose markers — neither applies to status's silent version probe. A follow-up could thread a "silent" mode through `shell.Run`, but for this PR a direct `exec.Command` is cleaner.

### repoSection

```go
func repoSection() Section {
	s := defaultRepoSeam()
	return Section{
		Name: "repo",
		Run: func(rt *config.Runtime) []Row {
			return []Row{repoRow(rt, s)}
		},
	}
}

type repoSeam struct {
	lastCommit func(repoPath string) (sha, subject, isoDate string, err error)
	isClean    func(repoPath string) (bool, error)
	stat       func(string) (os.FileInfo, error)
}

func repoRow(rt *config.Runtime, s repoSeam) Row {
	path, source, err := rt.ResolveRepoPath()
	if err != nil {
		return Row{Marker: output.MarkerFail, Name: "repo", Detail: err.Error()}
	}
	if path == "" {
		return Row{
			Marker:  output.MarkerUnknown,
			Name:    "repo",
			Detail:  "not configured (embed-fallback mode)",
			Verbose: "no source matched (--repo flag, MACHEIM_REPO, ~/.config/macheim/config.yaml, ~/src/macheim, ~/code/macheim)",
		}
	}
	if _, err := s.stat(path); err != nil {
		return Row{Marker: output.MarkerFail, Name: "repo", Detail: fmt.Sprintf("[%s] %s (missing)", source, path)}
	}
	sha, subject, isoDate, err := s.lastCommit(path)
	if err != nil {
		return Row{Marker: output.MarkerOK, Name: "repo", Detail: fmt.Sprintf("[%s] %s", source, path), Verbose: fmt.Sprintf("git log -1 failed: %v", err)}
	}
	clean, _ := s.isClean(path) // ignore err — fall through as "dirty"
	cleanLabel := "dirty"
	if clean {
		cleanLabel = "clean"
	}
	shortSHA := sha
	if len(sha) >= 7 {
		shortSHA = sha[:7]
	}
	return Row{
		Marker:  output.MarkerOK,
		Name:    "repo",
		Detail:  fmt.Sprintf("[%s] %s @ %s (%s)", source, path, shortSHA, cleanLabel),
		Verbose: fmt.Sprintf("%s — %s", subject, isoDate),
	}
}

func defaultRepoSeam() repoSeam {
	return repoSeam{
		lastCommit: gitrepo.LastCommit,
		isClean:    gitrepo.IsClean,
		stat:       os.Stat,
	}
}
```

### driftSection

```go
func driftSection() Section {
	return Section{
		Name: "drift",
		Run: func(_ *config.Runtime) []Row {
			return []Row{
				{Marker: output.MarkerUnknown, Name: "drift:brew", Detail: "not implemented (see #14)", Hidden: true},
				{Marker: output.MarkerUnknown, Name: "drift:dotfiles", Detail: "not implemented (see #17 / #18)", Hidden: true},
				{Marker: output.MarkerUnknown, Name: "drift:macos", Detail: "deferred", Hidden: true},
			}
		},
	}
}
```

`Hidden: true` triggers the `--quiet` filter in `Run`.

## 8. `cmd/status.go` (replace stub)

```go
package cmd

import (
	"context"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/status"
	"github.com/urfave/cli/v3"
)

func statusCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Read-only summary of drift between this Mac and the repo",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return status.Run(rt, cmd.Root().Writer)
		},
	}
}
```

## 9. `cmd/status_test.go` (smoke)

```go
package cmd

import (
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestStatus_RunsAndProducesOutput exercises cmd -> status.Run wiring.
// Status never returns an ExitCoder (it always succeeds), but we still
// swap cli.OsExiter as belt-and-suspenders in case a future check decides
// to fail.
func TestStatus_RunsAndProducesOutput(t *testing.T) {
	var captured int
	prevExiter := cli.OsExiter
	cli.OsExiter = func(code int) { captured = code }
	t.Cleanup(func() { cli.OsExiter = prevExiter })

	stdout, _, _, err := runRoot(t, "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != 0 {
		t.Errorf("status should not call OsExiter; got captured=%d", captured)
	}
	// At minimum, the brew, repo, and at least one drift line appear.
	for _, want := range []string{"brew", "repo"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q\nfull:\n%s", want, stdout)
		}
	}
}
```

## 10. Testing strategy

- **`internal/output/output_test.go`** — table tests on Row + Summary outputs; UseColor truth table.
- **`internal/doctor/render_test.go`** — existing assertions stay byte-identical. Verifies the refactor.
- **`internal/status/sections_test.go`** — table-driven per section:
  - brewSection: arm64+present, arm64+missing, amd64+present, amd64+missing, unsupported-arch, version-fails-but-binary-exists
  - repoSection: embed-fallback, configured+writable+last-commit, configured+missing, lastCommit-fails-but-stat-ok
  - driftSection: just emits three Hidden rows; assert by length + names
- **`cmd/status_test.go`** — integration smoke.
- All tests pure-Go, no real exec of brew, no real git.

## 11. Risk assessment

| Risk | Likelihood | Mitigation |
|---|---|---|
| Doctor's `render_test.go` byte assertions break during the refactor | Medium | Run render_test.go after the doctor refactor; if any diff appears, adjust `output.Row` until byte-identical. The promise is byte-stability. |
| `brew --version` is slow under load and stalls status output | Low | Wrap in a goroutine with a 2s context deadline; on timeout fall back to "version unavailable". (Defer this hardening to a follow-up unless the bare `exec.Command` is visibly slow.) |
| Status PR rebases break against fast-moving main | Low | Path-disjoint with the just-merged 5 PRs; no rebase pain expected. |
| `?` rows misread as actual "broken" rows | Medium | Detail text is explicit ("not implemented (see #14)" / "deferred"); see §2 marker-semantics rationale. |

## 12. Backward compatibility

`status` was a "not implemented yet (see issue #11)" stub on main. Replacing it with real output is additive. No flag changes; `--quiet`/`--verbose`/`--no-color` already inherit from root.

## 13. Dependencies (on other sub-issues)

- **Blocked by:** #4, #5, #6, #8, #9, #10 — all merged. The just-merged set is the foundation this consumes.
- **Blocks:** none directly; later commands (`bootstrap` orchestration) may consume status's internals.

## 14. Out of scope (deferred)

- Real drift detection — `drift:brew` lands in #14 (brew diff engine); `drift:dotfiles` in #17/#18; `drift:macos` deferred indefinitely.
- JSON output mode — YAGNI; not requested by any consumer.
- `brew --version` timeout / cancellation hardening — defer until a real stall is observed.
- Threading a "silent" mode through `shell.Run` — covered locally by `exec.Command` for this single call site.

## 15. Acceptance verification

| AC | From | Verifying command | Expected |
|---|---|---|---|
| Brew section (installed?, path, version) | #55 | `./dist/macheim status` | row beginning with `✓ brew` and ending with `(<version>)` on a healthy mac |
| Repo + last-commit + fallback-mode section | #56 | `./dist/macheim --repo $(pwd) status` | row beginning with `✓ repo` with `[flag] <path> @ <sha> (clean|dirty)`; bare `./dist/macheim status` shows `? repo  not configured (embed-fallback mode)` |
| Per-module drift placeholders | #57 | `./dist/macheim status` | three `?` rows: `drift:brew`, `drift:dotfiles`, `drift:macos` |
| `--quiet` hides drift placeholders | this design | `./dist/macheim --quiet status` | no `?` rows |
| `--verbose` adds probe lines | this design | `./dist/macheim --verbose status` | each row followed by an indented context line |
| `--no-color` strips ANSI | this design | `./dist/macheim --no-color status \| cat` (piping forces non-TTY) | no `\033[` escapes |
| Always exits 0 | #11 (read-only) | `./dist/macheim status; echo $?` | exit 0 |

Plus regressions:

| Check | Command | Expected |
|---|---|---|
| Build | `make build` | exit 0 |
| Tests | `make test` | exit 0 (includes new `internal/output` + `internal/status` packages, doctor render_test unchanged) |
| Lint | `make lint` | `0 issues` |
| CI | GitHub Actions on PR | all matrix legs pass |

## 16. Commit plan

Four commits (each independently passes `make lint && make test`):

1. `refactor(doctor): extract internal/output for shared row rendering` — `internal/output/output.go` + test; `internal/doctor/render.go` calls into it; assertions in `render_test.go` byte-identical.
2. `feat(status): add brew + repo sections behind testable seams` — `internal/status/sections.go` (brew + repo only) + `sections_test.go`; `internal/status/status.go` minus driftSection.
3. `feat(status): add drift placeholder section + Run orchestrator` — `driftSection` + `DefaultSections` + `Run`; tests cover the orchestrator's `--quiet` hide logic.
4. `feat(cmd): replace status stub with real handler` — `cmd/status.go` replaces stub; `cmd/status_test.go` smoke.

Closing line on commit 4:

```
Closes #11
Closes #55
Closes #56
Closes #57
```
