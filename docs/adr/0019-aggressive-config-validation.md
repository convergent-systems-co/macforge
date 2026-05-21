# ADR 0019 — Aggressive config validation (static + runtime)

| Status      | Accepted                                                                  |
|-------------|---------------------------------------------------------------------------|
| Date        | 2026-05-21                                                                |
| Issue       | [#13](https://github.com/convergent-systems-co/macforge/issues/13)        |
| Supersedes  | —                                                                         |
| Refines     | [ADR-0014](0014-keychain-naming-convention.md) §"opt-out: `allow_nonstandard`" — formalizes when the validator honors it |

> Numbered 0019 because [ADR-0018](0018-peer-subtree-named-workstation.md) was claimed mid-flight by the workstation-peer rename. Issue #13 was opened before that rename landed, but the implementation PR for #13 lands after, so this ADR slots in here.

## Context

A user (the maintainer, on this repo) edited their global `macforge.yaml` so that `keychain.name` referenced one Apple Team ID while the top-level `team` field referenced a different one. The config layered cleanly: `config.Load` accepted it, every identity verb worked (those derive the keychain name from `team` via `keychain.DefaultName(cfg.Team, "signing")`), and `macforge keychain create` worked. Twenty minutes later, `macforge sign` failed with `MF-SIGN-002 — no identity matches team <X>` because the signing service was reading `cfg.Keychain.Name` directly while everything else was deriving from `cfg.Team`.

This is the **fail-late** anti-pattern: configuration disagreements are detected only at the point of use, with error messages that don't name the configuration field that's wrong. Generalized:

- `team` can be empty and `config.Load` doesn't notice.
- `keychain.name` can be malformed (doesn't match `macforge-<TEAM>-<PURPOSE>`) and `config.Load` doesn't notice.
- `keychain.name`'s `<TEAM>` segment can disagree with `cfg.team` and `config.Load` doesn't notice.
- `keychain.unlock` can reference an unset env var and only `keychain unlock` (deep in the call stack) notices.
- The configured keychain file can be missing on disk and only the first verb that touches `security find-identity` notices.

ADR-0015 established that the config lives in one global file with an optional project-local override. ADR-0014 established the keychain naming convention. Neither says how the config validator enforces them at load time.

## Decision

Three coordinated changes:

### 1. Strict static validation in `config.Load`

`internal/config/load.go::validate(cfg)` is extended with these rules. All return `MF-CONFIG-001` with hints pointing at the offending field; the existing version and `keychain.unlock`-prefix checks stay.

| Rule | Effect |
|------|--------|
| `team` non-empty | Empty `team` → reject. Hint points at `macforge apple init --team <TEAM>`. |
| `keychain.name` matches `macforge-<TEAM>-<PURPOSE>` regex (when set) | Delegates to `keychain.ValidateName`. Single source of truth for the convention. |
| `keychain.name`'s `<TEAM>` segment equals `cfg.team` (when set) | The #13 reproducer. Hint suggests `keychain.name: macforge-<cfg.team>-signing`. |
| `keychain.allow_nonstandard: true` | Escape hatch — skips BOTH `keychain.name` checks. Operator opt-in for unusual setups (e.g., a CI keychain that doesn't follow MacForge naming but is intentionally wired). |

### 2. `config.ResolveKeychainName(cfg *Config) string`

One helper, in `internal/config/resolve.go`, that every Apple verb consumes to read the keychain name:

- If `cfg.Keychain.Name` is set → returns it verbatim (the static validator already proved it well-formed or `allow_nonstandard`-permitted).
- Otherwise → returns `keychain.DefaultName(cfg.Team, "signing")`.

The resolver presumes a previously-validated config and does NOT re-validate. Validation is one-shot at Load; the resolver is just the runtime read path.

This commit closes the original #13 bug: `signing.Service.Sign` now reads `config.ResolveKeychainName(cfg)` instead of `cfg.Keychain.Name`. A `grep` for `cfg.Keychain.Name` in `internal/signing/` returns only test comments.

### 3. `macforge apple config validate` verb

A new verb under the `apple` subtree (per ADR-0017's subsystem-specific reading: this verb is Apple-specific, not cross-subsystem, since the rules it enforces are Apple-Developer-ID-shaped). It does NOT add a second `config.Load` invocation; `newRuntime("apple.config.validate", true)` loads as part of its normal cliRuntime construction. If Load fails, the verb surfaces the failure through `rt.emit` like any other verb.

If Load succeeds, the verb walks the loaded `*config.Config` field-by-field and emits per-rule check lines:

```
✓ config schema version = 1
✓ team = 57RGSN9YLM
✓ keychain.name = macforge-57RGSN9YLM-signing (canonical, matches team 57RGSN9YLM)
✓ keychain.unlock = env:MACFORGE_KEYCHAIN_PASSWORD
✗ $MACFORGE_KEYCHAIN_PASSWORD is unset
  hint: export MACFORGE_KEYCHAIN_PASSWORD="<your-keychain-password>"
✗ keychain macforge-57RGSN9YLM-signing not found on disk
  hint: run `macforge apple keychain create` to create it
✓ sign.hardened_runtime = true
✓ sign.timestamp = true
○ sign.entitlements = (unset; project-shaped, set in ./macforge.yaml)
○ notarize.asc_profile = macforge-prod (notarytool profile presence not checked this pass)

2 errors, 0 warnings
```

Markers:

| Marker | Meaning |
|--------|---------|
| `✓` | check passed |
| `✗` | check failed; verb exits non-zero |
| `○` | informational; the value is shown but the field isn't validated this pass (the verb makes no claim about its correctness) |
| `!` | warning (reserved; not used in this PR) |

Runtime checks (added on top of Load's static checks):

- **env-var presence** for `keychain.unlock: env:VAR`. Uses `os.LookupEnv` — never reads the value, only checks presence (`Common.md §4.1`).
- **keychain file reachability** for the resolved keychain name. Delegates to `internal/apple/security.Client.HasKeychain` (`security show-keychain-info <name>`). When `security` itself is unavailable (non-darwin, no PATH), this surfaces as a red check rather than crashing — keeps CI Linux runs honest.

`keyring:` references emit a `○` (not yet validated this pass; integration with `internal/keychain.Secret` is a follow-up).

### Audit + JSON envelope

- Audit `command` field: `apple.config.validate` (per ADR-0017).
- JSON envelope `schema`: `macforge.v1.apple.config.validate`.

Both are locked by tests in `cmd/macforge/config_cmd_test.go`.

### Out of scope (deferred)

- **Per-verb runtime preflight** (issue #13's section "3. Per-verb runtime preflight"). The validate verb runs the preflight as a one-shot; pushing it into every keychain-touching verb is a separate concern.
- **Unknown-key warnings** at the YAML level (#13's section "5. warn on unknown top-level keys"). Needs viper API research; not blocking #13.
- **`cmd/macforge/keychain_cmd.go` and `identity_cmd.go` sweep.** Both already derive the keychain name via `keychain.DefaultName(rt.cfg.Team, "signing")` — structurally correct but not routed through the resolver. A follow-up commit can route them; doesn't change behavior for the closed bug.

## Consequences

### Positive

- **Fail-fast on inconsistent config.** The #13 reproducer is rejected at Load with a hint pointing at the exact field. Twenty-minute discovery loops collapse to seconds.
- **One source of truth for the keychain name.** `ResolveKeychainName` is the only sanctioned read path; the static validator is the only sanctioned write path.
- **Operator-visible self-check.** `macforge apple config validate` answers "is my setup ready?" without sacrificing a real artifact to find out.
- **The escape hatch is explicit.** `keychain.allow_nonstandard: true` is the single, documented opt-out; the schema field already existed (introduced for `apple keychain create --allow-nonstandard`), this ADR just gives it teeth in `Load`.

### Negative

- **Existing configs without `team` now fail.** Anyone running v0.1-dev with an empty `team` field gets a hard error at next `macforge` invocation. v0.1 has one user, this is acceptable; v1.0 would require a deprecation cycle.
- **Existing configs with mismatched `keychain.name` now fail.** Same caveat. The hint tells the operator exactly what to change.
- **Linux/non-macOS CI cannot get an all-green `apple config validate`.** The `security show-keychain-info` runtime check surfaces as red. This is acceptable — that verb is operator-tooling, not CI tooling. CI doesn't need to "validate the config" — it `go test`s.

### Neutral

- **The resolver does not re-validate.** Two layers (validator at Load, reader at runtime) keep concerns separate; the resolver is therefore safe to call on configs that originated outside `Load` (e.g., a literal-built `*Config` in a test).
- **The internal regex `keychain.nameRE` is re-used through `keychain.ValidateName`.** No duplication; if the convention ever shifts, one edit propagates.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Warn instead of error on `keychain.name`/`team` mismatch** | Backward-compatible | The original bug is a silent override that produces a misleading failure 20 min later. Warnings don't gate misconfig. | Rejected — the maintainer (issue #13 reporter) explicitly chose error over warn |
| **Push validation into every verb (no central validator)** | No new logic at Load | Each verb re-implements the same checks; drift is inevitable; the "validate" verb has to be built anyway | Rejected |
| **Number the ADR 0018** | Sequential within the issue | Already taken by `0018-peer-subtree-named-workstation.md` | Rejected; 0019 |
| **No `ResolveKeychainName` helper; just fix `signing.go`'s read** | Smaller diff | The bug recurs at every new `cfg.Keychain.Name` read site (verify, future package/notarize verbs) | Rejected — the helper is structural, not cosmetic |
| **Strict regex inline in `validate()`** (duplicate of `keychain.nameRE`) | Removes `keychain` import from `config` | Two regex sources drift apart. Already-tested `keychain.ValidateName` is the canonical owner. | Rejected — import edge is one-way (`config` → `keychain`), no cycle |
| This ADR | Strict + central + one helper + one verb | Three coordinated changes land in one PR | **Chosen** |

## Implementation

- `internal/config/resolve.go` (new) — `ResolveKeychainName(cfg)`.
- `internal/config/load.go` — three new validate rules; reuses `keychain.ValidateName`.
- `internal/signing/signing.go` — replace direct `cfg.Keychain.Name` reads with `config.ResolveKeychainName(cfg)`.
- `cmd/macforge/config_cmd.go` (new) — `newConfigCmd()` + `newConfigValidateCmd()` + the result type.
- `cmd/macforge/apple_cmd.go` — register `newConfigCmd()`.
- Tests in `internal/config/`, `internal/signing/` (integration), `cmd/macforge/`.
- New FakeRunner fixtures `testdata/security/find-identity-one-derived/` and `testdata/codesign/sign-derived-keychain/` for the resolver integration tests.

## Links

- Issue: [#13](https://github.com/convergent-systems-co/macforge/issues/13)
- Related: [ADR-0014](0014-keychain-naming-convention.md), [ADR-0015](0015-single-global-config-xdg.md), [ADR-0017](0017-apple-command-namespace.md).
