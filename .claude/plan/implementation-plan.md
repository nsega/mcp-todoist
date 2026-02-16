# Plan: Extend mcp-todoist for Full GTD+Zettelkasten Workflow

## Context

The current mcp-todoist MCP server has only 5 basic task CRUD tools in a single 528-line `main.go`. The user's GTD+Zettelkasten workflow (Todoist + Obsidian + Notebook LM) requires full Todoist API coverage: projects, sections, labels, comments, and higher-level GTD operations like inbox processing and weekly reviews. Additionally, [entire.io](https://docs.entire.io) will be introduced to capture AI agent sessions alongside git commits.

---

## Phase 0: Entire.io Setup

1. Install entire CLI: `brew install entireio/tap/entire`
2. Enable in project: `cd /Users/naokisega/src/github.com/nsega/mcp-todoist && entire enable`
3. Verify git hooks: `ls -la .git/hooks/`
4. Add `.entire/` to `.gitignore` if any local state files are created

---

## Phase 1: Restructure to Modular Architecture (zero behavior change)

Refactor the monolithic `main.go` into a proper Go package layout:

```
mcp-todoist/
├── main.go                          # ~30 lines: env, client, server, RegisterAll, run
├── internal/
│   ├── models/                      # Shared data types (no logic, no deps)
│   │   ├── task.go                  # Task, DueDate structs
│   │   ├── project.go
│   │   ├── section.go
│   │   ├── label.go
│   │   └── comment.go
│   ├── todoist/                     # Pure API client (no MCP awareness)
│   │   ├── client.go               # Client struct, do() method, Option pattern
│   │   ├── tasks.go                # CreateTask, GetTasks, GetTask, UpdateTask, DeleteTask, CloseTask, ReopenTask
│   │   ├── projects.go
│   │   ├── sections.go
│   │   ├── labels.go
│   │   └── comments.go
│   └── tools/                      # MCP tool handlers
│       ├── register.go             # RegisterAll(server, client) wiring
│       ├── tasks.go
│       ├── projects.go
│       ├── sections.go
│       ├── labels.go
│       ├── comments.go
│       └── gtd.go                  # Higher-level GTD composite tools
```

**Steps:**
1. Create `internal/models/task.go` — extract `Task`, `DueDate` from `main.go:26-43`
2. Create `internal/todoist/client.go` — extract `makeAPIRequest` from `main.go:110-145` into `Client.do()` method with `NewClient(token, ...Option)`, `WithHTTPClient()`, `WithBaseURL()` options
3. Create `internal/todoist/tasks.go` — task API methods on `*Client`
4. Create `internal/tools/tasks.go` — migrate 5 existing handlers, receiving `*todoist.Client` via closure
5. Create `internal/tools/register.go` — `RegisterAll()` function
6. Rewrite `main.go` to thin entry point
7. Verify: `make build && make test` — all existing behavior preserved

**Files modified:** `main.go` (rewrite), `go.mod` (no change expected)
**Files created:** all `internal/` files listed above

---

## Phase 2: Full Todoist REST API v2 Coverage

### 2a. Enhance Existing Task Tools
- Expand `Task` model with all API fields: `SectionID`, `Labels`, `ParentID`, `Order`, `IsCompleted`, `URL`, `CommentCount`, `CreatorID`, `AssigneeID`, `Duration`
- Add `task_id` parameter as alternative to `task_name` on update/delete/complete tools (skip name search when ID is provided)
- Expand `create_task` params: `project_id`, `section_id`, `parent_id`, `labels[]`, `assignee_id`
- Expand `update_task` params: same additions
- Add `todoist_reopen_task` tool — POST `/tasks/{id}/reopen`

### 2b. Project Tools (7 new tools)

| Tool | Description | Key Params |
|------|-------------|------------|
| `todoist_get_projects` | List all projects | — |
| `todoist_get_project` | Get single project | `project_id` |
| `todoist_create_project` | Create project | `name`, `parent_id`, `color`, `is_favorite`, `view_style` |
| `todoist_update_project` | Update project | `project_id`, `name`, `color`, `is_favorite` |
| `todoist_delete_project` | Delete project | `project_id` |
| `todoist_archive_project` | Archive project | `project_id` |
| `todoist_unarchive_project` | Unarchive project | `project_id` |

### 2c. Section Tools (4 new tools)

