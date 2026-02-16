package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

// router dispatches to handlers based on method+path prefix.
type router struct {
	mu       sync.Mutex
	handlers []routeEntry
}

type routeEntry struct {
	method  string
	prefix  string
	handler http.HandlerFunc
}

func newRouter() *router {
	return &router{}
}

func (rt *router) handle(method, prefix string, h http.HandlerFunc) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.handlers = append(rt.handlers, routeEntry{method, prefix, h})
}

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mu.Lock()
	handlers := make([]routeEntry, len(rt.handlers))
	copy(handlers, rt.handlers)
	rt.mu.Unlock()

	// Match longest prefix first for specificity.
	var best *routeEntry
	for i := range handlers {
		e := &handlers[i]
		if r.Method == e.method && strings.HasPrefix(r.URL.Path, e.prefix) {
			if best == nil || len(e.prefix) > len(best.prefix) {
				best = e
			}
		}
	}
	if best != nil {
		best.handler(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func setupTest(t *testing.T, rt *router) (*mcp.ClientSession, func()) {
	t.Helper()
	apiSrv := httptest.NewServer(rt)
	client := todoist.NewClient("test-token", todoist.WithBaseURL(apiSrv.URL))

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	RegisterAll(mcpServer, client)

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()

	sSession, err := mcpServer.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	cSession, err := mcpClient.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		_ = cSession.Close()
		_ = sSession.Close()
		apiSrv.Close()
	}
	return cSession, cleanup
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]interface{}) *mcp.CallToolResult {
	t.Helper()
	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s) error: %v", name, err)
	}
	return result
}

func resultText(r *mcp.CallToolResult) string {
	for _, c := range r.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// --- Task tool tests ---

func TestCreateTaskTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"1","content":"Test task","priority":2}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_create_task", map[string]interface{}{
		"content":  "Test task",
		"priority": 2,
	})
	text := resultText(result)
	if !strings.Contains(text, "Task created") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestGetTasksTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"Task A","priority":1},{"id":"2","content":"Task B","priority":3}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_get_tasks", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "Task A") || !strings.Contains(text, "Task B") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestCompleteTaskTool_byID(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/tasks/42/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_complete_task", map[string]interface{}{
		"task_id": "42",
	})
	text := resultText(result)
	if !strings.Contains(text, "Successfully completed") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestDeleteTaskTool_notFound(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_delete_task", map[string]interface{}{
		"task_name": "nonexistent",
	})
	if !result.IsError {
		t.Error("expected IsError = true")
	}
	if !strings.Contains(resultText(result), "Could not find") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

func TestReopenTaskTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/tasks/7/reopen", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_reopen_task", map[string]interface{}{
		"task_id": "7",
	})
	text := resultText(result)
	if !strings.Contains(text, "Successfully reopened") {
		t.Errorf("unexpected result: %s", text)
	}
}

// --- Project tool tests ---

func TestGetProjectsTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/projects", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"p1","name":"Work"},{"id":"p2","name":"Personal","inbox_project":true}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_get_projects", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "Work") || !strings.Contains(text, "[Inbox]") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestCreateProjectTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/projects", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"p3","name":"New Project"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_create_project", map[string]interface{}{
		"name": "New Project",
	})
	if !strings.Contains(resultText(result), "Project created") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

func TestDeleteProjectTool(t *testing.T) {
	rt := newRouter()
	rt.handle("DELETE", "/projects/p1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_delete_project", map[string]interface{}{
		"project_id": "p1",
	})
	if !strings.Contains(resultText(result), "Successfully deleted project") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

// --- Label tool tests ---

func TestGetLabelsTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/labels", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"l1","name":"waiting","is_favorite":true}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_get_labels", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "waiting") || !strings.Contains(text, "[Favorite]") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestCreateLabelTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/labels", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"l2","name":"next-action"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_create_label", map[string]interface{}{
		"name": "next-action",
	})
	if !strings.Contains(resultText(result), "Label created") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

// --- Section tool tests ---

func TestGetSectionsTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/sections", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"s1","name":"Backlog","project_id":"p1"}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_get_sections", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "Backlog") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestCreateSectionTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/sections", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"s2","name":"In Progress","project_id":"p1"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_create_section", map[string]interface{}{
		"name":       "In Progress",
		"project_id": "p1",
	})
	if !strings.Contains(resultText(result), "Section created") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

// --- Comment tool tests ---

func TestCreateCommentTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/comments", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"c1","content":"A note","task_id":"42"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_create_comment", map[string]interface{}{
		"content": "A note",
		"task_id": "42",
	})
	if !strings.Contains(resultText(result), "Comment created") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

func TestGetCommentsTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/comments", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"c1","content":"A comment","task_id":"42","posted_at":"2025-01-15T10:30:00Z"}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_get_comments", map[string]interface{}{
		"task_id": "42",
	})
	if !strings.Contains(resultText(result), "A comment") {
		t.Errorf("unexpected result: %s", resultText(result))
	}
}

// --- GTD tool tests ---

func TestInboxReviewTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/projects", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"inbox1","name":"Inbox","inbox_project":true}],"next_cursor":""}`))
	})
	rt.handle("GET", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"Old task","project_id":"inbox1","added_at":"2020-01-01T00:00:00Z"}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_inbox_review", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "Inbox Review") || !strings.Contains(text, "Old task") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestWeeklyReviewTool(t *testing.T) {
	rt := newRouter()
	rt.handle("GET", "/projects", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"p1","name":"Work"}],"next_cursor":""}`))
	})
	rt.handle("GET", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"A task","project_id":"p1"}],"next_cursor":""}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_weekly_review", map[string]interface{}{})
	text := resultText(result)
	if !strings.Contains(text, "Weekly Review") || !strings.Contains(text, "Work") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestBulkCreateTasksTool(t *testing.T) {
	rt := newRouter()
	callCount := 0
	rt.handle("POST", "/tasks", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_, _ = w.Write([]byte(`{"id":"` + string(rune('0'+callCount)) + `","content":"task"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_bulk_create_tasks", map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"content": "Task 1"},
			{"content": "Task 2"},
		},
	})
	text := resultText(result)
	if !strings.Contains(text, "2 created") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestMoveTaskTool(t *testing.T) {
	rt := newRouter()
	rt.handle("POST", "/tasks/42", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"42","content":"Moved task","project_id":"p2"}`))
	})
	cs, cleanup := setupTest(t, rt)
	defer cleanup()

	result := callTool(t, cs, "todoist_move_task", map[string]interface{}{
		"task_id":    "42",
		"project_id": "p2",
	})
	text := resultText(result)
	if !strings.Contains(text, "Successfully moved") {
		t.Errorf("unexpected result: %s", text)
	}
}
