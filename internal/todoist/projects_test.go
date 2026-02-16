package todoist

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetProjects(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"100","name":"Inbox","is_inbox_project":true}]`))
	})
	defer srv.Close()

	projects, err := c.GetProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || !projects[0].IsInboxProject {
		t.Errorf("unexpected projects: %+v", projects)
	}
}

func TestGetProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/200" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"200","name":"Work"}`))
	})
	defer srv.Close()

	p, err := c.GetProject("200")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Work" {
		t.Errorf("name = %q", p.Name)
	}
}

func TestCreateProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["name"] != "New Project" {
			t.Errorf("name = %v", body["name"])
		}
		_, _ = w.Write([]byte(`{"id":"300","name":"New Project"}`))
	})
	defer srv.Close()

	p, err := c.CreateProject(map[string]interface{}{"name": "New Project"})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "300" {
		t.Errorf("id = %q", p.ID)
	}
}

func TestDeleteProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/projects/300" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.DeleteProject("300"); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/300/archive" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.ArchiveProject("300"); err != nil {
		t.Fatal(err)
	}
}

func TestUnarchiveProject(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/projects/300/unarchive" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.UnarchiveProject("300"); err != nil {
		t.Fatal(err)
	}
}
