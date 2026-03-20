package mcpbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// ToolRegistry provides a unified interface to all tools across MCP servers.
type ToolRegistry struct {
	bridge           *Bridge
	tools            map[string]*RegistryTool // key: "server_originalName"
	toolsByServer    map[string][]string      // key: serverName, value: list of tool IDs
	mu               sync.RWMutex
	logger           *slog.Logger
	lastRefresh      time.Time
	refreshInterval  time.Duration
}

// RegistryTool represents a tool with routing information.
type RegistryTool struct {
	ServerName   string          `json:"server_name"`
	OriginalName string          `json:"original_name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
	CreatedAt    time.Time       `json:"created_at"`
	Hash         string          `json:"hash"`
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry(bridge *Bridge, logger *slog.Logger) *ToolRegistry {
	if logger == nil {
		logger = slog.Default()
	}
	return &ToolRegistry{
		bridge:          bridge,
		tools:           make(map[string]*RegistryTool),
		toolsByServer:   make(map[string][]string),
		logger:          logger,
		refreshInterval: 30 * time.Second,
	}
}

// Refresh fetches the latest tools from all MCP servers.
func (tr *ToolRegistry) Refresh(ctx context.Context) error {
	registry, err := tr.bridge.GetUnifiedRegistry(ctx)
	if err != nil {
		return fmt.Errorf("failed to get unified registry: %w", err)
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.tools = make(map[string]*RegistryTool)
	tr.toolsByServer = make(map[string][]string)

	// Track collisions
	collisions := make(map[string][]string)

	for _, toolInfo := range registry.AllTools {
		toolID := fmt.Sprintf("%s_%s", toolInfo.Server, toolInfo.OriginalName)

		// Check for collisions
		if original, exists := collisions[toolInfo.OriginalName]; exists {
			collisions[toolInfo.OriginalName] = append(original, toolInfo.Server)
		} else {
			collisions[toolInfo.OriginalName] = []string{toolInfo.Server}
		}

		registryTool := &RegistryTool{
			ServerName:   toolInfo.Server,
			OriginalName: toolInfo.OriginalName,
			Description:  toolInfo.Description,
			InputSchema:  toolInfo.InputSchema,
			CreatedAt:    time.Now(),
			Hash:         hashTool(&toolInfo),
		}

		tr.tools[toolID] = registryTool
		tr.toolsByServer[toolInfo.Server] = append(tr.toolsByServer[toolInfo.Server], toolID)
	}

	// Log collisions
	for toolName, servers := range collisions {
		if len(servers) > 1 {
			tr.logger.Warn("tool_name_collision", slog.String("tool", toolName), slog.Any("servers", servers))
		}
	}

	tr.lastRefresh = time.Now()
	tr.logger.Info("registry refreshed", slog.Int("tool_count", len(tr.tools)))

	return nil
}

// GetTool retrieves a tool by its full ID (server_originalName).
func (tr *ToolRegistry) GetTool(toolID string) *RegistryTool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.tools[toolID]
}

// GetToolsByServer retrieves all tools for a specific server.
func (tr *ToolRegistry) GetToolsByServer(serverName string) []*RegistryTool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	toolIDs := tr.toolsByServer[serverName]
	tools := make([]*RegistryTool, 0, len(toolIDs))
	for _, toolID := range toolIDs {
		if tool, exists := tr.tools[toolID]; exists {
			tools = append(tools, tool)
		}
	}
	return tools
}

// ListAllTools returns all available tools sorted by server and name.
func (tr *ToolRegistry) ListAllTools() []*RegistryTool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	tools := make([]*RegistryTool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}

	// Sort by server name, then tool name
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].ServerName != tools[j].ServerName {
			return tools[i].ServerName < tools[j].ServerName
		}
		return tools[i].OriginalName < tools[j].OriginalName
	})

	return tools
}

// GetToolForLLM returns a tool formatted for LLM consumption.
func (tr *ToolRegistry) GetToolForLLM(toolID string) map[string]interface{} {
	tool := tr.GetTool(toolID)
	if tool == nil {
		return nil
	}

	// Parse input schema
	var schema interface{}
	_ = json.Unmarshal(tool.InputSchema, &schema)

	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        toolID,
			"description": tool.Description,
			"parameters":  schema,
		},
	}
}

// GetAllToolsForLLM returns all tools formatted for OpenAI function calling.
func (tr *ToolRegistry) GetAllToolsForLLM() []map[string]interface{} {
	tools := tr.ListAllTools()
	result := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		toolID := fmt.Sprintf("%s_%s", tool.ServerName, tool.OriginalName)
		result[i] = tr.GetToolForLLM(toolID)
	}
	return result
}

// ExecuteTool calls a tool on its server and returns the result.
func (tr *ToolRegistry) ExecuteTool(ctx context.Context, toolID string, params json.RawMessage) (interface{}, error) {
	tool := tr.GetTool(toolID)
	if tool == nil {
		return nil, fmt.Errorf("tool %q not found", toolID)
	}

	result, err := tr.bridge.CallTool(ctx, tool.ServerName, tool.OriginalName, params)
	if err != nil {
		tr.logger.Error("tool_execution_failed", slog.String("tool_id", toolID), slog.Any("error", err))
		return nil, err
	}

	// Extract text from response
	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return nil, nil
}

// GetStats returns statistics about the registry.
func (tr *ToolRegistry) GetStats() map[string]interface{} {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	serverStats := make(map[string]int)
	for server, toolIDs := range tr.toolsByServer {
		serverStats[server] = len(toolIDs)
	}

	return map[string]interface{}{
		"total_tools":     len(tr.tools),
		"total_servers":   len(tr.toolsByServer),
		"by_server":       serverStats,
		"last_refresh":    tr.lastRefresh.String(),
	}
}

// hashTool creates a deterministic hash of a tool for change detection.
func hashTool(tool *ToolInfo) string {
	// Simple hash combining server, name, and description
	h := tool.Server + ":" + tool.OriginalName + ":" + tool.Description
	// In production, could use sha256 or similar
	return fmt.Sprintf("%x", len(h)) // Placeholder; implement proper hashing as needed
}

// SyncWithBridge periodically refreshes the registry from the bridge.
func (tr *ToolRegistry) SyncWithBridge(ctx context.Context) error {
	ticker := time.NewTicker(tr.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := tr.Refresh(ctx); err != nil {
				tr.logger.Error("sync_failed", slog.Any("error", err))
			}
		}
	}
}
