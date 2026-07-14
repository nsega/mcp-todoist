package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetComments returns comments for a task or project.
// Exactly one of taskID or projectID should be non-empty.
func (c *Client) GetComments(ctx context.Context, taskID, projectID string) ([]models.Comment, error) {
	endpoint := "/comments"
	values := url.Values{}
	if taskID != "" {
		values.Set("task_id", taskID)
	} else if projectID != "" {
		values.Set("project_id", projectID)
	}
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	data, err := c.do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var page PaginatedResponse[models.Comment]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}
	return page.Results, nil
}

// CreateComment creates a new comment.
func (c *Client) CreateComment(ctx context.Context, body map[string]any) (*models.Comment, error) {
	data, err := c.do(ctx, "POST", "/comments", body)
	if err != nil {
		return nil, err
	}

	var comment models.Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment: %w", err)
	}
	return &comment, nil
}

// UpdateComment updates an existing comment.
func (c *Client) UpdateComment(ctx context.Context, id string, body map[string]any) (*models.Comment, error) {
	data, err := c.do(ctx, "POST", "/comments/"+id, body)
	if err != nil {
		return nil, err
	}

	var comment models.Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment: %w", err)
	}
	return &comment, nil
}

// DeleteComment deletes a comment.
func (c *Client) DeleteComment(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/comments/"+id, nil)
	return err
}
