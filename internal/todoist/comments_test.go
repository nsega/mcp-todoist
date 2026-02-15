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
		_, _ = w.Write([]byte(`[{"id":"c1","content":"A comment","task_id":"42"}]`))
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

	cm, err := c.CreateComment(map[string]interface{}{"content": "New comment", "task_id": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if cm.ID != "c2" {
		t.Errorf("id = %q", cm.ID)
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
