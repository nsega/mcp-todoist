# CLAUDE.md

## Project Overview

Go MCP server for Todoist. Communicates via stdio (stdout = MCP protocol, stderr = logs).
Built with [go-sdk v1.4.0](https://github.com/modelcontextprotocol/go-sdk).

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
- `go build`, `go test -race`, `go vet`, `staticcheck`, `go fix -diff`
- `golangci-lint v2.10.1` (separate job)

## Environment

- Requires `TODOIST_API_TOKEN` env var
- Go 1.26.2+
