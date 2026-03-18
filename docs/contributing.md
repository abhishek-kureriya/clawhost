# Contributing to ClawHost 🤝

We love contributions! ClawHost is an open source project that thrives on community involvement. Whether you're fixing bugs, adding features, improving documentation, or sharing feedback, every contribution makes a difference.

## 🚀 **Quick Start**

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a feature branch
4. **Make** your changes
5. **Test** your changes
6. **Submit** a pull request

```bash
# Quick setup
git clone https://github.com/yourusername/clawhost.git
cd clawhost
go mod tidy
go test ./...
```

## 📋 **Ways to Contribute**

### 🐛 **Bug Fixes**
- Check [existing issues](https://github.com/yourusername/clawhost/issues)
- Look for [`good first issue`](https://github.com/yourusername/clawhost/labels/good%20first%20issue) labels
- Follow our bug report template

### ✨ **New Features**
- Review [feature requests](https://github.com/yourusername/clawhost/discussions/categories/ideas)
- Propose new ideas in GitHub Discussions
- Start with [`help wanted`](https://github.com/yourusername/clawhost/labels/help%20wanted) issues

### 📚 **Documentation**
- Fix typos and improve clarity
- Add missing API documentation
- Create tutorials and examples
- Translate to other languages

### 🔧 **Infrastructure**
- Improve CI/CD workflows
- Add new cloud providers
- Enhance monitoring and metrics
- Optimize performance

## 🎯 **Priority Areas**

We especially welcome contributions in these areas:

| Area | Description | Difficulty |
|------|-------------|------------|
| **Cloud Providers** | Add AWS, GCP, Azure support | Medium |
| **Monitoring** | Enhanced metrics and alerting | Medium |
| **Security** | Authentication, encryption, auditing | Hard |
| **Performance** | Optimization, caching, scaling | Medium |
| **Documentation** | Guides, API docs, examples | Easy |
| **Testing** | Unit tests, integration tests | Easy-Medium |

## 🛠 **Development Setup**

### Prerequisites

- **Go 1.23+**: [Download here](https://golang.org/dl/)
- **Docker**: For testing and development
- **PostgreSQL**: For database testing
- **Git**: For version control

### Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/clawhost.git
cd clawhost

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your settings

# Run tests
go test ./...

# Start the core API server
go run core/cmd/main.go

# Start the commercial service (if developing commercial features)
go run hosting-service/cmd/main.go
```

### Environment Variables

```bash
# Required for development
export HETZNER_API_TOKEN="your_hetzner_token"
export DATABASE_URL="postgres://user:pass@localhost/clawhost"
export JWT_SECRET="your_jwt_secret"

# Optional
export STRIPE_SECRET_KEY="sk_test_..."
export WEBHOOK_SECRET="whsec_..."
export LOG_LEVEL="debug"
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./core/provisioning

# Run integration tests (requires Docker)
make test-integration

# Benchmark tests
go test -bench=. ./core/monitoring
```

## 📝 **Code Guidelines**

### Go Style

We follow standard Go conventions:

- Use `gofmt` and `goimports` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html) principles
- Write clear, self-documenting code
- Use meaningful variable and function names

### Code Structure

```go
// Good: Clear, documented function
// CreateHetznerServer provisions a new server on Hetzner Cloud
// and returns the server details once it's running.
func CreateHetznerServer(ctx context.Context, config ServerConfig) (*Server, error) {
    if config.Name == "" {
        return nil, fmt.Errorf("server name is required")
    }
    
    // Implementation...
}

// Bad: Unclear, undocumented function
func makeServer(c ServerConfig) *Server {
    // Implementation...
}
```

### Error Handling

```go
// Good: Descriptive errors
if err := validateConfig(config); err != nil {
    return nil, fmt.Errorf("invalid configuration: %w", err)
}

// Good: Context-aware errors
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

if err := client.CreateServer(ctx, config); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return nil, fmt.Errorf("server creation timed out after 30s: %w", err)
    }
    return nil, fmt.Errorf("failed to create server: %w", err)
}
```

### Testing

```go
// Good: Table-driven tests
func TestValidateServerConfig(t *testing.T) {
    tests := []struct {
        name    string
        config  ServerConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: ServerConfig{
                Name:       "test-server",
                ServerType: "cx11",
                Location:   "nbg1",
            },
            wantErr: false,
        },
        {
            name: "missing name",
            config: ServerConfig{
                ServerType: "cx11",
                Location:   "nbg1",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateServerConfig(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateServerConfig() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🔄 **Pull Request Process**

### Before Submitting

1. **Create an issue** (for non-trivial changes)
2. **Fork and branch** from `main`
3. **Write tests** for new functionality
4. **Update documentation** if needed
5. **Run tests** locally
6. **Check code quality** with linters

### PR Checklist

- [ ] Code follows Go conventions and project style
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Documentation updated (if applicable)
- [ ] Commit messages are clear and descriptive
- [ ] PR description explains the changes
- [ ] Related issue referenced (if applicable)

### Commit Messages

Use clear, descriptive commit messages:

```bash
# Good
git commit -m "feat: add AWS provider for multi-cloud support"
git commit -m "fix: handle timeout errors in Hetzner provisioning"
git commit -m "docs: add API examples for instance management"

# Use conventional commit format
# Type: feat, fix, docs, style, refactor, test, chore
```

### PR Description Template

```markdown
## Description
Brief description of what this PR does.

## Changes
- List of changes made
- Another change

## Testing
How you tested these changes.

## Related Issues
Closes #123
References #456

## Screenshots (if applicable)
[Add screenshots for UI changes]
```

## 📊 **Review Process**

### What We Look For

- **Correctness**: Does the code work as intended?
- **Clarity**: Is the code readable and well-documented?
- **Testing**: Are there adequate tests?
- **Performance**: Any performance implications?
- **Security**: Any security concerns?
- **Maintainability**: Is the code easy to maintain?

### Timeline

- **Initial Review**: Within 2-3 business days
- **Follow-up Reviews**: Within 1 business day
- **Merge**: After approval and CI passes

### Who Reviews

- **Core Team**: For significant changes
- **Domain Experts**: For specialized areas
- **Community**: Anyone can provide feedback

## 🏆 **Recognition**

### Contributor Benefits

- **GitHub Badge**: Special contributor badge
- **Hall of Fame**: Listed in our community page
- **Early Access**: Preview new features
- **Swag**: T-shirts, stickers, and goodies
- **Conference Tickets**: Free tickets to events

### Types of Contributions

- 🐛 **Bug Fixes**: Critical for stability
- ✨ **Features**: New functionality
- 📚 **Documentation**: Helps everyone learn
- 🎨 **Design**: UI/UX improvements
- 🔧 **Infrastructure**: DevOps and tooling
- 🌍 **Community**: Helping others, organizing events

## 🤔 **Getting Help**

### Before You Start

1. **Read the docs**: Check existing documentation
2. **Search issues**: Look for existing discussions
3. **Check examples**: Review example code
4. **Ask questions**: Use GitHub Discussions

### Where to Get Help

- **GitHub Discussions**: For general questions
- **Discord**: Real-time chat and support
- **Issues**: For bugs and feature requests
- **Email**: community@clawhost.com for sensitive topics

### Mentorship Program

New to open source? We offer mentorship:

- **Pair Programming**: Work with experienced contributors
- **Code Reviews**: Learn best practices
- **Project Guidance**: Help choosing what to work on
- **Career Advice**: Open source career development

Interested? Email mentor@clawhost.com

## 📜 **Code of Conduct**

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). We are committed to providing a welcoming and inclusive environment for all contributors.

### Key Points

- **Be respectful**: Treat everyone with kindness
- **Be inclusive**: Welcome people of all backgrounds
- **Be constructive**: Provide helpful feedback
- **Be professional**: Maintain a professional demeanor

## 🏗 **Project Structure**

### Repository Layout

```
clawhost/
├── core/                   # Open source core (MIT License)
│   ├── provisioning/       # Cloud provider integrations
│   ├── monitoring/         # Metrics and health checks
│   ├── api/               # Core API server
│   └── cmd/               # CLI tools
├── hosting-service/       # Commercial service (Proprietary)
│   ├── dashboard/         # Customer dashboard
│   ├── billing/          # Stripe billing
│   └── support/          # Support system
├── examples/             # Usage examples
├── docs/                 # Documentation
├── community/            # Community resources
└── .github/              # GitHub workflows
```

### Core Packages

- **`core/provisioning`**: Cloud provider integrations (Hetzner, AWS, etc.)
- **`core/monitoring`**: Metrics collection and health monitoring
- **`core/api`**: Core REST API for instance management
- **`hosting-service/`**: Commercial features (dashboard, billing, support)

## 🚨 **Security**

For security issues, please email security@clawhost.com instead of creating public issues. See our [Security Policy](SECURITY.md) for details.

### Reporting Guidelines

- **Sensitive Issues**: Email security@clawhost.com
- **General Bugs**: Create GitHub issues
- **Feature Requests**: Use GitHub Discussions

## 📞 **Contact**

- **General Questions**: GitHub Discussions
- **Bugs**: GitHub Issues  
- **Security**: security@clawhost.com
- **Community**: community@clawhost.com
- **Business**: hello@clawhost.com

---

**Thank you for contributing to ClawHost!** 🙏

*Together, we're building the future of OpenClaw hosting.*