# ADR 0015 — Layered Config: Global Base at XDG Path, Optional Project Override

| Status      | Accepted                                     |
|-------------|----------------------------------------------|
| Date        | 2026-05-21                                   |
| Supersedes  | Parts of [ADR-0005](0005-state-and-config-layout.md) (user-level path; filename) |
| Issues      | [#1](https://github.com/convergent-systems-co/macforge/issues/1), [#4](https://github.com/convergent-systems-co/macforge/issues/4) |

## Context

[ADR-0005](0005-state-and-config-layout.md) defined a two-tier config model where the user-level file lived at `~/Library/Application Support/MacForge/config.yaml`. That location was non-discoverable and used a different filename than the project-local file. In practice, almost every macOS developer-tool keeps its config under `~/.config/<tool>/`, and a single Apple Team ID per organization is the norm.

The user requested:
- One global file at `~/.config/macforge/macforge.yaml`.
- Optional project-local override at `./macforge.yaml` that carries ONLY project-specific fields.
- Pattern: load global, override with project — like `~/.gitconfig` + `.git/config`.

## Decision

MacForge uses a **layered config**:

- **Global** (REQUIRED): `${XDG_CONFIG_HOME:-$HOME/.config}/macforge/macforge.yaml`. Holds identity-shaped fields (team, keychain, signing identity, ASC profile). Created once by `macforge init`. Without it, every verb except `init` and `version` errors with `MF-CONFIG-002`.
- **Project-local** (OPTIONAL): `./macforge.yaml` in the working directory. Hand-authored. Should carry only project-shaped fields (entitlements path, package formats, publish target). Not enforced — anyone can override anything — but the convention is documented.
- **Audit log** unchanged: `./.macforge/audit/<UTC>.jsonl` — per-project per ADR-0005.

### Resolution order (highest priority first)

```
1. CLI flag                          --team-id, --entitlements, --config, ...
2. Environment variable              MACFORGE_TEAM, MACFORGE_KEYCHAIN_PASSWORD, ...
3. ./macforge.yaml                   project override (OPTIONAL — silent if missing)
4. ~/.config/macforge/macforge.yaml  global base (REQUIRED)
5. Built-in defaults
```

The `--config <path>` flag overrides the global path only; project-local is always cwd-relative. Environment variables use the `MACFORGE_` prefix with dots replaced by underscores (e.g., `MACFORGE_KEYCHAIN_LOCK_TIMEOUT=600` for `keychain.lock_timeout`).

### Field classification — what lives where

Convention. Not enforced by the loader in v0.1; reviewers + docs are the gate.

| Field                              | Global ✓ | Project ✓ | Why                                                          |
|------------------------------------|:--------:|:---------:|--------------------------------------------------------------|
| `version`                          | ✓        |           | Schema version; global owns the source of truth              |
| `team`                             | ✓        |           | One Apple Team ID per developer/org                          |
| `identity.signing`                 | ✓        |           | Same cert across all projects                                |
| `identity.installer`               | ✓        |           | Same installer cert across projects (when used)              |
| `keychain.name`                    | ✓        |           | The dedicated keychain is identity-scoped                    |
| `keychain.unlock`                  | ✓        |           | Same env-var reference for all signing in the org            |
| `keychain.allow_nonstandard`       | ✓        |           | Operator preference, not project-shaped                      |
| `keychain.lock_timeout`            | ✓        |           | Operator preference                                          |
| `keychain.persist_unlock`          | ✓        |           | Operator preference                                          |
| `notarize.asc_profile`             | ✓        |           | One App Store Connect profile per account                    |
| `audit.user_mirror`                | ✓        |           | Operator preference                                          |
| `audit.bodies`                     | ✓        |           | Operator preference                                          |
| `sign.hardened_runtime`            |          | ✓         | Apps that link entitlements that disable HR may need override |
| `sign.timestamp`                   |          | ✓         | Almost always true; rare to override                         |
| `sign.entitlements`                |          | ✓         | The plist is the project's own                               |
| `notarize.wait`                    |          | ✓         | Per-project release-cadence choice                           |
| `notarize.staple`                  |          | ✓         | Almost always true for .app/.pkg/.dmg                        |
| `package.formats`                  |          | ✓         | `.dmg` vs `.zip` is per-project                              |
| `publish.github.repo`              |          | ✓         | One per project obviously                                    |
| `publish.github.draft`             |          | ✓         | Per-project release process                                  |

A project file with no overrides is valid (empty file works); a project file with `team:` will override the global team — not recommended, but allowed.

### Example layout

```
~/.config/macforge/macforge.yaml         (global; ~150 bytes)
─────────────────────────────────────────
version: 1
team: XYZ1234567
identity:
  signing: developer-id-application
keychain:
  name: macforge-XYZ1234567-signing
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD
notarize:
  asc_profile: macforge-prod


~/code/myapp/macforge.yaml               (project override; ~80 bytes)
─────────────────────────────────────────
sign:
  entitlements: ./Entitlements.plist
package:
  formats: [zip, dmg]
publish:
  github:
    repo: convergent-systems-co/myapp
```

## Consequences

### Positive

- **XDG-compliant, discoverable.** Matches gh, cargo, kubectl, helm, fly.
- **Single source of truth for identity-shaped fields.** No accidental drift across projects.
- **Project overrides remain possible** for the genuinely-per-project knobs.
- **Filename consistency** — `macforge.yaml` at both layers, no `config.yaml` ambiguity.
- **Migration path is one command:** `macforge init --team <TEAM>` writes the global; old `./macforge.yaml` files become optional overrides (and most can be deleted since their team / keychain entries are now redundant).

### Negative

- **No formal enforcement** of the project-vs-global field convention in v0.1. A project file can override `team`, and that may confuse the operator. Documented; could be enforced via a `--strict-layering` flag later.
- **Migration cost for v0.1-dev testers** — their existing `./macforge.yaml` files now layer ON TOP of a global that doesn't exist yet. They need to run `macforge init --team <T>` once and optionally trim their project files.

### Neutral

- `internal/config/paths_darwin.go` and `paths_other.go` removed; uniform path resolution.
- `internal/config.LoadOptions` is `{GlobalPath, ProjectPath}` — both layered explicitly.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Layered (global at XDG + optional project override)** — this ADR | Single SoT for identity; per-project knobs work; git-like model | Convention not enforced | **Chosen** |
| Single global only (no project layer) | One mental model; simplest | Loses per-project entitlements / formats / publish target | Rejected by maintainer |
| Two-tier with `~/Library/Application Support/MacForge/config.yaml` (ADR-0005 original) | macOS-canonical Apple path | Non-discoverable; inconsistent filename | Superseded |
| Profile-based config (`--profile <name>` selects a section) | Multi-team via profile switching | Premature; one-team-per-org is the dominant case | Deferred |

## Implementation Notes

- `internal/config/paths.go`: `UserConfigDir()` honors `$XDG_CONFIG_HOME`; `ConfigPath()` returns the global file path; `ProjectAuditDir(cwd)` unchanged.
- `internal/config/load.go`: `Load()` reads global as base then merges project on top; missing project file is silent; missing global file is `MF-CONFIG-002` with hint.
- `cmd/macforge/init_cmd.go`: writes the global, never the project. `MkdirAll`s the parent dir. Refuses overwrite.
- `cmd/macforge/runtime.go`: `macforgeYAMLPath()` returns `config.ConfigPath()` (the global location).
- Build-tag files `paths_darwin.go` / `paths_other.go` removed; the path is platform-uniform now.

## Migration for v0.1-dev testers

```bash
# 1. Create the global config:
macforge init --team <YOUR_TEAM_ID>

# 2. (Optional) Trim your project-local macforge.yaml to project-only fields.
#    The example below assumes Entitlements + package formats are the only
#    project-specific knobs:
cat > /your/project/macforge.yaml <<'YAML'
sign:
  entitlements: ./Entitlements.plist
package:
  formats: [zip]
YAML

# 3. (Optional) Delete the project-local file entirely if you don't need
#    overrides — global will be the only source.
rm /your/project/macforge.yaml
```

## Links

- Supersedes [ADR-0005 §State and config layout](0005-state-and-config-layout.md) (user-level path and filename only; audit-log + keychain locations unchanged).
- Closes [issue #1](https://github.com/convergent-systems-co/macforge/issues/1), [issue #4](https://github.com/convergent-systems-co/macforge/issues/4).
- Reference: [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).
