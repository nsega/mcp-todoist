package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetProjectsInput struct{}
type GetProjectsOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to retrieve"`
}
type GetProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateProjectInput struct {
	Name       string `json:"name" jsonschema:"Name of the project"`
	ParentID   string `json:"parent_id,omitempty" jsonschema:"Parent project ID (optional)"`
	Color      string `json:"color,omitempty" jsonschema:"Color of the project (optional)"`
	IsFavorite bool   `json:"is_favorite,omitempty" jsonschema:"Whether the project is a favorite (optional)"`
	ViewStyle  string `json:"view_style,omitempty" jsonschema:"View style: list or board (optional)"`
}
type CreateProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateProjectInput struct {
	ProjectID  string `json:"project_id" jsonschema:"The project ID to update"`
	Name       string `json:"name,omitempty" jsonschema:"New name (optional)"`
	Color      string `json:"color,omitempty" jsonschema:"New color (optional)"`
	IsFavorite *bool  `json:"is_favorite,omitempty" jsonschema:"Set favorite status (optional)"`
}
type UpdateProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to delete"`
}
type DeleteProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ArchiveProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to archive"`
}
type ArchiveProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UnarchiveProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to unarchive"`
}
type UnarchiveProjectOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func registerProjectTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_projects",
		Description: "List all Todoist projects",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetProjectsInput) (*mcp.CallToolResult, GetProjectsOutput, error) {
		projects, err := c.GetProjects()
		if err != nil {
			return nil, GetProjectsOutput{}, err
		}

		if len(projects) == 0 {
			msg := "No projects found"
			return textResult(msg, false), GetProjectsOutput{Success: true, Message: msg}, nil
		}

		var lines []string
		for _, p := range projects {
			line := fmt.Sprintf("- %s (ID: %s)", p.Name, p.ID)
			if p.IsInboxProject {
				line += " [Inbox]"
			}
			if p.IsFavorite {
				line += " [Favorite]"
			}
			lines = append(lines, line)
		}
		msg := strings.Join(lines, "\n")
		return textResult(msg, false), GetProjectsOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_project",
		Description: "Get a single Todoist project by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetProjectInput) (*mcp.CallToolResult, GetProjectOutput, error) {
		p, err := c.GetProject(input.ProjectID)
		if err != nil {
			return nil, GetProjectOutput{}, err
		}

		msg := fmt.Sprintf("Project: %s\nID: %s\nColor: %s\nFavorite: %v\nShared: %v\nInbox: %v",
			p.Name, p.ID, p.Color, p.IsFavorite, p.IsShared, p.IsInboxProject)
		if p.URL != "" {
			msg += fmt.Sprintf("\nURL: %s", p.URL)
		}
		return textResult(msg, false), GetProjectOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_project",
		Description: "Create a new Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateProjectInput) (*mcp.CallToolResult, CreateProjectOutput, error) {
		body := map[string]interface{}{"name": input.Name}
		if input.ParentID != "" {
			body["parent_id"] = input.ParentID
		}
		if input.Color != "" {
			body["color"] = input.Color
		}
		if input.IsFavorite {
			body["is_favorite"] = true
		}
		if input.ViewStyle != "" {
			body["view_style"] = input.ViewStyle
		}

		p, err := c.CreateProject(body)
		if err != nil {
			return nil, CreateProjectOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Project created: %s (ID: %s)", p.Name, p.ID)
		return textResult(msg, false), CreateProjectOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_project",
		Description: "Update an existing Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateProjectInput) (*mcp.CallToolResult, UpdateProjectOutput, error) {
		body := map[string]interface{}{}
		if input.Name != "" {
			body["name"] = input.Name
		}
		if input.Color != "" {
			body["color"] = input.Color
		}
		if input.IsFavorite != nil {
			body["is_favorite"] = *input.IsFavorite
		}

		p, err := c.UpdateProject(input.ProjectID, body)
		if err != nil {
			return nil, UpdateProjectOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Project updated: %s (ID: %s)", p.Name, p.ID)
		return textResult(msg, false), UpdateProjectOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_project",
		Description: "Delete a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteProjectInput) (*mcp.CallToolResult, DeleteProjectOutput, error) {
		if err := c.DeleteProject(input.ProjectID); err != nil {
			return nil, DeleteProjectOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted project: %s", input.ProjectID)
		return textResult(msg, false), DeleteProjectOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_archive_project",
		Description: "Archive a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ArchiveProjectInput) (*mcp.CallToolResult, ArchiveProjectOutput, error) {
		if err := c.ArchiveProject(input.ProjectID); err != nil {
			return nil, ArchiveProjectOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully archived project: %s", input.ProjectID)
		return textResult(msg, false), ArchiveProjectOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_unarchive_project",
		Description: "Unarchive a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UnarchiveProjectInput) (*mcp.CallToolResult, UnarchiveProjectOutput, error) {
		if err := c.UnarchiveProject(input.ProjectID); err != nil {
			return nil, UnarchiveProjectOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully unarchived project: %s", input.ProjectID)
		return textResult(msg, false), UnarchiveProjectOutput{Success: true, Message: msg}, nil
	})
}
