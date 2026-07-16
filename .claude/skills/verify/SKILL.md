---
name: verify
description: Runtime verification recipe for mcp-todoist - drive the stdio MCP binary against a fake Todoist API in a Linux container to observe wire behavior (headers, retries) without touching the real API.
---

# Verifying mcp-todoist changes at the runtime surface

The surface is the stdio MCP binary. The client hard-codes
`https://api.todoist.com`, and macOS Go ignores `SSL_CERT_FILE`, so wire
observation needs a Linux container with a fake TLS Todoist.

## Recipe (worked 2026-07-15)

1. Build linux binary from repo root: `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $SCRATCH/mcp-todoist-linux .`
2. Fake server: Go program that self-signs a CA, writes it to `/work/ca.pem`, serves TLS as `api.todoist.com` on `127.0.0.1:443`, logs `method path x-request-id -> status` per request, and injects failures (500 on first POST /api/v1/tasks, 429 + Retry-After on first GET). See the session scratchpad `fake-todoist/main.go` pattern.
3. Run: `docker run --rm --add-host api.todoist.com:127.0.0.1 -v $SCRATCH:/work alpine sh /work/run-inside.sh` where the script starts the fake, then pipes newline-delimited JSON-RPC (initialize, notifications/initialized, tools/call ...) into the binary with `TODOIST_API_TOKEN=dummy SSL_CERT_FILE=/work/ca.pem`, with 1-4s sleeps between messages.
4. Evidence: `/work/requests.log` (wire) plus the binary's stdout (MCP responses).

## Gotchas

- Pipelined stdin works; the go-sdk processes messages in order and exits on EOF, so keep stdin open (sleeps) until responses land.
- Docker Desktop may block on a privileged-helper GUI prompt at first start; the `orbstack` docker context on this machine is stale (socket missing).
- `GODEBUG=http2debug=2` on the darwin binary dumps decoded header fields to stderr if you ever need real-API observation, but creating real tasks needs explicit user permission.
- `todoist_create_task` output has no task ID; use `todoist_delete_task` with `task_id` from a get_tasks call (or fake-server-known IDs) for cleanup.
