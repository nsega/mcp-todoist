# Todoist MCP Server (Go)

A Model Context Protocol (MCP) server for Todoist, written in Go. This server enables Claude and other MCP clients to interact with your Todoist tasks using natural language.

This is a Go rewrite of the original [TypeScript implementation](https://github.com/abhiz123/todoist-mcp-server), built with the official [go-sdk v1.1.0](https://github.com/modelcontextprotocol/go-sdk).

## Features

- **Natural Language Task Management**: Create, update, complete, and delete tasks using everyday language
- **Smart Task Search**: Locate tasks via partial name matching
- **Flexible Filtering**: Organize tasks by due date, priority, and other attributes
- **Rich Task Details**: Support for descriptions, deadlines, and priority levels (1-4)
- **Intuitive Error Handling**: Clear feedback throughout operations

## Available Tools

### 1. `todoist_create_task`
Creates new tasks with optional description, due date, and priority level.

**Parameters:**
- `content` (required): The content/title of the task
- `description` (optional): Detailed description of the task
- `due_string` (optional): Natural language due date like 'tomorrow', 'next Monday', 'Jan 23'
- `priority` (optional): Task priority from 1 (normal) to 4 (urgent)

**Example:** "Create high priority task 'Fix bug' with description 'Critical performance issue'"

### 2. `todoist_get_tasks`
Retrieves and filters tasks using natural language date filtering and priority/project filtering.

**Parameters:**
- `project_id` (optional): Filter tasks by project ID
- `filter` (optional): Natural language filter like 'today', 'tomorrow', 'next week', 'priority 1', 'overdue'
- `priority` (optional): Filter by priority level (1-4)
- `limit` (optional): Maximum number of tasks to return (default: 10)

**Example:** "Show high priority tasks due this week"

### 3. `todoist_update_task`
Modifies existing tasks found via partial name matching. Can update content, description, due date, or priority.

**Parameters:**
- `task_name` (required): Name/content of the task to search for and update
- `content` (optional): New content/title for the task
- `description` (optional): New description for the task
- `due_string` (optional): New due date in natural language
- `priority` (optional): New priority level from 1 (normal) to 4 (urgent)

**Example:** "Update meeting task to be due next Monday"

### 4. `todoist_complete_task`
Marks tasks as finished using natural language search.

**Parameters:**
- `task_name` (required): Name/content of the task to search for and complete

**Example:** "Mark the documentation task as complete"

### 5. `todoist_delete_task`
Removes tasks by name with confirmation messages.

**Parameters:**
- `task_name` (required): Name/content of the task to search for and delete

**Example:** "Delete the PR review task"

## Prerequisites

- Go 1.25.4 or later
- A Todoist account
- Todoist API token

## Getting Your Todoist API Token

1. Log in to [Todoist](https://todoist.com)
2. Go to Settings → Integrations → Developer
3. Copy your API token from the "API token" section

## Installation

### From Source

```bash
git clone https://github.com/nsega/mcp-todoist.git
cd mcp-todoist
make build
```

The binary will be available at `build/mcp-todoist`.

### Using Go Install

```bash
go install github.com/nsega/mcp-todoist@latest
```

## Usage

### Running Directly

```bash
export TODOIST_API_TOKEN="your_api_token_here"
./build/mcp-todoist
```

### Using Make

```bash
make run TODOIST_API_TOKEN=your_api_token_here
```

## Configuration with Claude Desktop

Add the following to your Claude Desktop configuration file:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "todoist": {
      "command": "/path/to/mcp-todoist",
      "env": {
        "TODOIST_API_TOKEN": "your_api_token_here"
      }
    }
  }
}
```

Replace `/path/to/mcp-todoist` with the actual path to your built binary.

## Using the Todoist Tools

Once configured with Claude Desktop or another MCP client, you can interact with your Todoist tasks using natural language. Here are comprehensive examples for each tool:

### Creating Tasks

**Simple task:**
```
Create a task "Buy groceries"
```

**Task with description:**
```
Create a task "Review PR #123" with description "Check code quality and test coverage"
```

**Task with due date:**
```
Create a task "Team meeting" due tomorrow
```

**High priority task with all details:**
```
Create a high priority task "Deploy to production" with description "Deploy v2.0.0 release" due next Friday
```

**Priority levels:**
- Priority 1 (normal/default)
- Priority 2 (medium)
- Priority 3 (high)
- Priority 4 (urgent)

### Getting Tasks

**Get all tasks (default limit 10):**
```
Show me my tasks
```

**Get tasks with custom limit:**
```
Show me 20 tasks
```

**Filter by priority:**
```
Show me all priority 4 tasks
```

**Filter by due date:**
```
Show me tasks due today
Show me tasks due this week
Show me overdue tasks
```

**Combine filters:**
```
Show me high priority tasks due tomorrow
```

**Get tasks for a specific project:**
```
Show me tasks in project ID 2203306141
```

### Updating Tasks

**Update task name:**
```
Update the task "Buy groceries" to "Buy groceries and supplies"
```

**Change due date:**
```
Update the "Team meeting" task to be due next Monday
```

**Change priority:**
```
Update the "Deploy to production" task to priority 4
```

**Update multiple fields:**
```
Update the "Review PR" task with new description "Focus on security aspects" and make it due tomorrow with priority 3
```

**Note:** Task updates use partial name matching, so you only need to include enough of the task name to uniquely identify it.

### Completing Tasks

**Complete a task:**
```
Mark the "Buy groceries" task as complete
Complete the task "Team meeting"
```

**Partial name matching:**
```
Complete the "groceries" task
```

This will find and complete any task containing "groceries" in the name.

### Deleting Tasks

**Delete a task:**
```
Delete the "Old task" task
Remove the task "Cancelled meeting"
```

**Partial name matching:**
```
Delete the task containing "old"
```

**Note:** Like updates and completions, deletions use partial name matching for convenience.

### Natural Language Examples

The server is designed to work with natural language queries through Claude:

```
"Add a new task to buy milk tomorrow"
→ Creates task "Buy milk" due tomorrow

