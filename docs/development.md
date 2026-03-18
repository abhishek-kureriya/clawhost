# Development Guide

Use this guide for local development of ClawHost core and hosting-service components.

## Prerequisites

- Go 1.23+
- Docker and Docker Compose
- PostgreSQL (or containerized DB)
- Hetzner API token for provisioning tests

## Local Setup

1. Clone repository
2. Copy environment file (`.env.example` -> `.env`)
3. Install dependencies with `go mod tidy`
4. Start dependencies with `docker-compose up -d`

## Run Services

- Core API: `go run core/cmd/main.go`
- Commercial service: `go run hosting-service/cmd/main.go`

## Run Tests

- All tests: `go test ./...`
- Core-only tests: `go test ./core/...`

## Development Workflow

1. Create a feature branch.
2. Make minimal focused changes.
3. Run tests for affected areas.
4. Update docs if behavior changed.
5. Open a PR using the repository template.

## Reference

- [Contributing](contributing.md)
- [Installation](installation.md)
- [Launch Guide](launch-guide.md)
