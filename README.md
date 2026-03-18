# ClawHost - Open Source OpenClaw Hosting for Self-Hosted AI

ClawHost is an open source OpenClaw hosting platform for teams that want to self-host AI agents on their own machines or servers. It provides a Go-based core API, provisioning automation, monitoring, and a basic local dashboard.

If you are searching for **OpenClaw self-hosting**, **OpenClaw deployment**, **open source AI hosting**, or **self-managed AI agent platform**, this repository is the OSS foundation.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Community](https://img.shields.io/badge/Community-Welcome-green?style=for-the-badge)](community/)

## 🎆 **Open Source Strategy Overview**

ClawHost follows an **open source core + commercial service** model, similar to GitLab, Supabase, and other successful companies:

### 🔓 **Open Source Core** (`/core/`)
- **Server Provisioning**: Cloud provider integrations (Hetzner, AWS, etc.)
- **Instance Monitoring**: Metrics collection and health checks
- **Core API**: Basic instance management and monitoring
- **OpenClaw Integration**: Automated installation and configuration

### 💼 **Commercial Hosting Service** (`/hosting-service/`)
- **Customer Dashboard**: Professional UI for instance management
- **Billing & Subscriptions**: Stripe integration and payment processing
- **Premium Support**: Ticket system and dedicated support
- **Enterprise Features**: Advanced monitoring, backups, scaling

---

## 🚀 **Quick Start**

📚 For detailed guides and references, see [docs/README.md](docs/README.md).

### Open Source Local Mode (Recommended)

```bash
# Clone repository
git clone https://github.com/abhishek-kureriya/clawhost.git
cd clawhost

# Build and run local OSS core
make build
./clawhost

# Verify API and open dashboard
curl http://localhost:8080/health
# http://localhost:8080/dashboard
```

### Using the Open Source Core

```bash
# Clone the repository
git clone https://github.com/abhishek-kureriya/clawhost.git
cd clawhost

# Install dependencies
go mod tidy

# Run the core API server
go run core/cmd/main.go
```

### Using Our Hosted Service

Skip the setup entirely! Get started with managed OpenClaw hosting:

🌐 Managed service details available on request.

---

## 🏗️ **Architecture**

### Product Flow

```
┌─ Marketing Site ─┐    ┌─ Onboarding App ─┐    ┌─ Customer Dashboard ─┐
│ • Landing page   │    │ • Sign up wizard │    │ • Manage OpenClaw    │
│ • Pricing        │ -> │ • Payment setup  │ -> │ • Analytics          │
│ • Testimonials   │    │ • AI configuration│   │ • Platform settings  │
└──────────────────┘    └──────────────────┘    └──────────────────────┘
        │
        ▼
      ┌─ Provisioning API ─┐
      │ • Create server     │
      │ • Install OpenClaw  │
      │ • Connect platforms │
      └─────────────────────┘
```

### Repository Structure

```
clawhost/
├── cli/                  # `clawhost` command-line tool
├── core/                 🔓 OPEN SOURCE
│   ├── provisioning/     # Cloud provider integrations
│   ├── monitoring/       # Metrics & health checks
│   ├── api/             # Core management API
│   └── cmd/             # Core API server entrypoint
├── hosting-service/   💼 COMMERCIAL
│   ├── marketing-site/  # Landing, pricing, testimonials
│   ├── onboarding/      # Signup, billing setup, AI config
│   ├── dashboard/       # Customer dashboard and analytics
│   ├── billing/         # Stripe integration
│   └── support/         # Ticket system
├── docs/               # Documentation
├── examples/           # Usage examples
└── community/          # Community resources
```

## 📋 **Core Features (Open Source)**

### ☁️ **Multi-Cloud Provisioning**
- **Hetzner Cloud**: Cost-effective European servers
- **DigitalOcean**: Global availability
- **AWS EC2**: Enterprise-grade infrastructure
- **Custom Providers**: Extend with your own cloud APIs

### 📈 **Instance Monitoring**
- Real-time metrics collection (CPU, Memory, Disk, Network)
- Health checks and uptime monitoring
- Log aggregation and analysis
- Alert system for issues

### 🤖 **OpenClaw Integration**
- Automated OpenClaw installation via cloud-init
- LLM provider configuration (OpenAI, Anthropic, Google)
- Messaging platform setup (WhatsApp, Telegram, Discord)
- Custom AI personality configuration

---

## 🎯 **Use Cases**

### 👥 **Self-Hosting Teams**
```bash
# Deploy OpenClaw to your own infrastructure
go run core/cmd/provision.go \
  --provider hetzner \
  --server-type cx21 \
  --location nbg1 \
  --llm-provider openai \
  --api-key $OPENAI_API_KEY
```

### 🏢 **SaaS Providers**
Build your own OpenClaw hosting service:
```bash
# Use the core as your foundation
import "github.com/yourusername/clawhost/core/provisioning"
import "github.com/yourusername/clawhost/core/monitoring"

# Add your own billing, UI, and customer management
```

### 🎨 **Custom Integrations**
Extend ClawHost for specific needs:
```go
// Add custom cloud providers
type MyCloudProvider struct{}
func (p *MyCloudProvider) CreateServer(config ServerConfig) (*ServerInfo, error) {
    // Your implementation
}
```

---

## 🛠️ **Development**

### Prerequisites
- Go 1.23 or higher
- Docker & Docker Compose
- Cloud provider API tokens (only for remote provisioning workflows)

### Local Setup
```bash
# Clone repository
git clone https://github.com/yourusername/clawhost
cd clawhost

# Install dependencies
go mod tidy

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Start development services
docker-compose up -d

# Run core API
go run core/cmd/main.go
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run core tests only
go test ./core/...

# Run with coverage
go test -cover ./...
```

---

## 📄 **Documentation**

- **[Documentation Index](docs/README.md)** - Start here for all docs
- **[Open Source CLI (Local / Self-Managed)](docs/cli.md)** - Simple OSS command flow
- **[Launch Guide](docs/launch-guide.md)** - Step-by-step launch and production rollout
- **[Installation Guide](docs/installation.md)** - Local and production installation
- **[Core API Reference](docs/core-api.md)** - Open source API documentation
- **[Full API Reference](docs/api-reference.md)** - Complete endpoint catalog
- **[Provisioning Guide](docs/provisioning.md)** - Multi-cloud deployment
- **[Monitoring Guide](docs/monitoring.md)** - Metrics and health checks
- **[Development Guide](docs/development.md)** - Local development workflow
- **[CLI Reference](docs/cli.md)** - Command reference for `clawhost`
- **[Contributing Guide](docs/contributing.md)** - How to contribute
- **[Security Policy](docs/security.md)** - Vulnerability reporting and hardening

---

## 🆘 **Commercial vs Open Source**

| Feature | Open Source Core | Commercial Service |
|---------|------------------|--------------------|
| **Server Provisioning** | ✅ Full | ✅ Enhanced |
| **Basic Monitoring** | ✅ Included | ✅ + Advanced |
| **OpenClaw Installation** | ✅ Automated | ✅ + Optimized |
| **Multi-Cloud Support** | ✅ Yes | ✅ + Managed |
| **Web Dashboard** | ❌ No | ✅ Professional UI |
| **Billing Integration** | ❌ No | ✅ Stripe + Invoicing |
| **Customer Support** | 💬 Community | ✅ 24/7 Dedicated |
| **Automatic Backups** | ❌ Manual | ✅ Scheduled |
| **SSL Certificates** | ❌ Manual | ✅ Automated |
| **Scaling & Load Balancing** | ❌ DIY | ✅ Managed |
| **SLA Guarantees** | ❌ None | ✅ 99.9% Uptime |

---

## 🌍 **Community**

### Getting Help
- **[Community Forum](community/forum.md)** - Ask questions and share ideas
- **[Discord Server](https://discord.gg/clawhost)** - Real-time chat
- **[GitHub Issues](https://github.com/yourusername/clawhost/issues)** - Bug reports and feature requests
- **[Documentation](docs/)** - Comprehensive guides

### Contributing
- **[Contributing Guide](docs/contributing.md)** - How to get started
- **[Code of Conduct](docs/code-of-conduct.md)** - Community guidelines
- **[Development Setup](docs/development.md)** - Local development guide

### Ecosystem
- **[Community Plugins](community/plugins/)** - User-contributed extensions
- **[Templates](community/templates/)** - Deployment templates
- **[Integrations](community/integrations/)** - Third-party integrations

---

## 💰 **Managed Service**

Don't want to manage infrastructure? Try our hosted service:

### 🎆 **ClawHost Managed Plans**
- **🌱 Starter**: €49/month - 1 AI instance, basic support
- **💼 Professional**: €99/month - 3 AI instances, priority support
- **🏢 Enterprise**: €199/month - 10 AI instances, 24/7 support

Free trial information available on request.

---

## 📦 **Examples & Tutorials**

- **[Basic Deployment](examples/basic-deployment/)** - Deploy your first OpenClaw instance
- **[Cleanup](examples/cleanup/)** - Safely delete instances and stop server billing

---

## 🔐 **License & Legal**

- **Core**: [MIT License](LICENSE) - Free for commercial use
- **Hosting Service**: Proprietary - Commercial license required
- **Trademarks**: "ClawHost" is a trademark of [Your Company]

---

## 🚀 **Roadmap**

### Open Source Core
- [x] ✅ Hetzner Cloud provisioning
- [x] ✅ Basic monitoring and metrics
- [x] ✅ OpenClaw automated installation
- [ ] 🚧 AWS EC2 provider
- [ ] 🚧 Google Cloud provider
- [ ] 🚧 Kubernetes deployment
- [ ] 🚧 Prometheus integration
- [ ] 🚧 Terraform modules

### Commercial Service
- [x] ✅ Customer dashboard UI
- [x] ✅ Stripe billing integration
- [x] ✅ Support ticket system
- [ ] 🚧 Advanced monitoring dashboard
- [ ] 🚧 Automated backups
- [ ] 🚧 Multi-region deployments
- [ ] 🚧 Enterprise SSO

---

**Built with ❤️ for the OpenClaw community**

[Documentation](docs/) • [Community](community/)