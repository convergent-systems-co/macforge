# ADR 0003 — Apple-Tool Boundary: Shell-Out Only

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge must interact with Apple's signing, notarization, and verification toolchain. The available surfaces are:

1. Apple's CLI tools: `codesign`, `security`, `xcrun notarytool`, `productbuild`, `pkgbuild`, `spctl`, `stapler`.
2. Apple's native frameworks: `Security.framework`, `CryptoKit`, `CoreFoundation` (accessible from Go only via cgo).
3. Apple's network APIs: App Store Connect REST API (for notarization).

The choice of boundary shapes everything else: build complexity (cgo? darwin-only?), testability (mockable? fixture-driven?), auditability (string-parsing vs. structured), and CI portability (which OSes can run which tests).

## Decision

**MacForge shells out to Apple's CLI tools exclusively.** Every Apple invocation flows through `internal/apple/runner.go`'s `Runner` interface. Two implementations exist: `ExecRunner` (real `os/exec`) and `FakeRunner` (replays fixtures from `testdata/`).

- No cgo.
- No direct Security.framework calls.
- No direct App Store Connect REST calls (we use `notarytool`, which calls the API for us).

Per-tool wrappers under `internal/apple/{codesign,security,notarytool,productbuild,spctl}` translate structured Go inputs to argv, and parse stdout/stderr into structured Go outputs.

## Consequences

### Positive

- **Auditability is trivial.** Every Apple call is a process spawn with explicit argv. The audit log records argv, exit code, and stdout/stderr hashes. Reproducing a MacForge run by hand is just reading the audit log.
- **No cgo.** Cross-compile from any host. CI doesn't need a darwin builder for unit and tier-2 tests.
- **One mockable seam.** `Runner` interface is the only `os/exec` consumer. Faking is mechanical.
- **Mirrors what humans do.** Anyone debugging can copy a command line and run it by hand — the audit log is a runbook.
- **Apple's CLI evolves more slowly than its frameworks.** Lower churn surface.

### Negative

- **Stdout/stderr parsing is brittle.** Apple changes wording between macOS releases (e.g. `codesign --display` output). Wrappers must defend against this; fixtures track the expected shape per macOS version we support.
- **Notarization uses `notarytool`, which is itself a wrapper.** We're one layer further from the API. Acceptable cost — `notarytool` is the documented supported interface.
- **No structured introspection of keychains.** We get text out of `security`, not objects.

### Neutral

- Shelling out from Go is cheap (microseconds). The dominant latency is the Apple tool itself, not the spawn cost.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Shell-out only** | Auditable, no cgo, fakeable, no darwin build dependency | stdout-parsing brittleness | **Chosen** |
| Native Security.framework via cgo | Structured data, deterministic | darwin-only build, cgo toolchain dependency, harder to audit (opaque framework calls), App Store Connect still needs HTTPS → partial coverage anyway | Rejected — fails CI portability and auditability tests |
| Hybrid: shell-out + direct App Store Connect REST | Structured notarization payloads, full auditability of signing path | Two integration patterns, JWT auth surface, duplication of `notarytool`'s job | Rejected — complexity not justified at v0.x |
| Hybrid: native keychain + shell-out signing | Tight control over keychain operations | cgo plus dual patterns; loses single-boundary benefit | Rejected — same problem, dressed differently |

## Links

- Spec §6: [Apple-Tool Boundary](../superpowers/specs/2026-05-21-macforge-architecture-design.md#6-apple-tool-boundary)
- Related: [ADR-0011](0011-error-model-and-codes.md) (error propagation through the boundary), [ADR-0012](0012-audit-log-schema.md) (audit hook in `Runner`), [ADR-0013](0013-testing-strategy-three-tiers.md) (FakeRunner enables tier 2)
