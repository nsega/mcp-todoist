package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	todoistAPIURL = "https://api.todoist.com/rest/v2"
)

var (
	apiToken string
)

// Task represents a Todoist task
type Task struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	ProjectID   string    `json:"project_id"`
	Priority    int       `json:"priority"`
	Due         *DueDate  `json:"due,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// DueDate represents a task's due date
type DueDate struct {
	Date      string `json:"date"`
	String    string `json:"string"`
	Datetime  string `json:"datetime,omitempty"`
	Recurring bool   `json:"recurring"`
	Timezone  string `json:"timezone,omitempty"`
}

// CreateTaskInput defines the input for creating a task
type CreateTaskInput struct {
	Content     string `json:"content" jsonschema:"The content/title of the task"`
	Description string `json:"description,omitempty" jsonschema:"Detailed description of the task (optional)"`
	DueString   string `json:"due_string,omitempty" jsonschema:"Natural language due date like 'tomorrow', 'next Monday', 'Jan 23' (optional)"`
	Priority    int    `json:"priority,omitempty" jsonschema:"Task priority from 1 (normal) to 4 (urgent) (optional)"`
}

// CreateTaskOutput defines the output for creating a task
type CreateTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetTasksInput defines the input for getting tasks
type GetTasksInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Filter tasks by project ID (optional)"`
	Filter    string `json:"filter,omitempty" jsonschema:"Natural language filter like 'today', 'tomorrow', 'next week', 'priority 1', 'overdue' (optional)"`
	Priority  int    `json:"priority,omitempty" jsonschema:"Filter by priority level (1-4) (optional)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of tasks to return (optional, default 10)"`
}

// GetTasksOutput defines the output for getting tasks
type GetTasksOutput struct {
	Tasks []Task `json:"tasks"`
}

// UpdateTaskInput defines the input for updating a task
type UpdateTaskInput struct {
	TaskName    string `json:"task_name" jsonschema:"Name/content of the task to search for and update"`
	Content     string `json:"content,omitempty" jsonschema:"New content/title for the task (optional)"`
	Description string `json:"description,omitempty" jsonschema:"New description for the task (optional)"`
	DueString   string `json:"due_string,omitempty" jsonschema:"New due date in natural language like 'tomorrow', 'next Monday' (optional)"`
	Priority    int    `json:"priority,omitempty" jsonschema:"New priority level from 1 (normal) to 4 (urgent) (optional)"`
}

// UpdateTaskOutput defines the output for updating a task
type UpdateTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteTaskInput defines the input for deleting a task
type DeleteTaskInput struct {
	TaskName string `json:"task_name" jsonschema:"Name/content of the task to search for and delete"`
}

// DeleteTaskOutput defines the output for deleting a task
type DeleteTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CompleteTaskInput defines the input for completing a task
type CompleteTaskInput struct {
	TaskName string `json:"task_name" jsonschema:"Name/content of the task to search for and complete"`
}

// CompleteTaskOutput defines the output for completing a task
type CompleteTaskOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// makeAPIRequest makes an HTTP request to the Todoist API
func makeAPIRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, todoistAPIURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// createTaskHandler implements the todoist_create_task tool
func createTaskHandler(ctx context.Context, req *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, CreateTaskOutput, error) {
	// Build request body
	requestBody := map[string]interface{}{
		"content": input.Content,
	}

	if input.Description != "" {
		requestBody["description"] = input.Description
	}

	if input.DueString != "" {
		requestBody["due_string"] = input.DueString
	}

	if input.Priority > 0 && input.Priority <= 4 {
		requestBody["priority"] = input.Priority
	}

	// Make API request
	respBody, err := makeAPIRequest("POST", "/tasks", requestBody)
	if err != nil {
		return nil, CreateTaskOutput{Success: false, Message: err.Error()}, err
	}

	// Parse response
	var task Task
	if err := json.Unmarshal(respBody, &task); err != nil {
		return nil, CreateTaskOutput{Success: false, Message: err.Error()}, err
	}

	// Build response message
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

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: msg,
			},
		},
	}

	output := CreateTaskOutput{
		Success: true,
		Message: msg,
	}

	return result, output, nil
}

