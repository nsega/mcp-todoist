package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ActionOutput is the shared structured output for all tools: a success
// flag plus the same human-readable message returned as text content.
type ActionOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func textResult(msg string, isError bool) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: isError,
	}
}
