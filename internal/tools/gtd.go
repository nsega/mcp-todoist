package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/models"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

// --- Inbox Review ---

type InboxReviewInput struct{}
type InboxReviewOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Weekly Review ---

type WeeklyReviewInput struct{}
type WeeklyReviewOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Move Task ---

type MoveTaskInput struct {
	TaskID    string `json:"task_id,omitempty" jsonschema:"Task ID to move (preferred over task_name)"`
	TaskName  string `json:"task_name,omitempty" jsonschema:"Name of the task to search for and move"`
	ProjectID string `json:"project_id,omitempty" jsonschema:"Destination project ID (optional)"`
	SectionID string `json:"section_id,omitempty" jsonschema:"Destination section ID (optional)"`
}
type MoveTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Bulk Create Tasks ---

type BulkTaskItem struct {
	Content     string   `json:"content" jsonschema:"Task content/title"`
	Description string   `json:"description,omitempty" jsonschema:"Task description (optional)"`
	DueString   string   `json:"due_string,omitempty" jsonschema:"Due date in natural language (optional)"`
	Priority    int      `json:"priority,omitempty" jsonschema:"Priority 1-4 (optional)"`
	ProjectID   string   `json:"project_id,omitempty" jsonschema:"Project ID (optional)"`
	SectionID   string   `json:"section_id,omitempty" jsonschema:"Section ID (optional)"`
	Labels      []string `json:"labels,omitempty" jsonschema:"Labels (optional)"`
}

