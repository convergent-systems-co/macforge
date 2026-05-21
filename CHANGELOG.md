# Changelog

All notable changes to MacForge are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Versions follow
SemVer.

## [Unreleased]

### Fixed

- **`keychain create` now adds the new keychain to the user's keychain search list.** Previously the `.keychain-db` file was created on disk but not registered, so `find-identity` and `codesign`'s automatic identity resolution couldn't see imported certs. Symmetric removal on `keychain delete`. ([#2](https://github.com/convergent-systems-co/macforge/issues/2))
- **CRITICAL: audit log no longer leaks `Invocation.Redact` secrets.** Per-call-site secrets (keychain passwords, `--p12-password`, ASC credentials) declared in `Invocation.Redact` are now masked from `probe_payload` BEFORE the event is written. Previously the Redact list became informational tags only; the raw payload (with passwords inline) was persisted to the JSONL audit log. ([#3](https://github.com/convergent-systems-co/macforge/issues/3))

### Changed (breaking — config layout)

- **Config layered: global at XDG path + optional project override.** Per [ADR-0015](docs/adr/0015-single-global-config-xdg.md). ([#1](https://github.com/convergent-systems-co/macforge/issues/1), [#4](https://github.com/convergent-systems-co/macforge/issues/4))
  - Global: `${XDG_CONFIG_HOME:-$HOME/.config}/macforge/macforge.yaml` (REQUIRED; identity-shaped fields).
  - Project: `./macforge.yaml` (OPTIONAL override; project-shaped fields).
  - `macforge init` now writes the **global** file. Project-local files are user-authored.

  **Migration for v0.1-dev testers:** previous `./macforge.yaml` files become optional override files. Run `macforge init --team <TEAM>` once to create the global; optionally trim project files to just override fields, or delete them if not needed.

## [v0.1.0] — 2026-05-21

The bootstrap release. Foundation + first vertical slice (identity →
keychain → sign → verify). Note: this release is **hand-signed and
hand-notarized** per the bootstrap procedure documented in
[`docs/adr/0007-release-dogfood-and-bootstrap.md`](docs/adr/0007-release-dogfood-and-bootstrap.md);
subsequent releases are signed by the previous MacForge.

### Added

- `macforge init` — scaffold `macforge.yaml` keyed on a team ID.
- `macforge keychain {create, delete, list, unlock}` — manage dedicated
  `macforge-<TEAM>-<PURPOSE>` keychains.
- `macforge identity {import, list, status}` — import Developer ID
  certificates, list signing identities, read cert validity windows.
- `macforge sign <path>` — sign one or more artifacts.
- `macforge verify <path>` — codesign + spctl verification.
- `macforge audit` — placeholder (full implementation in v0.2).
- JSONL audit log at `./.macforge/audit/<UTC>.jsonl` with field
  vocabulary mirroring `~/.ai/Common.md §5.2`.
- Dual-mode output (`--output human|json`) with stable schemas.
- 14 architectural ADRs in `docs/adr/`.
- Three-tier testing: unit, FakeRunner integration, darwin+e2e.

### Stubbed (planned for v0.2)

- `macforge identity {create, rotate, export}`.
- `macforge package`, `macforge notarize`, `macforge publish`, `macforge release`.

### Security

- Inline passwords in `macforge.yaml` are rejected at config-load.
- `login.keychain` operations are refused.
- Audit-log redaction is applied before write per `Common.md §4`.
