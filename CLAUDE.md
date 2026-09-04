# CLAUDE.md

## Project Overview

Go MCP server for Todoist. Communicates via stdio (stdout = MCP protocol, stderr = logs).
Built with [go-sdk v1.7.0](https://github.com/modelcontextprotocol/go-sdk).

## Architecture

Three-layer design, all under `internal/`:

```
main.go                → Entry point, slog setup, server bootstrap
internal/tools/        → MCP tool handlers (register.go wires all tools)
internal/todoist/      → HTTP API client (no MCP awareness)
internal/models/       → Shared data types
```

- `tools/` depends on `todoist/` and `models/`
- `todoist/` depends only on `models/`
- `models/` has zero dependencies

## Commands

```bash
make build       # Build to build/mcp-todoist
make test        # Run tests with -race
make lint        # golangci-lint
make check       # fmt + vet + lint + test
make coverage    # Tests with HTML coverage report
```

## Conventions

### Logging

Uses `log/slog` with `JSONHandler` writing to `os.Stderr`. No `log.Fatal` — use `slog.Error` + `os.Exit(1)`.
Lowercase log messages. Pass errors as structured attributes: `slog.Error("msg", "error", err)`.

### Error Handling

- API client methods return `(T, error)` — no panics
- Tool handlers return MCP error results, never crash the server

### Testing

- Tests use `net/http/httptest` for API client testing
- Test assertions use `testing.T.Fatal`/`Fatalf` — no external test libraries
- All tests run with `-race` flag

## CI

GitHub Actions runs on every push/PR to main:
- `go build`, `go test -race`, `go vet`, `staticcheck v0.8.1`, `go fix -diff`
- `golangci-lint v2.13.2` (separate job)

Both jobs read their Go version from `go.mod` via `go-version-file`, so the Go version lives
in one place. `setup-go` v6+ exports `GOTOOLCHAIN=local`, so tool versions are pinned rather
than `@latest`.

Raising the `go` directive is not sufficient on its own: both tool pins have a floor tied to
it, because a linter built with an older Go than the module targets refuses to run. Check
them in the same commit as any Go bump:

- `GOLANGCI_LINT_VERSION` in the `Makefile`, which is canonical. CI reads it via
  `make print-golangci-lint-version`, so `make lint` and the Lint job cannot drift apart.
- `STATICCHECK_VERSION` in the workflow.

Job names carry no version, keeping branch-protection contexts stable across upgrades.
Module and build caching comes from `setup-go` (on by default); do not add an
`actions/cache` step over `GOMODCACHE`/`GOCACHE`, as it collides with that on restore.

## Environment

- Requires `TODOIST_API_TOKEN` env var
- Go 1.27.1+
