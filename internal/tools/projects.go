package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetProjectsInput struct{}
type GetProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to retrieve"`
}
type CreateProjectInput struct {
	Name       string `json:"name" jsonschema:"Name of the project"`
	ParentID   string `json:"parent_id,omitempty" jsonschema:"Parent project ID (optional)"`
	Color      string `json:"color,omitempty" jsonschema:"Color of the project (optional)"`
	IsFavorite bool   `json:"is_favorite,omitempty" jsonschema:"Whether the project is a favorite (optional)"`
	ViewStyle  string `json:"view_style,omitempty" jsonschema:"View style: list or board (optional)"`
}
type UpdateProjectInput struct {
	ProjectID  string `json:"project_id" jsonschema:"The project ID to update"`
	Name       string `json:"name,omitempty" jsonschema:"New name (optional)"`
	Color      string `json:"color,omitempty" jsonschema:"New color (optional)"`
	IsFavorite *bool  `json:"is_favorite,omitempty" jsonschema:"Set favorite status (optional)"`
}
type DeleteProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to delete"`
}
type ArchiveProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to archive"`
}
type UnarchiveProjectInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to unarchive"`
}

func registerProjectTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_projects",
		Description: "List all Todoist projects",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetProjectsInput) (*mcp.CallToolResult, ActionOutput, error) {
		projects, err := c.GetProjects(ctx)
		if err != nil {
			return nil, ActionOutput{}, err
		}

		if len(projects) == 0 {
			msg := "No projects found"
			return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
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
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_project",
		Description: "Get a single Todoist project by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		p, err := c.GetProject(ctx, input.ProjectID)
		if err != nil {
			return nil, ActionOutput{}, err
		}

		msg := fmt.Sprintf("Project: %s\nID: %s\nColor: %s\nFavorite: %v\nShared: %v\nInbox: %v",
			p.Name, p.ID, p.Color, p.IsFavorite, p.IsShared, p.IsInboxProject)
		if p.URL != "" {
			msg += fmt.Sprintf("\nURL: %s", p.URL)
		}
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_project",
		Description: "Create a new Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		createReq := todoist.CreateProjectRequest{Name: input.Name}
		if input.ParentID != "" {
			createReq.ParentID = new(input.ParentID)
		}
		if input.Color != "" {
			createReq.Color = new(input.Color)
		}
		if input.IsFavorite {
			createReq.IsFavorite = new(true)
		}
		if input.ViewStyle != "" {
			createReq.ViewStyle = new(input.ViewStyle)
		}
		p, err := c.CreateProject(ctx, createReq)
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Project created: %s (ID: %s)", p.Name, p.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_project",
		Description: "Update an existing Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		updateReq := todoist.UpdateProjectRequest{}
		if input.Name != "" {
			updateReq.Name = new(input.Name)
		}
		if input.Color != "" {
			updateReq.Color = new(input.Color)
		}
		if input.IsFavorite != nil {
			updateReq.IsFavorite = input.IsFavorite
		}
		p, err := c.UpdateProject(ctx, input.ProjectID, updateReq)
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Project updated: %s (ID: %s)", p.Name, p.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_project",
		Description: "Delete a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		if err := c.DeleteProject(ctx, input.ProjectID); err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted project: %s", input.ProjectID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_archive_project",
		Description: "Archive a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ArchiveProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		if err := c.ArchiveProject(ctx, input.ProjectID); err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully archived project: %s", input.ProjectID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_unarchive_project",
		Description: "Unarchive a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UnarchiveProjectInput) (*mcp.CallToolResult, ActionOutput, error) {
		if err := c.UnarchiveProject(ctx, input.ProjectID); err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully unarchived project: %s", input.ProjectID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})
}
