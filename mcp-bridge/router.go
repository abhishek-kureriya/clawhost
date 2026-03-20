package mcpbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Bridge is the main JSON-RPC router that manages multiple MCP servers.
type Bridge struct {
	servers map[string]*MCPServer
	mu      sync.RWMutex
	logger  *slog.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewBridge creates a new MCP bridge instance.
func NewBridge(logger *slog.Logger) *Bridge {
	if logger == nil {
		logger = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Bridge{
		servers: make(map[string]*MCPServer),
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterServer registers an MCP server with the bridge.
func (b *Bridge) RegisterServer(server *MCPServer) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.servers[server.Name]; exists {
		return fmt.Errorf("server %q already registered", server.Name)
	}

	if err := server.Start(b.ctx); err != nil {
		return fmt.Errorf("failed to start server %q: %w", server.Name, err)
	}

	b.servers[server.Name] = server
	b.logger.Info("server_registered", slog.String("name", server.Name))
	return nil
}

// RouteRequest routes a JSON-RPC request to the appropriate MCP server.
func (b *Bridge) RouteRequest(ctx context.Context, serverName string, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	b.mu.RLock()
	server, exists := b.servers[serverName]
	b.mu.RUnlock()

	if !exists {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: fmt.Sprintf("server %q not found", serverName),
			},
		}, nil
	}

	// Add timeout if not present
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	resp, err := server.SendRequest(ctx, req)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: fmt.Sprintf("internal error: %v", err),
			},
		}, nil
	}

	return resp, nil
}

// GetToolsList retrieves all available tools from a specific MCP server.
func (b *Bridge) GetToolsList(ctx context.Context, serverName string) (*ToolsListResponse, error) {
	b.mu.RLock()
	_, exists := b.servers[serverName]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %q not found", serverName)
	}

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  json.RawMessage(`{}`),
	}

	resp, err := b.RouteRequest(ctx, serverName, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", resp.Error.Message)
	}

	var toolsResp ToolsListResponse
	if err := json.Unmarshal(resp.Result, &toolsResp); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &toolsResp, nil
}

// CallTool invokes a tool on a specific MCP server.
func (b *Bridge) CallTool(ctx context.Context, serverName, toolName string, params json.RawMessage) (*ToolCallResponse, error) {
	b.mu.RLock()
	_, exists := b.servers[serverName]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %q not found", serverName)
	}

	// Build tools/call request
	callParams := map[string]interface{}{
		"name": toolName,
		"arguments": json.RawMessage(params),
	}
	callParamsBytes, _ := json.Marshal(callParams)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  callParamsBytes,
	}

	resp, err := b.RouteRequest(ctx, serverName, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", resp.Error.Message)
	}

	var toolResp ToolCallResponse
	if err := json.Unmarshal(resp.Result, &toolResp); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return &toolResp, nil
}

// GetUnifiedRegistry returns a unified registry of all tools across all servers.
func (b *Bridge) GetUnifiedRegistry(ctx context.Context) (*UnifiedRegistry, error) {
	b.mu.RLock()
	serverNames := make([]string, 0, len(b.servers))
	for name := range b.servers {
		serverNames = append(serverNames, name)
	}
	b.mu.RUnlock()

	registry := &UnifiedRegistry{
		Servers: make(map[string]ServerTools),
		AllTools: make([]ToolInfo, 0),
	}

	// Fetch tools from all servers concurrently
	type serverToolsResult struct {
		name  string
		tools *ToolsListResponse
		err   error
	}

	resultChan := make(chan serverToolsResult, len(serverNames))
	for _, name := range serverNames {
		go func(serverName string) {
			tools, err := b.GetToolsList(ctx, serverName)
			resultChan <- serverToolsResult{serverName, tools, err}
		}(name)
	}

	// Collect results with timeout
	collectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	collected := 0
	for collected < len(serverNames) {
		select {
		case result := <-resultChan:
			collected++
			if result.err != nil {
				b.logger.Warn("failed_to_fetch_tools", slog.String("server", result.name), slog.Any("error", result.err))
				continue
			}

			serverTools := ServerTools{
				ServerName: result.name,
				Tools:      result.tools.Tools,
			}
			registry.Servers[result.name] = serverTools

			// Add to unified list with server prefix
			for _, tool := range result.tools.Tools {
				unified := ToolInfo{
					Name:        fmt.Sprintf("%s_%s", result.name, tool.Name),
					Server:      result.name,
					OriginalName: tool.Name,
					Description: tool.Description,
					InputSchema: tool.InputSchema,
				}
				registry.AllTools = append(registry.AllTools, unified)
			}

		case <-collectCtx.Done():
			b.logger.Warn("timeout collecting tools from MCP servers")
			return registry, nil
		}
	}

	return registry, nil
}

// HealthCheck returns the health status of all registered servers.
func (b *Bridge) HealthCheck() map[string]bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	status := make(map[string]bool)
	for name, server := range b.servers {
		status[name] = server.IsHealthy()
	}
	return status
}

// Close gracefully shuts down all MCP servers.
func (b *Bridge) Close(ctx context.Context) error {
	b.cancel()

	b.mu.Lock()
	defer b.mu.Unlock()

	var lastErr error
	for name, server := range b.servers {
		if err := server.Stop(ctx); err != nil {
			b.logger.Error("failed to stop server", slog.String("name", name), slog.Any("error", err))
			lastErr = err
		}
	}

	return lastErr
}

// Tool-related response structures
type ToolsListResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type ToolCallResponse struct {
	Content []ToolContent `json:"content"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// UnifiedRegistry aggregates tools from all MCP servers
type UnifiedRegistry struct {
	Servers  map[string]ServerTools `json:"servers"`
	AllTools []ToolInfo             `json:"all_tools"`
}

type ServerTools struct {
	ServerName string            `json:"server_name"`
	Tools      []ToolDefinition  `json:"tools"`
}

type ToolInfo struct {
	Name        string          `json:"name"`
	Server      string          `json:"server"`
	OriginalName string         `json:"original_name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}
