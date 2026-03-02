package todoist

import (
	"net/http"
	"testing"
)

func TestGetComments(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("task_id"); q != "42" {
			t.Errorf("task_id = %q", q)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"c1","content":"A comment","task_id":"42"}],"next_cursor":""}`))
	})
	defer srv.Close()

	comments, err := c.GetComments("42", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 || comments[0].Content != "A comment" {
		t.Errorf("unexpected comments: %+v", comments)
	}
}

func TestCreateComment(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"id":"c2","content":"New comment","task_id":"42"}`))
	})
	defer srv.Close()

	cm, err := c.CreateComment(map[string]any{"content": "New comment", "task_id": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if cm.ID != "c2" {
		t.Errorf("id = %q", cm.ID)
	}
}

func TestGetComments_noParams(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		// When neither taskID nor projectID is set, the URL must not
		// have a trailing '?' (which the old code produced).
		if r.URL.RawQuery != "" {
			t.Errorf("expected empty query string, got %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	defer srv.Close()

	_, err := c.GetComments("", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetComments_byProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/comments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("project_id"); q != "p1" {
			t.Errorf("project_id = %q", q)
		}
		if q := r.URL.Query().Get("task_id"); q != "" {
			t.Errorf("unexpected task_id = %q", q)
		}
		_, _ = w.Write([]byte(`{"results":[],"next_cursor":""}`))
	})
	defer srv.Close()

	_, err := c.GetComments("", "p1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteComment(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/comments/c1" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.DeleteComment("c1"); err != nil {
		t.Fatal(err)
	}
}
