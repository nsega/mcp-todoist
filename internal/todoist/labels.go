package todoist

import (
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetLabels returns all personal labels.
func (c *Client) GetLabels() ([]models.Label, error) {
	data, err := c.do("GET", "/labels", nil)
	if err != nil {
		return nil, err
	}

	var labels []models.Label
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}
	return labels, nil
}

// CreateLabel creates a new personal label.
func (c *Client) CreateLabel(body map[string]interface{}) (*models.Label, error) {
	data, err := c.do("POST", "/labels", body)
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
func (c *Client) UpdateLabel(id string, body map[string]interface{}) (*models.Label, error) {
	data, err := c.do("POST", "/labels/"+id, body)
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
func (c *Client) DeleteLabel(id string) error {
	_, err := c.do("DELETE", "/labels/"+id, nil)
	return err
}
