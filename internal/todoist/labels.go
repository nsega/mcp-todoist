package todoist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetLabels returns all personal labels.
func (c *Client) GetLabels(ctx context.Context) ([]models.Label, error) {
	data, err := c.do(ctx, "GET", "/labels", nil)
	if err != nil {
		return nil, err
	}

	var page PaginatedResponse[models.Label]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}
	return page.Results, nil
}

// CreateLabelRequest is the request body for creating a label.
type CreateLabelRequest struct {
	Name       string  `json:"name"`
	Color      *string `json:"color,omitempty"`
	IsFavorite *bool   `json:"is_favorite,omitempty"`
}

// UpdateLabelRequest is the request body for updating a label.
type UpdateLabelRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// CreateLabel creates a new personal label.
func (c *Client) CreateLabel(ctx context.Context, req CreateLabelRequest) (*models.Label, error) {
	data, err := c.do(ctx, "POST", "/labels", req)
	if err != nil {
		return nil, err
	}

	var label models.Label
	if err := json.Unmarshal(data, &label); err != nil {
		return nil, fmt.Errorf("failed to parse label: %w", err)
	}
	return &label, nil
}

// UpdateLabel updates an existing label.
func (c *Client) UpdateLabel(ctx context.Context, id string, req UpdateLabelRequest) (*models.Label, error) {
	data, err := c.do(ctx, "POST", "/labels/"+id, req)
	if err != nil {
		return nil, err
	}

	var label models.Label
	if err := json.Unmarshal(data, &label); err != nil {
		return nil, fmt.Errorf("failed to parse label: %w", err)
	}
	return &label, nil
}

// DeleteLabel deletes a label.
func (c *Client) DeleteLabel(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/labels/"+id, nil)
	return err
}
