# Todoist MCP Server (Go)

[![Go Version](https://img.shields.io/badge/Go-1.25.7-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server for Todoist, written in Go. This server enables Claude and other MCP clients to interact with your Todoist tasks, projects, sections, labels, and comments using natural language. Includes GTD workflow tools for inbox processing and weekly reviews.

This is a Go rewrite of the original [TypeScript implementation](https://github.com/abhiz123/todoist-mcp-server), built with the official [go-sdk v1.3.0](https://github.com/modelcontextprotocol/go-sdk).

## Features

- **Full Todoist API Coverage**: 29 tools covering tasks, projects, sections, labels, and comments
- **GTD Workflow Support**: Inbox review, weekly review, task moving, and bulk creation
- **Smart Task Search**: Locate tasks via exact or partial name matching
- **Flexible Filtering**: Organize tasks by due date, priority, project, and more
- **Task ID Support**: Use task IDs directly or search by name

## Available Tools

### Task Tools (6)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `todoist_create_task` | Create a new task | `content`, `description`, `due_string`, `priority`, `project_id`, `section_id`, `parent_id`, `labels`, `assignee_id` |
| `todoist_get_tasks` | List tasks with filters | `project_id`, `filter`, `priority`, `limit` |
| `todoist_update_task` | Update a task by ID or name | `task_id`/`task_name`, `content`, `description`, `due_string`, `priority`, `labels`, `assignee_id` |
| `todoist_delete_task` | Delete a task | `task_id`/`task_name` |
| `todoist_complete_task` | Mark a task as complete | `task_id`/`task_name` |
| `todoist_reopen_task` | Reopen a completed task | `task_id`/`task_name` |

### Project Tools (7)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `todoist_get_projects` | List all projects | — |
| `todoist_get_project` | Get a single project | `project_id` |
| `todoist_create_project` | Create a project | `name`, `parent_id`, `color`, `is_favorite`, `view_style` |
| `todoist_update_project` | Update a project | `project_id`, `name`, `color`, `is_favorite` |
| `todoist_delete_project` | Delete a project | `project_id` |
| `todoist_archive_project` | Archive a project | `project_id` |
| `todoist_unarchive_project` | Unarchive a project | `project_id` |

### Section Tools (4)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `todoist_get_sections` | List sections | `project_id` (optional) |
| `todoist_create_section` | Create a section | `name`, `project_id`, `order` |
| `todoist_update_section` | Update a section | `section_id`, `name` |
| `todoist_delete_section` | Delete a section | `section_id` |

### Label Tools (4)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `todoist_get_labels` | List all labels | — |
| `todoist_create_label` | Create a label | `name`, `color`, `is_favorite` |
| `todoist_update_label` | Update a label | `label_id`, `name`, `color` |
| `todoist_delete_label` | Delete a label | `label_id` |

### Comment Tools (4)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `todoist_get_comments` | List comments | `task_id` or `project_id` |
| `todoist_create_comment` | Add a comment | `content`, `task_id` or `project_id` |
| `todoist_update_comment` | Update a comment | `comment_id`, `content` |
| `todoist_delete_comment` | Delete a comment | `comment_id` |

### GTD Workflow Tools (4)

| Tool | Description | How It Works |
|------|-------------|--------------|
| `todoist_inbox_review` | Inbox processing view | Auto-detects inbox project, groups tasks by age (today/this week/older) |
| `todoist_weekly_review` | Weekly review summary | Aggregates: projects with task counts, overdue tasks, tasks with no due date |
| `todoist_move_task` | Move task to project/section | `task_id`/`task_name`, `project_id`, `section_id` |
| `todoist_bulk_create_tasks` | Batch create tasks | `tasks[]` array with content, description, due_string, priority, project_id, section_id, labels |

## Prerequisites

- Go 1.25.7 or later
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

## Example Workflows

### GTD Inbox Processing

```
Review my Todoist inbox
→ Runs todoist_inbox_review, shows tasks grouped by age

Move the "research API" task to project Work, section Backlog
→ Runs todoist_move_task with project and section IDs
```

### Weekly Review

```
Run my weekly review
→ Runs todoist_weekly_review, shows project summaries, overdue tasks, undated tasks
```

### Batch Task Creation

```
Create these tasks in my "Reading List" project:
- Read "Thinking, Fast and Slow"
- Read "Deep Work"
- Read "Atomic Habits"
→ Runs todoist_bulk_create_tasks with 3 items
```

### Project Management

```
Show me all my projects
→ Lists all projects with IDs and inbox/favorite status

Create a new project called "Q1 Goals" with board view
→ Creates project with view_style: board
```

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

### Running All Checks

```bash
make check
```

### Building for Multiple Platforms

```bash
make build-all
```

## Project Structure

```
mcp-todoist/
├── main.go                          # Thin entry point
├── internal/
│   ├── models/                      # Shared data types
│   │   ├── task.go
│   │   ├── project.go
│   │   ├── section.go
│   │   ├── label.go
│   │   └── comment.go
│   ├── todoist/                     # API client (no MCP awareness)
│   │   ├── client.go
│   │   ├── tasks.go
│   │   ├── projects.go
│   │   ├── sections.go
│   │   ├── labels.go
│   │   └── comments.go
│   └── tools/                       # MCP tool handlers
│       ├── register.go
│       ├── tasks.go
│       ├── projects.go
│       ├── sections.go
│       ├── labels.go
│       ├── comments.go
│       └── gtd.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## API Integration

This server uses the [Todoist REST API v2](https://developer.todoist.com/rest/v2/) with full coverage of:

- Tasks: CRUD, complete, reopen, search by name or ID
- Projects: CRUD, archive, unarchive
- Sections: CRUD within projects
- Labels: CRUD for personal labels
- Comments: CRUD on tasks and projects
- GTD: Inbox review, weekly review, task moving, bulk creation

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original TypeScript implementation by [abhiz123](https://github.com/abhiz123/todoist-mcp-server)
- Built with [go-sdk v1.3.0](https://github.com/modelcontextprotocol/go-sdk)
- GitHub Actions and Makefile structure inspired by [mcp-obsidian](https://github.com/nsega/mcp-obsidian)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/nsega/mcp-todoist/issues) on GitHub.
