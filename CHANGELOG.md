# Changelog

All notable changes to MacForge are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Versions follow
SemVer.

## [Unreleased]

### Added

- **`macforge apple config validate`** — runs every static rule `config.Load` enforces plus runtime checks (env-var presence for `keychain.unlock`, keychain file reachability via `security show-keychain-info`). Per-rule `✓` / `✗` / `○` output; exits non-zero on any red. JSON envelope schema `macforge.v1.apple.config.validate`. See [ADR-0019](docs/adr/0019-aggressive-config-validation.md). ([#13](https://github.com/convergent-systems-co/macforge/issues/13))
- **`config.ResolveKeychainName(cfg)`** — single resolver every Apple verb consumes for the keychain name. Returns `cfg.Keychain.Name` if set, else `keychain.DefaultName(cfg.Team, "signing")`. Replaces direct `cfg.Keychain.Name` reads in signing. ([#13](https://github.com/convergent-systems-co/macforge/issues/13))
- **`macforge identity create`** — generate an RSA-2048 keypair + PKCS#10 CSR for the Apple Developer ID portal, AND bundle the private key in an encrypted PKCS#12 backup, AND import the key into the configured macforge keychain. The private key never touches disk unencrypted. Flags: `--cn` (required), `--org`, `--email`, `--country`, `--out` (path prefix; default `./identity` → `./identity.csr` + `./identity.p12`), `--password` (optional; generated if omitted and shown once on stdout), `--keychain`. Promotes the stub from v0.2 — closes the v0.1 gap where users were told to import a Developer ID cert but the tool couldn't help them generate the CSR to get one.
- **`macforge identity rotate`** — archive the current keychain identities to an encrypted PKCS#12, then generate a fresh RSA-2048 key + CSR (like `create`). Both old and new keys remain in the keychain afterward (Apple allows multiple valid Developer ID certs per team). Same flags as `create` plus `--archive-out` (default `./identity-archive-<UTC>.p12`), `--archive-password`, `--no-archive`.
- **`macforge identity export`** — write all identities from the configured keychain to an AES-encrypted PKCS#12 backup via `security export`. Flags: `--keychain`, `--out`, `--password`. Output file is `chmod 0600`. Useful for password-manager backups and seeding CI runners.

### Fixed

- **CRITICAL: signing no longer silently consumes a stale `keychain.name`.** `internal/signing.Service.Sign` now reads through `config.ResolveKeychainName`; `config.Load` strict-validates that, when set, `keychain.name` matches `macforge-<TEAM>-<PURPOSE>` AND that its `<TEAM>` segment equals top-level `team`. Mismatches now fail at Load with `MF-CONFIG-001` + a hint pointing at the offending field rather than failing 20 minutes later inside `sign` with a misleading `MF-SIGN-002`. Escape hatch: `keychain.allow_nonstandard: true` opts out of both `keychain.name` checks. ([#13](https://github.com/convergent-systems-co/macforge/issues/13))
- **`keychain create` now adds the new keychain to the user's keychain search list.** Previously the `.keychain-db` file was created on disk but not registered, so `find-identity` and `codesign`'s automatic identity resolution couldn't see imported certs. Symmetric removal on `keychain delete`. ([#2](https://github.com/convergent-systems-co/macforge/issues/2))
- **CRITICAL: audit log no longer leaks `Invocation.Redact` secrets.** Per-call-site secrets (keychain passwords, `--password`, ASC credentials) declared in `Invocation.Redact` are now masked from `probe_payload` BEFORE the event is written. Previously the Redact list became informational tags only; the raw payload (with passwords inline) was persisted to the JSONL audit log. ([#3](https://github.com/convergent-systems-co/macforge/issues/3))

### Changed (breaking — CLI structure)

- **All Apple-platform verbs moved under `macforge apple <verb>`.** Per [ADR-0017](docs/adr/0017-apple-command-namespace.md). Preparation for macheim merging in as a peer subtree. ([#14](https://github.com/convergent-systems-co/macforge/issues/14))
  - `macforge sign` → `macforge apple sign`
  - `macforge keychain create` → `macforge apple keychain create`
  - `macforge identity create` → `macforge apple identity create`
  - ... and the same for every other verb (`init`, `verify`, `package`, `notarize`, `publish`, `release`, `audit`).
  - `macforge version` stays at the root (universal).
  - JSON envelope `schema` field: `macforge.v1.sign` → `macforge.v1.apple.sign`; audit `command` field: `sign` → `apple.sign`. Same shape for every verb.
  - Hard cutover; no backwards-compat aliases.

### Changed (breaking — audit log layout)

- **Audit log moved to `~/.macforge/audit/<UTC-date-time>Z-<trace>.jsonl`, one file per `macforge` invocation.** Per [ADR-0016](docs/adr/0016-audit-log-per-invocation-user-home.md). ([#5](https://github.com/convergent-systems-co/macforge/issues/5))
  - **Was:** `./.macforge/audit/<UTC-date>.jsonl` (project-local, daily-rotated).
  - **Now:** `~/.macforge/audit/2026-05-21T15-30-22Z-<26char-ULID>.jsonl` (user-home, per-invocation).
  - Format and schema (JSONL, ADR-0012 vocabulary) unchanged; only location and rotation changed.

  **Migration:** existing `.macforge/` directories in project dirs are now unused; safe to delete. Existing audit files at the old path will no longer be appended to.

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
