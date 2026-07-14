# Staged Tech Debt Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate six identified debt items in mcp-todoist across five risk-ordered PRs: repo hygiene, context propagation plus typed API requests, handler dedup, update field-clearing, and retry/rate-limit handling.

**Architecture:** Three-layer Go MCP server (`internal/tools` MCP handlers → `internal/todoist` HTTP client → `internal/models` types). All changes preserve this layering. Stages 1-3 are behavior-preserving refactors verified by the existing test suite; stages 4-5 add behavior with new table-style httptest coverage.

**Tech Stack:** Go 1.26.2, `github.com/modelcontextprotocol/go-sdk` v1.6.1, standard library only (`net/http`, `net/http/httptest`, `testing`).

**Spec:** `docs/superpowers/specs/2026-07-14-tech-debt-refactor-design.md`

## Global Constraints

- No new dependencies. Standard library only.
- `make check` (fmt, vet, golangci-lint, `go test -race`) must pass at the end of every task.
- Tests use `net/http/httptest` and plain `testing` with `t.Fatal`/`t.Fatalf`. No test libraries.
- Logging via `log/slog`, lowercase messages, errors as attributes.
- Each stage is one branch and one PR. Both GitHub Actions jobs ("Build and Test", "golangci-lint") must be green before merge. After opening each PR, STOP and ask the user to review and merge before starting the next stage.
- Commit messages follow Conventional Commits 1.0.0 and must end with:
  `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>` and
  `Claude-Session: https://claude.ai/code/session_015j22Jr9L3mzY3wNohfvWcu`
- No em-dashes in any prose, comment, or commit message. Use commas, colons, parentheses, or sentence splits.

## Stage → Task map

| Stage (PR) | Branch | Tasks |
|------------|--------|-------|
| 1: Repo hygiene | `refactor/stage-1-repo-hygiene` | Task 1 |
| 2: Client API reshape | `refactor/stage-2-client-api` | Tasks 2, 3, 4 |
| 3: Handler dedup | `refactor/stage-3-handler-dedup` | Tasks 5, 6 |
| 4: Update clearing semantics | `feat/stage-4-update-clearing` | Task 7 |
| 5: Retry and rate limits | `feat/stage-5-client-retries` | Task 8 |

Each stage branches from up-to-date `main` (`git checkout main && git pull` first).

---

### Task 1: Repo hygiene (all of Stage 1)

**Files:**
- Delete (tracked): `mcp-todoist` (11MB binary at repo root)
- Delete (disk): `.go.mod.~undo-tree~`
- Modify: `.gitignore` (add editor-artifact pattern)
- Modify: `CLAUDE.md` (go-sdk version reference)
- Modify: `README.md` (go-sdk version references)

**Interfaces:**
- Consumes: nothing
- Produces: nothing (no code changes)

- [ ] **Step 1: Create the stage branch**

```bash
git checkout main && git pull
git checkout -b refactor/stage-1-repo-hygiene
```

- [ ] **Step 2: Remove the committed binary and editor artifact**

`.gitignore` already lists `mcp-todoist` (line 10); the binary was committed before that rule existed, so it stays tracked until removed explicitly. The undo-tree file may or may not be tracked; handle both.

```bash
git rm -f mcp-todoist
git rm -f --ignore-unmatch '.go.mod.~undo-tree~'
rm -f '.go.mod.~undo-tree~'
```

- [ ] **Step 3: Add the editor-artifact pattern to .gitignore**

In `.gitignore`, under the `# Editor/IDE` section at the bottom, add:

```
*.~undo-tree~
```

- [ ] **Step 4: Verify removal**

Run: `git ls-files | grep -E '(^mcp-todoist$|undo-tree)'`
Expected: no output.

- [ ] **Step 5: Commit the artifact removal**

```bash
git add .gitignore
git commit -m "chore: remove committed binary and editor artifact"
```

(Append the two required footer lines to this and every commit in this plan.)

- [ ] **Step 6: Find stale version references**

Run: `grep -n "v1\.4\.0" CLAUDE.md README.md`
Expected: at least one hit in each file, e.g. CLAUDE.md's line
`Built with [go-sdk v1.4.0](https://github.com/modelcontextprotocol/go-sdk).`

- [ ] **Step 7: Update every hit to v1.6.1**

