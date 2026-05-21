# Testing conventions

macheim's test suite lives alongside the code (`*_test.go` in each package).
This document describes the patterns the codebase uses and the gate that
separates unit from integration tests.

## Locations

- Unit tests: `internal/<pkg>/*_test.go`. Always run by `make test`.
- Fixture files: `internal/<pkg>/testdata/`. Standard Go convention — `go test` excludes `testdata/` from compilation.
- Integration tests: gated by `MACHEIM_INTEGRATION=1` (see below).

## Patterns

### Table-driven tests

Default to the table pattern. See `internal/config/runtime_test.go::TestRuntime_ResolveRepoPath` for a worked example.

### Seam injection for OS-touching code

Functions that exec binaries, hit the filesystem, or read the env take an
optional seam struct so tests substitute fakes. See `internal/doctor/checks.go`
for the established pattern.

### Test parallelism

Use `t.Parallel()` for unit tests with no shared state. Tests that call
`t.Setenv` MUST NOT call `t.Parallel()` — the env helper panics in parallel
contexts. When a function calls `t.Setenv`, neither it nor its subtests are
parallel-safe.

## Integration tests

Slow, network-dependent, or macOS-specific tests gate on an env var:

```go
func TestInstallBrew(t *testing.T) {
    if os.Getenv("MACHEIM_INTEGRATION") != "1" {
        t.Skip("integration test; set MACHEIM_INTEGRATION=1 to run")
    }
    // ...
}
```

CI (`.github/workflows/ci.yml`) runs unit tests only; integration tests run
locally on demand.
