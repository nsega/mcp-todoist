# Session Context

Session ID: eac1bebd-3aa4-4d57-b5df-dd1ada4cc4da
Commit Message: Migrate logging to log/slog by refering how to use log/slog via context7

## Prompts

### Prompt 1

Migrate logging to log/slog by refering how to use log/slog via context7

## Summary

Build succeeds and all tests pass. Here's a summary of the migration:

**Changes in `main.go`:**

| Before | After |
|--------|-------|
| `import "log"` + `"fmt"` | `import "log/slog"` |
| `log.Fatal("Error: ...")` | `slog.Error("...") + os.Exit(1)` |
| `fmt.Fprintf(os.Stderr, "...starting...\n")` | `slog.Info("Todoist MCP Server starting")` |
| `log.Fatalf("Server error: %v", err)` | `slog.Error("server error", "error", err) + os.Exit(1)` |

**Key decisions:**
- **JSONHandler to stderr** — structured JSON output is ideal for an MCP server since stdout is reserved for the MCP protocol
- **`slog.Error` + `os.Exit(1)`** instead of `log.Fatal` — `log/slog` intentionally doesn't provide `Fatal`/`Panic` methods; explicit exit is the recommended pattern
- **Structured error attribute** — `"error", err` passes the error as a structured key-value pair rather than interpolating it into the message string
- **No changes to `gtd.go`** — those `fmt.Fprintf` calls build report strings for tool output, not logging
