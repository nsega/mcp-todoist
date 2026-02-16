package todoist

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := NewClient("test-token", WithBaseURL(srv.URL))
	return c, srv
}

func TestNewClient_defaults(t *testing.T) {
	c := NewClient("tok")
	if c.token != "tok" {
		t.Errorf("token = %q, want %q", c.token, "tok")
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
}

func TestNewClient_withOptions(t *testing.T) {
	hc := &http.Client{}
	c := NewClient("tok", WithHTTPClient(hc), WithBaseURL("http://example.com"))
	if c.httpClient != hc {
		t.Error("WithHTTPClient not applied")
	}
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestDo_setsAuthHeader(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Authorization = %q", auth)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	_, err := c.do("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDo_errorStatus(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid token"}`))
	})
	defer srv.Close()

	_, err := c.do("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestDo_sendsJSONBody(t *testing.T) {
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	_, err := c.do("POST", "/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatal(err)
	}
}
