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
		if !strings.HasPrefix(r.URL.Path, "/tasks") {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"1","content":"Test task","priority":1}]`))
	})
	defer srv.Close()

	tasks, err := c.GetTasks("", "")
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

func TestGetTasks_withFilters(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("project_id") != "123" {
			t.Errorf("project_id = %q", q.Get("project_id"))
		}
		if q.Get("filter") != "today" {
			t.Errorf("filter = %q", q.Get("filter"))
		}
		_, _ = w.Write([]byte(`[]`))
	})
	defer srv.Close()

	_, err := c.GetTasks("123", "today")
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
		var body map[string]interface{}
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

	task, err := c.CreateTask(map[string]interface{}{"content": "New task"})
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

	task, err := c.UpdateTask("10", map[string]interface{}{"content": "Updated"})
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

func TestFindTaskByName_exactMatch(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"1","content":"Buy groceries"},
			{"id":"2","content":"Buy groceries and milk"}
		]`))
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
		_, _ = w.Write([]byte(`[{"id":"3","content":"Weekly team meeting"}]`))
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
		_, _ = w.Write([]byte(`[{"id":"1","content":"Some task"}]`))
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
