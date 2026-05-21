# ADR 0011 — Error Model: Sentinels + Stable Codes

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge errors travel through three audiences:

1. **Humans on a terminal** — they need actionable messages with a hint.
2. **CI scripts and downstream tools** — they need stable, parseable codes.
3. **The audit log** — it needs structured fields (code, op, details).

Plain Go errors (`fmt.Errorf("oops")`) satisfy none of these. `errors.Is/As` against sentinel errors handle (2) and (3) but not (1). A full error system with rich codes, hints, and structured details covers all three.

## Decision

A two-layer error model:

### Layer 1 — Stable codes

Every distinct error condition gets a stable code from the namespace `MF-<SUBSYSTEM>-NNN`. Codes are part of MacForge's public contract — once published, never re-meaninged.

Subsystem prefixes (initial set):

| Prefix      | Subsystem                          |
|-------------|------------------------------------|
| `MF-IDENT`  | identity (cert / key / CSR)        |
| `MF-KEYCHAIN` | keychain lifecycle               |
| `MF-SIGN`   | signing                            |
| `MF-PACKAGE`| packaging                          |
| `MF-NOTARIZE` | notarization                     |
| `MF-VERIFY` | verification                       |
| `MF-PUBLISH`| publishing                         |
| `MF-RELEASE`| release pipeline orchestration     |
| `MF-TOOL`   | Apple-tool boundary (missing tool, non-zero exit) |
| `MF-CONFIG` | config loading and validation      |
| `MF-AUDIT`  | audit-log writer                   |
| `MF-OUTPUT` | output rendering                   |
| `MF-CI`     | CI provider integration            |

Codes start at `001` per subsystem. New conditions append. Retired codes stay reserved.

### Layer 2 — The `Error` type

```go
// internal/errors/error.go
package errors

type Error struct {
    Code    string            // e.g. "MF-NOTARIZE-001"
    Op      string            // e.g. "notarize.Submit"
    Msg     string            // human-readable, terse
    Hint    string            // optional — remediation suggestion
    Details map[string]any    // structured context (NEVER secrets — apply redaction)
    Cause   error             // underlying cause; supports errors.Unwrap
}

func (e *Error) Error() string { return fmt.Sprintf("[%s] %s: %s", e.Code, e.Op, e.Msg) }
func (e *Error) Unwrap() error { return e.Cause }

// Sentinel roots per subsystem — for errors.Is matching at coarse granularity.
var (
    ErrIdentity = stdErrors.New("identity")
    ErrKeychain = stdErrors.New("keychain")
    ErrSigning  = stdErrors.New("signing")
    ErrPackage  = stdErrors.New("package")
    ErrNotarize = stdErrors.New("notarize")
    ErrVerify   = stdErrors.New("verify")
    ErrPublish  = stdErrors.New("publish")
    ErrTool     = stdErrors.New("tool")
    ErrConfig   = stdErrors.New("config")
)

// Constructors keep the code/op/sentinel matching tight.
func NewSigning(code, op, msg string, opts ...Option) *Error { ... }
func NewNotarize(code, op, msg string, opts ...Option) *Error { ... }
// ...
```

`errors.Is(err, errs.ErrSigning)` matches anything in the signing subsystem. `errors.As(err, &mfErr)` extracts the structured form. Code-level matching is `err.Code == "MF-SIGN-001"` after the `As` cast.

### Layer 3 — Output envelope

The output renderer (see [ADR-0006](0006-output-format-dual.md)) shapes the `*Error` into either human text or JSON envelope:

```json
{
  "ok": false,
  "schema": "macforge.v1.error",
  "trace": "01HVQK...",
  "command": "notarize",
  "code": "MF-NOTARIZE-001",
  "op": "notarize.Submit",
  "message": "Apple rejected the submission: invalid signing identity",
  "hint": "Run `macforge identity status` to confirm the cert is not expired",
  "details": {
    "submission_uuid": "abc-123",
    "log_url": "https://..."
  }
}
```

### Wrapping discipline

- Apple-tool wrappers (`apple/codesign` etc.) emit `MF-TOOL-NNN` errors.
- Subsystem callers (`signing.Sign`) wrap with `MF-SIGN-NNN`. The original `MF-TOOL` error is preserved as `Cause` so `errors.As(err, &mfErr)` and `errors.Unwrap` both work.
- Each `Op` is the **caller site**, not the failing tool: `signing.Sign`, not `codesign`.

### Exit codes

- `0` — success.
- `1` — generic failure (CLI parse error, unexpected panic recovery).
- `2..255` — reserved; future subsystem-coded exits if proven useful.

The structured code in stderr / JSON envelope is the primary failure signal; exit codes stay coarse.

## Consequences

### Positive

- **Stable contract for CI consumers.** Code values become part of the public API and follow SemVer rules — new codes are minor, retiring a code is major.
- **`Hint` field** turns errors into mini-runbooks.
- **`Cause` chain preserves full diagnostic** — Apple-tool stderr is never lost.
- **`errors.Is` matches at the subsystem level**, `As` extracts the code — Go-idiomatic.

### Negative

- **Discipline cost.** Every new error condition must be added to a constants file with a fresh code. Reviewers must enforce.
- **Hint text must be maintained.** Stale hints are worse than absent hints; a hint that points to a deleted command is a defect.

### Neutral

- The error system is part of `internal/errors` — also `internal`. If we ever expose a Go library API, error types lift out then.

## Redaction

Per `Common.md §4`:

- `Details` MUST NOT contain raw secrets. Apply `Invocation.Redact` substrings before populating `Details`.
- `Cause.Error()` strings that came from Apple tools MUST pass through the same redactor.
- Tests assert that no `MF-*` error in the suite carries a recognizable secret-pattern substring.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Sentinel roots + stable codes + Error type** | Three-audience coverage; Go-idiomatic; stable contract | Discipline cost on every new condition | **Chosen** |
| `fmt.Errorf("...")` + sentinels only | Minimal | No stable code surface; CI consumers must regex error strings — fragile | Rejected |
| `github.com/pkg/errors` style | Stack traces, simple wrap | Pre-go-1.13 idiom; `errors.Is/As` is the modern replacement | Rejected — go-1.13+ stdlib is sufficient |
| gRPC-style status codes (numeric enum) | Compact, widely known | Numeric codes are opaque without a lookup table; `MF-*` codes are self-describing | Rejected — readability wins |

## Links

- Spec §9: [Error Model](../superpowers/specs/2026-05-21-macforge-architecture-design.md#9-error-model)
- Related: [ADR-0003](0003-apple-tool-boundary-shell-out.md), [ADR-0006](0006-output-format-dual.md), [ADR-0012](0012-audit-log-schema.md)
