package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetLabelsInput struct{}
type CreateLabelInput struct {
	Name       string `json:"name" jsonschema:"Name of the label"`
	Color      string `json:"color,omitempty" jsonschema:"Color of the label (optional)"`
	IsFavorite bool   `json:"is_favorite,omitempty" jsonschema:"Whether the label is a favorite (optional)"`
}
type UpdateLabelInput struct {
	LabelID string `json:"label_id" jsonschema:"The label ID to update"`
	Name    string `json:"name,omitempty" jsonschema:"New name (optional)"`
	Color   string `json:"color,omitempty" jsonschema:"New color (optional)"`
}
type DeleteLabelInput struct {
	LabelID string `json:"label_id" jsonschema:"The label ID to delete"`
}

func registerLabelTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_labels",
		Description: "List all personal labels",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetLabelsInput) (*mcp.CallToolResult, ActionOutput, error) {
		labels, err := c.GetLabels(ctx)
		if err != nil {
			return nil, ActionOutput{}, err
		}

		if len(labels) == 0 {
			msg := "No labels found"
			return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
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
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_label",
		Description: "Create a new personal label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateLabelInput) (*mcp.CallToolResult, ActionOutput, error) {
		createReq := todoist.CreateLabelRequest{Name: input.Name}
		if input.Color != "" {
			createReq.Color = new(input.Color)
		}
		if input.IsFavorite {
			createReq.IsFavorite = new(true)
		}
		l, err := c.CreateLabel(ctx, createReq)
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Label created: %s (ID: %s)", l.Name, l.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_label",
		Description: "Update an existing label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateLabelInput) (*mcp.CallToolResult, ActionOutput, error) {
		updateReq := todoist.UpdateLabelRequest{}
		if input.Name != "" {
			updateReq.Name = new(input.Name)
		}
		if input.Color != "" {
			updateReq.Color = new(input.Color)
		}
		l, err := c.UpdateLabel(ctx, input.LabelID, updateReq)
		if err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Label updated: %s (ID: %s)", l.Name, l.ID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_label",
		Description: "Delete a label",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteLabelInput) (*mcp.CallToolResult, ActionOutput, error) {
		if err := c.DeleteLabel(ctx, input.LabelID); err != nil {
			return nil, ActionOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted label: %s", input.LabelID)
		return textResult(msg, false), ActionOutput{Success: true, Message: msg}, nil
	})
}
