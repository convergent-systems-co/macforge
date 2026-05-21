# Changelog

All notable changes to MacForge are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Versions follow
SemVer.

## [Unreleased]

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