Edit each line found in Step 6, replacing `v1.4.0` with `v1.6.1`. Keep link URLs unchanged (they don't embed the version). Re-run the grep; expected: no output.

- [ ] **Step 8: Run checks**

Run: `make check`
Expected: `All checks passed!`

- [ ] **Step 9: Commit the docs sync**

```bash
git add CLAUDE.md README.md
git commit -m "docs: sync go-sdk version references to v1.6.1"
```

- [ ] **Step 10: Open the PR and watch CI**

```bash
git push -u origin refactor/stage-1-repo-hygiene
gh pr create --title "chore: stage 1 repo hygiene" --body "Stage 1 of the tech debt refactor (see docs/superpowers/specs/2026-07-14-tech-debt-refactor-design.md). Removes the committed 11MB binary and editor artifact, adds the ignore pattern, and syncs go-sdk version references to v1.6.1. No code changes."
gh pr checks --watch
```

Expected: all checks pass. Then STOP: ask the user to review and merge before Stage 2.

---

### Task 2: Propagate context.Context through the API client (Stage 2, part 1)

**Files:**
- Modify: `internal/todoist/client.go` (the `do` method, lines 53-93)
- Modify: `internal/todoist/tasks.go`, `projects.go`, `sections.go`, `labels.go`, `comments.go` (every method)
- Modify: `internal/tools/tasks.go`, `gtd.go`, `projects.go`, `sections.go`, `labels.go`, `comments.go` (every client call site)
- Modify: `internal/todoist/client_test.go`, `tasks_test.go`, `projects_test.go`, `sections_test.go`, `labels_test.go`, `comments_test.go` (every method call site)

**Interfaces:**
- Consumes: existing `Client` methods.
- Produces: every `Client` method (and unexported helpers `do`, `getTasksPage`, `getFilteredTasksPage`) takes `ctx context.Context` as its first parameter. `do` uses `http.NewRequestWithContext`. Bodies stay `map[string]any` in this task; Task 3 changes them. Also produces `resolveTaskID(ctx context.Context, c *todoist.Client, id, name string) (string, string, error)` in `internal/tools/tasks.go`.

- [ ] **Step 1: Create the stage branch**

```bash
git checkout main && git pull
git checkout -b refactor/stage-2-client-api
```

- [ ] **Step 2: Change `do` to accept and use a context**

In `internal/todoist/client.go`, add `"context"` to imports and change the signature and request construction:

```go
// do executes an HTTP request against the Todoist API and returns the
// response body bytes. For responses with no content (204) it returns nil.
func (c *Client) do(ctx context.Context, method, endpoint string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	// rest of the method is unchanged
```

- [ ] **Step 3: Add ctx to every client method**

Mechanical sweep over the five resource files. Every method gains `ctx context.Context` as first parameter and passes it to `c.do(ctx, ...)` (or to the page helper it calls). Add `"context"` to each file's imports. The complete list of new signatures:

`internal/todoist/tasks.go`:
```go
func (c *Client) getTasksPage(ctx context.Context, projectID, cursor string) ([]models.Task, string, error)
func (c *Client) GetTasks(ctx context.Context, projectID string) ([]models.Task, error)
func (c *Client) getFilteredTasksPage(ctx context.Context, query, cursor string) ([]models.Task, string, error)
func (c *Client) GetTasksByFilter(ctx context.Context, query string) ([]models.Task, error)
func (c *Client) GetTask(ctx context.Context, id string) (*models.Task, error)
func (c *Client) CreateTask(ctx context.Context, body map[string]any) (*models.Task, error)
func (c *Client) UpdateTask(ctx context.Context, id string, body map[string]any) (*models.Task, error)
func (c *Client) MoveTask(ctx context.Context, id string, body map[string]any) (*models.Task, error)
func (c *Client) DeleteTask(ctx context.Context, id string) error
func (c *Client) CloseTask(ctx context.Context, id string) error
func (c *Client) ReopenTask(ctx context.Context, id string) error
func (c *Client) FindTaskByName(ctx context.Context, name string) (*models.Task, error)
```

`internal/todoist/projects.go`:
```go
func (c *Client) GetProjects(ctx context.Context) ([]models.Project, error)
func (c *Client) GetProject(ctx context.Context, id string) (*models.Project, error)
func (c *Client) CreateProject(ctx context.Context, body map[string]any) (*models.Project, error)
func (c *Client) UpdateProject(ctx context.Context, id string, body map[string]any) (*models.Project, error)
func (c *Client) DeleteProject(ctx context.Context, id string) error
func (c *Client) ArchiveProject(ctx context.Context, id string) error
func (c *Client) UnarchiveProject(ctx context.Context, id string) error
```

`internal/todoist/sections.go`:
```go
func (c *Client) GetSections(ctx context.Context, projectID string) ([]models.Section, error)
func (c *Client) CreateSection(ctx context.Context, body map[string]any) (*models.Section, error)
func (c *Client) UpdateSection(ctx context.Context, id string, body map[string]any) (*models.Section, error)
func (c *Client) DeleteSection(ctx context.Context, id string) error
```

`internal/todoist/labels.go`:
```go
func (c *Client) GetLabels(ctx context.Context) ([]models.Label, error)
func (c *Client) CreateLabel(ctx context.Context, body map[string]any) (*models.Label, error)
func (c *Client) UpdateLabel(ctx context.Context, id string, body map[string]any) (*models.Label, error)
func (c *Client) DeleteLabel(ctx context.Context, id string) error
```

`internal/todoist/comments.go`:
```go
func (c *Client) GetComments(ctx context.Context, taskID, projectID string) ([]models.Comment, error)
func (c *Client) CreateComment(ctx context.Context, body map[string]any) (*models.Comment, error)
func (c *Client) UpdateComment(ctx context.Context, id string, body map[string]any) (*models.Comment, error)
func (c *Client) DeleteComment(ctx context.Context, id string) error
```

Representative complete example (apply the same shape everywhere):

```go
// GetTasks returns the first page of active tasks, optionally filtered by project.
func (c *Client) GetTasks(ctx context.Context, projectID string) ([]models.Task, error) {
	tasks, _, err := c.getTasksPage(ctx, projectID, "")
	return tasks, err
}
```

Internal pagination loops (`GetTasksByFilter`, `FindTaskByName`) pass ctx into each page call.

- [ ] **Step 4: Verify the todoist package compiles (tools will still be broken)**

Run: `go build ./internal/todoist/`
Expected: success.

- [ ] **Step 5: Update every call site in internal/tools**

Every handler already receives `ctx context.Context` as its first closure parameter; pass it through. Update `resolveTaskID` in `internal/tools/tasks.go:93` to:

```go
func resolveTaskID(ctx context.Context, c *todoist.Client, id, name string) (string, string, error) {
	if id != "" {
		return id, "", nil
	}
	if name == "" {
		return "", "", fmt.Errorf("either task_id or task_name is required")
	}
	task, err := c.FindTaskByName(ctx, name)
	if err != nil {
		return "", "", err
	}
	if task == nil {
		return "", "", nil // not found
	}
	return task.ID, task.Content, nil
}
```

Call-site checklist (add `ctx` as first argument to each):
- `tools/tasks.go`: `resolveTaskID` (5 calls), `c.CreateTask`, `c.GetTasksByFilter`, `c.GetTasks`, `c.UpdateTask`, `c.DeleteTask`, `c.CloseTask`, `c.ReopenTask`
- `tools/gtd.go`: `c.GetProjects` (2 calls), `c.GetTasks` (2 calls), `c.GetTasksByFilter`, `resolveTaskID`, `c.MoveTask`, `c.CreateTask` (bulk loop)
- `tools/projects.go`: `c.GetProjects`, `c.GetProject`, `c.CreateProject`, `c.UpdateProject`, `c.DeleteProject`, `c.ArchiveProject`, `c.UnarchiveProject`
- `tools/sections.go`: `c.GetSections`, `c.CreateSection`, `c.UpdateSection`, `c.DeleteSection`
- `tools/labels.go`: `c.GetLabels`, `c.CreateLabel`, `c.UpdateLabel`, `c.DeleteLabel`
- `tools/comments.go`: `c.GetComments`, `c.CreateComment`, `c.UpdateComment`, `c.DeleteComment`

- [ ] **Step 6: Update every call site in the todoist tests**

In `internal/todoist/*_test.go`, add `"context"` to imports where missing and pass `context.Background()` as the first argument to every client method call. Example transformation:

```go
// before
tasks, err := c.GetTasks("")
// after
tasks, err := c.GetTasks(context.Background(), "")
```

Do not change any assertion. Find all call sites with:
`grep -n "c\.\(Get\|Create\|Update\|Delete\|Move\|Close\|Reopen\|Find\|Archive\|Unarchive\)" internal/todoist/*_test.go`

- [ ] **Step 7: Run checks**

Run: `make check`
Expected: `All checks passed!` with the same test names passing as on main.

- [ ] **Step 8: Commit**

```bash
git add internal/
git commit -m "refactor(todoist): propagate context through the API client"
```

---

### Task 3: Typed request structs for tasks (Stage 2, part 2)

**Files:**
- Modify: `internal/todoist/client.go` (add `Ptr` helper)
- Modify: `internal/todoist/tasks.go` (structs + `CreateTask`/`UpdateTask`/`MoveTask` signatures)
- Modify: `internal/tools/tasks.go` (create/update handlers), `internal/tools/gtd.go` (move + bulk-create handlers)
- Modify: `internal/todoist/tasks_test.go` (body construction in `TestCreateTask`, `TestUpdateTask`, `TestMoveTask` and any other map-body call)

**Interfaces:**
- Consumes: ctx-aware methods from Task 2.
- Produces (exact, later tasks depend on these):

```go
// internal/todoist/client.go
// Ptr returns a pointer to v. Convenience for building request structs.
func Ptr[T any](v T) *T { return &v }

// internal/todoist/tasks.go
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

// Labels is *[]string so a pointer to an empty slice can express
// "clear all labels"; omitempty on a plain slice would drop it.
type UpdateTaskRequest struct {
	Content      *string   `json:"content,omitempty"`
	Description  *string   `json:"description,omitempty"`
	DueString    *string   `json:"due_string,omitempty"`
	Priority     *int      `json:"priority,omitempty"`
	Labels       *[]string `json:"labels,omitempty"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
	DeadlineDate *string   `json:"deadline_date,omitempty"`
}

