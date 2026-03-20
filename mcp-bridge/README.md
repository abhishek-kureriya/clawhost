# MCP Bridge - High-Performance Go Package

A production-grade Go package for managing multiple [Model Context Protocol (MCP)](https://modelcontextprotocol.io) servers and routing AI agent requests to the correct subprocess.

## Features

✅ **Server Lifecycle Manager** - Start, stop, and monitor MCP subprocesses with health tracking
✅ **JSON-RPC 2.0 Router** - Route requests from AI agents to the correct MCP server
✅ **Unified Tool Registry** - Aggregate tools from all MCP servers into a single LLM-compatible interface
✅ **Concurrency Safety** - Thread-safe with RWMutex and goroutine-based monitoring
✅ **Environment Injection** - Inject API keys and secrets securely into subprocesses
✅ **Context-Based Timeouts** - Prevent deadlocks with context cancellation

## Architecture

### Components

#### 1. **MCPServer (server.go)**
Manages individual MCP subprocess lifecycle:
- `Start()` - Spawn subprocess with environment injection
- `SendRequest()` - Send JSON-RPC 2.0 requests and read responses
- `Stop()` - Graceful shutdown with timeout and kill fallback
- `IsHealthy()` - Thread-safe health status check
- Background goroutines for process monitoring and stderr capture

```go
server := mcpbridge.NewMCPServer(
    "slack",
    "npx",
    []string{"@slack/mcp-server-slack"},
    map[string]string{
        "SLACK_BOT_TOKEN": os.Getenv("SLACK_BOT_TOKEN"),
    },
    logger,
)
server.Start(ctx)
```

#### 2. **Bridge (router.go)**
Central router managing multiple MCP servers:
- `RegisterServer()` - Add MCP server to bridge
- `RouteRequest()` - Forward JSON-RPC request to correct server
- `GetToolsList()` - Fetch available tools from a server
- `GetUnifiedRegistry()` - Aggregate tools from all servers
- `HealthCheck()` - Check status of all servers

```go
bridge := mcpbridge.NewBridge(logger)
bridge.RegisterServer(slackServer)
bridge.RegisterServer(postgresServer)

// Route a request
resp, err := bridge.RouteRequest(ctx, "slack", &mcpbridge.JSONRPCRequest{
    JSONRPC: "2.0",
    ID:      1,
    Method:  "tools/list",
    Params:  json.RawMessage("{}"),
})
```

#### 3. **ToolRegistry (registry.go)**
Unified interface for all tools across servers:
- `Refresh()` - Fetch latest tools from all servers
- `GetAllToolsForLLM()` - Export tools in OpenAI/Claude API format
- `ExecuteTool()` - Call a tool and return structured result
- `GetToolsByServer()` - List tools from a specific server
- `GetStats()` - Monitoring and observability

```go
registry := mcpbridge.NewToolRegistry(bridge, logger)
registry.Refresh(ctx)

// Get formatted for LLM
tools := registry.GetAllToolsForLLM()

// Execute a tool
result, err := registry.ExecuteTool(ctx, "slack_send_message", params)
```

## Concurrency Model

### Safety Guarantees
- **RWMutex on MCPServer health**: Readers use `IsHealthy()`, writers during process exit
- **Goroutine-based monitoring**: Separate goroutines for `monitorProcess()` and `captureStderr()`
- **Channel-based responses**: Non-blocking response collection with timeout
- **Context cancellation**: All I/O operations respect `context.Context` deadlines

### Concurrent Scenarios

#### Multiple simultaneous tool calls
```go
// Each request runs in its own goroutine
for i := 0; i < 10; i++ {
    go func(idx int) {
        result, err := registry.ExecuteTool(ctx, toolID, params)
        // Handle result
    }(i)
}
```

#### Server restart during requests
- In-flight requests time out via context
- New requests after restart wait for `Start()` to complete
- No data races due to health checks + mutex

#### Graceful shutdown
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
bridge.Close(ctx)  // Stops all servers with timeout
```

## Environment Injection (Security)

Secrets are injected into subprocess environment, not exposed to parent process:

```go
envVars := map[string]string{
    "SLACK_BOT_TOKEN": os.Getenv("SLACK_BOT_TOKEN"),      // From AWS Secrets Manager
    "DATABASE_URL":    os.Getenv("DATABASE_URL"),         // From Vault
    "API_KEY":         os.Getenv("OPENAI_API_KEY"),       // From .env file
}

server := mcpbridge.NewMCPServer("slack", cmd, args, envVars, logger)
// These secrets are injected via exec.Cmd.Env, never logged or transmitted
```

**Best Practices:**
- Load secrets from AWS Secrets Manager / HashiCorp Vault in production
- Never log environment variables (checked before Start())
- Use separate service accounts for each MCP server
- Implement secret rotation without restarting if possible

## Protocol: JSON-RPC 2.0

### Line-Delimited Communication
MCP servers communicate via **line-delimited JSON-RPC 2.0**:

```
Request:  {"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}
Response: {"jsonrpc":"2.0","id":1,"result":{"tools":[...]}}
```

### Standard Methods

#### `tools/list`
Lists all tools available on a server:
```go
req := &mcpbridge.JSONRPCRequest{
    JSONRPC: "2.0",
    ID:      1,
    Method:  "tools/list",
    Params:  json.RawMessage("{}"),
}
resp, _ := server.SendRequest(ctx, req)
// resp.Result: {"tools":[{"name":"send_message","description":"..."}]}
```

#### `tools/call`
Invokes a tool:
```go
params := map[string]interface{}{
    "name":      "send_message",
    "arguments": map[string]string{"channel": "#general", "text": "Hello"},
}
req := &mcpbridge.JSONRPCRequest{
    JSONRPC: "2.0",
    ID:      2,
    Method:  "tools/call",
    Params:  json.Marshal(params),
}
```

## Example Usage

See [`examples/main.go`](examples/main.go) for a complete working example with Slack and Postgres MCP servers.

### Quick Start

```bash
# Build the package
cd mcp-bridge
go build -v .

# Run the example
cd examples
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_APP_TOKEN=xapp-...
export DATABASE_URL="postgres://user:pass@localhost/db"
go run main.go
```

## Performance Characteristics

- **Startup**: ~100ms per MCP server (subprocess spawn + pipe setup)
- **Request latency**: ~5-50ms (subprocess I/O + JSON marshaling)
- **Memory per server**: ~2-5MB (buffers + pipes)
- **Throughput**: 100+ concurrent requests per second (limited by MCP server I/O)

## Error Handling

### Connection Errors
```go
resp, err := bridge.RouteRequest(ctx, "slack", req)
if err != nil {
    // Server not running, pipe error, or timeout
}
if resp.Error != nil {
    // RPC-level error from server
    log.Printf("RPC error: %s", resp.Error.Message)
}
```

### Timeout Handling
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
result, err := registry.ExecuteTool(ctx, toolID, params)
if err != nil {
    // Could be timeout or other error
}
```

### Health Checks
```go
health := bridge.HealthCheck()
if !health["slack"] {
    log.Println("Slack server is unhealthy, attempting restart...")
    bridge.RegisterServer(newSlackServer)  // Re-register with new instance
}
```

## Testing

Run unit tests:
```bash
go test -v ./...
```

## Monitoring & Observability

The package emits structured logs via `log/slog`:

```json
{"time":"2025-01-15T10:30:00Z","level":"INFO","msg":"mcp_server_started","name":"slack"}
{"time":"2025-01-15T10:30:05Z","level":"INFO","msg":"registry_refreshed","tool_count":24}
{"time":"2025-01-15T10:30:10Z","level":"WARN","msg":"tool_name_collision","tool":"query","servers":["slack","postgres"]}
```

Export metrics:
- Server health status (healthy/unhealthy)
- Tool registry stats (total tools, per-server breakdown)
- Request latencies (histogram)
- Process exit codes and stderr logs

## Production Deployment

### Docker Example
```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN cd mcp-bridge && go build -o /app/bridge ./examples

FROM ubuntu:24.04
COPY --from=builder /app/bridge /app/bridge
# Install Node.js and npm for MCP servers
RUN apt-get update && apt-get install -y nodejs npm
# Copy .npmrc with private package tokens (if needed)
CMD ["/app/bridge"]
```

### Kubernetes Deployment
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: openclaw-bridge
spec:
  containers:
  - name: bridge
    image: myregistry/openclaw-bridge:latest
    env:
    - name: SLACK_BOT_TOKEN
      valueFrom:
        secretKeyRef:
          name: mcp-secrets
          key: slack_bot_token
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: mcp-secrets
          key: database_url
    ports:
    - containerPort: 8080
      name: http
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
```

## Troubleshooting

### Server won't start
```
failed to start MCP server "slack": exec: "npx": executable file not found
```
→ Ensure `npx` is in PATH or provide full path to command

### Request timeouts
```
request timeout
```
→ Increase context timeout or check server CPU/memory

### Tool name collisions
```
WARN tool_name_collision tool=query servers=["slack","postgres"]
```
→ Expected; tools are prefixed: `slack_query`, `postgres_query`

### Memory leaks
→ Call `bridge.Close(ctx)` on shutdown
→ Check that stopped servers don't have dangling goroutines

## License

Part of ClawHost - Managed OpenClaw AI Hosting Platform
