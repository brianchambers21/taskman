# Task Management System - Root Makefile

# Variables
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Directories
API_DIR := taskman-api
MCP_DIR := taskman-mcp
TEST_DIR := test

.PHONY: all build build-api build-mcp clean test test-all test-e2e deploy help setup-dev

all: clean deps test build ## Build and test everything

help: ## Show this help message
	@echo 'Task Management System - Root Makefile'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: build-api build-mcp ## Build all components

build-api: ## Build API server
	@echo "Building API server..."
	@cd $(API_DIR) && make build
	@echo "✓ API server build completed"

build-mcp: ## Build MCP server
	@echo "Building MCP server..."
	@cd $(MCP_DIR) && make build
	@echo "✓ MCP server build completed"

build-dev: ## Build for development
	@echo "Building for development..."
	@cd $(API_DIR) && make build-dev
	@cd $(MCP_DIR) && make build-dev
	@echo "✓ Development builds completed"

build-docker: ## Build Docker images for all components
	@echo "Building Docker images..."
	@cd $(API_DIR) && make build-docker
	@cd $(MCP_DIR) && make build-docker
	@echo "✓ Docker builds completed"

# Dependencies
deps: ## Download dependencies for all modules
	@echo "Downloading dependencies..."
	@cd $(API_DIR) && make deps
	@cd $(MCP_DIR) && make deps
	@go mod download  # Root module for E2E tests
	@echo "✓ All dependencies updated"

# Testing
test: test-api test-mcp ## Run all unit and integration tests

test-all: test test-e2e ## Run all tests including end-to-end

test-api: ## Run API tests
	@echo "Running API tests..."
	@cd $(API_DIR) && make test
	@echo "✓ API tests completed"

test-mcp: ## Run MCP tests  
	@echo "Running MCP tests..."
	@cd $(MCP_DIR) && make test
	@echo "✓ MCP tests completed"

test-e2e: ## Run end-to-end tests
	@echo "Running end-to-end tests..."
	@echo "Note: This requires API server to be running"
	@go test -v ./$(TEST_DIR)/...
	@echo "✓ End-to-end tests completed"

test-coverage: ## Generate coverage reports for all components
	@echo "Generating coverage reports..."
	@cd $(API_DIR) && make test-coverage
	@cd $(MCP_DIR) && make test-coverage
	@echo "✓ Coverage reports generated"

# Code quality
lint: ## Run linters for all components
	@echo "Running linters..."
	@cd $(API_DIR) && make lint
	@cd $(MCP_DIR) && make lint
	@echo "✓ Linting completed"

fmt: ## Format code for all components
	@echo "Formatting code..."
	@cd $(API_DIR) && make fmt
	@cd $(MCP_DIR) && make fmt
	@go fmt ./$(TEST_DIR)/...
	@echo "✓ Code formatting completed"

vet: ## Run go vet for all components
	@echo "Running go vet..."
	@cd $(API_DIR) && make vet
	@cd $(MCP_DIR) && make vet
	@go vet ./$(TEST_DIR)/...
	@echo "✓ Go vet completed"

# Database operations
setup-db: ## Setup database with Docker Compose
	@echo "Setting up database..."
	@docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 5
	@cd $(API_DIR) && make migrate-up
	@echo "✓ Database setup completed"

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@cd $(API_DIR) && make migrate-up
	@echo "✓ Migrations completed"

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@cd $(API_DIR) && make migrate-down
	@echo "✓ Rollback completed"

# Development environment
setup-dev: install-tools setup-env setup-db ## Setup complete development environment
	@echo "✓ Development environment setup completed"

setup-env: ## Setup environment files
	@echo "Setting up environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env from .env.example"; fi
	@echo "✓ Environment setup completed"

install-tools: ## Install development tools for all components
	@echo "Installing development tools..."
	@cd $(API_DIR) && make install-tools
	@cd $(MCP_DIR) && make install-tools
	@echo "✓ Development tools installed"

