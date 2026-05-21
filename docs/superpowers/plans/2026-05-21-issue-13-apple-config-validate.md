# Plan: #13 — Aggressive config validation

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate the class of bugs where stale or inconsistent config (e.g., `keychain.name` referencing the wrong team) silently passes `config.Load` and only surfaces 20 minutes later inside a verb. Push the validation up to `config.Load` for static checks, introduce a single resolver helper for all keychain-name reads, and ship a `macforge apple config validate` verb that runs static + runtime checks at once.

**Architecture:** Reuse the existing `config.Load → validate(cfg)` plumbing, the `keychain` package's `nameRE` + `DefaultName`, the `apple/security` `HasKeychain` runtime check, the standard `cliRuntime` + `Outputter` plumbing, and the existing `mferrors.CodeConfigInvalid` (MF-CONFIG-001) error code. No new packages; one new file per OWNS list.

**Issue:** [#13](https://github.com/convergent-systems-co/macforge/issues/13)
**Branch:** `fix/13-apple-config-validate`
**Base:** `main`

## Acceptance Criteria

- [ ] strict `config.Load` rejects missing `team` with `MF-CONFIG-001` + hint
- [ ] strict `config.Load` rejects `keychain.name` that doesn't match `macforge-<TEAM>-<PURPOSE>` regex when `allow_nonstandard: false`
- [ ] strict `config.Load` rejects `keychain.name` whose `<TEAM>` segment doesn't equal `cfg.team` (the original bug from #13)
- [ ] strict `config.Load` honors `keychain.allow_nonstandard: true` and skips the team-consistency check
- [ ] `config.ResolveKeychainName(cfg)` returns expected values in 3 cases (set+valid, set+nonstandard+allow=true, unset → derive from team)
- [ ] `signing.Service.Sign` consumes `ResolveKeychainName` (no direct `cfg.Keychain.Name` reads anywhere in the codebase except inside the resolver itself)
- [ ] `macforge apple config validate` runs static+runtime checks, exits non-zero on any red, prints `✓` / `✗` / `○` per check
- [ ] JSON envelope schema `macforge.v1.apple.config.validate`
- [ ] Audit `command` field reads `apple.config.validate`
- [ ] ADR-0018 authored and merged in same PR
- [ ] All existing tests stay green; new tests cover every new acceptance criterion above

## Seed Commit (lands first, all sub-tasks depend on it)

The seed introduces the `ResolveKeychainName` helper as a self-contained new file. T1 then strengthens `validate()` (uses the same helper for the team-consistency check); T2 and T3 then depend on both.

**OWNS:**
- `internal/config/resolve.go` (NEW) — `ResolveKeychainName(cfg *Config) string`
- `internal/config/resolve_test.go` (NEW) — table-driven unit test for the 3 cases

**Notes for the seed:**
- The function is intentionally permissive: when `Name` is non-empty it returns `Name` verbatim. The *validation* that `Name` is well-formed and consistent with `Team` is the job of `validate()` (T1). The resolver presumes a previously-validated config and never re-validates.
- When `Name` is empty it returns `keychain.DefaultName(cfg.Team, "signing")`.
- Import path: `import "github.com/convergent-systems-co/macforge/internal/keychain"`.

## Sub-tasks (parallel-safe after seed lands)

### T1 — Strict static validation in Load

**OWNS:**
- `internal/config/load.go` — extended `validate()` function
- `internal/config/load_test.go` — new tests for missing team, keychain.name regex mismatch, team-segment mismatch, allow_nonstandard escape

**DEPENDS_ON:** seed

**ACCEPTANCE:**
- 4+ new test cases pass:
  - `TestLoad_RejectsMissingTeam` — `team:` empty → error with `MF-CONFIG-001`
  - `TestLoad_RejectsBadKeychainNameRegex` — `keychain.name: my-weird-name` → error
  - `TestLoad_RejectsTeamMismatch` — `team: XYZ1234567`, `keychain.name: macforge-ABC9876543-signing` → error pointing at the mismatch
  - `TestLoad_AllowsNonstandardWhenOptedIn` — same as above + `keychain.allow_nonstandard: true` → no error
- All existing tests in `load_test.go` stay green (note: `TestLoad_EnvOverridesBoth` and `TestLoad_DefaultGlobalHonorsXDG` and `TestLoad_RejectsInlinePassword` need to keep working; verify they don't rely on an empty `team`)

**Implementation hint:**
```go
// in validate():
if strings.TrimSpace(cfg.Team) == "" {
    return mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.validate",
        "team is required",
        mferrors.WithHint("Set top-level `team: <APPLE_TEAM_ID>` in macforge.yaml"))
}
if cfg.Keychain.Name != "" && !cfg.Keychain.AllowNonstandard {
    // 1. regex check against macforge-<TEAM>-<PURPOSE>
    // 2. extract <TEAM> segment, compare to cfg.Team
    // both failures return MF-CONFIG-001 with a hint pointing at the field
}
```
- The regex is the same one `keychain.nameRE` enforces. Either re-import + use `keychain.ValidateName(cfg.Keychain.Name)` (recommended — single source of truth) or duplicate a stricter inline check. Prefer the re-import.
- For the team-segment check, split on `-`: `parts := strings.Split(name, "-")`; team is `parts[1]` after the literal `macforge`. Compare to `cfg.Team`.

### T2 — `apple config validate` cobra verb

**OWNS:**
- `cmd/macforge/config_cmd.go` (NEW) — `newConfigCmd()` + `newConfigValidateCmd()` + the result type
- `cmd/macforge/config_cmd_test.go` (NEW) — happy + sad paths via cobra Execute
- `cmd/macforge/apple_cmd.go` — register `newConfigCmd()` in the `AddCommand(...)` list

**DEPENDS_ON:** seed, T1 (for the strict validate logic)

**ACCEPTANCE:**
- `macforge apple config validate` builds and runs
- Output shape matches issue body sample (`✓` / `✗` / `○` lines, ordered per-check, hints under reds)
- Exit code: 0 when all green, non-zero when any red (`MF-CONFIG-001`)
- `--output json` produces an envelope with `schema = macforge.v1.apple.config.validate`
- Audit `command` field = `apple.config.validate`
- Test cases:
  - happy path (well-formed config) → exit 0, all `✓`
  - bad keychain.name/team mismatch → exit non-zero, has `✗ keychain.name ... — does not match team ...`
  - missing `MACFORGE_KEYCHAIN_PASSWORD` env when `keychain.unlock: env:MACFORGE_KEYCHAIN_PASSWORD` → exit non-zero, has `✗ $MACFORGE_KEYCHAIN_PASSWORD is unset`

**Implementation hint:**
- The verb does NOT call `config.Load`-from-scratch a second time — `newRuntime("apple.config.validate", true)` already loads & validates. If Load rejects on a hard structural error (parse failure, missing required field), the verb naturally surfaces that via `rt.emit`. The validate verb's *additional* runtime checks (env-var reachability, keychain file presence) run AFTER Load succeeds.
- For runtime keychain file check, use `internal/apple/security` `Client.HasKeychain(ctx, name)` — needs an `apple.Runner` (use `newRunnerWithAudit(rt)`).
- For env-var reachability, parse `cfg.Keychain.Unlock`: if it starts with `env:`, strip the prefix and call `os.LookupEnv`. `keyring:` references are reported as `○` informational for now (no `keyring` lookup wired yet — that's a follow-up).
- Result type:
  ```go
  type configValidateResult struct {
      Checks []validateCheck `json:"checks"`
      Errors int             `json:"errors"`
      Warnings int           `json:"warnings"`
  }
  type validateCheck struct {
      Status string `json:"status"` // "ok" | "fail" | "info"
      Label  string `json:"label"`
      Hint   string `json:"hint,omitempty"`
  }
  func (r configValidateResult) SchemaName() string { return "macforge.v1.apple.config.validate" }
  func (r configValidateResult) HumanLines() []string { /* ✓/✗/○ formatting */ }
  ```
- The RunE returns a non-nil `*mferrors.Error` when `Errors > 0` so the renderer exits non-zero. But the *result* (the check list) still needs to be printed. Use `rt.emit(result, runErr)` — the existing emit prints the failure envelope. For the success path (`Errors == 0`), call `rt.emit(result, nil)`.
- If you need finer-grained "print the list BEFORE returning the err" behavior, look at how `sign_cmd.go` emits even on error: it always calls `rt.emit(result, signErr)` and the renderer takes care of both paths.

### T3 — Wire signing to ResolveKeychainName

**OWNS:**
- `internal/signing/signing.go` — replace direct `cfg.Keychain.Name` reads with `config.ResolveKeychainName(cfg)`
- `internal/signing/signing_test.go` — add tests verifying resolver is honored

**DEPENDS_ON:** seed

**ACCEPTANCE:**
- The two `cfg.Keychain.Name` reads in `Sign` (line 41 `FindIdentities` arg, line 53 `SignOptions.Keychain`) both go through `config.ResolveKeychainName(cfg)`.
- `grep -n "cfg.Keychain.Name" internal/signing` returns nothing.
- New test cases:
  - `TestSign_UsesResolvedKeychainName_FromEmptyName` — config with empty `Keychain.Name` and team `XYZ1234567` → `Sign` calls security with `macforge-XYZ1234567-signing`
  - `TestSign_UsesResolvedKeychainName_FromExplicitName` — config with `Keychain.Name: macforge-XYZ1234567-custom` + `AllowNonstandard: false` (still valid because regex matches and team matches) → `Sign` calls security with `macforge-XYZ1234567-custom`
- Existing tests stay green.

**Implementation hint:**
- The existing unit tests use the package-internal `selectIdentityByTeam` helper. To test the `Sign` flow you'll need a `apple.FakeRunner` (already used in `signing_integration_test.go`) — check that file for the shape. If FakeRunner-based tests are integration-tagged, mark the new tests `//go:build integration` to match.
- **Strict scope:** do NOT touch `cmd/macforge/keychain_cmd.go` or `cmd/macforge/identity_cmd.go`. Those use `keychain.DefaultName(rt.cfg.Team, "signing")` directly. T3 is signing only. If the larger "no direct reads anywhere" sweep needs more files, T3 calls it out as a follow-up; it does not silently broaden scope (Code.md §11.4 bug-report-separately rule).

### T4 — Docs (ADR-0018 + README + CHANGELOG)

**OWNS:**
- `docs/adr/0018-aggressive-config-validation.md` (NEW)
- `README.md` — add `apple config validate` to the verb table + a brief mention of validation in the Configuration layering section
- `CHANGELOG.md` — entry under `[Unreleased]` → `### Added` for the verb and the resolver helper; `### Fixed` for the stale-keychain-name detection (Closes #13)

**DEPENDS_ON:** seed

**ACCEPTANCE:**
- ADR follows MADR shape (Status / Date / Context / Decision / Consequences / Alternatives / Implementation / Links — see ADR-0017 for the canonical shape)
- README table gains a `macforge apple config validate` row marked ✓
- CHANGELOG `[Unreleased]` mentions the verb and the resolver + the bug-fix

## Test Strategy

- **Tier 1 (unit, no tags):** all new tests in `internal/config/` (resolver + strict Load) and `cmd/macforge/config_cmd_test.go`.
- **Tier 2 (integration, `//go:build integration`):** signing-resolver tests in `internal/signing/` using `apple.FakeRunner` (matches existing `signing_integration_test.go` pattern). The `apple config validate` runtime check tests can use FakeRunner or `os.Setenv` (env-var checks need real env, not FakeRunner).
- **Adversarial focus:**
  - `validate()` MUST NOT fail-open on missing team. A config with `team: ""` and nothing else should return an error, not silently succeed.
  - Resolver MUST NOT silently prefer a wrong `Name` over the derived default — but this is the validator's job; the resolver presumes a validated config.
  - Exit codes MUST match: 0 on all-green, non-zero on any red.

## Notes for coders

- ALL command strings to `newRuntime(...)` use the `apple.config.validate` shape (per ADR-0017). DO NOT use `config.validate` (no `apple.` prefix).
- Result types' `SchemaName()` return `macforge.v1.apple.config.validate`.
- Existing `cmd/macforge/runtime.go` `cliRuntime` already provides `Output`/`emit` plumbing — reuse it.
- For runtime checks needing the security CLI (keychain file presence), use the existing `internal/apple/security` package's `HasKeychain` method on the `Client`.
- Bug-fix-separately rule (Code.md §11.4): if anyone hits unexpected behavior (e.g., a third place that reads `cfg.Keychain.Name` and isn't in their OWNS), escalate to TL-1; do not silently broaden scope.
- One logical change per commit (Code.md §11.2). Each sub-task = one commit. Conventional Commits prefix: `feat:` for T1/T2/T3, `docs:` for T4, `feat:` for the seed (it adds a new exported function).
