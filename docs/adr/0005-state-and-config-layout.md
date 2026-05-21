# ADR 0005 — State and Config Layout: Hybrid

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge runs in two modes:

- **Interactive / local development** — a developer on their Mac configuring identities, signing dev builds.
- **CI** — GitHub Actions / GitLab / Azure DevOps; hermetic, ephemeral, no user home directory worth persisting across runs.

The state model must serve both. Project-scoped config that can be committed and reviewed; ephemeral per-run state that stays out of git; user-scoped state that survives across projects but is never required.

## Decision

Hybrid layout:

```
<repo>/
├── macforge.yaml                  committed — authoritative project config
└── .macforge/                     gitignored — ephemeral per-project state
    ├── audit/<UTC>.jsonl
    ├── runs/<trace>/              per-run scratch
    └── cache/                     downloaded notarization logs, ASC metadata

~/Library/Application Support/MacForge/
├── config.yaml                    global defaults; team registry
├── identities/<team>.json         metadata only (no secrets)
├── audit/                         optional cross-project mirror (opt-in)
└── credentials/                   ASC profile metadata (NEVER raw values)

~/Library/Keychains/macforge-<team>-<purpose>.keychain-db   (managed by macOS)
```

**Precedence** (viper layering, highest first):

1. CLI flag
2. Environment variable (`MACFORGE_*`)
3. `./macforge.yaml`
4. `~/Library/Application Support/MacForge/config.yaml`
5. Built-in defaults

The user-level layer is **always optional**. CI can run with zero user-level state. A fresh checkout + valid `macforge.yaml` + the right env vars is sufficient.

## Consequences

### Positive

- **CI-native.** Pure project config + env vars is a complete run; no user-home prerequisites.
- **Audit ships with the artifact.** `./.macforge/audit/` is right next to the binary that was signed; archive it as a release artifact.
- **Multi-team and cross-project work survives** via the user-level layer.
- **`macforge.yaml` is reviewable** in PRs — config changes go through the same gate as code.

### Negative

- **Two locations** means two mental models. Documentation must be explicit about which lives where.
- **Audit log path inside `.macforge/`** means it's gitignored by default — by design, but worth a callout in onboarding.

### Neutral

- macOS uses `~/Library/Application Support/MacForge/` per Apple's Standard Directories convention. We don't use XDG paths because we are macOS-only (the build chain runs on Linux/Windows; the runtime semantics belong to macOS).

## `macforge.yaml` schema (v1)

```yaml
version: 1
team: XYZ1234567
identity:
  signing: developer-id-application
  installer: developer-id-installer        # optional
keychain:
  name: macforge-XYZ1234567-signing        # default convention; see ADR-0014
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD   # never inline
sign:
  hardened_runtime: true
  timestamp: true
  entitlements: ./Entitlements.plist
notarize:
  asc_profile: macforge-prod               # stored via `notarytool store-credentials`
  wait: true
  staple: true
package:
  formats: [zip, dmg]
publish:
  github:
    repo: convergent-systems-co/myapp
    draft: false
```

Schema is versioned at the top level. Migrations are explicit (no silent upgrades).

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Hybrid (project + user)** | CI-native, audit ships with artifact, multi-team supported | Two locations to document | **Chosen** |
| Project-local only | Single mental model; maximum hermeticity | Multi-team support means duplicate config per project; no cross-project history | Rejected — loses multi-team and audit-history value |
| User-level only (XDG / macOS-native) | Familiar to brew users | Project config can't ship in the repo; reviewability lost | Rejected — fails the CI-native requirement |
| Configurable root via `$MACFORGE_HOME` | Maximum flexibility | More paths to document and test; YAGNI | Rejected — premature; reconsider if needed |

## Links

- Spec §7: [State Layout](../superpowers/specs/2026-05-21-macforge-architecture-design.md#7-state-layout)
- Related: [ADR-0012](0012-audit-log-schema.md), [ADR-0014](0014-keychain-naming-convention.md)
