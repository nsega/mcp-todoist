package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/nsega/mcp-todoist/internal/models"
)

// filteredTasksResponse wraps the /tasks/filter endpoint response.
// The documented key is "items", but we also accept "results" as a fallback
// in case the API returns the same shape as the /tasks endpoint.
type filteredTasksResponse struct {
	Items      []models.Task `json:"items"`
	Results    []models.Task `json:"results"`
	NextCursor string        `json:"next_cursor"`
}

// tasks returns whichever array is populated.
func (r *filteredTasksResponse) tasks() []models.Task {
	if len(r.Items) > 0 {
		return r.Items
	}
	return r.Results
}

// getTasksPage returns one page of active tasks plus the cursor for the next page.
func (c *Client) getTasksPage(ctx context.Context, projectID, cursor string) ([]models.Task, string, error) {
	endpoint := "/tasks"
	values := url.Values{}
	if projectID != "" {
		values.Set("project_id", projectID)
	}
	if cursor != "" {
		values.Set("cursor", cursor)
	}
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	data, err := c.do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	var page PaginatedResponse[models.Task]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, "", fmt.Errorf("failed to parse tasks: %w", err)
	}
	return page.Results, page.NextCursor, nil
}

// GetTasks returns the first page of active tasks, optionally filtered by project.
func (c *Client) GetTasks(ctx context.Context, projectID string) ([]models.Task, error) {
	tasks, _, err := c.getTasksPage(ctx, projectID, "")
	return tasks, err
}

// getFilteredTasksPage returns one page of filtered tasks plus the cursor for the next page.
func (c *Client) getFilteredTasksPage(ctx context.Context, query, cursor string) ([]models.Task, string, error) {
	values := url.Values{}
	values.Set("query", query)
	if cursor != "" {
		values.Set("cursor", cursor)
	}
	endpoint := "/tasks/filter?" + values.Encode()

	slog.Debug("todoist filter request", "endpoint", endpoint)

	data, err := c.do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	slog.Debug("todoist filter response", "body", string(data))

	var resp filteredTasksResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("failed to parse filtered tasks: %w", err)
	}
	return resp.tasks(), resp.NextCursor, nil
}

// GetTasksByFilter returns tasks matching a Todoist filter query using the
// dedicated /tasks/filter endpoint. It paginates up to maxFindPages pages.
func (c *Client) GetTasksByFilter(ctx context.Context, query string) ([]models.Task, error) {
	var all []models.Task
	cursor := ""
	for range maxFindPages {
		tasks, nextCursor, err := c.getFilteredTasksPage(ctx, query, cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, tasks...)
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return all, nil
}

// GetTask returns a single task by ID.
func (c *Client) GetTask(ctx context.Context, id string) (*models.Task, error) {
	data, err := c.do(ctx, "GET", "/tasks/"+id, nil)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return &task, nil
}

// CreateTaskRequest is the request body for creating a task.
type CreateTaskRequest struct {
	Content     string   `json:"content"`
	Description *string  `json:"description,omitempty"`
	DueString   *string  `json:"due_string,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	ProjectID   *string  `json:"project_id,omitempty"`
	SectionID   *string  `json:"section_id,omitempty"`
	ParentID    *string  `json:"parent_id,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	AssigneeID  *string  `json:"assignee_id,omitempty"`
}

// UpdateTaskRequest is the request body for updating a task.
//
// Labels is *[]string so a pointer to an empty slice can express
// "clear all labels"; omitempty on a plain slice would drop it.
type UpdateTaskRequest struct {
	Content      *string   `json:"content,omitempty"`
	Description  *string   `json:"description,omitempty"`
	DueString    *string   `json:"due_string,omitempty"`
	Priority     *int      `json:"priority,omitempty"`
	Labels       *[]string `json:"labels,omitempty"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
	DeadlineDate *string   `json:"deadline_date,omitempty"`
}

// MoveTaskRequest moves a task. Set exactly one field.
type MoveTaskRequest struct {
	ProjectID *string `json:"project_id,omitempty"`
	SectionID *string `json:"section_id,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
}

// CreateTask creates a new task.
func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (*models.Task, error) {
	data, err := c.do(ctx, "POST", "/tasks", req)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return &task, nil
}

// UpdateTask updates an existing task.
func (c *Client) UpdateTask(ctx context.Context, id string, req UpdateTaskRequest) (*models.Task, error) {
	data, err := c.do(ctx, "POST", "/tasks/"+id, req)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return &task, nil
}

// MoveTask moves a task to another project, section, or parent.
// The request must set exactly one of ProjectID, SectionID, ParentID.
func (c *Client) MoveTask(ctx context.Context, id string, req MoveTaskRequest) (*models.Task, error) {
	data, err := c.do(ctx, "POST", "/tasks/"+id+"/move", req)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return &task, nil
}

// DeleteTask deletes a task.
func (c *Client) DeleteTask(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/tasks/"+id, nil)
	return err
}

// CloseTask marks a task as complete.
func (c *Client) CloseTask(ctx context.Context, id string) error {
	_, err := c.do(ctx, "POST", "/tasks/"+id+"/close", nil)
	return err
}

// ReopenTask reopens a completed task.
func (c *Client) ReopenTask(ctx context.Context, id string) error {
	_, err := c.do(ctx, "POST", "/tasks/"+id+"/reopen", nil)
	return err
}

var collapseWS = regexp.MustCompile(`\s+`)

// normalizeWhitespace trims and collapses runs of whitespace to a single space.
func normalizeWhitespace(s string) string {
	return collapseWS.ReplaceAllString(strings.TrimSpace(s), " ")
}

// maxFindPages limits the number of pages FindTaskByName will fetch
// to avoid unbounded iteration (20 pages × 50 tasks/page = 1000 tasks).
const maxFindPages = 20

// FindTaskByName searches for a task by partial name matching across all pages.
// Exact matches take priority over partial matches. Returns nil if no match is found.
func (c *Client) FindTaskByName(ctx context.Context, name string) (*models.Task, error) {
	norm := strings.ToLower(normalizeWhitespace(name))

	var partial *models.Task
	cursor := ""
	for range maxFindPages {
		tasks, nextCursor, err := c.getTasksPage(ctx, "", cursor)
		if err != nil {
			return nil, err
		}

		for i := range tasks {
			content := strings.ToLower(normalizeWhitespace(tasks[i].Content))
			if content == norm {
				return &tasks[i], nil // exact match — return immediately
			}
			if partial == nil && strings.Contains(content, norm) {
				t := tasks[i]
				partial = &t
			}
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return partial, nil
}
