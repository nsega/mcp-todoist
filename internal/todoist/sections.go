package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetSections returns sections, optionally filtered by project.
func (c *Client) GetSections(ctx context.Context, projectID string) ([]models.Section, error) {
	endpoint := "/sections"
	if projectID != "" {
		values := url.Values{}
		values.Set("project_id", projectID)
		endpoint += "?" + values.Encode()
	}

	data, err := c.do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var page PaginatedResponse[models.Section]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse sections: %w", err)
	}
	return page.Results, nil
}

// CreateSectionRequest is the request body for creating a section.
type CreateSectionRequest struct {
	Name         string `json:"name"`
	ProjectID    string `json:"project_id"`
	SectionOrder *int   `json:"section_order,omitempty"`
}

// UpdateSectionRequest is the request body for updating a section.
type UpdateSectionRequest struct {
	Name string `json:"name"`
}

// CreateSection creates a new section.
func (c *Client) CreateSection(ctx context.Context, req CreateSectionRequest) (*models.Section, error) {
	data, err := c.do(ctx, "POST", "/sections", req)
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
func (c *Client) UpdateSection(ctx context.Context, id string, req UpdateSectionRequest) (*models.Section, error) {
	data, err := c.do(ctx, "POST", "/sections/"+id, req)
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
func (c *Client) DeleteSection(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/sections/"+id, nil)
	return err
}
