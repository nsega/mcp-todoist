package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

// --- Input / Output types ---

type CreateTaskInput struct {
	Content     string   `json:"content" jsonschema:"The content/title of the task"`
	Description string   `json:"description,omitempty" jsonschema:"Detailed description of the task (optional)"`
	DueString   string   `json:"due_string,omitempty" jsonschema:"Natural language due date like 'tomorrow', 'next Monday', 'Jan 23' (optional)"`
	Priority    int      `json:"priority,omitempty" jsonschema:"Task priority from 1 (normal) to 4 (urgent) (optional)"`
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"Project ID to create the task in (optional)"`
	SectionID   string   `json:"section_id,omitempty" jsonschema:"Section ID to create the task in (optional)"`
	ParentID    string   `json:"parent_id,omitempty" jsonschema:"Parent task ID for sub-tasks (optional)"`
	Labels      []string `json:"labels,omitempty" jsonschema:"Labels to apply to the task (optional)"`
	AssigneeID  string   `json:"assignee_id,omitempty" jsonschema:"User ID to assign the task to (optional)"`
}

type CreateTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetTasksInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Filter tasks by project ID (optional)"`
	Filter    string `json:"filter,omitempty" jsonschema:"Natural language filter like 'today', 'tomorrow', 'next week', 'priority 1', 'overdue' (optional)"`
	Priority  int    `json:"priority,omitempty" jsonschema:"Filter by priority level (1-4) (optional)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of tasks to return (optional, default 10)"`
}

type GetTasksOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateTaskInput struct {
	TaskID      string   `json:"task_id,omitempty" jsonschema:"Task ID to update (preferred over task_name)"`
	TaskName    string   `json:"task_name,omitempty" jsonschema:"Name/content of the task to search for and update"`
	Content     string   `json:"content,omitempty" jsonschema:"New content/title for the task (optional)"`
	Description string   `json:"description,omitempty" jsonschema:"New description for the task (optional)"`
	DueString   string   `json:"due_string,omitempty" jsonschema:"New due date in natural language (optional)"`
	Priority    int      `json:"priority,omitempty" jsonschema:"New priority level from 1 (normal) to 4 (urgent) (optional)"`
	Labels      []string `json:"labels,omitempty" jsonschema:"New labels for the task (optional)"`
	AssigneeID  string   `json:"assignee_id,omitempty" jsonschema:"User ID to assign the task to (optional)"`
}

type UpdateTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteTaskInput struct {
	TaskID   string `json:"task_id,omitempty" jsonschema:"Task ID to delete (preferred over task_name)"`
	TaskName string `json:"task_name,omitempty" jsonschema:"Name/content of the task to search for and delete"`
}

type DeleteTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CompleteTaskInput struct {
	TaskID   string `json:"task_id,omitempty" jsonschema:"Task ID to complete (preferred over task_name)"`
	TaskName string `json:"task_name,omitempty" jsonschema:"Name/content of the task to search for and complete"`
}

type CompleteTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ReopenTaskInput struct {
	TaskID   string `json:"task_id,omitempty" jsonschema:"Task ID to reopen (preferred over task_name)"`
	TaskName string `json:"task_name,omitempty" jsonschema:"Name/content of the task to search for and reopen"`
}

type ReopenTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- helpers ---

func resolveTaskID(c *todoist.Client, id, name string) (string, string, error) {
	if id != "" {
		return id, "", nil
	}
	if name == "" {
		return "", "", fmt.Errorf("either task_id or task_name is required")
	}
	task, err := c.FindTaskByName(name)
	if err != nil {
		return "", "", err
	}
	if task == nil {
		return "", "", nil // not found
	}
	return task.ID, task.Content, nil
}

func textResult(msg string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: isError,
	}
}

// --- registrations ---

