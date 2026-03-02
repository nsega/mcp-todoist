package todoist

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGetTasks(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/tasks" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"Test task","priority":1}],"next_cursor":""}`))
	})
	defer srv.Close()

	tasks, err := c.GetTasks("")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks", len(tasks))
	}
	if tasks[0].Content != "Test task" {
		t.Errorf("content = %q", tasks[0].Content)
	}
}

func TestGetTasks_withProjectID(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("project_id") != "123" {
			t.Errorf("project_id = %q", q.Get("project_id"))
		}
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	defer srv.Close()

	_, err := c.GetTasks("123")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/42" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"42","content":"Single task"}`))
	})
	defer srv.Close()

	task, err := c.GetTask("42")
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != "42" {
		t.Errorf("id = %q", task.ID)
	}
}

func TestCreateTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["content"] != "New task" {
			t.Errorf("content = %v", body["content"])
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"99","content":"New task"}`))
	})
	defer srv.Close()

	task, err := c.CreateTask(map[string]any{"content": "New task"})
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != "99" {
		t.Errorf("id = %q", task.ID)
	}
}

func TestUpdateTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tasks/10" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"10","content":"Updated"}`))
	})
	defer srv.Close()

	task, err := c.UpdateTask("10", map[string]any{"content": "Updated"})
	if err != nil {
		t.Fatal(err)
	}
	if task.Content != "Updated" {
		t.Errorf("content = %q", task.Content)
	}
}

func TestDeleteTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/tasks/5" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.DeleteTask("5"); err != nil {
		t.Fatal(err)
	}
}

func TestCloseTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tasks/7/close" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.CloseTask("7"); err != nil {
		t.Fatal(err)
	}
}

func TestReopenTask(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tasks/7/reopen" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.ReopenTask("7"); err != nil {
		t.Fatal(err)
	}
}

// --- GetTasksByFilter tests ---

func TestGetTasksByFilter(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/filter" {
			t.Errorf("path = %s, want /tasks/filter", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("query"); got != "today" {
			t.Errorf("query = %q, want %q", got, "today")
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"1","content":"Due today","priority":1}],"next_cursor":""}`))
	})
	defer srv.Close()

	tasks, err := c.GetTasksByFilter("today")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks", len(tasks))
	}
	if tasks[0].Content != "Due today" {
		t.Errorf("content = %q", tasks[0].Content)
	}
}

func TestGetTasksByFilter_specialChars(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks/filter" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("query"); got != "today & @deep-research" {
			t.Errorf("query = %q, want %q", got, "today & @deep-research")
		}
		// Only one parameter (query), so no literal & should appear in raw query.
		if strings.Count(r.URL.RawQuery, "&") != 0 {
			t.Errorf("raw query contains unencoded &: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"items":[],"next_cursor":""}`))
	})
	defer srv.Close()

	_, err := c.GetTasksByFilter("today & @deep-research")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTasksByFilter_hashFilter(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("query"); got != "priority 1 & #Work" {
			t.Errorf("query = %q, want %q", got, "priority 1 & #Work")
		}
		_, _ = w.Write([]byte(`{"items":[],"next_cursor":""}`))
	})
	defer srv.Close()

	_, err := c.GetTasksByFilter("priority 1 & #Work")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTasksByFilter_labelFilter(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("query"); got != "@deep-research" {
			t.Errorf("query = %q, want %q", got, "@deep-research")
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"10","content":"Research task","labels":["deep-research"]}],"next_cursor":""}`))
	})
	defer srv.Close()

	tasks, err := c.GetTasksByFilter("@deep-research")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != "10" {
		t.Errorf("expected 1 task with id=10, got %+v", tasks)
	}
}

func TestGetTasksByFilter_apiError(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid filter"}`))
	})
	defer srv.Close()

	_, err := c.GetTasksByFilter(")))invalid(((")
	if err == nil {
		t.Fatal("expected error for bad filter, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 in error, got: %v", err)
	}
}

func TestGetTasksByFilter_emptyResult(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[],"next_cursor":""}`))
	})
	defer srv.Close()

	tasks, err := c.GetTasksByFilter("@nonexistent-label")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

// --- FindTaskByName tests ---

func TestFindTaskByName_exactMatch(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[
			{"id":"1","content":"Buy groceries"},
			{"id":"2","content":"Buy groceries and milk"}
		],"next_cursor":""}`))
	})
	defer srv.Close()

	task, err := c.FindTaskByName("Buy groceries")
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "1" {
		t.Errorf("expected exact match id=1, got %+v", task)
	}
}