// getTasksHandler implements the todoist_get_tasks tool
func getTasksHandler(ctx context.Context, req *mcp.CallToolRequest, input GetTasksInput) (*mcp.CallToolResult, GetTasksOutput, error) {
	// Build query parameters
	endpoint := "/tasks"
	params := []string{}

	if input.ProjectID != "" {
		params = append(params, fmt.Sprintf("project_id=%s", input.ProjectID))
	}

	if input.Filter != "" {
		params = append(params, fmt.Sprintf("filter=%s", input.Filter))
	}

	if len(params) > 0 {
		endpoint += "?" + strings.Join(params, "&")
	}

	// Make API request
	respBody, err := makeAPIRequest("GET", endpoint, nil)
	if err != nil {
		return nil, GetTasksOutput{}, err
	}

	// Parse response
	var tasks []Task
	if err := json.Unmarshal(respBody, &tasks); err != nil {
		return nil, GetTasksOutput{}, err
	}

	// Apply additional filters
	filteredTasks := tasks
	if input.Priority > 0 && input.Priority <= 4 {
		filtered := []Task{}
		for _, task := range filteredTasks {
			if task.Priority == input.Priority {
				filtered = append(filtered, task)
			}
		}
		filteredTasks = filtered
	}

	// Apply limit
	limit := input.Limit
	if limit == 0 {
		limit = 10 // Default limit
	}
	if len(filteredTasks) > limit {
		filteredTasks = filteredTasks[:limit]
	}

	// Build response message
	var msg string
	if len(filteredTasks) == 0 {
		msg = "No tasks found matching the criteria"
	} else {
		taskList := []string{}
		for _, task := range filteredTasks {
			taskStr := fmt.Sprintf("- %s", task.Content)
			if task.Description != "" {
				taskStr += fmt.Sprintf("\n  Description: %s", task.Description)
			}
			if task.Due != nil && task.Due.String != "" {
				taskStr += fmt.Sprintf("\n  Due: %s", task.Due.String)
			}
			if task.Priority > 0 {
				taskStr += fmt.Sprintf("\n  Priority: %d", task.Priority)
			}
			taskList = append(taskList, taskStr)
		}
		msg = strings.Join(taskList, "\n\n")
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: msg,
			},
		},
	}

	output := GetTasksOutput{
		Tasks: filteredTasks,
	}

	return result, output, nil
}

// findTaskByName searches for a task by partial name matching
func findTaskByName(taskName string) (*Task, error) {
	respBody, err := makeAPIRequest("GET", "/tasks", nil)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(respBody, &tasks); err != nil {
		return nil, err
	}

	// Search for matching task (case-insensitive partial match)
	taskNameLower := strings.ToLower(taskName)
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.Content), taskNameLower) {
			return &task, nil
		}
	}

	return nil, nil
}

// updateTaskHandler implements the todoist_update_task tool
func updateTaskHandler(ctx context.Context, req *mcp.CallToolRequest, input UpdateTaskInput) (*mcp.CallToolResult, UpdateTaskOutput, error) {
	// Find the task
	task, err := findTaskByName(input.TaskName)
	if err != nil {
		return nil, UpdateTaskOutput{Success: false, Message: err.Error()}, err
	}

	if task == nil {
		msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: msg,
				},
			},
			IsError: true,
		}
		return result, UpdateTaskOutput{Success: false, Message: msg}, nil
	}

	// Build update data
	updateData := map[string]interface{}{}
	if input.Content != "" {
		updateData["content"] = input.Content
	}
	if input.Description != "" {
		updateData["description"] = input.Description
	}
	if input.DueString != "" {
		updateData["due_string"] = input.DueString
	}
	if input.Priority > 0 && input.Priority <= 4 {
		updateData["priority"] = input.Priority
	}

	// Make API request
	respBody, err := makeAPIRequest("POST", "/tasks/"+task.ID, updateData)
	if err != nil {
		return nil, UpdateTaskOutput{Success: false, Message: err.Error()}, err
	}

	// Parse response
	var updatedTask Task
	if err := json.Unmarshal(respBody, &updatedTask); err != nil {
		return nil, UpdateTaskOutput{Success: false, Message: err.Error()}, err
	}

	// Build response message
	msg := fmt.Sprintf("Task \"%s\" updated:\nNew Title: %s", task.Content, updatedTask.Content)
	if updatedTask.Description != "" {
		msg += fmt.Sprintf("\nNew Description: %s", updatedTask.Description)
	}
	if updatedTask.Due != nil && updatedTask.Due.String != "" {
		msg += fmt.Sprintf("\nNew Due Date: %s", updatedTask.Due.String)
	}
	if updatedTask.Priority > 0 {
		msg += fmt.Sprintf("\nNew Priority: %d", updatedTask.Priority)
	}

	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: msg,
			},
		},
	}

	output := UpdateTaskOutput{
		Success: true,
		Message: msg,
	}

	return result, output, nil
}

