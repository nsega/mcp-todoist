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

- Go 1.23 or later
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