// MoveTaskRequest moves a task. Set exactly one field.
type MoveTaskRequest struct {
	ProjectID *string `json:"project_id,omitempty"`
	SectionID *string `json:"section_id,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
}

func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error)
func (c *Client) UpdateTask(ctx context.Context, id string, req UpdateTaskRequest) (*models.Task, error)
func (c *Client) MoveTask(ctx context.Context, id string, req MoveTaskRequest) (*models.Task, error)
```

Method bodies only change `body` to `req` in the `c.do(ctx, ...)` call; JSON marshaling of nil-pointer omitempty fields produces the same wire bytes as the omitted map keys did.

- [ ] **Step 1: Add the `Ptr` helper and the three structs, change the three method signatures**

Apply the definitions from the Interfaces block above verbatim.

- [ ] **Step 2: Rewrite the create-task handler body construction**

In `internal/tools/tasks.go`, inside `todoist_create_task`:

```go
req2 := todoist.CreateTaskRequest{Content: input.Content}
if input.Description != "" {
	req2.Description = todoist.Ptr(input.Description)
}
if input.DueString != "" {
	req2.DueString = todoist.Ptr(input.DueString)
}
if input.Priority > 0 && input.Priority <= 4 {
	req2.Priority = todoist.Ptr(input.Priority)
}
if input.ProjectID != "" {
	req2.ProjectID = todoist.Ptr(input.ProjectID)
}
if input.SectionID != "" {
	req2.SectionID = todoist.Ptr(input.SectionID)
}
if input.ParentID != "" {
	req2.ParentID = todoist.Ptr(input.ParentID)
}
if len(input.Labels) > 0 {
	req2.Labels = input.Labels
}
if input.AssigneeID != "" {
	req2.AssigneeID = todoist.Ptr(input.AssigneeID)
}

task, err := c.CreateTask(ctx, req2)
```

(Name the variable `req2` only if `req *mcp.CallToolRequest` shadows; the closure parameter is named `req`, so it does. Keep `req2` or rename thoughtfully, e.g. `createReq`.)

- [ ] **Step 3: Rewrite the update-task handler body construction**

Same guard conditions as today, so behavior is unchanged this stage:

```go
updateReq := todoist.UpdateTaskRequest{}
if input.Content != "" {
	updateReq.Content = todoist.Ptr(input.Content)
}
if input.Description != "" {
	updateReq.Description = todoist.Ptr(input.Description)
}
if input.DueString != "" {
	updateReq.DueString = todoist.Ptr(input.DueString)
}
if input.Priority > 0 && input.Priority <= 4 {
	updateReq.Priority = todoist.Ptr(input.Priority)
}
if len(input.Labels) > 0 {
	updateReq.Labels = todoist.Ptr(input.Labels)
}
if input.AssigneeID != "" {
	updateReq.AssigneeID = todoist.Ptr(input.AssigneeID)
}
if input.DeadlineDate != "" {
	updateReq.DeadlineDate = todoist.Ptr(input.DeadlineDate)
}

updated, err := c.UpdateTask(ctx, id, updateReq)
```

- [ ] **Step 4: Rewrite the move and bulk-create handlers in gtd.go**

Move (validation logic unchanged):

```go
moveReq := todoist.MoveTaskRequest{}
dests := 0
if input.ProjectID != "" {
	moveReq.ProjectID = todoist.Ptr(input.ProjectID)
	dests++
}
if input.SectionID != "" {
	moveReq.SectionID = todoist.Ptr(input.SectionID)
	dests++
}
if input.ParentID != "" {
	moveReq.ParentID = todoist.Ptr(input.ParentID)
	dests++
}
if dests != 1 {
	msg := "exactly one of project_id, section_id, parent_id must be set"
	return textResult(msg, true), MoveTaskOutput{Success: false, Message: msg}, nil
}

_, err = c.MoveTask(ctx, id, moveReq)
```

Bulk create, inside the loop over `input.Tasks`:

```go
createReq := todoist.CreateTaskRequest{Content: item.Content}
if item.Description != "" {
	createReq.Description = todoist.Ptr(item.Description)
}
if item.DueString != "" {
	createReq.DueString = todoist.Ptr(item.DueString)
}
if item.Priority > 0 && item.Priority <= 4 {
	createReq.Priority = todoist.Ptr(item.Priority)
}
if item.ProjectID != "" {
	createReq.ProjectID = todoist.Ptr(item.ProjectID)
}
if item.SectionID != "" {
	createReq.SectionID = todoist.Ptr(item.SectionID)
}
if len(item.Labels) > 0 {
	createReq.Labels = item.Labels
}

task, err := c.CreateTask(ctx, createReq)
```

- [ ] **Step 5: Update the todoist task tests' body construction**

In `internal/todoist/tasks_test.go`, replace map literals with structs. The httptest handlers decode the request body as `map[string]any` and their assertions stay untouched, which is exactly what proves wire-format equivalence:

```go
// TestCreateTask
task, err := c.CreateTask(context.Background(), CreateTaskRequest{Content: "New task"})
// TestUpdateTask
task, err := c.UpdateTask(context.Background(), "10", UpdateTaskRequest{Content: Ptr("Updated")})
// TestMoveTask
task, err := c.MoveTask(context.Background(), "10", MoveTaskRequest{ProjectID: Ptr("42")})
```

Sweep the rest of the file (and `client_test.go` if it builds bodies) for remaining `map[string]any{` arguments to these three methods.

- [ ] **Step 6: Run checks**

Run: `make check`
Expected: `All checks passed!` The `internal/tools` tests pass unmodified because the JSON reaching the fake API is byte-equivalent.

- [ ] **Step 7: Commit**

```bash
git add internal/
git commit -m "refactor(todoist): replace task request maps with typed structs"
```

---

### Task 4: Typed request structs for projects, sections, labels, comments (Stage 2, part 3)

**Files:**
- Modify: `internal/todoist/projects.go`, `sections.go`, `labels.go`, `comments.go`
- Modify: `internal/tools/projects.go`, `sections.go`, `labels.go`, `comments.go`
- Modify: `internal/todoist/projects_test.go`, `sections_test.go`, `labels_test.go`, `comments_test.go`

**Interfaces:**
- Consumes: `todoist.Ptr` from Task 3, ctx-aware methods from Task 2.
- Produces (exact):

```go
// internal/todoist/projects.go
type CreateProjectRequest struct {
	Name       string  `json:"name"`
	ParentID   *string `json:"parent_id,omitempty"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
	ViewStyle  *string `json:"view_style,omitempty"`
}
type UpdateProjectRequest struct {
	Name       *string `json:"name,omitempty"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
}
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*models.Project, error)
func (c *Client) UpdateProject(ctx context.Context, id string, req UpdateProjectRequest) (*models.Project, error)

// internal/todoist/sections.go
type CreateSectionRequest struct {
	Name         string `json:"name"`
	ProjectID    string `json:"project_id"`
	SectionOrder *int   `json:"section_order,omitempty"`
}
type UpdateSectionRequest struct {
	Name string `json:"name"`
}
func (c *Client) CreateSection(ctx context.Context, req CreateSectionRequest) (*models.Section, error)
func (c *Client) UpdateSection(ctx context.Context, id string, req UpdateSectionRequest) (*models.Section, error)

// internal/todoist/labels.go
type CreateLabelRequest struct {
	Name       string  `json:"name"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
}
type UpdateLabelRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}
func (c *Client) CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error)
func (c *Client) UpdateLabel(ctx context.Context, id string, req UpdateLabelRequest) (*models.Label, error)

