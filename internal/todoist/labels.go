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

// CreateLabel creates a new personal label.
func (c *Client) CreateLabel(ctx context.Context, body map[string]any) (*models.Label, error) {
	data, err := c.do(ctx, "POST", "/labels", body)
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
func (c *Client) UpdateLabel(ctx context.Context, id string, body map[string]any) (*models.Label, error) {
	data, err := c.do(ctx, "POST", "/labels/"+id, body)
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
