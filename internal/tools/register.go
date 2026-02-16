package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
)

// RegisterAll registers all MCP tools on the server.
func RegisterAll(s *mcp.Server, c *todoist.Client) {
	registerTaskTools(s, c)
	registerProjectTools(s, c)
	registerSectionTools(s, c)
	registerLabelTools(s, c)
	registerCommentTools(s, c)
	registerGTDTools(s, c)
}
