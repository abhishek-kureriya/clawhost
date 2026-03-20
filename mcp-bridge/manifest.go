package mcpbridge

import (
	"fmt"
	"log/slog"
	"os"
)

// ServiceConfig defines how an MCP server is spawned and configured.
type ServiceConfig struct {
	Name          string            `json:"name"`
	DisplayName   string            `json:"display_name"`
	Description   string            `json:"description"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	EnvVars       []string          `json:"env_vars"`        // Required environment variable names
	Optional      bool              `json:"optional"`        // If true, missing env vars don't block startup
	Category      string            `json:"category"`        // "communication", "database", "cloud", "dev", "search"
	ToolsProvided []string          `json:"tools_provided"`  // List of key tools this service provides
}

// ServiceManifest defines all available MCP services.
type ServiceManifest struct {
	Services map[string]*ServiceConfig `json:"services"`
}

// BigFiveManifest returns the pre-configured "Big 5" services for ClawHost.
// These are the most common services users will want to integrate.
func BigFiveManifest() *ServiceManifest {
	return &ServiceManifest{
		Services: map[string]*ServiceConfig{
			"github": {
				Name:        "github",
				DisplayName: "GitHub",
				Description: "Code repository management, search, issues, and pull requests",
				Command:     "npx",
				Args: []string{
					"-y",
					"@modelcontextprotocol/server-github",
				},
				EnvVars: []string{"GITHUB_PERSONAL_ACCESS_TOKEN"},
				Optional: false,
				Category: "dev",
				ToolsProvided: []string{
					"search_code",
					"create_pull_request",
					"list_issues",
					"get_issue",
					"create_issue",
					"list_repos",
				},
			},
			"slack": {
				Name:        "slack",
				DisplayName: "Slack",
				Description: "Team messaging, channels, and thread management",
				Command:     "npx",
				Args: []string{
					"-y",
					"@modelcontextprotocol/server-slack",
				},
				EnvVars: []string{"SLACK_BOT_TOKEN", "SLACK_APP_TOKEN"},
				Optional: false,
				Category: "communication",
				ToolsProvided: []string{
					"list_channels",
					"post_message",
					"reply_to_thread",
					"search_messages",
					"get_user_info",
					"list_users",
				},
			},
			"postgres": {
				Name:        "postgres",
				DisplayName: "PostgreSQL",
				Description: "Database querying, schema inspection, and data analysis",
				Command:     "npx",
				Args: []string{
					"-y",
					"@modelcontextprotocol/server-postgres",
				},
				EnvVars: []string{"PG_CONNECTION_STRING"},
				Optional: false,
				Category: "database",
				ToolsProvided: []string{
					"query",
					"list_tables",
					"describe_table",
					"analyze_table",
					"get_schema",
				},
			},
			"google-drive": {
				Name:        "google-drive",
				DisplayName: "Google Drive",
				Description: "File storage, document reading, and search integration",
				Command:     "npx",
				Args: []string{
					"-y",
					"@modelcontextprotocol/server-google-drive",
				},
				EnvVars: []string{"GOOGLE_CREDENTIALS_JSON"},
				Optional: true,
				Category: "cloud",
				ToolsProvided: []string{
					"list_files",
					"read_file",
					"search_files",
					"create_file",
					"share_file",
				},
			},
			"tavily": {
				Name:        "tavily",
				DisplayName: "Tavily (Web Search)",
				Description: "Real-time web search and context retrieval",
				Command:     "npx",
				Args: []string{
					"-y",
					"@modelcontextprotocol/server-tavily",
				},
				EnvVars: []string{"TAVILY_API_KEY"},
				Optional: true,
				Category: "search",
				ToolsProvided: []string{
					"search",
					"get_search_context",
					"search_news",
					"search_recent",
				},
			},
		},
	}
}

// StartFromManifest initializes and starts MCP servers from a manifest based on enabled services.
func StartFromManifest(bridge *Bridge, manifest *ServiceManifest, enabled map[string]bool, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	for serviceName, serviceConfig := range manifest.Services {
		// Skip if not enabled
		if enabled[serviceName] == false {
			logger.Info("service_skipped", slog.String("service", serviceName))
			continue
		}

		// Check if required environment variables are present
		missingVars := []string{}
		envMap := make(map[string]string)

		for _, envVar := range serviceConfig.EnvVars {
			value := os.Getenv(envVar)
			if value == "" {
				missingVars = append(missingVars, envVar)
			} else {
				envMap[envVar] = value
			}
		}

		// If required vars are missing and service is not optional, skip or error
		if len(missingVars) > 0 && !serviceConfig.Optional {
			logger.Warn("service_skipped_missing_env",
				slog.String("service", serviceName),
				slog.Any("missing_vars", missingVars),
			)
			continue
		}

		// Create and register the server
		server := NewMCPServer(
			serviceConfig.Name,
			serviceConfig.Command,
			serviceConfig.Args,
			envMap,
			logger,
		)

		if err := bridge.RegisterServer(server); err != nil {
			logger.Error("failed_to_register_service",
				slog.String("service", serviceName),
				slog.Any("error", err),
			)
			continue
		}

		logger.Info("service_registered",
			slog.String("service", serviceName),
			slog.String("display_name", serviceConfig.DisplayName),
			slog.Any("tools", serviceConfig.ToolsProvided),
		)
	}

	return nil
}

// ValidateManifest checks that all required environment variables for enabled services are set.
func ValidateManifest(manifest *ServiceManifest, enabled map[string]bool) map[string][]string {
	missingByService := make(map[string][]string)

	for serviceName, serviceConfig := range manifest.Services {
		if enabled[serviceName] == false {
			continue
		}

		for _, envVar := range serviceConfig.EnvVars {
			if os.Getenv(envVar) == "" {
				missingByService[serviceName] = append(missingByService[serviceName], envVar)
			}
		}
	}

	return missingByService
}

// GetServiceInfo returns detailed information about a specific service.
func GetServiceInfo(manifest *ServiceManifest, serviceName string) (*ServiceConfig, error) {
	if config, exists := manifest.Services[serviceName]; exists {
		return config, nil
	}
	return nil, fmt.Errorf("service %q not found in manifest", serviceName)
}

// ListServices returns all available services grouped by category.
func ListServices(manifest *ServiceManifest) map[string][]*ServiceConfig {
	categorized := make(map[string][]*ServiceConfig)

	for _, config := range manifest.Services {
		categorized[config.Category] = append(categorized[config.Category], config)
	}

	return categorized
}