# Services management
start-db: ## Start database only
	@echo "Starting database..."
	@docker-compose up -d postgres
	@echo "✓ Database started"

start-api: build-dev start-db ## Start API server
	@echo "Starting API server..."
	@cd $(API_DIR) && make run &
	@echo "✓ API server started"

start-mcp: build-dev ## Start MCP server
	@echo "Starting MCP server..."
	@cd $(MCP_DIR) && make run &
	@echo "✓ MCP server started"

start-all: start-db ## Start all services
	@echo "Starting all services..."
	@./scripts/deploy-system.sh start
	@echo "✓ All services started"

stop-all: ## Stop all services
	@echo "Stopping all services..."
	@./scripts/deploy-system.sh stop
	@echo "✓ All services stopped"

# Deployment
deploy-staging: build test ## Deploy to staging environment
	@echo "Deploying to staging..."
	@./scripts/deploy-system.sh deploy
	@echo "✓ Staging deployment completed"

deploy-prod: build test-all ## Deploy to production environment
	@echo "Deploying to production..."
	@echo "⚠️  Production deployment requires proper credentials and approval!"
	@read -p "Continue with production deployment? [y/N] " confirm && [ "$$confirm" = "y" ]
	@./scripts/deploy-system.sh deploy
	@echo "✓ Production deployment completed"

# Monitoring and observability
setup-monitoring: ## Setup Google Cloud monitoring for all components
	@echo "Setting up monitoring..."
	@cd $(API_DIR) && make setup-monitoring
	@cd $(MCP_DIR) && make setup-monitoring
	@echo "✓ Monitoring setup completed"

# Health checks
health: ## Check health of all services
	@echo "Checking system health..."
	@./scripts/deploy-system.sh health
	@echo "✓ Health check completed"

health-api: ## Check API server health
	@echo "Checking API health..."
	@cd $(API_DIR) && make health
	@echo "✓ API health check completed"

health-mcp: ## Check MCP server health
	@echo "Checking MCP health..."
	@cd $(MCP_DIR) && make health
	@echo "✓ MCP health check completed"

# Utilities
clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	@cd $(API_DIR) && make clean
	@cd $(MCP_DIR) && make clean
	@rm -f coverage.out coverage.html
	@echo "✓ Clean completed"

status: ## Show status of all services
	@echo "System status:"
	@./scripts/deploy-system.sh status

logs-api: ## Show API server logs
	@echo "API server logs:"
	@tail -f $(API_DIR)/logs/api.log

logs-mcp: ## Show MCP server logs
	@echo "MCP server logs:"
	@tail -f $(MCP_DIR)/logs/mcp.log

logs-db: ## Show database logs
	@echo "Database logs:"
	@docker-compose logs -f postgres

# Security
security-scan: ## Run security scans for all components
	@echo "Running security scans..."
	@cd $(API_DIR) && make security-scan
	@cd $(MCP_DIR) && make security-scan
	@echo "✓ Security scans completed"

# Documentation
docs: ## Generate documentation for all components
	@echo "Generating documentation..."
	@cd $(API_DIR) && make docs
	@cd $(MCP_DIR) && make docs-mcp
	@echo "✓ Documentation generated"

# Version information
version: ## Show version information
	@echo "Task Management System"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo ""
	@echo "Component versions:"
	@cd $(API_DIR) && make version
	@cd $(MCP_DIR) && make version

# Quick development workflow
dev: clean setup-dev build-dev start-all ## Complete development setup and start
	@echo "✓ Development environment ready!"
	@echo ""
	@echo "Services running:"
	@echo "  - Database: PostgreSQL on port 5433"
	@echo "  - API: http://localhost:8080"
	@echo "  - MCP: stdio/http mode"
	@echo ""
	@echo "Use 'make status' to check service status"
	@echo "Use 'make test-e2e' to run end-to-end tests"