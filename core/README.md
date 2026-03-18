# ClawHost Core - Open Source Components

This directory contains the **open source core** of ClawHost - the foundational components for managing OpenClaw AI instances across multiple cloud providers.

## 🔓 **What's Open Source**

The core provides everything needed to **self-host and manage OpenClaw instances**:

- ☁️ **Multi-cloud provisioning** (Hetzner, AWS, DigitalOcean)
- 📈 **Instance monitoring** and health checks
- 🤖 **OpenClaw automation** (installation, configuration)
- 🔌 **Core API** for instance management

## 🏗️ **Architecture**

```
core/
├── provisioning/     # Cloud provider integrations
│   ├── hetzner.go    # Hetzner Cloud provider
│   ├── aws.go        # AWS EC2 provider (coming soon)
│   └── openclaw.go   # OpenClaw installation automation
├── monitoring/       # Metrics and health monitoring
│   ├── metrics.go    # Instance metrics collection
│   └── health.go     # Health check system
├── api/             # Core management API
│   └── server.go     # HTTP API server
└── cmd/             # Command-line tools
    ├── main.go       # Core API server
    └── provision.go  # Provisioning CLI tool
```

## 🚀 **Quick Start**

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Set Environment Variables
```bash
export HETZNER_API_TOKEN="your_hetzner_token"
export OPENAI_API_KEY="your_openai_key"  # Optional
```

### 3. Start Core API Server
```bash
go run core/cmd/main.go
```

### 4. Provision Your First Instance
```bash
# Using the CLI tool
go run core/cmd/provision.go \
  --provider hetzner \
  --server-type cx11 \
  --location nbg1 \
  --llm-provider openai

# Or via API
curl -X POST http://localhost:8080/api/v1/provision \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "hetzner",
    "server_type": "cx11",
    "location": "nbg1",
    "openclaw_config": {
      "llm_provider": "openai",
      "llm_model": "gpt-3.5-turbo"
    }
  }'
```

## 📚 **API Reference**

### Core Endpoints

#### Health Check
```http
GET /health
```

#### Instance Status
```http
GET /api/v1/instances/{id}/status
GET /api/v1/instances/{id}/metrics
GET /api/v1/instances/{id}/health
GET /api/v1/instances/{id}/logs
```

#### Provisioning
```http
GET /api/v1/provision/status/{job_id}
```

## 🌍 **Cloud Providers**

### Hetzner Cloud ✅ **Ready**
```go
import "github.com/yourusername/clawhost/core/provisioning"

provider := provisioning.NewHetznerProvider(apiToken)
server, err := provider.CreateServer(ctx, provisioning.ServerConfig{
    Name:       "my-openclaw-bot",
    ServerType: "cx11",
    Location:   "nbg1",
    UserData:   openclawCloudInit,
})
```

### AWS EC2 🚧 **Coming Soon**
```go
// Will be available in v0.2.0
provider := provisioning.NewAWSProvider(accessKey, secretKey, region)
```

### Custom Provider 🔧 **DIY**
```go
// Implement the CloudProvider interface
type MyProvider struct{}

func (p *MyProvider) CreateServer(ctx context.Context, config ServerConfig) (*ServerInfo, error) {
    // Your cloud provider logic here
    return &ServerInfo{...}, nil
}
```

## 📈 **Monitoring**

### Collect Metrics
```go
import "github.com/yourusername/clawhost/core/monitoring"

collector := monitoring.NewBasicMetricsCollector()
metrics, err := collector.CollectMetrics("instance-123")

fmt.Printf("CPU Usage: %.1f%%\n", metrics.CPUUsage)
fmt.Printf("Memory Usage: %.1f%%\n", metrics.MemoryUsage)
```

### Health Checks
```go
healthChecker := monitoring.NewBasicHealthChecker()
status, err := healthChecker.CheckHealth("instance-123")

if status.Healthy {
    fmt.Println("All systems operational")
} else {
    fmt.Printf("Issues detected: %s\n", status.Message)
}
```

## 🤖 **OpenClaw Integration**

### Automatic Installation
```go
import "github.com/yourusername/clawhost/core/provisioning"

config := provisioning.OpenClawConfig{
    LLMProvider:       "openai",
    LLMModel:          "gpt-4",
    PersonalityPrompt: "You are a helpful customer service bot.",
    BusinessKnowledge: "We sell eco-friendly products...",
}

cloudInit := provisioning.GenerateCloudInitScript(config)
// Use this in your server provisioning
```

### Supported LLM Providers
- 🤖 **OpenAI**: GPT-3.5, GPT-4, GPT-4 Turbo
- 🌯 **Anthropic**: Claude 3 (Haiku, Sonnet, Opus)
- 🌍 **Google**: Gemini Pro, Gemini Ultra
- 🔧 **Custom**: Bring your own LLM API

## 🛠️ **Development**

### Running Tests
```bash
# Test core components
go test ./core/...

# Test with coverage
go test -cover ./core/...

# Test specific component
go test ./core/provisioning -v
```

### Adding a New Cloud Provider

1. **Implement the interface**:
```go
// core/provisioning/myprovider.go
type MyCloudProvider struct{
    apiKey string
}

func (p *MyCloudProvider) CreateServer(ctx context.Context, config ServerConfig) (*ServerInfo, error) {
    // Implementation here
}
```

2. **Add tests**:
```go
// core/provisioning/myprovider_test.go
func TestMyProviderCreateServer(t *testing.T) {
    // Test implementation
}
```

3. **Update documentation**

## 📎 **Examples**

Check the [examples directory](../examples/) for:
- **[Basic Deployment](../examples/basic-deployment/)** - Get started quickly
- **[Multi-Instance Setup](../examples/multi-instance/)** - Manage multiple bots
- **[Custom Monitoring](../examples/monitoring/)** - Advanced metrics
- **[Provider Integration](../examples/custom-provider/)** - Add new clouds

## 🤝 **Contributing**

We welcome contributions to the open source core!

### Priority Areas
- 🌍 **New cloud providers** (AWS, GCP, Azure)
- 📈 **Enhanced monitoring** (Prometheus integration)
- 🔧 **Developer tools** (CLI improvements)
- 📝 **Documentation** (guides and examples)

### Development Process
1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `go test ./core/...`
5. Submit a pull request

## 📝 **License**

The core components are **MIT licensed** - free for commercial use!

## 🆘 **Support**

- 💬 **[Community Forum](../community/forum.md)** - Ask questions
- 🐛 **[GitHub Issues](https://github.com/yourusername/clawhost/issues)** - Report bugs
- 📚 **[Documentation](../docs/)** - Detailed guides

---

**Ready to build something amazing? Start with the core!** 🚀