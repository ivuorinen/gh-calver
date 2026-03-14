# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

gh-calver is a GitHub CLI extension for Calendar Versioning (CalVer) release management. It produces versions in `YYYY.MM.MICRO` format where MICRO resets each month and increments on collision. Written in pure Go with zero external dependencies.

## Commands

```bash
make test          # run all tests
make test-race     # tests with race detector
make test-verbose  # verbose test output
make test-cov      # tests with coverage summary
make lint          # go vet + staticcheck
make build         # build for current platform
make build-all     # cross-compile all platforms
make install       # build + install as gh extension
make uninstall     # remove locally installed extension
make clean         # remove build artifacts
make release       # tag + push + goreleaser (clean main only)
make all           # lint, test, build
```

Run a single test:
```bash
go test -run TestFunctionName ./cmd/
go test -run TestFunctionName ./internal/calver/
go test -run TestFunctionName ./internal/git/
```

## Architecture

Three layers with clear boundaries:

- **`internal/calver/`** — Pure version logic: parsing, sorting, calculating next version. No I/O. The `Version` struct holds Year, Month, Micro, Prefix fields.
- **`internal/git/`** — Git/GitHub operations behind a `Client` interface (`client.go`). Real implementation in `git.go` wraps `git` and `gh` CLI commands.
- **`cmd/`** — CLI commands (current, next, release, bump, list). Each command gets its own `flag.FlagSet`. Global state (`ops` client, `nowFn` time function) lives in `deps.go`.

Entry point: `main.go` → `cmd.Execute()` (dispatcher in `root.go`).

Release builds are configured via `.goreleaser.yaml` and triggered by `make release`.

## Testing Patterns

- Tests use Go's standard `testing` package with table-driven tests
- Commands are tested via a `mockOps` struct in `cmd/testhelpers_test.go` that implements `git.Client`
- Time is injected via `nowFn` in `deps.go`; tests set it to a fixed value (`2025-03-15`)
- Stdout capture helper: `captureStdout()` in testhelpers
- The release command tests verify atomic behavior: tag creation → push → release, with cleanup on failure

## Key Conventions

- No external Go dependencies — stdlib only
- Optional `v` prefix on versions controlled by `--prefix v` flag
- `flag.ErrHelp` is treated as success (exit 0), not an error
- Errors are wrapped with context using `fmt.Errorf("...: %w", err)`
