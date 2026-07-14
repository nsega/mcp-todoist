package todoist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetProjects returns all projects.
func (c *Client) GetProjects(ctx context.Context) ([]models.Project, error) {
	data, err := c.do(ctx, "GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var page PaginatedResponse[models.Project]
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	return page.Results, nil
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(ctx context.Context, id string) (*models.Project, error) {
	data, err := c.do(ctx, "GET", "/projects/"+id, nil)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	return &project, nil
}

// CreateProject creates a new project.
func (c *Client) CreateProject(ctx context.Context, body map[string]any) (*models.Project, error) {
	data, err := c.do(ctx, "POST", "/projects", body)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	return &project, nil
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, id string, body map[string]any) (*models.Project, error) {
	data, err := c.do(ctx, "POST", "/projects/"+id, body)
	if err != nil {
		return nil, err
	}

	var project models.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	return &project, nil
}

// DeleteProject deletes a project.
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/projects/"+id, nil)
	return err
}

// ArchiveProject archives a project.
func (c *Client) ArchiveProject(ctx context.Context, id string) error {
	_, err := c.do(ctx, "POST", "/projects/"+id+"/archive", nil)
	return err
}

// UnarchiveProject unarchives a project.
func (c *Client) UnarchiveProject(ctx context.Context, id string) error {
	_, err := c.do(ctx, "POST", "/projects/"+id+"/unarchive", nil)
	return err
}
