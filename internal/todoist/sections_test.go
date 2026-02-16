package todoist

import (
	"net/http"
	"testing"
)

func TestGetSections(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sections" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("project_id"); q != "123" {
			t.Errorf("project_id = %q", q)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"s1","name":"Backlog","project_id":"123"}],"next_cursor":""}`))
	})
	defer srv.Close()

	sections, err := c.GetSections("123")
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 1 || sections[0].Name != "Backlog" {
		t.Errorf("unexpected sections: %+v", sections)
	}
}

func TestCreateSection(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sections" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"s2","name":"In Progress","project_id":"123"}`))
	})
	defer srv.Close()

	sec, err := c.CreateSection(map[string]interface{}{"name": "In Progress", "project_id": "123"})
	if err != nil {
		t.Fatal(err)
	}
	if sec.ID != "s2" {
		t.Errorf("id = %q", sec.ID)
	}
}

func TestDeleteSection(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/sections/s1" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.DeleteSection("s1"); err != nil {
		t.Fatal(err)
	}
}
