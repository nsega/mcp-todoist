package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetSectionsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Filter sections by project ID (optional)"`
}
type CreateSectionInput struct {
	Name      string `json:"name" jsonschema:"Name of the section"`
	ProjectID string `json:"project_id" jsonschema:"Project ID the section belongs to"`
	Order     int    `json:"order,omitempty" jsonschema:"Order among other sections (optional)"`
}
type UpdateSectionInput struct {
	SectionID string `json:"section_id" jsonschema:"The section ID to update"`
	Name      string `json:"name" jsonschema:"New name for the section"`
}
type DeleteSectionInput struct {
	SectionID string `json:"section_id" jsonschema:"The section ID to delete"`
}

func registerSectionTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_sections",
		Description: "List sections, optionally filtered by project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSectionsInput) (*mcp.CallToolResult, ActionOutput, error) {
		sections, err := c.GetSections(ctx, input.ProjectID)
		if err != nil {
			return nil, ActionOutput{}, err
		}

		if len(sections) == 0 {
			msg := "No sections found"
			return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
		}

		var lines []string
		for _, sec := range sections {
			lines = append(lines, fmt.Sprintf("- %s (ID: %s, Project: %s)", sec.Name, sec.ID, sec.ProjectID))
		}
		msg := strings.Join(lines, "\n")
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_section",
		Description: "Create a new section in a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateSectionInput) (*mcp.CallToolResult, ActionOutput, error) {
		createReq := todoist.CreateSectionRequest{Name: input.Name, ProjectID: input.ProjectID}
		if input.Order > 0 {
			createReq.SectionOrder = new(input.Order)
		}
		sec, err := c.CreateSection(ctx, createReq)
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Section created: %s (ID: %s)", sec.Name, sec.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_section",
		Description: "Update an existing section name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateSectionInput) (*mcp.CallToolResult, ActionOutput, error) {
		sec, err := c.UpdateSection(ctx, input.SectionID, todoist.UpdateSectionRequest{Name: input.Name})
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Section updated: %s (ID: %s)", sec.Name, sec.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_section",
		Description: "Delete a section from a Todoist project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteSectionInput) (*mcp.CallToolResult, ActionOutput, error) {
		if err := c.DeleteSection(ctx, input.SectionID); err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted section: %s", input.SectionID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})
}