"What are my urgent tasks?"
→ Shows all priority 4 tasks

"I finished the documentation task"
→ Marks task containing "documentation" as complete

"Move my meeting to next week"
→ Updates the task containing "meeting" with new due date

"I don't need that old PR task anymore"
→ Deletes task containing "old PR"
```

### Tips for Best Results

1. **Be specific with task names** - When creating tasks, use clear, descriptive names for easier searching later
2. **Use partial matching wisely** - You can update/complete/delete tasks using just part of the name, but make sure it's unique enough
3. **Natural dates work** - Use phrases like "tomorrow", "next Monday", "in 3 days", "Jan 23"
4. **Priority is optional** - If you don't specify priority, tasks default to priority 1 (normal)
5. **Filters are flexible** - Combine multiple filters (priority, date, project) to find exactly what you need

## Development

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Code Coverage

```bash
make coverage
```

### Linting

```bash
make lint
```

### Formatting

```bash
make fmt
```

### Running All Checks

```bash
make check
```

### Building for Multiple Platforms

```bash
make build-all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Project Structure

```
mcp-todoist/
├── main.go                 # Main server implementation
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── Makefile               # Build automation
├── README.md              # This file
├── LICENSE                # MIT License
├── .github/
│   └── workflows/
│       └── build-and-test.yml  # CI/CD workflow
└── build/                 # Build output directory
```

## API Integration

This server uses the [Todoist REST API v2](https://developer.todoist.com/rest/v2/) directly via HTTP requests. The implementation includes:

- Task creation with full metadata support
- Task retrieval with filtering capabilities
- Task updates with partial name matching
- Task completion and deletion
- Comprehensive error handling

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original TypeScript implementation by [abhiz123](https://github.com/abhiz123/todoist-mcp-server)
- Built with [go-sdk v1.1.0](https://github.com/modelcontextprotocol/go-sdk)
- GitHub Actions and Makefile structure inspired by [mcp-obsidian](https://github.com/nsega/mcp-obsidian)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/nsega/mcp-todoist/issues) on GitHub.
