package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	mcp "clawhost/mcp-bridge"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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

	// Define which services are enabled based on environment variables
	enabled := map[string]bool{
		"github":       os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN") != "",
		"slack":        os.Getenv("SLACK_BOT_TOKEN") != "",
		"postgres":     os.Getenv("PG_CONNECTION_STRING") != "",
		"google-drive": os.Getenv("GOOGLE_CREDENTIALS_JSON") != "",
		"tavily":       os.Getenv("TAVILY_API_KEY") != "",
	}

	// Validate that required environment variables are present
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

	// Fetch available tools from all services
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := registry.Refresh(ctx); err != nil {
		logger.Error("failed_to_refresh_registry", slog.Any("error", err))
		cancel()
		os.Exit(1)
	}
	cancel()

	stats := registry.GetStats()
	logger.Info("registry_refreshed", slog.Any("stats", stats))

	// Display available services
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("MCP BRIDGE - BIG 5 SERVICES")
	fmt.Println(strings.Repeat("=", 60))

	services := mcp.ListServices(manifest)
	for category, configs := range services {
		fmt.Printf("\n%s\n", categoryEmoji(category))
		for _, config := range configs {
			status := "❌"
			if enabled[config.Name] {
				status = "✅"
			}
			fmt.Printf("  %s %s\n", status, config.DisplayName)
			fmt.Printf("     Description: %s\n", config.Description)
			fmt.Printf("     Tools: %v\n", config.ToolsProvided)
		}
	}

	// Display health status
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SERVER HEALTH")
	fmt.Println(strings.Repeat("=", 60))

	health := bridge.HealthCheck()
	for server, status := range health {
		statusStr := "✅ healthy"
		if !status {
			statusStr = "❌ unhealthy"
		}
		fmt.Printf("  %s: %s\n", server, statusStr)
	}

	// Display available tools
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("AVAILABLE TOOLS FOR LLM")
	fmt.Println(strings.Repeat("=", 60))

	allTools := registry.ListAllTools()
	for _, tool := range allTools {
		fmt.Printf("  • %s_%s: %s\n", tool.ServerName, tool.OriginalName, tool.Description)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Bridge is ready! Connect your LLM to use these tools.")
	fmt.Println(strings.Repeat("=", 60) + "\n")
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
