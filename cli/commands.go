package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yourusername/clawhost/core/provisioning"
)

const (
	stateFileName = "deployments.json"
	stateDirName  = ".clawhost"
)

type deploymentState struct {
	Deployments map[string]deployment `json:"deployments"`
}

type deployment struct {
	Name       string    `json:"name"`
	Provider   string    `json:"provider"`
	ServerID   string    `json:"server_id"`
	PublicIP   string    `json:"public_ip"`
	Status     string    `json:"status"`
	ServerType string    `json:"server_type"`
	Location   string    `json:"location"`
	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func runCLI(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		printCLIUsage()
		return nil
	case "init":
		return runInit(args[1:])
	case "deploy":
		return runDeploy(args[1:])
	case "status":
		return runStatus(args[1:])
	case "logs":
		return runLogs(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "backup":
		return runBackup(args[1:])
	case "restore":
		return runRestore(args[1:])
	case "destroy":
		return runDestroy(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printCLIUsage() {
	fmt.Println("clawhost <command> [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init      Initialize new deployment")
	fmt.Println("  deploy    Deploy OpenClaw instance")
	fmt.Println("  status    Check deployment status")
	fmt.Println("  logs      View logs")
	fmt.Println("  update    Update OpenClaw version")
	fmt.Println("  backup    Create backup")
	fmt.Println("  restore   Restore from backup")
	fmt.Println("  destroy   Remove deployment")
	fmt.Println("")
	fmt.Println("If no command is provided, clawhost starts the Core API server.")
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment profile name")
	provider := fs.String("provider", "hetzner", "Cloud provider")
	serverType := fs.String("server-type", "cx11", "Server type")
	location := fs.String("location", "nbg1", "Server location")
	llmProvider := fs.String("llm-provider", "openai", "LLM provider")
	llmModel := fs.String("llm-model", "gpt-4o-mini", "LLM model")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing config file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	configPath := ".clawhost.yaml"
	if _, err := os.Stat(configPath); err == nil && !*overwrite {
		return fmt.Errorf("%s already exists (use --overwrite)", configPath)
	}

	config := fmt.Sprintf("name: %s\nprovider: %s\nserver_type: %s\nlocation: %s\nllm_provider: %s\nllm_model: %s\n", *name, *provider, *serverType, *location, *llmProvider, *llmModel)
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	if _, ok := state.Deployments[*name]; !ok {
		now := time.Now().UTC()
		state.Deployments[*name] = deployment{
			Name:       *name,
			Provider:   *provider,
			ServerType: *serverType,
			Location:   *location,
			Status:     "initialized",
			Version:    "openclaw-latest",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := saveState(state); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Initialized deployment profile %q\n", *name)
	fmt.Printf("Config written to %s\n", configPath)
	return nil
}

func runDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	serverType := fs.String("server-type", "cx11", "Hetzner server type")
	location := fs.String("location", "nbg1", "Hetzner location")
	sshKey := fs.String("ssh-key", "", "Hetzner SSH key name")
	llmProvider := fs.String("llm-provider", "openai", "LLM provider")
	llmModel := fs.String("llm-model", "gpt-4o-mini", "LLM model")
	personality := fs.String("personality", "You are a helpful AI assistant.", "AI personality prompt")
	businessKnowledge := fs.String("business-knowledge", "", "Business knowledge prompt")
	waitReady := fs.Bool("wait", true, "Wait for server to reach running")
	timeout := fs.Duration("timeout", 15*time.Minute, "Wait timeout")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	apiToken := strings.TrimSpace(os.Getenv("HETZNER_API_TOKEN"))
	if apiToken == "" {
		return errors.New("HETZNER_API_TOKEN is required")
	}

	provider := provisioning.NewHetznerProvider(apiToken)
	cloudInit := provisioning.GenerateCloudInitScript(provisioning.OpenClawConfig{
		LLMProvider:       *llmProvider,
		LLMModel:          *llmModel,
		PersonalityPrompt: *personality,
		BusinessKnowledge: *businessKnowledge,
		EnvironmentVars:   map[string]string{},
	})

	ctx := context.Background()
	server, err := provider.CreateServer(ctx, provisioning.ServerConfig{
		Name:       fmt.Sprintf("openclaw-%s-%d", *name, time.Now().Unix()),
		ServerType: *serverType,
		Location:   *location,
		SSHKeyName: *sshKey,
		UserData:   cloudInit,
		Labels: map[string]string{
			"app":        "openclaw",
			"deployment": *name,
		},
	})
	if err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	if *waitReady {
		if err := provider.WaitForServer(ctx, server.ID, "running", *timeout); err != nil {
			return fmt.Errorf("wait for running: %w", err)
		}
		refreshed, getErr := provider.GetServer(ctx, server.ID)
		if getErr == nil {
			server = refreshed
		}
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	createdAt := now
	if prev, ok := state.Deployments[*name]; ok && !prev.CreatedAt.IsZero() {
		createdAt = prev.CreatedAt
	}
	state.Deployments[*name] = deployment{
		Name:       *name,
		Provider:   "hetzner",
		ServerID:   server.ID,
		PublicIP:   server.PublicIP,
		Status:     server.Status,
		ServerType: *serverType,
		Location:   *location,
		Version:    "openclaw-latest",
		CreatedAt:  createdAt,
		UpdatedAt:  now,
	}
	if err := saveState(state); err != nil {
		return err
	}

	fmt.Printf("✅ Deployment created: %s\n", *name)
	fmt.Printf("Server ID: %s\n", server.ID)
	fmt.Printf("Public IP: %s\n", server.PublicIP)
	fmt.Printf("Status: %s\n", server.Status)
	return nil
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	name := fs.String("name", "", "Deployment name")
	all := fs.Bool("all", false, "Show all deployments")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	if len(state.Deployments) == 0 {
		fmt.Println("No deployments found. Run: clawhost init && clawhost deploy")
		return nil
	}

	if *all {
		names := make([]string, 0, len(state.Deployments))
		for deploymentName := range state.Deployments {
			names = append(names, deploymentName)
		}
		sort.Strings(names)
		for _, deploymentName := range names {
			d := state.Deployments[deploymentName]
			printStatusLine(d)
		}
		return nil
	}

	target := *name
	if target == "" {
		target = "default"
	}
	d, ok := state.Deployments[target]
	if !ok {
		return fmt.Errorf("deployment %q not found", target)
	}

	apiToken := strings.TrimSpace(os.Getenv("HETZNER_API_TOKEN"))
	if apiToken != "" && d.ServerID != "" {
		provider := provisioning.NewHetznerProvider(apiToken)
		server, getErr := provider.GetServer(context.Background(), d.ServerID)
		if getErr == nil {
			d.Status = server.Status
			d.PublicIP = server.PublicIP
			d.UpdatedAt = time.Now().UTC()
			state.Deployments[target] = d
			_ = saveState(state)
		}
	}

	statusJSON, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(statusJSON))
	return nil
}

func runLogs(args []string) error {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	apiURL := fs.String("api-url", "http://localhost:8080", "ClawHost Core API URL")
	limit := fs.Int("limit", 100, "Log lines limit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	d, ok := state.Deployments[*name]
	if !ok {
		return fmt.Errorf("deployment %q not found", *name)
	}
	if d.ServerID == "" {
		return fmt.Errorf("deployment %q has no server_id yet", *name)
	}

	url := fmt.Sprintf("%s/api/v1/instances/%s/logs?limit=%d", strings.TrimRight(*apiURL, "/"), d.ServerID, *limit)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("fetch logs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("logs request failed (%d): %s", resp.StatusCode, string(body))
	}

	fmt.Println(string(body))
	return nil
}

func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	version := fs.String("version", "latest", "Target OpenClaw version")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	d, ok := state.Deployments[*name]
	if !ok {
		return fmt.Errorf("deployment %q not found", *name)
	}

	d.Version = *version
	d.UpdatedAt = time.Now().UTC()
	state.Deployments[*name] = d
	if err := saveState(state); err != nil {
		return err
	}

	fmt.Printf("✅ Marked deployment %q for OpenClaw version %q\n", *name, *version)
	fmt.Println("Note: This updates desired version metadata; rollout automation can apply it.")
	return nil
}

func runBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	output := fs.String("output", "", "Backup output file path")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	statePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	if *output == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return homeErr
		}
		backupDir := filepath.Join(home, stateDirName, "backups")
		if mkErr := os.MkdirAll(backupDir, 0o755); mkErr != nil {
			return mkErr
		}
		*output = filepath.Join(backupDir, fmt.Sprintf("deployments-%s.json", time.Now().UTC().Format("20060102-150405")))
	}

	content, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("read state file: %w", err)
	}
	if err := os.WriteFile(*output, content, 0o644); err != nil {
		return fmt.Errorf("write backup file: %w", err)
	}

	fmt.Printf("✅ Backup created: %s\n", *output)
	return nil
}

func runRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	input := fs.String("input", "", "Backup input file path")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*input) == "" {
		return errors.New("--input is required")
	}

	content, err := os.ReadFile(*input)
	if err != nil {
		return fmt.Errorf("read backup file: %w", err)
	}

	var state deploymentState
	if err := json.Unmarshal(content, &state); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}
	if state.Deployments == nil {
		state.Deployments = map[string]deployment{}
	}

	if err := saveState(state); err != nil {
		return err
	}
	fmt.Printf("✅ Restore completed from %s\n", *input)
	return nil
}

