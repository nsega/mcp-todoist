package todoist

import (
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetSections returns sections, optionally filtered by project.
func (c *Client) GetSections(projectID string) ([]models.Section, error) {
	endpoint := "/sections"
	if projectID != "" {
		endpoint += "?project_id=" + projectID
	}

	data, err := c.do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var sections []models.Section
	if err := json.Unmarshal(data, &sections); err != nil {
		return nil, fmt.Errorf("failed to parse sections: %w", err)
	}
	return sections, nil
}

// CreateSection creates a new section.
func (c *Client) CreateSection(body map[string]interface{}) (*models.Section, error) {
	data, err := c.do("POST", "/sections", body)
	if err != nil {
		return nil, err
	}

	var section models.Section
	if err := json.Unmarshal(data, &section); err != nil {
		return nil, fmt.Errorf("failed to parse section: %w", err)
	}
	return &section, nil
}

// UpdateSection updates an existing section.
func (c *Client) UpdateSection(id string, body map[string]interface{}) (*models.Section, error) {
	data, err := c.do("POST", "/sections/"+id, body)
	if err != nil {
		return nil, err
	}

	var section models.Section
	if err := json.Unmarshal(data, &section); err != nil {
		return nil, fmt.Errorf("failed to parse section: %w", err)
	}
	return &section, nil
}

// DeleteSection deletes a section.
func (c *Client) DeleteSection(id string) error {
	_, err := c.do("DELETE", "/sections/"+id, nil)
	return err
}