// deleteTaskHandler implements the todoist_delete_task tool
func deleteTaskHandler(ctx context.Context, req *mcp.CallToolRequest, input DeleteTaskInput) (*mcp.CallToolResult, DeleteTaskOutput, error) {
	// Find the task
	task, err := findTaskByName(input.TaskName)
	if err != nil {
		return nil, DeleteTaskOutput{Success: false, Message: err.Error()}, err
	}

	if task == nil {
		msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: msg,
				},
			},
			IsError: true,
		}
		return result, DeleteTaskOutput{Success: false, Message: msg}, nil
	}

	// Delete the task
	_, err = makeAPIRequest("DELETE", "/tasks/"+task.ID, nil)
	if err != nil {
		return nil, DeleteTaskOutput{Success: false, Message: err.Error()}, err
	}

	msg := fmt.Sprintf("Successfully deleted task: \"%s\"", task.Content)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: msg,
			},
		},
	}

	output := DeleteTaskOutput{
		Success: true,
		Message: msg,
	}

	return result, output, nil
}

// completeTaskHandler implements the todoist_complete_task tool
func completeTaskHandler(ctx context.Context, req *mcp.CallToolRequest, input CompleteTaskInput) (*mcp.CallToolResult, CompleteTaskOutput, error) {
	// Find the task
	task, err := findTaskByName(input.TaskName)
	if err != nil {
		return nil, CompleteTaskOutput{Success: false, Message: err.Error()}, err
	}

	if task == nil {
		msg := fmt.Sprintf("Could not find a task matching \"%s\"", input.TaskName)
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: msg,
				},
			},
			IsError: true,
		}
		return result, CompleteTaskOutput{Success: false, Message: msg}, nil
	}

	// Complete the task
	_, err = makeAPIRequest("POST", "/tasks/"+task.ID+"/close", nil)
	if err != nil {
		return nil, CompleteTaskOutput{Success: false, Message: err.Error()}, err
	}

	msg := fmt.Sprintf("Successfully completed task: \"%s\"", task.Content)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: msg,
			},
		},
	}

	output := CompleteTaskOutput{
		Success: true,
		Message: msg,
	}

	return result, output, nil
}

func main() {
	// Check for API token
	apiToken = os.Getenv("TODOIST_API_TOKEN")
	if apiToken == "" {
		log.Fatal("Error: TODOIST_API_TOKEN environment variable is required")
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "todoist-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "todoist_create_task",
		Description: "Create a new task in Todoist with optional description, due date, and priority",
	}, createTaskHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todoist_get_tasks",
		Description: "Get a list of tasks from Todoist with various filters",
	}, getTasksHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todoist_update_task",
		Description: "Update an existing task in Todoist by searching for it by name and then updating it",
	}, updateTaskHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todoist_delete_task",
		Description: "Delete a task from Todoist by searching for it by name",
	}, deleteTaskHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "todoist_complete_task",
		Description: "Mark a task as complete by searching for it by name",
	}, completeTaskHandler)

	// Log server start to stderr (stdout is used for MCP communication)
	fmt.Fprintf(os.Stderr, "Todoist MCP Server starting...\n")

	// Run the server with stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
