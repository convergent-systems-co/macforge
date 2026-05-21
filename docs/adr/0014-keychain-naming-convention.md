# ADR 0014 — Keychain Naming Convention

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

`GOALS.md` is explicit on two security principles:

> Dedicated keychains only. No `login.keychain` usage.

> Multi-team support.

That means MacForge creates and manages its own keychains, one (or more) per team identity. The naming convention matters because:

- Operators need to recognize MacForge-managed keychains at a glance.
- The convention must encode enough information to avoid collisions across teams and purposes.
- Naming is referenced from `macforge.yaml` and from audit logs — it must be deterministic.
- The convention must be macOS-legal (filesystem-safe characters).

The keychain itself is a file in `~/Library/Keychains/`. macOS appends `-db` to user-created keychains created via `security create-keychain`.

## Decision

The default keychain name is:

```
macforge-<TEAM_ID>-<PURPOSE>.keychain-db
```

Where:

- `<TEAM_ID>` is the 10-character Apple Developer Team ID (e.g., `XYZ1234567`). Case preserved.
- `<PURPOSE>` is a short slug. Initial set: `signing`, `notarize`, `dev`. Slug rules below.
- The `-db` suffix is appended automatically by macOS; we treat the full file basename (with suffix) as the canonical reference.

**Examples:**

```
~/Library/Keychains/macforge-XYZ1234567-signing.keychain-db
~/Library/Keychains/macforge-XYZ1234567-dev.keychain-db
~/Library/Keychains/macforge-ABC9876543-signing.keychain-db
```

### Slug rules

- Lowercase ASCII letters, digits, and hyphens only.
- Length 3–32 characters.
- Must not start or end with a hyphen.
- Reserved slugs: `signing`, `notarize`, `dev`, `release`, `staging`. Other slugs are allowed for custom use cases but must pass the validation regex `^[a-z0-9](?:[a-z0-9-]{1,30}[a-z0-9])?$`.

### Configuration

Users can override the full keychain name in `macforge.yaml`:

```yaml
keychain:
  name: macforge-XYZ1234567-signing       # default: macforge-<team>-<signing|...>
```

Setting `keychain.name` to anything not matching the convention requires `keychain.allow_nonstandard: true` and is logged as a `decision` event in the audit log.

### Unlock policy

- Keychain passwords MUST come from `env:VAR_NAME` or `keyring:macforge:<id>` references in `macforge.yaml`. Inline passwords are rejected at config-load time.
- MacForge never displays keychain passwords (per `Common.md §4`).
- MacForge sets the keychain to **lock on sleep** and **lock after 1 hour** by default (`security set-keychain-settings -l -t 3600`). Override via `keychain.lock_timeout` (seconds; `0` disables).

### Lifecycle commands

```
macforge keychain create   create a new managed keychain with the default name
macforge keychain delete   delete a managed keychain (refuses login.keychain)
macforge keychain list     list MacForge-managed keychains (filter on macforge- prefix)
macforge keychain unlock   unlock for the current shell session
```

Direct `macforge` commands (`identity create`, `identity import`, `sign`) operate against the keychain named in `macforge.yaml`. They unlock-and-operate; they do not leave the keychain unlocked beyond the command's lifetime unless `keychain.persist_unlock: true`.

## Consequences

### Positive

- **Visually identifiable.** `macforge-` prefix on every managed keychain.
- **Team-safe multi-tenancy.** Two teams on the same machine never collide.
- **Audit-friendly.** Audit-log entries can include `keychain` as a field that matches the file basename verbatim.
- **`login.keychain` is never touched** — explicit refusal in `keychain delete` and any operation that would mutate the login keychain.

### Negative

- **Operators with pre-existing keychains** must rename or import-then-recreate. We provide `macforge keychain import-existing --as macforge-<team>-<purpose>` to ease migration.
- **Name length** can grow. `macforge-XYZ1234567-signing.keychain-db` is 35 chars — well under macOS filename limits.

### Neutral

- The convention is documentation, not crypto. Anyone can create a keychain named `macforge-anything-fake.keychain-db`; MacForge's commands validate the file's keychain status (`security list-keychain-domains -d user`) before trusting it.

## Multi-team workflow

```yaml
# macforge.yaml — team A
team: XYZ1234567
keychain:
  name: macforge-XYZ1234567-signing
```

```yaml
# macforge.yaml — team B (sibling project)
team: ABC9876543
keychain:
  name: macforge-ABC9876543-signing
```

Both keychains coexist in `~/Library/Keychains/`. `macforge identity list` aggregates across all `macforge-*` keychains it finds.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **`macforge-<TEAM>-<PURPOSE>.keychain-db`** | Visually identifiable, collision-safe, audit-friendly, supports multi-team | Slightly long names | **Chosen** |
| `<PURPOSE>.keychain-db` (no prefix) | Short | Collisions with non-MacForge keychains; no team safety | Rejected |
| Generated UUID names | Strict collision avoidance | Unreadable; humans can't tell what they're looking at; awful debugging | Rejected — readability matters |
| Project-local keychains under `./.macforge/keychains/` | Stays out of `~/Library/` | Fights macOS keychain conventions; `security` defaults don't search there; `codesign` lookup gymnastics | Rejected — fight Apple, lose |
| Single MacForge keychain holding all identities | Simplest | Loses team isolation; can't share a keychain across machines safely | Rejected — defeats `GOALS.md` multi-team principle |

## Links

- Spec §7: [State Layout](../superpowers/specs/2026-05-21-macforge-architecture-design.md#7-state-layout)
- Reference: `Common.md §4` (secret handling), `GOALS.md` (Security section)
- Related: [ADR-0005](0005-state-and-config-layout.md)