type BulkCreateTasksInput struct {
	Tasks []BulkTaskItem `json:"tasks" jsonschema:"Array of tasks to create"`
}
type BulkCreateTasksOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func registerGTDTools(s *mcp.Server, c *todoist.Client) {
	// --- todoist_inbox_review ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_inbox_review",
		Description: "Get all inbox tasks grouped by age (today, this week, older) for GTD inbox processing",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input InboxReviewInput) (*mcp.CallToolResult, InboxReviewOutput, error) {
		// Find inbox project.
		projects, err := c.GetProjects()
		if err != nil {
			return nil, InboxReviewOutput{}, err
		}

		var inboxID string
		for _, p := range projects {
			if p.IsInboxProject {
				inboxID = p.ID
				break
			}
		}
		if inboxID == "" {
			msg := "Could not find inbox project"
			return textResult(msg, true), InboxReviewOutput{Success: false, Message: msg}, nil
		}

		tasks, err := c.GetTasks(inboxID, "")
		if err != nil {
			return nil, InboxReviewOutput{}, err
		}

		if len(tasks) == 0 {
			msg := "Inbox is empty! Nothing to process."
			return textResult(msg, false), InboxReviewOutput{Success: true, Message: msg}, nil
		}

		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		weekAgo := todayStart.AddDate(0, 0, -7)

		var todayTasks, weekTasks, olderTasks []models.Task
		for _, t := range tasks {
			switch {
			case t.CreatedAt.After(todayStart):
				todayTasks = append(todayTasks, t)
			case t.CreatedAt.After(weekAgo):
				weekTasks = append(weekTasks, t)
			default:
				olderTasks = append(olderTasks, t)
			}
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "## Inbox Review (%d tasks)\n\n", len(tasks))

		writeGroup := func(title string, group []models.Task) {
			fmt.Fprintf(&sb, "### %s (%d)\n", title, len(group))
			if len(group) == 0 {
				sb.WriteString("(none)\n")
			}
			for _, t := range group {
				sb.WriteString(fmt.Sprintf("- %s (ID: %s)", t.Content, t.ID))
				if t.Priority > 1 {
					sb.WriteString(fmt.Sprintf(" [P%d]", t.Priority))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}

		writeGroup("Added Today", todayTasks)
		writeGroup("Added This Week", weekTasks)
		writeGroup("Older", olderTasks)

		msg := sb.String()
		return textResult(msg, false), InboxReviewOutput{Success: true, Message: msg}, nil
	})

	// --- todoist_weekly_review ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_weekly_review",
		Description: "Comprehensive weekly review: projects with task counts, overdue tasks, tasks with no due date",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input WeeklyReviewInput) (*mcp.CallToolResult, WeeklyReviewOutput, error) {
		projects, err := c.GetProjects()
		if err != nil {
			return nil, WeeklyReviewOutput{}, err
		}

		allTasks, err := c.GetTasks("", "")
		if err != nil {
			return nil, WeeklyReviewOutput{}, err
		}

		// Count tasks per project.
		projectCounts := map[string]int{}
		for _, t := range allTasks {
			projectCounts[t.ProjectID]++
		}

		// Overdue tasks.
		overdueTasks, err := c.GetTasks("", "overdue")
		if err != nil {
			overdueTasks = nil // non-fatal
		}

		// Tasks with no due date.
		var noDueTasks []models.Task
		for _, t := range allTasks {
			if t.Due == nil {
				noDueTasks = append(noDueTasks, t)
			}
		}

		var sb strings.Builder
		sb.WriteString("## Weekly Review\n\n")

		// Projects summary.
		sb.WriteString("### Projects\n")
		for _, p := range projects {
			count := projectCounts[p.ID]
			tag := ""
			if p.IsInboxProject {
				tag = " [Inbox]"
			}
			fmt.Fprintf(&sb, "- %s%s: %d active tasks\n", p.Name, tag, count)
		}
		sb.WriteString("\n")

		// Overdue.
		fmt.Fprintf(&sb, "### Overdue Tasks (%d)\n", len(overdueTasks))
		for _, t := range overdueTasks {
			due := ""
			if t.Due != nil {
				due = t.Due.Date
			}
			fmt.Fprintf(&sb, "- %s (due: %s, ID: %s)\n", t.Content, due, t.ID)
		}
		sb.WriteString("\n")

		// No due date.
		fmt.Fprintf(&sb, "### No Due Date (%d)\n", len(noDueTasks))
		for _, t := range noDueTasks {
			fmt.Fprintf(&sb, "- %s (ID: %s)\n", t.Content, t.ID)
		}

		msg := sb.String()
		return textResult(msg, false), WeeklyReviewOutput{Success: true, Message: msg}, nil
	})

	// --- todoist_move_task ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_move_task",
		Description: "Move a task to a different project and/or section",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input MoveTaskInput) (*mcp.CallToolResult, MoveTaskOutput, error) {
		id, originalName, err := resolveTaskID(c, input.TaskID, input.TaskName)
		if err != nil {
			return nil, MoveTaskOutput{Success: false, Message: err.Error()}, err
		}
		if id == "" {
			msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
			return textResult(msg, true), MoveTaskOutput{Success: false, Message: msg}, nil
		}

		body := map[string]interface{}{}
		if input.ProjectID != "" {
			body["project_id"] = input.ProjectID
		}
		if input.SectionID != "" {
			body["section_id"] = input.SectionID
		}

		_, err = c.UpdateTask(id, body)
		if err != nil {
			return nil, MoveTaskOutput{Success: false, Message: err.Error()}, err
		}

		label := originalName
		if label == "" {
			label = id
		}
		msg := fmt.Sprintf("Successfully moved task \"%s\"", label)
		if input.ProjectID != "" {
			msg += fmt.Sprintf(" to project %s", input.ProjectID)
		}
		if input.SectionID != "" {
			msg += fmt.Sprintf(" section %s", input.SectionID)
		}
		return textResult(msg, false), MoveTaskOutput{Success: true, Message: msg}, nil
	})

	// --- todoist_bulk_create_tasks ---
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todoist_bulk_create_tasks",
		Description: "Create multiple tasks at once. Useful for batch processing from knowledge capture or project planning",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input BulkCreateTasksInput) (*mcp.CallToolResult, BulkCreateTasksOutput, error) {
		var created, failed int
		var lines []string

		for _, item := range input.Tasks {
			body := map[string]interface{}{"content": item.Content}
			if item.Description != "" {
				body["description"] = item.Description
			}
			if item.DueString != "" {
				body["due_string"] = item.DueString
			}
			if item.Priority > 0 && item.Priority <= 4 {
				body["priority"] = item.Priority
			}
			if item.ProjectID != "" {
				body["project_id"] = item.ProjectID
			}
			if item.SectionID != "" {
				body["section_id"] = item.SectionID
			}
			if len(item.Labels) > 0 {
				body["labels"] = item.Labels
			}

			task, err := c.CreateTask(body)
			if err != nil {
				failed++
				lines = append(lines, fmt.Sprintf("FAILED: %s — %s", item.Content, err.Error()))
				continue
			}
			created++
			lines = append(lines, fmt.Sprintf("OK: %s (ID: %s)", task.Content, task.ID))
		}

		msg := fmt.Sprintf("Bulk create: %d created, %d failed\n\n%s", created, failed, strings.Join(lines, "\n"))
		success := failed == 0
		return textResult(msg, !success), BulkCreateTasksOutput{Success: success, Message: msg}, nil
	})
}
