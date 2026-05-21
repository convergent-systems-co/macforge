# ADR 0001 — Language: Go

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge wraps Apple's release toolchain (`codesign`, `security`, `notarytool`, `productbuild`, `spctl`) and ships a single CLI binary plus a GitHub Action. The tool must:

- Cross-compile from any host to `darwin/amd64` and `darwin/arm64` cheaply.
- Distribute as a single static binary with no runtime.
- Have an ergonomic shell-out / process-handling story (every Apple call is `os/exec`).
- Be approachable for a broad open-source contributor base.
- Survive a decade-plus maintenance window without a treadmill of major-version migrations.

Candidates considered: Go, Rust, Swift, Python.

## Decision

**Go 1.26+** (currently `go 1.26.3` in `go.mod`).

The `go.mod` module path is `github.com/convergent-systems-co/macforge`.

## Consequences

### Positive

- Single static binary; trivial cross-compile (`GOOS=darwin GOARCH=arm64 go build`).
- `os/exec`, `context`, `errors.Is/As`, `slog`, `encoding/json` are first-class stdlib.
- Strong ecosystem for CLI tooling (cobra, viper, goreleaser).
- Stable language; v1 compatibility promise reduces long-term maintenance risk.
- Familiar to ops engineers (kubectl, gh, helm, hugo, terraform are all Go).
- Easy to onboard contributors.

### Negative

- Native macOS Security.framework integration would require cgo (we choose not to — see ADR-0003).
- Generics still relatively young; some library APIs feel pre-generic.
- Reflection-heavy serialization paths cost CPU vs. Rust.

### Neutral

- The tool's hot path is process spawning and I/O, not CPU; Go's runtime cost is invisible compared to Apple-tool latency.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Go** | Single binary, cross-compile, stdlib breadth, ops familiarity, decade stability | Generics young; cgo for native Apple APIs is awkward | **Chosen** |
| Rust | Memory safety, performance, expressive type system, single binary | Steeper learning curve; build times; smaller ops-tooling community; macOS framework FFI also non-trivial | Rejected — onboarding cost not justified for an I/O-bound tool |
| Swift | Native access to Security.framework, CryptoKit, Foundation | macOS/Linux Swift toolchain awkward on CI; smaller cross-platform CLI ecosystem; Apple-platform lock-in fights the open-source-first principle | Rejected — exactly the wrong language for cross-platform CI tooling |
| Python | Fastest iteration | Distribution is hostile (PyInstaller / pyOxidizer / `pipx`); GIL irrelevant but startup latency is a CI tax; long-term type maturity uneven | Rejected — distribution and durability fail the Decade Test |

## Links

- Spec: [`../superpowers/specs/2026-05-21-macforge-architecture-design.md`](../superpowers/specs/2026-05-21-macforge-architecture-design.md)
- Related: [ADR-0002](0002-project-layout-cmd-internal.md), [ADR-0003](0003-apple-tool-boundary-shell-out.md), [ADR-0004](0004-dependency-stack.md)
