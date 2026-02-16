package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
	"github.com/nsega/mcp-todoist/internal/tools"
)

func main() {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		log.Fatal("Error: TODOIST_API_TOKEN environment variable is required")
	}

	client := todoist.NewClient(token)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "todoist-mcp-server",
		Version: "1.0.0",
	}, nil)

	tools.RegisterAll(server, client)

	fmt.Fprintf(os.Stderr, "Todoist MCP Server starting...\n")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
