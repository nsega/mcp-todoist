# Staged Technical Debt Refactor: Design

Date: 2026-07-14
Status: Approved pending user review

## Background

An audit of the codebase (~1,900 lines of source, ~1,600 of tests, all green) found the
three-layer architecture healthy but identified six concrete debt items. This design
addresses all six in five risk-ordered stages. A sweeping restructure was considered and
rejected: the architecture does not need it.

The debt items:

1. `context.Context` is dropped: handlers receive `ctx` but `internal/todoist/client.go`
   uses `http.NewRequest`, so cancellation and timeouts never propagate.
2. Stringly-typed API bodies: handlers build `map[string]any` instead of typed request
   structs, so field names are not compile-checked.
3. Handler repetition: the resolve-ID, not-found, label-fallback flow is copy-pasted
   across five task tools, and every tool declares an identical `{Success, Message}`
   output struct.
4. Update semantics: empty-string checks make it impossible to clear a description,
   due date, or labels via `todoist_update_task`.
5. No rate-limit or retry handling in the HTTP client.
6. Repo hygiene: an 11MB compiled binary and a `.go.mod.~undo-tree~` file are committed
   at the repo root; CLAUDE.md and README say go-sdk v1.4.0 while go.mod is on v1.6.1.

## Staging strategy

Each stage is one PR. Stages are ordered so that signatures change exactly once
(context and typed requests land together) and behavior changes come last, isolated
from mechanical refactors.

| Stage | Content | Behavior change |
|-------|---------|-----------------|
| 1 | Repo hygiene | No |
| 2 | Client API reshape: context + typed requests | No |
| 3 | Handler dedup | No |
| 4 | Update semantics: field clearing | Yes |
| 5 | Retry and rate-limit handling | Yes |

## Stage 1: Repo hygiene

- `git rm --cached` the `mcp-todoist` binary and `.go.mod.~undo-tree~`.
- Add both patterns to `.gitignore`.
- Update CLAUDE.md and README go-sdk references from v1.4.0 to v1.6.1.
- No code or test changes.

## Stage 2: Client API reshape

Every `Client` method gains `ctx context.Context` as its first parameter, threaded down
to `do()`, which switches to `http.NewRequestWithContext`. Tool handlers already receive
`ctx` from the MCP SDK and pass it through.

The `map[string]any` bodies become typed request structs living in `internal/todoist`
next to the methods that use them. They are API-shaped, not domain models, so they do
not belong in `models/`.

```go
type CreateTaskRequest struct {
    Content     string   `json:"content"`
    Description *string  `json:"description,omitempty"`
    DueString   *string  `json:"due_string,omitempty"`
    Priority    *int     `json:"priority,omitempty"`
    ProjectID   *string  `json:"project_id,omitempty"`
    SectionID   *string  `json:"section_id,omitempty"`
    ParentID    *string  `json:"parent_id,omitempty"`
    Labels      []string `json:"labels,omitempty"`
    AssigneeID  *string  `json:"assignee_id,omitempty"`
}
```

Pointer fields give three states: `nil` means absent, pointer-to-zero means "explicitly
send the zero value". Stage 4 depends on this distinction. `CreateTask`, `UpdateTask`,
and `MoveTask` take these structs; handlers build them instead of maps. Existing
httptest tests update mechanically. This is the largest diff of the plan but is
compiler-verified end to end.

## Stage 3: Handler dedup

Two extractions in `internal/tools`, both pure refactors:

1. One shared `ActionOutput{Success bool; Message string}` type replaces the per-tool
   output structs whose JSON schemas are already identical.
2. The repeated resolve, not-found, act, label-fallback flow in `tasks.go` (update,
   delete, complete, reopen) collapses into a helper that takes the action as a
   closure, shaped like
   `runTaskAction(ctx, c, id, name, verb string, action func(ctx, id) error)`.
   Update keeps its custom body-building but reuses the resolve and label parts.

Wire behavior and user-visible messages stay byte-identical, verified by the existing
`tools_test.go` assertions.

## Stage 4: Update semantics (behavior change)

`UpdateTaskInput` fields that can be cleared become pointers: `Description`,
`DueString`, `Labels`. Semantics:

- Field absent in the tool call: not touched (as today).
- `description: ""`: sends `""`, clearing the description.
- `due_string: "no date"` passes through (Todoist's native clear syntax); `""` is
  also mapped to `"no date"` so MCP clients do not need Todoist trivia.
- `labels: []`: sends `[]`, removing all labels. Note: `omitempty` on a plain
  `[]string` would drop an empty slice, so `UpdateTaskRequest` declares
  `Labels *[]string` (nil means absent, pointer to empty slice means clear).

The jsonschema descriptions document the clearing behavior so MCP clients can discover
it. New tests cover each clear path.

## Stage 5: Retry and rate-limit handling (behavior change)

Implemented inside `client.do` only; nothing above it changes.

- On 429: read `Retry-After` (seconds), wait, retry. Cap the wait at 30 seconds; if
  the header is missing or unparseable, fall back to the backoff schedule.
- On 5xx: retry with fixed backoff (250ms, then 1s), max 3 attempts total.
- Waits select on `ctx.Done()` so cancellation interrupts a sleep (this is why
  stage 2 lands first).
- Other 4xx statuses fail immediately, as today.

## Regression verification

Regression safety is gated on tests and the GitHub Actions CI jobs. For every stage:

1. **Local gate:** `make check` (fmt, vet, golangci-lint, `go test -race`) must pass
   before the PR is opened.
2. **CI gate:** both GitHub Actions jobs must be green on the PR before merge:
   - "Build and Test": `go build`, `go test -race` with coverage upload,
     `go vet`, `staticcheck`, `go fix -diff`.
   - "golangci-lint" (v2.10.1).
3. **No-behavior-change stages (1, 2, 3):** the existing test suite must pass
   unmodified in spirit. Stage 2 updates test call sites mechanically (adding `ctx`,
   swapping maps for structs) but every assertion on request paths, bodies, and
   response handling is preserved. Stage 3 changes no test assertions at all.
4. **Behavior-change stages (4, 5):** new table-driven httptest cases cover each new
   path before merge:
   - Stage 4: clearing description, due date (both `"no date"` and `""` spellings),
     and labels; absent fields remain untouched.
   - Stage 5: 429-then-200 succeeds, persistent 500 fails after 3 attempts,
     `Retry-After` is honored and capped, context cancellation during backoff
     returns promptly.
5. Coverage is uploaded to Codecov by CI on every PR; stages 4 and 5 must not
   reduce total coverage.

## Out of scope

- New tools or new Todoist API surface.
- Changes to the MCP transport, logging setup, or `models/` package.
- New dependencies. The retry logic uses only the standard library.

## Risks

- Stage 2 touches every client method and call site. Mitigation: it is purely
  signature-level, the compiler verifies every call site, and the unchanged test
  assertions verify the wire format.
- Stage 4 changes how existing MCP clients' calls are interpreted only in cases that
  previously did nothing (sending an explicit empty value). The absent-field path is
  unchanged, so current callers are unaffected.
