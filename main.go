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

	slog.Info("todoist MCP server starting", "version", "1.0.0")

	// go-sdk v1.8.0 caps inbound JSON-RPC frames at mcp.DefaultMaxLineLength (16 MiB) and
	// ends the session on overflow. Kept deliberately: no legitimate call to this server
	// approaches that size, and the cap bounds memory per frame.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
