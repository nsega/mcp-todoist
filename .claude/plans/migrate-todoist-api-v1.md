# Migrate mcp-todoist from Todoist REST API v2 to API v1

## Context
The Todoist REST API v2 has been deprecated (returns 410 Gone). The base URL was already changed to `https://api.todoist.com/api/v1`, but the v1 API has two breaking differences:
1. **List endpoints wrap responses** in `{"results": [...], "next_cursor": "..."}` instead of bare arrays
2. **Field names changed** in JSON responses (e.g., `is_completed` → `checked`, `created_at` → `added_at`)

All list-based tools are currently broken. Single-resource tools work but silently return zero-values for renamed fields.

## Step 1: Add `PaginatedResponse[T]` generic type
**File:** `internal/todoist/client.go`

Add after the `Client` struct:
```go
type PaginatedResponse[T any] struct {
    Results    []T    `json:"results"`
    NextCursor string `json:"next_cursor"`
}
```

## Step 2: Update model JSON tags

Keep all Go field names unchanged (prevents tool handler changes). Only change JSON tags.

**`internal/models/task.go`** — rename tags:
- `"order"` → `"child_order"` | `"comment_count"` → `"note_count"` | `"is_completed"` → `"checked"`
- `"creator_id"` → `"added_by_uid"` | `"assignee_id"` → `"responsible_uid"` | `"created_at"` → `"added_at"`
- Add new fields: `UserID`, `AssignedByUID`, `UpdatedAt`, `CompletedAt`, `IsDeleted`

**`internal/models/project.go`** — rename tags:
- `"order"` → `"child_order"` | `"is_inbox_project"` → `"inbox_project"`
- Remove `CommentCount` (not in v1), keep `URL`/`IsTeamInbox` with omitempty
- Add: `CreatorUID`, `CreatedAt`, `UpdatedAt`, `IsArchived`, `IsDeleted`, `Description`, `CanAssignTasks`

**`internal/models/section.go`** — rename tags:
- `"order"` → `"section_order"`
- Add: `UserID`, `AddedAt`, `UpdatedAt`, `IsArchived`, `IsDeleted`, `IsCollapsed`

**`internal/models/comment.go`** — add fields: `PostedUID`, `IsDeleted`

**`internal/models/label.go`** — no changes needed

## Step 3: Update list methods to unwrap paginated response

Same pattern in all 5 files — change `json.Unmarshal(data, &slice)` to unmarshal into `PaginatedResponse[T]` and return `.Results`:

- `internal/todoist/tasks.go` — `GetTasks()`
- `internal/todoist/projects.go` — `GetProjects()`
- `internal/todoist/sections.go` — `GetSections()`
- `internal/todoist/comments.go` — `GetComments()`
- `internal/todoist/labels.go` — `GetLabels()`

## Step 4: Fix request body key in sections tool

**`internal/tools/sections.go`** — change `body["order"]` → `body["section_order"]` in CreateSection

## Step 5: Update all test fixtures

**`internal/todoist/*_test.go`** (5 files) and **`internal/tools/tools_test.go`**:
- Wrap all list mock responses in `{"results":[...],"next_cursor":""}`
- Update JSON field names to match v1 (e.g., `"is_inbox_project"` → `"inbox_project"`, `"created_at"` → `"added_at"`)

## Step 6: Update README

Already partially done. Verify `README.md` references are correct.

## Verification
1. `go test ./...` — all unit tests pass
2. `make build` — binary compiles
3. Restart MCP server in Claude Code, run `/mcp` to verify connected
4. Call `todoist_get_projects`, `todoist_get_tasks`, `todoist_get_labels` to verify live API works
