package todoist

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

	_, err := c.do(context.Background(), "GET", "/test", nil)
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

	_, err := c.do(context.Background(), "GET", "/test", nil)
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

	_, err := c.do(context.Background(), "POST", "/test", map[string]string{"key": "val"})
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestDoSendsRequestIDOnPost(t *testing.T) {
	var reqID string
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		reqID = r.Header.Get("X-Request-Id")
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	if _, err := c.do(context.Background(), "POST", "/test", map[string]string{"key": "val"}); err != nil {
		t.Fatal(err)
	}
	if reqID == "" {
		t.Fatal("X-Request-Id header missing on POST")
	}
	if len(reqID) != 36 || strings.Count(reqID, "-") != 4 {
		t.Errorf("X-Request-Id = %q, want UUID format", reqID)
	}
}

func TestDoRequestIDStableAcrossRetries(t *testing.T) {
	var ids []string
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		ids = append(ids, r.Header.Get("X-Request-Id"))
		if len(ids) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	c.retryBackoff = []time.Duration{0, 0}

	if _, err := c.do(context.Background(), "POST", "/test", map[string]string{"key": "val"}); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("calls = %d, want 2", len(ids))
	}
	if ids[0] == "" {
		t.Fatal("X-Request-Id header missing on first attempt")
	}
	if ids[0] != ids[1] {
		t.Errorf("X-Request-Id changed across retries: %q then %q, want identical", ids[0], ids[1])
	}
}

func TestDoNoRequestIDOnGet(t *testing.T) {
	var reqID string
	got := false
	c, srv := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		reqID = r.Header.Get("X-Request-Id")
		got = true
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	if _, err := c.do(context.Background(), "GET", "/test", nil); err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Fatal("request never reached server")
	}
	if reqID != "" {
		t.Errorf("X-Request-Id = %q on GET, want empty (idempotency header is POST-only)", reqID)
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