func registerTaskTools(s *mcp.Server, c *todoist.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_create_task",
		Description: "Create a new task in Todoist with optional description, due date, priority, project, section, labels, and assignee",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
		body := map[string]interface{}{"content": input.Content}
		if input.Description != "" {
			body["description"] = input.Description
		}
		if input.DueString != "" {
			body["due_string"] = input.DueString
		}
		if input.Priority > 0 && input.Priority <= 4 {
			body["priority"] = input.Priority
		}
		if input.ProjectID != "" {
			body["project_id"] = input.ProjectID
		}
		if input.SectionID != "" {
			body["section_id"] = input.SectionID
		}
		if input.ParentID != "" {
			body["parent_id"] = input.ParentID
		}
		if len(input.Labels) > 0 {
			body["labels"] = input.Labels
		}
		if input.AssigneeID != "" {
			body["assignee_id"] = input.AssigneeID
		}

		task, err := c.CreateTask(body)
		if err != nil {
			return nil, CreateTaskOutput{Success: false, Message: err.Error()}, err
		}

		msg := fmt.Sprintf("Task created:\nTitle: %s", task.Content)
		if task.Description != "" {
			msg += fmt.Sprintf("\nDescription: %s", task.Description)
		}
		if task.Due != nil && task.Due.String != "" {
			msg += fmt.Sprintf("\nDue: %s", task.Due.String)
		}
		if task.Priority > 0 {
			msg += fmt.Sprintf("\nPriority: %d", task.Priority)
		}

		return textResult(msg, false), CreateTaskOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_get_tasks",
		Description: "Get a list of tasks from Todoist with various filters",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetTasksInput) (*mcp.CallToolResult, GetTasksOutput, error) {
		tasks, err := c.GetTasks(input.ProjectID, input.Filter)
		if err != nil {
			return nil, GetTasksOutput{}, err
		}

		// Apply priority filter.
		if input.Priority > 0 && input.Priority <= 4 {
			filtered := tasks[:0]
			for _, t := range tasks {
				if t.Priority == input.Priority {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
		}

		// Apply limit.
		limit := input.Limit
		if limit == 0 {
			limit = 10
		}
		if len(tasks) > limit {
			tasks = tasks[:limit]
		}

		var msg string
		if len(tasks) == 0 {
			msg = "No tasks found matching the criteria"
		} else {
			var lines []string
			for _, t := range tasks {
				s := fmt.Sprintf("- %s", t.Content)
				if t.Description != "" {
					s += fmt.Sprintf("\n  Description: %s", t.Description)
				}
				if t.Due != nil && t.Due.String != "" {
					s += fmt.Sprintf("\n  Due: %s", t.Due.String)
				}
				if t.Priority > 0 {
					s += fmt.Sprintf("\n  Priority: %d", t.Priority)
				}
				lines = append(lines, s)
			}
			msg = strings.Join(lines, "\n\n")
		}

		return textResult(msg, false), GetTasksOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_update_task",
		Description: "Update an existing task in Todoist by task_id or by searching by name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateTaskInput) (*mcp.CallToolResult, UpdateTaskOutput, error) {
		id, originalName, err := resolveTaskID(c, input.TaskID, input.TaskName)
		if err != nil {
			return nil, UpdateTaskOutput{Success: false, Message: err.Error()}, err
		}
		if id == "" {
			msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
			return textResult(msg, true), UpdateTaskOutput{Success: false, Message: msg}, nil
		}

		body := map[string]interface{}{}
		if input.Content != "" {
			body["content"] = input.Content
		}
		if input.Description != "" {
			body["description"] = input.Description
		}
		if input.DueString != "" {
			body["due_string"] = input.DueString
		}
		if input.Priority > 0 && input.Priority <= 4 {
			body["priority"] = input.Priority
		}
		if len(input.Labels) > 0 {
			body["labels"] = input.Labels
		}
		if input.AssigneeID != "" {
			body["assignee_id"] = input.AssigneeID
		}

		updated, err := c.UpdateTask(id, body)
		if err != nil {
			return nil, UpdateTaskOutput{Success: false, Message: err.Error()}, err
		}

		label := originalName
		if label == "" {
			label = id
		}
		msg := fmt.Sprintf("Task \"%s\" updated:\nNew Title: %s", label, updated.Content)
		if updated.Description != "" {
			msg += fmt.Sprintf("\nNew Description: %s", updated.Description)
		}
		if updated.Due != nil && updated.Due.String != "" {
			msg += fmt.Sprintf("\nNew Due Date: %s", updated.Due.String)
		}
		if updated.Priority > 0 {
			msg += fmt.Sprintf("\nNew Priority: %d", updated.Priority)
		}

		return textResult(msg, false), UpdateTaskOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_delete_task",
		Description: "Delete a task from Todoist by task_id or by searching by name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteTaskInput) (*mcp.CallToolResult, DeleteTaskOutput, error) {
		id, originalName, err := resolveTaskID(c, input.TaskID, input.TaskName)
		if err != nil {
			return nil, DeleteTaskOutput{Success: false, Message: err.Error()}, err
		}
		if id == "" {
			msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
			return textResult(msg, true), DeleteTaskOutput{Success: false, Message: msg}, nil
		}

		if err := c.DeleteTask(id); err != nil {
			return nil, DeleteTaskOutput{Success: false, Message: err.Error()}, err
		}

		label := originalName
		if label == "" {
			label = id
		}
		msg := fmt.Sprintf("Successfully deleted task: \"%s\"", label)
		return textResult(msg, false), DeleteTaskOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_complete_task",
		Description: "Mark a task as complete by task_id or by searching by name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CompleteTaskInput) (*mcp.CallToolResult, CompleteTaskOutput, error) {
		id, originalName, err := resolveTaskID(c, input.TaskID, input.TaskName)
		if err != nil {
			return nil, CompleteTaskOutput{Success: false, Message: err.Error()}, err
		}
		if id == "" {
			msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
			return textResult(msg, true), CompleteTaskOutput{Success: false, Message: msg}, nil
		}

		if err := c.CloseTask(id); err != nil {
			return nil, CompleteTaskOutput{Success: false, Message: err.Error()}, err
		}

		label := originalName
		if label == "" {
			label = id
		}
		msg := fmt.Sprintf("Successfully completed task: \"%s\"", label)
		return textResult(msg, false), CompleteTaskOutput{Success: true, Message: msg}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_reopen_task",
		Description: "Reopen a completed task by task_id or by searching by name",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ReopenTaskInput) (*mcp.CallToolResult, ReopenTaskOutput, error) {
		id, originalName, err := resolveTaskID(c, input.TaskID, input.TaskName)
		if err != nil {
			return nil, ReopenTaskOutput{Success: false, Message: err.Error()}, err
		}
		if id == "" {
			msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
			return textResult(msg, true), ReopenTaskOutput{Success: false, Message: msg}, nil
		}

		if err := c.ReopenTask(id); err != nil {
			return nil, ReopenTaskOutput{Success: false, Message: err.Error()}, err
		}

		label := originalName
		if label == "" {
			label = id
		}
		msg := fmt.Sprintf("Successfully reopened task: \"%s\"", label)
		return textResult(msg, false), ReopenTaskOutput{Success: true, Message: msg}, nil
	})
}