func runDestroy(args []string) error {
	fs := flag.NewFlagSet("destroy", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	force := fs.Bool("force", false, "Skip confirmation prompt")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	d, ok := state.Deployments[*name]
	if !ok {
		return fmt.Errorf("deployment %q not found", *name)
	}
	if d.ServerID == "" {
		return fmt.Errorf("deployment %q has no server_id", *name)
	}

	if !*force {
		fmt.Printf("Destroy deployment %q (server %s)? [y/N]: ", *name, d.ServerID)
		reader := bufio.NewReader(os.Stdin)
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			return readErr
		}
		answer := strings.ToLower(strings.TrimSpace(line))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	apiToken := strings.TrimSpace(os.Getenv("HETZNER_API_TOKEN"))
	if apiToken == "" {
		return errors.New("HETZNER_API_TOKEN is required")
	}

	provider := provisioning.NewHetznerProvider(apiToken)
	if err := provider.DeleteServer(context.Background(), d.ServerID); err != nil {
		return fmt.Errorf("destroy server: %w", err)
	}

	delete(state.Deployments, *name)
	if err := saveState(state); err != nil {
		return err
	}

	fmt.Printf("✅ Deployment %q destroyed\n", *name)
	return nil
}

func printStatusLine(d deployment) {
	fmt.Printf("- %s: status=%s server_id=%s ip=%s updated=%s\n", d.Name, d.Status, d.ServerID, d.PublicIP, d.UpdatedAt.Format(time.RFC3339))
}

func loadState() (deploymentState, error) {
	path, err := getStateFilePath()
	if err != nil {
		return deploymentState{}, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return deploymentState{}, err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		state := deploymentState{Deployments: map[string]deployment{}}
		if saveErr := saveState(state); saveErr != nil {
			return deploymentState{}, saveErr
		}
		return state, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return deploymentState{}, err
	}

	var state deploymentState
	if err := json.Unmarshal(content, &state); err != nil {
		return deploymentState{}, fmt.Errorf("parse state file: %w", err)
	}
	if state.Deployments == nil {
		state.Deployments = map[string]deployment{}
	}

	return state, nil
}

func saveState(state deploymentState) error {
	path, err := getStateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o644)
}

func getStateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, stateDirName, stateFileName), nil
}
