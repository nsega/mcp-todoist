package todoist

import (
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetComments returns comments for a task or project.
// Exactly one of taskID or projectID should be non-empty.
func (c *Client) GetComments(taskID, projectID string) ([]models.Comment, error) {
	endpoint := "/comments?"
	if taskID != "" {
		endpoint += "task_id=" + taskID
	} else if projectID != "" {
		endpoint += "project_id=" + projectID
	}

	data, err := c.do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var comments []models.Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}
	return comments, nil
}

// CreateComment creates a new comment.
func (c *Client) CreateComment(body map[string]interface{}) (*models.Comment, error) {
	data, err := c.do("POST", "/comments", body)
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
func (c *Client) UpdateComment(id string, body map[string]interface{}) (*models.Comment, error) {
	data, err := c.do("POST", "/comments/"+id, body)
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
func (c *Client) DeleteComment(id string) error {
	_, err := c.do("DELETE", "/comments/"+id, nil)
	return err
}
