.PHONY: all build clean test lint proto deps docker help

# Variables
BINARY_NAME=device-bridge
BINARY_CLI=device-bridge-cli
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"
PROTO_DIR=proto
OUT_DIR=bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

all: deps proto build

help: ## Show this help message
	@echo "$(COLOR_BOLD)Device Bridge v2 - Makefile$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

deps: ## Install dependencies
	@echo "$(COLOR_BLUE)Installing dependencies...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies installed$(COLOR_RESET)"

proto: ## Generate protobuf code
	@echo "$(COLOR_BLUE)Generating protobuf code...$(COLOR_RESET)"
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$(PROTO_DIR)/devicebridge/v1/*.proto; \
		echo "$(COLOR_GREEN)✓ Protobuf code generated$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ protoc not found, skipping proto generation$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install protoc: https://grpc.io/docs/protoc-installation/$(COLOR_RESET)"; \
	fi

build: ## Build the main daemon
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(OUT_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(OUT_DIR)/$(BINARY_NAME) \
		./cmd/bridge
	@echo "$(COLOR_GREEN)✓ Built: $(OUT_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

build-cli: ## Build the CLI tool
	@echo "$(COLOR_BLUE)Building $(BINARY_CLI)...$(COLOR_RESET)"
	@mkdir -p $(OUT_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) \
		-o $(OUT_DIR)/$(BINARY_CLI) \
		./cmd/bridge-cli
	@echo "$(COLOR_GREEN)✓ Built: $(OUT_DIR)/$(BINARY_CLI)$(COLOR_RESET)"

build-all: build build-cli ## Build all binaries

run: build ## Build and run the daemon
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	./$(OUT_DIR)/$(BINARY_NAME) --config configs/config.example.yaml

test: ## Run tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

test-coverage: test ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Generating coverage report...$(COLOR_RESET)"
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

test-integration: ## Run integration tests
	@echo "$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	$(GO) test -v -tags=integration ./test/integration/...
	@echo "$(COLOR_GREEN)✓ Integration tests complete$(COLOR_RESET)"

lint: ## Run linters
	@echo "$(COLOR_BLUE)Running linters...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not found$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install: https://golangci-lint.run/usage/install/$(COLOR_RESET)"; \
		$(GO) fmt ./...; \
		$(GO) vet ./...; \
	fi

fmt: ## Format code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	rm -rf $(OUT_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	docker build -t device-bridge:$(VERSION) .
	docker tag device-bridge:$(VERSION) device-bridge:latest
	@echo "$(COLOR_GREEN)✓ Docker image built$(COLOR_RESET)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	docker run --rm -it \
		--name device-bridge \
		-v $(PWD)/configs:/config \
		-p 50051:50051 \
		-p 8080:8080 \
		-p 9090:9090 \
		device-bridge:latest

install: build ## Install binary to /usr/local/bin
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME)...$(COLOR_RESET)"
	@sudo cp $(OUT_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(COLOR_GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(COLOR_RESET)"

uninstall: ## Uninstall binary from /usr/local/bin
	@echo "$(COLOR_BLUE)Uninstalling $(BINARY_NAME)...$(COLOR_RESET)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Uninstalled$(COLOR_RESET)"

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell $(GO) version)"

.DEFAULT_GOAL := help