| Tool | Description | Key Params |
|------|-------------|------------|
| `todoist_get_sections` | List sections | `project_id` |
| `todoist_create_section` | Create section | `name`, `project_id`, `order` |
| `todoist_update_section` | Update section | `section_id`, `name` |
| `todoist_delete_section` | Delete section | `section_id` |

### 2d. Label Tools (4 new tools)

| Tool | Description | Key Params |
|------|-------------|------------|
| `todoist_get_labels` | List all labels | — |
| `todoist_create_label` | Create label | `name`, `color`, `is_favorite` |
| `todoist_update_label` | Update label | `label_id`, `name`, `color` |
| `todoist_delete_label` | Delete label | `label_id` |

### 2e. Comment Tools (4 new tools)

| Tool | Description | Key Params |
|------|-------------|------------|
| `todoist_get_comments` | List comments | `task_id` or `project_id` |
| `todoist_create_comment` | Add comment | `content`, `task_id` or `project_id` |
| `todoist_update_comment` | Update comment | `comment_id`, `content` |
| `todoist_delete_comment` | Delete comment | `comment_id` |

---

## Phase 3: GTD Workflow Composite Tools (4 new tools)

These orchestrate multiple API calls to directly support the GTD+Zettelkasten workflow:

| Tool | Description | How it works |
|------|-------------|-------------|
| `todoist_inbox_review` | Get all inbox tasks grouped by age | Auto-detects inbox project via `is_inbox_project`, fetches tasks, groups by today/this week/older |
| `todoist_weekly_review` | Comprehensive weekly review summary | Aggregates: all projects with task counts, overdue tasks, recently completed tasks, tasks with no due date |
| `todoist_move_task` | Move task to different project/section | Name/ID lookup + update with `project_id`/`section_id` |
| `todoist_bulk_create_tasks` | Create multiple tasks at once | Accepts `tasks[]` array, creates each, returns summary. Useful for Knowledge→Action loop |

**Total: 29 tools** (6 task + 7 project + 4 section + 4 label + 4 comment + 4 GTD)

---

## Phase 4: Testing & Documentation

### Testing Strategy
- **API client tests** (`internal/todoist/*_test.go`): Use `net/http/httptest` — fake Todoist server returns canned JSON. Test request method/path/headers/body and response parsing. Cover error cases (401, 404, 500, malformed JSON).
- **Tool handler tests** (`internal/tools/*_test.go`): Define interfaces for `todoist.Client` methods, inject mocks. Verify MCP result content and `IsError` flag.
- **GTD tool tests**: Mock client tracking call sequences to verify orchestration logic.
- **Coverage targets**: `internal/todoist/` 90%+, `internal/tools/` 85%+

### Documentation
- Update `README.md` with all 29 tools documented
- Update CI pipeline if needed (existing `./...` patterns should auto-discover new packages)

---

## Verification

1. `make build` — compiles successfully
2. `make test` — all tests pass with race detector
3. `make lint` — no linting errors
4. Manual test: configure in Claude Desktop, exercise each tool category:
   - Create a project, add sections, create tasks in sections
   - Add labels and comments to tasks
   - Run `todoist_inbox_review` and `todoist_weekly_review`
   - Complete and reopen tasks
   - Use `todoist_bulk_create_tasks` to batch-create
5. `entire version` — verify entire.io is installed and hooks are active

---

## Key Files to Modify/Create

| File | Action |
|------|--------|
| `main.go` | Rewrite to thin entry point (~30 lines) |
| `internal/models/*.go` | Create — data types |
| `internal/todoist/client.go` | Create — HTTP client extracted from `main.go:110-145` |
| `internal/todoist/{tasks,projects,sections,labels,comments}.go` | Create — API methods |
| `internal/tools/{register,tasks,projects,sections,labels,comments,gtd}.go` | Create — MCP handlers |
| `internal/todoist/*_test.go` | Create — API client tests |
| `internal/tools/*_test.go` | Create — handler tests |
| `README.md` | Update — document all 29 tools |
| `.gitignore` | Update — add `.entire/` if needed |

## Existing Code to Reuse
- `makeAPIRequest` pattern (`main.go:110-145`) → becomes `Client.do()`
- `findTaskByName` logic (`main.go:294-315`) → enhanced with exact-match preference and multi-match error
- Handler patterns (`main.go:147-479`) → same structure, parameterized with client
- `Makefile` — no changes needed (`./...` covers new packages)
- CI workflow — no changes needed
