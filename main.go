package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nsega/mcp-todoist/internal/todoist"
	"github.com/nsega/mcp-todoist/internal/tools"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		slog.Error("TODOIST_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	client := todoist.NewClient(token)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "todoist-mcp-server",
		Version: "1.0.0",
	}, nil)

	tools.RegisterAll(server, client)

	slog.Info("Todoist MCP Server starting")

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
