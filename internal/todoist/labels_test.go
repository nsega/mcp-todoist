package todoist

import (
	"net/http"
	"testing"
)

func TestGetLabels(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/labels" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"l1","name":"urgent"}]`))
	})
	defer srv.Close()

	labels, err := c.GetLabels()
	if err != nil {
		t.Fatal(err)
	}
	if len(labels) != 1 || labels[0].Name != "urgent" {
		t.Errorf("unexpected labels: %+v", labels)
	}
}

func TestCreateLabel(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"id":"l2","name":"waiting"}`))
	})
	defer srv.Close()

	l, err := c.CreateLabel(map[string]interface{}{"name": "waiting"})
	if err != nil {
		t.Fatal(err)
	}
	if l.ID != "l2" {
		t.Errorf("id = %q", l.ID)
	}
}

func TestDeleteLabel(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/labels/l1" {
			t.Errorf("method=%s path=%s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.DeleteLabel("l1"); err != nil {
		t.Fatal(err)
	}
}
