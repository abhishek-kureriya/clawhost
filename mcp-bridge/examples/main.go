package main

import (
"context"
"encoding/json"
"fmt"
"log/slog"
"os"
"time"

mcp "clawhost/mcp-bridge"
)

func main() {
// Initialize logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
Level: slog.LevelDebug,
}))
slog.SetDefault(logger)

// Create bridge instance
bridge := mcp.NewBridge(logger)
defer func() {
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := bridge.Close(ctx); err != nil {
logger.Error("bridge_close_failed", slog.Any("error", err))
}
}()

// Load the Big 5 manifest
manifest := mcp.BigFiveManifest()

// Define which services are enabled
// In production, this would come from database or environment config
enabled := map[string]bool{
"github":       os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN") != "",
"slack":        os.Getenv("SLACK_BOT_TOKEN") != "",
"postgres":     os.Getenv("PG_CONNECTION_STRING") != "",
"google-drive": os.Getenv("GOOGLE_CREDENTIALS_JSON") != "",
"tavily":       os.Getenv("TAVILY_API_KEY") != "",
}

// Validate that required env vars are present
missing := mcp.ValidateManifest(manifest, enabled)
for service, vars := range missing {
if !manifest.Services[service].Optional {
logger.Warn("service_will_be_skipped",
slog.String("service", service),
slog.Any("missing_env_vars", vars),
)
}
}

// Initialize services from manifest
if err := mcp.StartFromManifest(bridge, manifest, enabled, logger); err != nil {
logger.Error("failed_to_start_services", slog.Any("error", err))
os.Exit(1)
}

// Initialize tool registry
registry := mcp.NewToolRegistry(bridge, logger)

// Fetch available tools
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
if err := registry.Refresh(ctx); err != nil {
logger.Error("failed to refresh registry", slog.Any("error", err))
cancel()
os.Exit(1)
}
cancel()

stats := registry.GetStats()
logger.Info("registry_refreshed", slog.Any("stats", stats))

// Display enabled services and their tools
fmt.Println("\n" + "="*60)
fmt.Println("MCP BRIDGE - BIG 5 SERVICES")
fmt.Println("="*60)

services := mcp.ListServices(manifest)
for category, configs := range services {
fmt.Printf("\n📦 %s\n", categoryEmoji(category))
for _, config := range configs {
status := "❌"
if enabled[config.Name] {
status = "✅"
}
fmt.Printf("  %s %s - %s\n", status, config.DisplayName, config.Description)
fmt.Printf("     Tools: %v\n", config.ToolsProvided)
}
}

// Print all available tools in LLM format
fmt.Println("\n" + "="*60)
fmt.Println("AVAILABLE TOOLS FOR LLM")
fmt.Println("="*60)

tools := registry.GetAllToolsForLLM()
for i, tool := range tools {
fmt.Printf("[%d] %v\n", i+1, tool["function"].(map[string]interface{})["name"])
}

// Display health status
fmt.Println("\n" + "="*60)
fmt.Println("SERVER HEALTH")
fmt.Println("="*60)

health := bridge.HealthCheck()
for server, status := range health {
statusStr := "✅ healthy"
if !status {
statusStr = "❌ unhealthy"
}
fmt.Printf("%s: %s\n", server, statusStr)
}

// Demonstrate concurrent requests
fmt.Println("\n" + "="*60)
fmt.Println("CONCURRENT REQUEST DEMO")
fmt.Println("="*60)

demonstrateConcurrency(context.Background(), bridge, registry, logger)

// Example: Execute a tool (if GitHub is available)
fmt.Println("\n" + "="*60)
fmt.Println("SAMPLE TOOL EXECUTION")
fmt.Println("="*60)

if health["github"] {
toolID := "github_search_code"
params := json.RawMessage(`{"query":"go context","max_results":5}`)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
result, err := registry.ExecuteTool(ctx, toolID, params)
cancel()

if err != nil {
logger.Error("tool_execution_failed", slog.String("tool", toolID), slog.Any("error", err))
} else {
fmt.Printf("Result: %v\n", result)
}
} else {
fmt.Println("⚠️  GitHub service not enabled/configured")
}
}

func categoryEmoji(category string) string {
emojis := map[string]string{
"communication": "💬 Communication",
"database":      "🗄️ Database",
"cloud":         "☁️ Cloud",
"dev":           "👨‍💻 Development",
"search":        "🔍 Search",
}
if emoji, ok := emojis[category]; ok {
return emoji
}
return category
}

// demonstrateConcurrency shows how the bridge handles concurrent requests.
func demonstrateConcurrency(baseCtx context.Context, bridge *mcp.Bridge, registry *mcp.ToolRegistry, logger *slog.Logger) {
const numRequests = 5

// Create a channel for results
resultChan := make(chan string, numRequests)

// Issue multiple concurrent requests
for i := 1; i <= numRequests; i++ {
go func(requestNum int) {
ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
defer cancel()

// Determine which server based on requestNum
serverName := "github"
if requestNum%2 == 0 {
serverName = "postgres"
} else if requestNum%3 == 0 {
serverName = "slack"
}

// Create a sample tools/list request
req := &mcp.JSONRPCRequest{
JSONRPC: "2.0",
ID:      int64(requestNum),
Method:  "tools/list",
Params:  json.RawMessage(`{}`),
}

resp, err := bridge.RouteRequest(ctx, serverName, req)
if err != nil {
resultChan <- fmt.Sprintf("Request %d (server: %s): ERROR - %v", requestNum, serverName, err)
return
}

if resp.Error != nil {
resultChan <- fmt.Sprintf("Request %d (server: %s): RPC Error - %s", requestNum, serverName, resp.Error.Message)
return
}

resultChan <- fmt.Sprintf("Request %d (server: %s): SUCCESS", requestNum, serverName)
}(i)
}

// Collect results
for i := 0; i < numRequests; i++ {
select {
case result := <-resultChan:
fmt.Printf("  %s\n", result)
case <-time.After(6 * time.Second):
fmt.Println("  ⏱️  Request timeout!")
}
}
}