// internal/todoist/comments.go
type CreateCommentRequest struct {
	Content   string  `json:"content"`
	TaskID    *string `json:"task_id,omitempty"`
	ProjectID *string `json:"project_id,omitempty"`
}
type UpdateCommentRequest struct {
	Content string `json:"content"`
}
func (c *Client) CreateComment(ctx context.Context, req CreateCommentRequest) (*models.Comment, error)
func (c *Client) UpdateComment(ctx context.Context, id string, req UpdateCommentRequest) (*models.Comment, error)
```

- [ ] **Step 1: Add structs and change signatures in the four client files**

Apply the Interfaces block verbatim. Method bodies still just pass `req` to `c.do(ctx, ...)`.

- [ ] **Step 2: Update the four tool files' body construction**

`internal/tools/projects.go` create:

```go
createReq := todoist.CreateProjectRequest{Name: input.Name}
if input.ParentID != "" {
	createReq.ParentID = todoist.Ptr(input.ParentID)
}
if input.Color != "" {
	createReq.Color = todoist.Ptr(input.Color)
}
if input.IsFavorite {
	createReq.IsFavorite = todoist.Ptr(true)
}
if input.ViewStyle != "" {
	createReq.ViewStyle = todoist.Ptr(input.ViewStyle)
}
p, err := c.CreateProject(ctx, createReq)
```

`internal/tools/projects.go` update:

```go
updateReq := todoist.UpdateProjectRequest{}
if input.Name != "" {
	updateReq.Name = todoist.Ptr(input.Name)
}
if input.Color != "" {
	updateReq.Color = todoist.Ptr(input.Color)
}
if input.IsFavorite != nil {
	updateReq.IsFavorite = input.IsFavorite
}
p, err := c.UpdateProject(ctx, input.ProjectID, updateReq)
```

`internal/tools/sections.go` create and update:

```go
createReq := todoist.CreateSectionRequest{Name: input.Name, ProjectID: input.ProjectID}
if input.Order > 0 {
	createReq.SectionOrder = todoist.Ptr(input.Order)
}
sec, err := c.CreateSection(ctx, createReq)
```

```go
sec, err := c.UpdateSection(ctx, input.SectionID, todoist.UpdateSectionRequest{Name: input.Name})
```

`internal/tools/labels.go` create and update:

```go
createReq := todoist.CreateLabelRequest{Name: input.Name}
if input.Color != "" {
	createReq.Color = todoist.Ptr(input.Color)
}
if input.IsFavorite {
	createReq.IsFavorite = todoist.Ptr(true)
}
l, err := c.CreateLabel(ctx, createReq)
```

```go
updateReq := todoist.UpdateLabelRequest{}
if input.Name != "" {
	updateReq.Name = todoist.Ptr(input.Name)
}
if input.Color != "" {
	updateReq.Color = todoist.Ptr(input.Color)
}
l, err := c.UpdateLabel(ctx, input.LabelID, updateReq)
```

`internal/tools/comments.go` create and update:

```go
createReq := todoist.CreateCommentRequest{Content: input.Content}
if input.TaskID != "" {
	createReq.TaskID = todoist.Ptr(input.TaskID)
}
if input.ProjectID != "" {
	createReq.ProjectID = todoist.Ptr(input.ProjectID)
}
cm, err := c.CreateComment(ctx, createReq)
```

```go
cm, err := c.UpdateComment(ctx, input.CommentID, todoist.UpdateCommentRequest{Content: input.Content})
```

- [ ] **Step 3: Update the four todoist test files**

Replace `map[string]any` arguments with the new structs, keeping every handler-side assertion untouched. Example: `c.CreateLabel(context.Background(), CreateLabelRequest{Name: "urgent"})`.

- [ ] **Step 4: Run checks**

Run: `make check`
Expected: `All checks passed!`

- [ ] **Step 5: Verify no maps remain in the client API**

Run: `grep -rn "map\[string\]any" internal/todoist/ internal/tools/ --include='*.go' | grep -v _test.go`
Expected: no output. (Test files may still decode bodies into maps; that's fine.)

- [ ] **Step 6: Commit, open the stage 2 PR, watch CI**

```bash
git add internal/
git commit -m "refactor(todoist): replace remaining request maps with typed structs"
git push -u origin refactor/stage-2-client-api
gh pr create --title "refactor: stage 2 client API reshape (context + typed requests)" --body "Stage 2 of the tech debt refactor. Every client method now takes ctx (http.NewRequestWithContext) and all map[string]any bodies are typed request structs. No behavior change; all existing test assertions preserved."
gh pr checks --watch
```

Expected: all checks pass. Then STOP: ask the user to review and merge before Stage 3.

---

### Task 5: Unify output structs into ActionOutput (Stage 3, part 1)

**Files:**
- Create: `internal/tools/output.go`
- Modify: `internal/tools/tasks.go`, `projects.go`, `sections.go`, `labels.go`, `comments.go`, `gtd.go`

**Interfaces:**
- Consumes: nothing new.
- Produces:

```go
// internal/tools/output.go
package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ActionOutput is the shared structured output for all tools: a success
// flag plus the same human-readable message returned as text content.
type ActionOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func textResult(msg string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: isError,
	}
}
```

- [ ] **Step 1: Create the stage branch**

```bash
git checkout main && git pull
git checkout -b refactor/stage-3-handler-dedup
```

- [ ] **Step 2: Create `internal/tools/output.go`**

Content as in the Interfaces block. Delete the old `textResult` from `tasks.go` (lines 110-115 pre-refactor).

- [ ] **Step 3: Delete the 29 per-tool output structs and substitute ActionOutput**

Delete these type declarations and replace every use (handler signature `(*mcp.CallToolResult, XOutput, error)` and every return literal) with `ActionOutput`:

- `tasks.go`: `CreateTaskOutput`, `GetTasksOutput`, `UpdateTaskOutput`, `DeleteTaskOutput`, `CompleteTaskOutput`, `ReopenTaskOutput`
- `projects.go`: `GetProjectsOutput`, `GetProjectOutput`, `CreateProjectOutput`, `UpdateProjectOutput`, `DeleteProjectOutput`, `ArchiveProjectOutput`, `UnarchiveProjectOutput`
- `sections.go`: `GetSectionsOutput`, `CreateSectionOutput`, `UpdateSectionOutput`, `DeleteSectionOutput`
- `labels.go`: `GetLabelsOutput`, `CreateLabelOutput`, `UpdateLabelOutput`, `DeleteLabelOutput`
- `comments.go`: `GetCommentsOutput`, `CreateCommentOutput`, `UpdateCommentOutput`, `DeleteCommentOutput`
- `gtd.go`: `InboxReviewOutput`, `WeeklyReviewOutput`, `MoveTaskOutput`, `BulkCreateTasksOutput`

(That's 29 declarations; the JSON shape of every one is already `{success, message}`, so the derived MCP output schema is unchanged.) Return literals keep field names: `ActionOutput{Success: true, Message: msg}`.

- [ ] **Step 4: Verify nothing references the old types**

Run: `grep -rn "Output{" internal/tools/*.go | grep -v "ActionOutput{"`
Expected: no output.

- [ ] **Step 5: Run checks**

Run: `make check`
Expected: `All checks passed!` (tool tests assert on text content and tool names only).

- [ ] **Step 6: Commit**

```bash
git add internal/tools/
git commit -m "refactor(tools): unify per-tool output structs into ActionOutput"
```

---

### Task 6: Extract task-action helpers (Stage 3, part 2)

**Files:**
- Modify: `internal/tools/tasks.go` (add `resolveTask`, `runTaskAction`; rewrite update/delete/complete/reopen handlers)
- Modify: `internal/tools/gtd.go` (move handler uses `resolveTask`)

**Interfaces:**
- Consumes: `resolveTaskID` (Task 2 signature), `ActionOutput`, `textResult` (Task 5).
- Produces:

```go
// resolveTask resolves a task by ID or name and returns a display label
// (the matched task name, falling back to the ID). taskID is "" when a
// name search found nothing.
func resolveTask(ctx context.Context, c *todoist.Client, id, name string) (taskID, label string, err error)

// runTaskAction runs the shared resolve, not-found, act, report flow for
// task tools whose success message is "Successfully <verb> task: ...".
func runTaskAction(ctx context.Context, c *todoist.Client, id, name, verb string, action func(context.Context, string) error) (*mcp.CallToolResult, ActionOutput, error)
```

- [ ] **Step 1: Add the two helpers to tasks.go**

```go
func resolveTask(ctx context.Context, c *todoist.Client, id, name string) (taskID, label string, err error) {
	taskID, originalName, err := resolveTaskID(ctx, c, id, name)
	if err != nil || taskID == "" {
		return taskID, "", err
	}
	label = originalName
	if label == "" {
		label = taskID
	}
	return taskID, label, nil
}

func runTaskAction(ctx context.Context, c *todoist.Client, id, name, verb string, action func(context.Context, string) error) (*mcp.CallToolResult, ActionOutput, error) {
	taskID, label, err := resolveTask(ctx, c, id, name)
	if err != nil {
		return nil, ActionOutput{Success: false, Message: err.Error()}, err
	}
	if taskID == "" {
		msg := fmt.Sprintf("Could not find a task matching \"%s\"", name)
		return textResult(msg, true), ActionOutput{Success: false, Message: msg}, nil
	}
	if err := action(ctx, taskID); err != nil {
		return nil, ActionOutput{Success: false, Message: err.Error()}, err
	}
	msg := fmt.Sprintf("Successfully %s task: \"%s\"", verb, label)
	return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
}
```

The message strings are copied verbatim from the current handlers so wire output stays byte-identical.

- [ ] **Step 2: Collapse the delete, complete, and reopen handlers**

Each handler body becomes a single call; `c.DeleteTask`, `c.CloseTask`, and `c.ReopenTask` already match `func(context.Context, string) error`:

```go
}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteTaskInput) (*mcp.CallToolResult, ActionOutput, error) {
	return runTaskAction(ctx, c, input.TaskID, input.TaskName, "deleted", c.DeleteTask)
})
```

Verbs: `"deleted"`, `"completed"`, `"reopened"`.

- [ ] **Step 3: Rewrite update and move to use resolveTask**

In the update handler, replace the resolve/not-found/label boilerplate:

```go
id, label, err := resolveTask(ctx, c, input.TaskID, input.TaskName)
if err != nil {
	return nil, ActionOutput{Success: false, Message: err.Error()}, err
}
if id == "" {
	msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
	return textResult(msg, true), ActionOutput{Success: false, Message: msg}, nil
}
// build updateReq exactly as before, then:
msg := fmt.Sprintf("Task \"%s\" updated:\nNew Title: %s", label, updated.Content)
```

Same substitution in the move handler in `gtd.go` (its success message construction is unchanged).

- [ ] **Step 4: Run checks**

Run: `make check`
Expected: `All checks passed!` If any `tools_test.go` assertion fails, the helper text drifted from the original; fix the helper, not the test.

- [ ] **Step 5: Commit, open the stage 3 PR, watch CI**

```bash
git add internal/tools/
git commit -m "refactor(tools): extract shared task action helpers"
git push -u origin refactor/stage-3-handler-dedup
gh pr create --title "refactor: stage 3 handler dedup" --body "Stage 3 of the tech debt refactor. Collapses 29 identical output structs into ActionOutput and extracts the repeated resolve/not-found/act/report flow into runTaskAction. Wire behavior byte-identical; no test assertions changed."
gh pr checks --watch
```

Expected: all checks pass. Then STOP: ask the user to review and merge before Stage 4.

---

### Task 7: Update field-clearing semantics (all of Stage 4, TDD)

**Files:**
- Modify: `internal/tools/tasks.go` (`UpdateTaskInput` fields + update handler mapping)
- Test: `internal/tools/tools_test.go`

**Interfaces:**
- Consumes: `UpdateTaskRequest` with `Labels *[]string` (Task 3), `resolveTask` (Task 6).
- Produces: new `UpdateTaskInput` field types (`Description *string`, `DueString *string`, `Labels *[]string`). Absent means untouched; empty means clear.

- [ ] **Step 1: Create the stage branch**

```bash
git checkout main && git pull
git checkout -b feat/stage-4-update-clearing
```

- [ ] **Step 2: Write the failing tests**

Append to `internal/tools/tools_test.go`. The helper router records the update request body for assertions:

```go
func captureUpdateBody(t *testing.T, rt *router, bodies *[]map[string]any) {
	t.Helper()
	rt.handle("POST", "/tasks/", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		*bodies = append(*bodies, body)
		_, _ = w.Write([]byte(`{"id":"10","content":"Task"}`))
	})
}

func TestUpdateTaskTool_ClearDescription(t *testing.T) {
	rt := newRouter()
	var bodies []map[string]any
	captureUpdateBody(t, rt, &bodies)
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	callTool(t, cs, "todoist_update_task", map[string]any{
		"task_id":     "10",
		"description": "",
	})

	if len(bodies) != 1 {
		t.Fatalf("got %d update requests", len(bodies))
	}
	desc, ok := bodies[0]["description"]
	if !ok {
		t.Fatal("description key missing: empty string should clear, not be dropped")
	}
	if desc != "" {
		t.Errorf("description = %v, want \"\"", desc)
	}
}

func TestUpdateTaskTool_ClearDueDate(t *testing.T) {
	for _, spelling := range []string{"", "no date"} {
		rt := newRouter()
		var bodies []map[string]any
		captureUpdateBody(t, rt, &bodies)
		cs, cleanup := setupTest(t, rt)

		callTool(t, cs, "todoist_update_task", map[string]any{
			"task_id":    "10",
			"due_string": spelling,
		})

		if len(bodies) != 1 {
			t.Fatalf("spelling %q: got %d update requests", spelling, len(bodies))
		}
		if got := bodies[0]["due_string"]; got != "no date" {
			t.Errorf("spelling %q: due_string = %v, want \"no date\"", spelling, got)
		}
		cleanup()
	}
}

func TestUpdateTaskTool_ClearLabels(t *testing.T) {
	rt := newRouter()
	var bodies []map[string]any
	captureUpdateBody(t, rt, &bodies)
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	callTool(t, cs, "todoist_update_task", map[string]any{
		"task_id": "10",
		"labels":  []string{},
	})

	if len(bodies) != 1 {
		t.Fatalf("got %d update requests", len(bodies))
	}
	labels, ok := bodies[0]["labels"]
	if !ok {
		t.Fatal("labels key missing: empty array should clear, not be dropped")
	}
	arr, ok := labels.([]any)
	if !ok || len(arr) != 0 {
		t.Errorf("labels = %v, want []", labels)
	}
}

func TestUpdateTaskTool_AbsentFieldsUntouched(t *testing.T) {
	rt := newRouter()
	var bodies []map[string]any
	captureUpdateBody(t, rt, &bodies)
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	callTool(t, cs, "todoist_update_task", map[string]any{
		"task_id": "10",
		"content": "Renamed",
	})

	if len(bodies) != 1 {
		t.Fatalf("got %d update requests", len(bodies))
	}
	for _, key := range []string{"description", "due_string", "labels"} {
		if _, ok := bodies[0][key]; ok {
			t.Errorf("%s key present, want absent", key)
		}
	}
}
```

Add `"encoding/json"` to the test file imports if not already there.

- [ ] **Step 3: Run the new tests to verify they fail**

Run: `go test ./internal/tools/ -race -run 'TestUpdateTaskTool_(ClearDescription|ClearDueDate|ClearLabels|AbsentFieldsUntouched)' -v`
Expected: `AbsentFieldsUntouched` passes; the three clear tests FAIL with "key missing" (today's empty-string guards drop the fields).

- [ ] **Step 4: Change the input types**

In `UpdateTaskInput`:

```go
Description *string   `json:"description,omitempty" jsonschema:"New description for the task. An empty string clears it (optional)"`
DueString   *string   `json:"due_string,omitempty" jsonschema:"New due date in natural language. An empty string or 'no date' removes the due date (optional)"`
Labels      *[]string `json:"labels,omitempty" jsonschema:"New labels for the task. An empty array removes all labels (optional)"`
```

`Content`, `Priority`, `AssigneeID`, `DeadlineDate` keep their current types and guards (clearing them is not meaningful in the Todoist API).

- [ ] **Step 5: Change the handler mapping**

Replace the three guards in the update handler:

```go
if input.Description != nil {
	updateReq.Description = input.Description
}
if input.DueString != nil {
	due := *input.DueString
	if due == "" {
		due = "no date" // Todoist's native syntax for removing a due date
	}
	updateReq.DueString = todoist.Ptr(due)
}
if input.Labels != nil {
	updateReq.Labels = input.Labels
}
```

- [ ] **Step 6: Run the new tests to verify they pass, then the full suite**

Run: `go test ./internal/tools/ -race -run TestUpdateTaskTool -v`
Expected: all PASS (including the pre-existing update test).
Run: `make check`
Expected: `All checks passed!`

- [ ] **Step 7: Commit, open the stage 4 PR, watch CI**

```bash
git add internal/tools/
git commit -m "feat(tools): support clearing description, due date, and labels on update"
git push -u origin feat/stage-4-update-clearing
gh pr create --title "feat: stage 4 update field clearing" --body "Stage 4 of the tech debt refactor. todoist_update_task now distinguishes absent from empty: empty description/labels clear the field, and an empty or 'no date' due_string removes the due date. Absent fields behave exactly as before."
gh pr checks --watch
```

Expected: all checks pass. Then STOP: ask the user to review and merge before Stage 5.

---

### Task 8: Retry and rate-limit handling in client.do (all of Stage 5, TDD)

**Files:**
- Modify: `internal/todoist/client.go`
- Test: `internal/todoist/client_test.go`

**Interfaces:**
- Consumes: ctx-aware `do` (Task 2).
- Produces: retrying `do` with unexported knobs the same-package tests may set:

```go
const maxAttempts = 3
const maxRetryAfter = 30 * time.Second

type Client struct {
	token        string
	baseURL      string
	httpClient   *http.Client
	retryBackoff []time.Duration // waits before attempts 2..n; tests may zero this
}

func parseRetryAfter(h string) (time.Duration, bool)
func sleepCtx(ctx context.Context, d time.Duration) error
```

- [ ] **Step 1: Create the stage branch**

```bash
git checkout main && git pull
git checkout -b feat/stage-5-client-retries
```

- [ ] **Step 2: Check existing tests for 5xx expectations**

Run: `grep -n "500\|StatusInternalServerError\|StatusBadRequest\|40[0-9]" internal/todoist/*_test.go`
Any existing test that returns a 5xx once and expects an immediate error will now be retried. For each such test, either switch the fake to a 4xx status (if it only tests "error surfaces") or zero the backoff and expect 3 requests. Note what you change in the commit body.

- [ ] **Step 3: Write the failing tests**

Append to `internal/todoist/client_test.go` (same package, so unexported fields and funcs are reachable):

```go
func TestDoRetriesOn429(t *testing.T) {
	var calls int
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	defer srv.Close()

	if _, err := c.GetTasks(context.Background(), ""); err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestDoFailsAfterMaxAttemptsOn500(t *testing.T) {
	var calls int
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()
	c.retryBackoff = []time.Duration{0, 0}

	_, err := c.GetTasks(context.Background(), "")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("err = %v, want status 500 mention", err)
	}
	if calls != maxAttempts {
		t.Errorf("calls = %d, want %d", calls, maxAttempts)
	}
}

func TestDoNoRetryOn4xx(t *testing.T) {
	var calls int
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
	})
	defer srv.Close()

	if _, err := c.GetTasks(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestDoContextCanceledDuringBackoff(t *testing.T) {
	var calls int
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()
	// default backoff (250ms) is longer than the deadline

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.GetTasks(ctx, "")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (canceled during first backoff)", calls)
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"0", 0, true},
		{"2", 2 * time.Second, true},
		{"60", maxRetryAfter, true}, // capped at 30s
		{"", 0, false},
		{"garbage", 0, false},
		{"-1", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseRetryAfter(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("parseRetryAfter(%q) = (%v, %v), want (%v, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
```

Add `"context"`, `"errors"`, `"strings"`, and `"time"` to the test imports as needed.

- [ ] **Step 4: Run the new tests to verify they fail**

Run: `go test ./internal/todoist/ -race -run 'TestDo(RetriesOn429|FailsAfterMaxAttemptsOn500|NoRetryOn4xx|ContextCanceledDuringBackoff)|TestParseRetryAfter' -v`
Expected: FAIL. `parseRetryAfter` is undefined (compile error) first; after stubbing, the retry tests fail because `do` never retries.

- [ ] **Step 5: Implement the retry loop**

In `internal/todoist/client.go`, add `"bytes"`, `"log/slog"`, and `"strconv"` to imports (drop `"strings"` if now unused) and replace `do` plus add the helpers:

```go
const maxAttempts = 3
const maxRetryAfter = 30 * time.Second

var defaultRetryBackoff = []time.Duration{250 * time.Millisecond, time.Second}
```

Add `retryBackoff []time.Duration` to the `Client` struct and set `retryBackoff: defaultRetryBackoff` in `NewClient`.

```go
// do executes an HTTP request against the Todoist API and returns the
// response body bytes. For responses with no content (204) it returns nil.
// 429 and 5xx responses are retried up to maxAttempts times; a 429's
// Retry-After header overrides the backoff schedule, capped at maxRetryAfter.
func (c *Client) do(ctx context.Context, method, endpoint string, body any) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBytes = data
	}

	var lastErr error
	var wait time.Duration
	for attempt := range maxAttempts {
		if attempt > 0 {
			if err := sleepCtx(ctx, wait); err != nil {
				return nil, err
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch {
		case resp.StatusCode == http.StatusNoContent:
			return nil, nil
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return respBody, nil
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
			wait = c.backoffFor(attempt)
			if resp.StatusCode == http.StatusTooManyRequests {
				if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
					wait = d
				}
			}
			slog.Debug("retrying todoist request", "status", resp.StatusCode, "attempt", attempt+1, "wait", wait.String())
		default:
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
	}
	return nil, lastErr
}

// backoffFor returns the wait before the attempt following `attempt`
// (0-indexed), reusing the last configured wait when out of range.
func (c *Client) backoffFor(attempt int) time.Duration {
	if len(c.retryBackoff) == 0 {
		return 0
	}
	if attempt < len(c.retryBackoff) {
		return c.retryBackoff[attempt]
	}
	return c.retryBackoff[len(c.retryBackoff)-1]
}

// parseRetryAfter parses a Retry-After header given in seconds. The
// second return is false when the header is absent or unusable.
func parseRetryAfter(h string) (time.Duration, bool) {
	if h == "" {
		return 0, false
	}
	secs, err := strconv.Atoi(strings.TrimSpace(h))
	if err != nil || secs < 0 {
		return 0, false
	}
	d := time.Duration(secs) * time.Second
	if d > maxRetryAfter {
		d = maxRetryAfter
	}
	return d, true
}

// sleepCtx waits for d or until ctx is done, whichever comes first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
```

Note the 204 case now precedes the generic 2xx return, preserving the original nil-body contract.

- [ ] **Step 6: Run the new tests, then the full suite**

Run: `go test ./internal/todoist/ -race -v`
Expected: all PASS, including the tests adjusted in Step 2.
Run: `make check`
Expected: `All checks passed!`

- [ ] **Step 7: Commit, open the stage 5 PR, watch CI**

```bash
git add internal/todoist/
git commit -m "feat(todoist): retry requests on 429 and 5xx with backoff"
git push -u origin feat/stage-5-client-retries
gh pr create --title "feat: stage 5 retry and rate-limit handling" --body "Stage 5 (final) of the tech debt refactor. client.do now retries 429 (honoring Retry-After, capped at 30s) and 5xx responses up to 3 attempts with 250ms/1s backoff, aborting promptly on context cancellation. Other 4xx statuses still fail immediately."
gh pr checks --watch
```

Expected: all checks pass. Then STOP: ask the user to review and merge. After merge, the refactor is complete; suggest deleting the merged branches.
