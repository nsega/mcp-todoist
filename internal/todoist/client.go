package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.todoist.com/api/v1"

const maxAttempts = 3
const maxRetryAfter = 30 * time.Second

var defaultRetryBackoff = []time.Duration{250 * time.Millisecond, time.Second}

// Client is an HTTP client for the Todoist API v1.
type Client struct {
	token        string
	baseURL      string
	httpClient   *http.Client
	retryBackoff []time.Duration // waits before attempts 2..n; tests may zero this
}

// PaginatedResponse wraps list endpoints in the Todoist API v1.
type PaginatedResponse[T any] struct {
	Results    []T    `json:"results"`
	NextCursor string `json:"next_cursor"`
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(cl *Client) { cl.baseURL = url }
}

// NewClient creates a new Todoist API client.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:        token,
		baseURL:      defaultBaseURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		retryBackoff: defaultRetryBackoff,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// do executes an HTTP request against the Todoist API and returns the
// response body bytes. For responses with no content (204) it returns nil.
// 429 and 5xx responses are retried up to maxAttempts times; a 429's
// Retry-After header overrides the backoff schedule, capped at maxRetryAfter.
func (c *Client) do(ctx context.Context, method, endpoint string, body any) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBytes = data
	}

	var lastErr error
	var wait time.Duration
	for attempt := range maxAttempts {
		if attempt > 0 {
			if err := sleepCtx(ctx, wait); err != nil {
				return nil, err
			}
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch {
		case resp.StatusCode == http.StatusNoContent:
			return nil, nil
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return respBody, nil
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
			wait = c.backoffFor(attempt)
			if resp.StatusCode == http.StatusTooManyRequests {
				if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
					wait = d
				}
			}
			slog.Debug("retrying todoist request", "status", resp.StatusCode, "attempt", attempt+1, "wait", wait.String())
		default:
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
	}
	return nil, lastErr
}

// backoffFor returns the wait before the attempt following `attempt`
// (0-indexed), reusing the last configured wait when out of range.
func (c *Client) backoffFor(attempt int) time.Duration {
	if len(c.retryBackoff) == 0 {
		return 0
	}
	if attempt < len(c.retryBackoff) {
		return c.retryBackoff[attempt]
	}
	return c.retryBackoff[len(c.retryBackoff)-1]
}

// parseRetryAfter parses a Retry-After header given in seconds. The
// second return is false when the header is absent or unusable.
func parseRetryAfter(h string) (time.Duration, bool) {
	if h == "" {
		return 0, false
	}
	secs, err := strconv.Atoi(strings.TrimSpace(h))
	if err != nil || secs < 0 {
		return 0, false
	}
	d := min(time.Duration(secs)*time.Second, maxRetryAfter)
	return d, true
}

// sleepCtx waits for d or until ctx is done, whichever comes first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
