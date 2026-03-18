# ClawHost Makefile

.PHONY: build run build-core run-core test clean docker-build docker-up docker-down dev lint lint-install format fmt-check

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOPATH_BIN=$(shell $(GOCMD) env GOPATH)/bin
GOLANGCI_LINT=$(GOPATH_BIN)/golangci-lint

# Binary names
BINARY_NAME=clawhost
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cli

build-core:
	cd core && $(GOBUILD) -o ../$(BINARY_NAME)-core -v ./cmd

run-core:
	cd core && $(GOBUILD) -o ../$(BINARY_NAME)-core -v ./cmd
	./$(BINARY_NAME)-core

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cli
	./$(BINARY_NAME)

# Run in development mode with hot reload (requires air)
dev:
	air

# Test the application
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Download dependencies
deps:
	$(GOGET) ./...
	$(GOMOD) tidy

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cli

# Docker commands
docker-build:
	docker build -t clawhost .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

# Setup development environment
setup-dev:
	cp .env.example .env
	echo "Please edit .env file with your configuration"
	make deps
	make docker-up
	sleep 10

# Production deployment helpers
deploy-staging:
	make build-linux
	# Add your staging deployment commands here

deploy-prod:
	make build-linux
	# Add your production deployment commands here

# Lint and format
lint: lint-install
	cd core && $(GOLANGCI_LINT) run ./...

lint-install:
	@test -x "$(GOLANGCI_LINT)" || $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

format:
	cd core && $(GOCMD) fmt ./...

fmt-check:
	cd core && test -z "$$($(GOCMD)fmt ./... )"

# Security scan
security:
	gosec ./...

# Generate API documentation
docs:
	swag init -g core/cmd/main.go

# Help
help:
	@echo "Available commands:"
	@echo "  build      - Build the application"
	@echo "  build-core - Build core API server binary"
	@echo "  run-core   - Build and run core API server"
	@echo "  run        - Build and run the application"
	@echo "  dev        - Run with hot reload (requires air)"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build files"
	@echo "  deps       - Download dependencies"
	@echo "  lint       - Run golangci-lint on core module"
	@echo "  lint-install - Install golangci-lint"
	@echo "  format     - Run go fmt on core module"
	@echo "  fmt-check  - Check if formatting changes are needed"
	@echo "  docker-up  - Start with Docker Compose"
	@echo "  docker-down- Stop Docker Compose"
	@echo "  setup-dev  - Setup development environment"