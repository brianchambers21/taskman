# Task Management MCP Server - Makefile

# Variables
BINARY_NAME := taskman-mcp
BUILD_DIR := ./bin
CMD_DIR := ./cmd
GO_FILES := $(shell find . -name "*.go" -type f | grep -v vendor/)
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Build targets
.PHONY: all build clean test test-unit test-integration test-mcp lint fmt vet deps help

all: clean deps test build ## Build everything

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build commands
build: $(BUILD_DIR)/$(BINARY_NAME) ## Build MCP server binary

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building MCP server..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	@echo "✓ MCP server built successfully"

# Development builds (local platform)
build-dev: ## Build for development (local platform)
	@echo "Building for development..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	@echo "✓ Development build completed"

# Docker builds
build-docker: ## Build Docker images
	@echo "Building Docker images..."
	@docker build -t taskman-mcp:$(VERSION) -t taskman-mcp:latest .
	@echo "✓ Docker images built successfully"

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# Testing
test: test-unit test-integration test-mcp ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./internal/...
	@echo "✓ Unit tests completed"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./test/...
	@echo "✓ Integration tests completed"

test-mcp: ## Run MCP-specific tests
	@echo "Running MCP functionality tests..."
	@go test -v -tags=mcp ./internal/tools/... ./internal/resources/... ./internal/prompts/...
	@echo "✓ MCP tests completed"

test-compliance: ## Run MCP protocol compliance tests
	@echo "Running MCP compliance tests..."
	@go test -v ./test/mcp_compliance_test.go
	@echo "✓ MCP compliance tests completed"

test-coverage: test-unit ## Generate test coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Go vet completed"

# MCP specific operations
mcp-stdio: build-dev ## Run MCP server in stdio mode
	@echo "Starting MCP server in stdio mode..."
	@TASKMAN_TRANSPORT_MODE=stdio $(BUILD_DIR)/$(BINARY_NAME)

mcp-http: build-dev ## Run MCP server in HTTP mode
	@echo "Starting MCP server in HTTP mode..."
	@TASKMAN_TRANSPORT_MODE=http $(BUILD_DIR)/$(BINARY_NAME)

mcp-both: build-dev ## Run MCP server in both stdio and HTTP modes
	@echo "Starting MCP server in both modes..."
	@TASKMAN_TRANSPORT_MODE=both $(BUILD_DIR)/$(BINARY_NAME)

# API connectivity tests
test-api-connection: ## Test connection to API server
	@echo "Testing API server connection..."
	@curl -f http://localhost:8080/api/v1/tasks > /dev/null || (echo "❌ API server not accessible" && exit 1)
	@echo "✅ API server connection successful"

# Deployment
deploy-staging: build test ## Deploy to staging
	@echo "Deploying to staging..."
	@./scripts/deploy.sh deploy
	@echo "✓ Staging deployment completed"

deploy-prod: build test ## Deploy to production
	@echo "Deploying to production..."
	@echo "⚠️  Make sure you have proper credentials and approval!"
	@read -p "Continue with production deployment? [y/N] " confirm && [ "$$confirm" = "y" ]
	@./scripts/deploy.sh deploy
	@echo "✓ Production deployment completed"

# Utilities
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Clean completed"

run: build-dev ## Run the MCP server locally
	@echo "Starting MCP server..."
	@$(BUILD_DIR)/$(BINARY_NAME)

run-dev: ## Run in development mode with live reload
	@echo "Starting development server with live reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed, running normally..."; \
		make run; \
	fi

# Health checks
health: ## Check MCP server health
	@echo "Checking MCP server health..."
	@./scripts/deploy.sh health
	@echo "✅ Health check completed"

# Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security scan..."; \
	fi

# Version info
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✓ Development tools installed"

# Monitoring setup
setup-monitoring: ## Setup Google Cloud monitoring
	@echo "Setting up Google Cloud monitoring..."
	@gcloud services enable monitoring.googleapis.com
	@gcloud services enable logging.googleapis.com
	@echo "✓ Google Cloud monitoring services enabled"

# MCP tools testing
test-tools: build-dev ## Test individual MCP tools
	@echo "Testing MCP tools..."
	@echo "Note: This requires a running API server"
	@make test-api-connection
	@go test -v ./internal/tools/integration_test.go
	@echo "✓ MCP tools testing completed"

# Generate MCP schema documentation
docs-mcp: ## Generate MCP schema documentation
	@echo "Generating MCP schema documentation..."
	@mkdir -p docs
	@echo "MCP Tools, Resources, and Prompts available:" > docs/mcp-schema.md
	@echo "This would generate documentation from MCP server capabilities"
	@echo "✓ MCP documentation generated"