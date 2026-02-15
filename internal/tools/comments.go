package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

type GetCommentsInput struct {
	TaskID    string `json:"task_id,omitempty" jsonschema:"Get comments for a task (provide task_id or project_id)"`
	ProjectID string `json:"project_id,omitempty" jsonschema:"Get comments for a project (provide task_id or project_id)"`
}
type GetCommentsOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CreateCommentInput struct {
	Content   string `json:"content" jsonschema:"Comment text content"`
	TaskID    string `json:"task_id,omitempty" jsonschema:"Task ID to comment on (provide task_id or project_id)"`
	ProjectID string `json:"project_id,omitempty" jsonschema:"Project ID to comment on (provide task_id or project_id)"`
}
type CreateCommentOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateCommentInput struct {
	CommentID string `json:"comment_id" jsonschema:"The comment ID to update"`
	Content   string `json:"content" jsonschema:"New comment text content"`
}
type UpdateCommentOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteCommentInput struct {
	CommentID string `json:"comment_id" jsonschema:"The comment ID to delete"`
}
type DeleteCommentOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func registerCommentTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_comments",
		Description: "List comments for a task or project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetCommentsInput) (*mcp.CallToolResult, GetCommentsOutput, error) {
		comments, err := c.GetComments(input.TaskID, input.ProjectID)
		if err != nil {
			return nil, GetCommentsOutput{}, err
		}

		if len(comments) == 0 {
			msg := "No comments found"
			return textResult(msg, false), GetCommentsOutput{Success: true, Message: msg}, nil
		}

		var lines []string
		for _, cm := range comments {
			lines = append(lines, fmt.Sprintf("- [%s] %s (ID: %s)", cm.PostedAt.Format("2006-01-02 15:04"), cm.Content, cm.ID))
		}
		msg := strings.Join(lines, "\n")
		return textResult(msg, false), GetCommentsOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_comment",
		Description: "Add a comment to a task or project",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCommentInput) (*mcp.CallToolResult, CreateCommentOutput, error) {
		body := map[string]interface{}{"content": input.Content}
		if input.TaskID != "" {
			body["task_id"] = input.TaskID
		}
		if input.ProjectID != "" {
			body["project_id"] = input.ProjectID
		}

		cm, err := c.CreateComment(body)
		if err != nil {
			return nil, CreateCommentOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Comment created (ID: %s): %s", cm.ID, cm.Content)
		return textResult(msg, false), CreateCommentOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_comment",
		Description: "Update an existing comment",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateCommentInput) (*mcp.CallToolResult, UpdateCommentOutput, error) {
		body := map[string]interface{}{"content": input.Content}
		cm, err := c.UpdateComment(input.CommentID, body)
		if err != nil {
			return nil, UpdateCommentOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Comment updated (ID: %s): %s", cm.ID, cm.Content)
		return textResult(msg, false), UpdateCommentOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_comment",
		Description: "Delete a comment",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteCommentInput) (*mcp.CallToolResult, DeleteCommentOutput, error) {
		if err := c.DeleteComment(input.CommentID); err != nil {
			return nil, DeleteCommentOutput{Success: false, Message: err.Error()}, err
		}
		msg := fmt.Sprintf("Successfully deleted comment: %s", input.CommentID)
		return textResult(msg, false), DeleteCommentOutput{Success: true, Message: msg}, nil
	})
}
