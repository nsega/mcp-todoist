package todoist

import (
	"encoding/json"
	"fmt"

	"github.com/nsega/mcp-todoist/internal/models"
)

// GetProjects returns all projects.
func (c *Client) GetProjects() ([]models.Project, error) {
	data, err := c.do("GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var projects []models.Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects: %w", err)
	}
	return projects, nil
}

// GetProject returns a single project by ID.
func (c *Client) GetProject(id string) (*models.Project, error) {
	data, err := c.do("GET", "/projects/"+id, nil)
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
func (c *Client) CreateProject(body map[string]interface{}) (*models.Project, error) {
	data, err := c.do("POST", "/projects", body)
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
func (c *Client) UpdateProject(id string, body map[string]interface{}) (*models.Project, error) {
	data, err := c.do("POST", "/projects/"+id, body)
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
func (c *Client) DeleteProject(id string) error {
	_, err := c.do("DELETE", "/projects/"+id, nil)
	return err
}

// ArchiveProject archives a project.
func (c *Client) ArchiveProject(id string) error {
	_, err := c.do("POST", "/projects/"+id+"/archive", nil)
	return err
}

// UnarchiveProject unarchives a project.
func (c *Client) UnarchiveProject(id string) error {
	_, err := c.do("POST", "/projects/"+id+"/unarchive", nil)
	return err
}
