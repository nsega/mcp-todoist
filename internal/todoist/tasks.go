package todoist

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetTasks returns active tasks, optionally filtered.
func (c *Client) GetTasks(projectID, filter string) ([]models.Task, error) {
	endpoint := "/tasks"
	values := url.Values{}
	if projectID != "" {
		values.Set("project_id", projectID)
	}
	if filter != "" {
		values.Set("filter", filter)
	}
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	data, err := c.do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var page PaginatedResponse[models.Task]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}
	return page.Results, nil
}

// GetTask returns a single task by ID.
func (c *Client) GetTask(id string) (*models.Task, error) {
	data, err := c.do("GET", "/tasks/"+id, nil)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}
	return &task, nil
}

// CreateTask creates a new task.
func (c *Client) CreateTask(body map[string]any) (*models.Task, error) {
	data, err := c.do("POST", "/tasks", body)
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
func (c *Client) UpdateTask(id string, body map[string]any) (*models.Task, error) {
	data, err := c.do("POST", "/tasks/"+id, body)
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
func (c *Client) DeleteTask(id string) error {
	_, err := c.do("DELETE", "/tasks/"+id, nil)
	return err
}

// CloseTask marks a task as complete.
func (c *Client) CloseTask(id string) error {
	_, err := c.do("POST", "/tasks/"+id+"/close", nil)
	return err
}

// ReopenTask reopens a completed task.
func (c *Client) ReopenTask(id string) error {
	_, err := c.do("POST", "/tasks/"+id+"/reopen", nil)
	return err
}

var collapseWS = regexp.MustCompile(`\s+`)

// normalizeWhitespace trims and collapses runs of whitespace to a single space.
func normalizeWhitespace(s string) string {
	return collapseWS.ReplaceAllString(strings.TrimSpace(s), " ")
}

// FindTaskByName searches for a task by partial name matching.
// Returns nil if no match is found.
func (c *Client) FindTaskByName(name string) (*models.Task, error) {
	tasks, err := c.GetTasks("", "")
	if err != nil {
		return nil, err
	}

	norm := strings.ToLower(normalizeWhitespace(name))

	// Prefer exact match first.
	for i := range tasks {
		if strings.ToLower(normalizeWhitespace(tasks[i].Content)) == norm {
			return &tasks[i], nil
		}
	}

	// Fall back to partial match.
	for i := range tasks {
		if strings.Contains(strings.ToLower(normalizeWhitespace(tasks[i].Content)), norm) {
			return &tasks[i], nil
		}
	}

	return nil, nil
}
