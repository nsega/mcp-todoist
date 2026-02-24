# Session Context

Session ID: eac1bebd-3aa4-4d57-b5df-dd1ada4cc4da
Commit Message: If necessary, update CLAUDE.md and README.md to improve the upcoming usa

## Prompts

### Prompt 1

if necessary, update CLAUDE.md and README.md to improve the upcoming usage.

## Summary

Created `CLAUDE.md` with:

- **Project overview** — what it is, how stdio is used
- **Architecture** — three-layer design with dependency direction
- **Commands** — build, test, lint, check, coverage
- **Conventions** — slog logging (JSONHandler to stderr, lowercase messages, structured attributes), error handling patterns, testing approach
- **CI** — what GitHub Actions checks run
- **Environment** — required env var, Go version

**README.md** — no changes needed. It's already comprehensive and nothing became inaccurate from the slog migration (logging is an internal detail, not user-facing).
