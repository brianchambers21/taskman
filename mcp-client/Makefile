# MCP Client Makefile

# Build variables
BINARY_NAME=mcp-client
BUILD_DIR=bin
MAIN_PKG=./cmd/client

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean test deps run help

all: clean deps build ## Build everything

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

build-dev: ## Build for development with debug info
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

test: ## Run unit tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-integration: ## Run integration tests (requires MCP server)
	@echo "Running integration tests..."
	$(GOTEST) -v ./test/integration/...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

run: build ## Build and run the client
	./$(BUILD_DIR)/$(BINARY_NAME)

run-dev: ## Run without building (development mode)
	$(GOCMD) run $(MAIN_PKG)

run-interactive: build ## Run in interactive mode
	./$(BUILD_DIR)/$(BINARY_NAME) -interactive

# Example commands
example-list-tools: build ## Example: List tools
	./$(BUILD_DIR)/$(BINARY_NAME) list-tools

example-execute-tool: build ## Example: Execute a tool
	./$(BUILD_DIR)/$(BINARY_NAME) execute-tool get_task_overview

example-intent: build ## Example: Process JSON intent
	./$(BUILD_DIR)/$(BINARY_NAME) -intent '{"method": "tools/list"}'

# Development helpers
fmt: ## Format Go code
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

lint: ## Run golint (if installed)
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed, skipping..."; \
	fi

check: fmt vet lint test ## Run all checks

install: build ## Install binary to GOPATH
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PKG)

# Docker helpers (if needed)
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME) .

# Help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)