func TestFindTaskByName_partialMatch(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"3","content":"Weekly team meeting"}],"next_cursor":""}`))
	})
	defer srv.Close()

	task, err := c.FindTaskByName("team meeting")
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "3" {
		t.Errorf("expected partial match, got %+v", task)
	}
}

func TestFindTaskByName_notFound(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"Some task"}],"next_cursor":""}`))
	})
	defer srv.Close()

	task, err := c.FindTaskByName("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if task != nil {
		t.Errorf("expected nil, got %+v", task)
	}
}

func TestFindTaskByName_whitespaceNormalization(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[
			{"id":"1","content":"  Buy   groceries  and   milk  "}
		],"next_cursor":""}`))
	})
	defer srv.Close()

	task, err := c.FindTaskByName("Buy groceries and milk")
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "1" {
		t.Errorf("expected whitespace-normalized match id=1, got %+v", task)
	}
}

func TestFindTaskByName_longTaskName(t *testing.T) {
	const longName = "Kubernetes Gateway API vs Ingress patterns for production GKE clusters"
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[
			{"id":"42","content":"` + longName + `"}
		],"next_cursor":""}`))
	})
	defer srv.Close()

	task, err := c.FindTaskByName(longName)
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "42" {
		t.Errorf("expected exact match id=42 for long name, got %+v", task)
	}
}

func TestFindTaskByName_pagination(t *testing.T) {
	page := 0
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		switch {
		case cursor == "" && page == 0:
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"1","content":"First page task"}
			],"next_cursor":"page2"}`))
		case cursor == "page2":
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"2","content":"Second page task"},
				{"id":"3","content":"TEST: verify MCP bug fixes"}
			],"next_cursor":""}`))
		default:
			t.Errorf("unexpected cursor = %q, page = %d", cursor, page)
			_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
		}
	})
	defer srv.Close()

	task, err := c.FindTaskByName("TEST: verify MCP bug fixes")
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "3" {
		t.Errorf("expected match on page 2 id=3, got %+v", task)
	}
	if page != 2 {
		t.Errorf("expected 2 pages fetched, got %d", page)
	}
}

func TestFindTaskByName_exactMatchPriorityAcrossPages(t *testing.T) {
	// Partial match on page 1, exact match on page 2 — exact should win.
	page := 0
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		switch {
		case cursor == "" && page == 0:
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"1","content":"Buy groceries and milk"}
			],"next_cursor":"page2"}`))
		case cursor == "page2":
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"2","content":"Buy groceries"}
			],"next_cursor":""}`))
		default:
			_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
		}
	})
	defer srv.Close()

	task, err := c.FindTaskByName("Buy groceries")
	if err != nil {
		t.Fatal(err)
	}
	if task == nil || task.ID != "2" {
		t.Errorf("expected exact match id=2 over partial id=1, got %+v", task)
	}
}

func TestFindTaskByName_notFoundAcrossPages(t *testing.T) {
	page := 0
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		switch {
		case cursor == "" && page == 0:
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"1","content":"Alpha task"}
			],"next_cursor":"page2"}`))
		case cursor == "page2":
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"2","content":"Beta task"}
			],"next_cursor":""}`))
		default:
			_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
		}
	})
	defer srv.Close()

	task, err := c.FindTaskByName("nonexistent across pages")
	if err != nil {
		t.Fatal(err)
	}
	if task != nil {
		t.Errorf("expected nil after exhausting all pages, got %+v", task)
	}
	if page != 2 {
		t.Errorf("expected both pages fetched, got %d", page)
	}
}

func TestFindTaskByName_apiErrorOnSecondPage(t *testing.T) {
	page := 0
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		switch {
		case cursor == "" && page == 0:
			page++
			_, _ = w.Write([]byte(`{"results":[
				{"id":"1","content":"First page task"}
			],"next_cursor":"page2"}`))
		case cursor == "page2":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		default:
			_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
		}
	})
	defer srv.Close()

	_, err := c.FindTaskByName("Target task")
	if err == nil {
		t.Fatal("expected error when second page fails, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}
