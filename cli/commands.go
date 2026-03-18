package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	stateFileName = "deployments.json"
	stateDirName  = ".clawhost"
)

type deploymentState struct {
	Deployments map[string]deployment `json:"deployments"`
}

type deployment struct {
	Name             string    `json:"name"`
	Provider         string    `json:"provider"`
	ServerID         string    `json:"server_id"`
	PublicIP         string    `json:"public_ip"`
	Status           string    `json:"status"`
	ServerType       string    `json:"server_type"`
	Location         string    `json:"location"`
	Version          string    `json:"version"`
	APIURL           string    `json:"api_url,omitempty"`
	OpenClawURL      string    `json:"openclaw_url,omitempty"`
	OpenClawStartCmd string    `json:"openclaw_start_cmd,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type cliConfig struct {
	Name             string
	Provider         string
	ServerType       string
	Location         string
	LLMProvider      string
	LLMModel         string
	APIURL           string
	OpenClawURL      string
	OpenClawStartCmd string
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
	case "upgrade":
		return runUpgrade(args[1:])
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
	fmt.Println("  init      Interactive setup wizard")
	fmt.Println("  deploy    Deploy OpenClaw locally")
	fmt.Println("  status    Check deployment status")
	fmt.Println("  logs      View logs")
	fmt.Println("  update    Update OpenClaw version")
	fmt.Println("  upgrade   Upgrade to latest OpenClaw version")
	fmt.Println("  backup    Create backup")
	fmt.Println("  restore   Restore from backup")
	fmt.Println("  destroy   Remove everything (with confirmation)")
	fmt.Println("")
	fmt.Println("If no command is provided, clawhost starts the Core API server.")
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment profile name")
	provider := fs.String("provider", "local", "Deployment provider")
	serverType := fs.String("server-type", "local", "Server type")
	location := fs.String("location", "localhost", "Server location")
	llmProvider := fs.String("llm-provider", "openai", "LLM provider")
	llmModel := fs.String("llm-model", "gpt-4o-mini", "LLM model")
	apiURL := fs.String("api-url", "http://localhost:8080", "Local ClawHost API URL")
	openclawURL := fs.String("openclaw-url", "http://localhost:3000", "Local OpenClaw URL")
	openclawStartCmd := fs.String("openclaw-start-cmd", "", "Command to start local OpenClaw agent (optional)")
	interactive := fs.Bool("interactive", true, "Run interactive setup wizard")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing config file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *interactive {
		fmt.Println("🧙 ClawHost Init Wizard (Local / Self-Managed)")
		reader := bufio.NewReader(os.Stdin)

		*name = prompt(reader, "Deployment name", *name)
		*provider = prompt(reader, "Provider", *provider)
		*serverType = prompt(reader, "Server type", *serverType)
		*location = prompt(reader, "Location", *location)
		*llmProvider = prompt(reader, "LLM provider", *llmProvider)
		*llmModel = prompt(reader, "LLM model", *llmModel)
		*apiURL = prompt(reader, "Local API URL", *apiURL)
		*openclawURL = prompt(reader, "OpenClaw URL", *openclawURL)
		*openclawStartCmd = prompt(reader, "OpenClaw start command (optional)", *openclawStartCmd)
	}

	configPath := ".clawhost.yaml"
	if _, err := os.Stat(configPath); err == nil && !*overwrite {
		return fmt.Errorf("%s already exists (use --overwrite)", configPath)
	}

	config := fmt.Sprintf("name: %s\nprovider: %s\nserver_type: %s\nlocation: %s\nllm_provider: %s\nllm_model: %s\napi_url: %s\nopenclaw_url: %s\nopenclaw_start_cmd: %s\n", *name, *provider, *serverType, *location, *llmProvider, *llmModel, *apiURL, *openclawURL, *openclawStartCmd)
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
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
		Name:             *name,
		Provider:         *provider,
		ServerID:         "local-" + *name,
		PublicIP:         "127.0.0.1",
		ServerType:       *serverType,
		Location:         *location,
		Status:           "initialized",
		Version:          "openclaw-latest",
		APIURL:           *apiURL,
		OpenClawURL:      *openclawURL,
		OpenClawStartCmd: *openclawStartCmd,
		CreatedAt:        createdAt,
		UpdatedAt:        now,
	}
	if err := saveState(state); err != nil {
		return err
	}

	fmt.Printf("✅ Initialized deployment profile %q\n", *name)
	fmt.Printf("Config written to %s\n", configPath)
	return nil
}

func runDeploy(args []string) error {
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	apiURL := fs.String("api-url", "", "Local ClawHost Core API URL")
	openclawURL := fs.String("openclaw-url", "", "Local OpenClaw URL")
	startCmd := fs.String("start-cmd", "", "Command to start OpenClaw before health check")
	timeout := fs.Duration("timeout", 60*time.Second, "Max wait for OpenClaw /health")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	cfg := loadCLIConfig()
	if strings.TrimSpace(*apiURL) == "" {
		*apiURL = cfg.APIURL
	}
	if strings.TrimSpace(*apiURL) == "" {
		*apiURL = "http://localhost:8080"
	}
	if strings.TrimSpace(*openclawURL) == "" {
		*openclawURL = cfg.OpenClawURL
	}
	if strings.TrimSpace(*openclawURL) == "" {
		*openclawURL = "http://localhost:3000"
	}
	if strings.TrimSpace(*startCmd) == "" {
		*startCmd = cfg.OpenClawStartCmd
	}

	coreHealthy := false
	if resp, err := http.Get(strings.TrimRight(*apiURL, "/") + "/health"); err == nil {
		_ = resp.Body.Close()
		if resp.StatusCode < http.StatusBadRequest {
			coreHealthy = true
		}
	}
	if !coreHealthy {
		return fmt.Errorf("core API not reachable at %s (start core first with ./clawhost or make run-core)", *apiURL)
	}

	if strings.TrimSpace(*startCmd) != "" {
		fmt.Printf("▶ Starting OpenClaw with command: %s\n", *startCmd)
		cmd := exec.Command("sh", "-lc", *startCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("openclaw start command failed: %w", err)
		}
	}

	openclawHealthy := waitForHealthyEndpoint(strings.TrimRight(*openclawURL, "/")+"/health", *timeout)
	if !openclawHealthy {
		return fmt.Errorf("OpenClaw not reachable at %s/health after %s", strings.TrimRight(*openclawURL, "/"), timeout.String())
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
	status := "running"
	state.Deployments[*name] = deployment{
		Name:             *name,
		Provider:         "local",
		ServerID:         "local-" + *name,
		PublicIP:         "127.0.0.1",
		Status:           status,
		ServerType:       "local",
		Location:         "localhost",
		Version:          "openclaw-latest",
		APIURL:           *apiURL,
		OpenClawURL:      *openclawURL,
		OpenClawStartCmd: *startCmd,
		CreatedAt:        createdAt,
		UpdatedAt:        now,
	}
	if err := saveState(state); err != nil {
		return err
	}

	fmt.Printf("✅ Local deployment configured: %s\n", *name)
	fmt.Printf("Core API: %s\n", *apiURL)
	fmt.Printf("OpenClaw URL: %s\n", *openclawURL)
	fmt.Printf("Status: %s\n", status)
	fmt.Println("✅ OpenClaw handshake successful")
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
			d = refreshLocalDeploymentStatus(d)
			state.Deployments[deploymentName] = d
			printStatusLine(d)
		}
		_ = saveState(state)
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

	d = refreshLocalDeploymentStatus(d)
	state.Deployments[target] = d
	_ = saveState(state)

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

func runUpgrade(args []string) error {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	name := fs.String("name", "default", "Deployment name")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	return runUpdate([]string{"--name", *name, "--version", "latest"})
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
	all := fs.Bool("all", true, "Remove all local ClawHost config and state")
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
		if *all {
			fmt.Printf("Remove deployment %q and delete local state/config files? [y/N]: ", *name)
		} else {
			fmt.Printf("Destroy deployment %q? [y/N]: ", *name)
		}
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

	delete(state.Deployments, *name)
	if err := saveState(state); err != nil {
		return err
	}

	if *all {
		if err := os.Remove(".clawhost.yaml"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		stateDir := filepath.Join(home, stateDirName)
		if err := os.RemoveAll(stateDir); err != nil {
			return err
		}
	}

	fmt.Printf("✅ Deployment %q destroyed\n", *name)
	if *all {
		fmt.Println("✅ Local ClawHost state/config removed")
	}
	return nil
}

func prompt(reader *bufio.Reader, label, defaultValue string) string {
	fmt.Printf("%s [%s]: ", label, defaultValue)
	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue
	}
	return value
}

func printStatusLine(d deployment) {
	fmt.Printf("- %s: status=%s server_id=%s ip=%s updated=%s\n", d.Name, d.Status, d.ServerID, d.PublicIP, d.UpdatedAt.Format(time.RFC3339))
}

func refreshLocalDeploymentStatus(d deployment) deployment {
	if !strings.HasPrefix(d.ServerID, "local-") {
		return d
	}

	cfg := loadCLIConfig()
	apiURL := strings.TrimSpace(d.APIURL)
	if apiURL == "" {
		apiURL = cfg.APIURL
	}
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	openclawURL := strings.TrimSpace(d.OpenClawURL)
	if openclawURL == "" {
		openclawURL = cfg.OpenClawURL
	}
	if openclawURL == "" {
		openclawURL = "http://localhost:3000"
	}

	coreHealthy := false
	if resp, reqErr := http.Get(strings.TrimRight(apiURL, "/") + "/health"); reqErr == nil {
		_ = resp.Body.Close()
		if resp.StatusCode < http.StatusBadRequest {
			coreHealthy = true
		}
	}

	openclawHealthy := false
	if resp, reqErr := http.Get(strings.TrimRight(openclawURL, "/") + "/health"); reqErr == nil {
		_ = resp.Body.Close()
		if resp.StatusCode < http.StatusBadRequest {
			openclawHealthy = true
		}
	}

	switch {
	case coreHealthy && openclawHealthy:
		d.Status = "running"
	case coreHealthy && !openclawHealthy:
		d.Status = "core-running-openclaw-down"
	case !coreHealthy && openclawHealthy:
		d.Status = "openclaw-running-core-down"
	default:
		d.Status = "local-configured"
	}

	d.APIURL = apiURL
	d.OpenClawURL = openclawURL

	d.UpdatedAt = time.Now().UTC()
	return d
}

func loadCLIConfig() cliConfig {
	config := cliConfig{
		Name:        "default",
		Provider:    "local",
		ServerType:  "local",
		Location:    "localhost",
		LLMProvider: "openai",
		LLMModel:    "gpt-4o-mini",
		APIURL:      "http://localhost:8080",
		OpenClawURL: "http://localhost:3000",
	}

	content, err := os.ReadFile(".clawhost.yaml")
	if err != nil {
		return config
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, ":") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			config.Name = value
		case "provider":
			config.Provider = value
		case "server_type":
			config.ServerType = value
		case "location":
			config.Location = value
		case "llm_provider":
			config.LLMProvider = value
		case "llm_model":
			config.LLMModel = value
		case "api_url":
			config.APIURL = value
		case "openclaw_url":
			config.OpenClawURL = value
		case "openclaw_start_cmd":
			config.OpenClawStartCmd = value
		}
	}

	return config
}

func waitForHealthyEndpoint(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < http.StatusBadRequest {
				return true
			}
		}
		time.Sleep(2 * time.Second)
	}

	return false
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
