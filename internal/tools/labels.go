package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetLabelsInput struct{}
type GetLabelsOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateLabelInput struct {
	Name       string `json:"name" jsonschema:"Name of the label"`
	Color      string `json:"color,omitempty" jsonschema:"Color of the label (optional)"`
	IsFavorite bool   `json:"is_favorite,omitempty" jsonschema:"Whether the label is a favorite (optional)"`
}
type CreateLabelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateLabelInput struct {
	LabelID string `json:"label_id" jsonschema:"The label ID to update"`
	Name    string `json:"name,omitempty" jsonschema:"New name (optional)"`
	Color   string `json:"color,omitempty" jsonschema:"New color (optional)"`
}
type UpdateLabelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteLabelInput struct {
	LabelID string `json:"label_id" jsonschema:"The label ID to delete"`
}
type DeleteLabelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func registerLabelTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_labels",
		Description: "List all personal labels",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetLabelsInput) (*mcp.CallToolResult, GetLabelsOutput, error) {
		labels, err := c.GetLabels()
		if err != nil {
			return nil, GetLabelsOutput{}, err
		}

		if len(labels) == 0 {
			msg := "No labels found"
			return textResult(msg, false), GetLabelsOutput{Success: true, Message: msg}, nil
		}

		var lines []string
		for _, l := range labels {
			line := fmt.Sprintf("- %s (ID: %s)", l.Name, l.ID)
			if l.IsFavorite {
				line += " [Favorite]"
			}
			lines = append(lines, line)
		}
		msg := strings.Join(lines, "\n")
		return textResult(msg, false), GetLabelsOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_label",
		Description: "Create a new personal label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateLabelInput) (*mcp.CallToolResult, CreateLabelOutput, error) {
		body := map[string]interface{}{"name": input.Name}
		if input.Color != "" {
			body["color"] = input.Color
		}
		if input.IsFavorite {
			body["is_favorite"] = true
		}

		l, err := c.CreateLabel(body)
		if err != nil {
			return nil, CreateLabelOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Label created: %s (ID: %s)", l.Name, l.ID)
		return textResult(msg, false), CreateLabelOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_label",
		Description: "Update an existing label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateLabelInput) (*mcp.CallToolResult, UpdateLabelOutput, error) {
		body := map[string]interface{}{}
		if input.Name != "" {
			body["name"] = input.Name
		}
		if input.Color != "" {
			body["color"] = input.Color
		}

		l, err := c.UpdateLabel(input.LabelID, body)
		if err != nil {
			return nil, UpdateLabelOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Label updated: %s (ID: %s)", l.Name, l.ID)
		return textResult(msg, false), UpdateLabelOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_label",
		Description: "Delete a label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteLabelInput) (*mcp.CallToolResult, DeleteLabelOutput, error) {
		if err := c.DeleteLabel(input.LabelID); err != nil {
			return nil, DeleteLabelOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted label: %s", input.LabelID)
		return textResult(msg, false), DeleteLabelOutput{Success: true, Message: msg}, nil
	})
}
