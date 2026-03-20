# Development Guide

Use this guide for local development of ClawHost core and hosting-service components.

## Core vs CLI (Local Dev)

- **CLI (`clawhost`)**: setup and deployment workflow (`init`, `deploy`, `status`, `upgrade`, `destroy`)
- **Core API (`make run-core`)**: monitoring and dashboard service (`/health`, `/dashboard`, `/api/v1/...`)

In local mode, this is the expected flow:

1. Start Core API
2. Run CLI setup/deploy commands
3. Monitor in Core dashboard

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

## Local Dev Quick Start (Open Source)

```bash
# 1) Start core monitoring API + dashboard
make run-core

# 2) In another terminal, build CLI
make build

# 3) Initialize and deploy local OpenClaw setup
./clawhost init
./clawhost deploy

# 4) Check status
./clawhost status --all
```

Dashboard:

- http://localhost:8080/dashboard

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
