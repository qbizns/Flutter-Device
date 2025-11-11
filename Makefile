.PHONY: all build test test-unit test-integration test-hardware test-coverage bench lint clean proto docker help

# Variables
BINARY_NAME=bridge
CLI_NAME=bridge-cli
GO=go
GOFLAGS=-v
LDFLAGS=-w -s
BUILD_DIR=bin
COVERAGE_FILE=coverage.out

all: lint test build

help:
	@echo "Device Bridge v2 - Makefile Commands"
	@echo ""
	@echo "  build              - Build binaries"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-hardware      - Run hardware tests (requires physical devices)"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  bench              - Run benchmarks"
	@echo "  lint               - Run linters"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts"
	@echo "  proto              - Generate protobuf code"
	@echo "  docker             - Build Docker image"
	@echo "  run                - Build and run bridge"
	@echo "  ci                 - Run CI pipeline locally"

build:
	@echo "Building binaries..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/bridge
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(CLI_NAME) ./cmd/bridge-cli
	@echo "✓ Build complete"

test: test-unit

test-unit:
	@echo "Running unit tests..."
	$(GO) test -v -race -short ./...
	@echo "✓ Unit tests passed"

test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -race -tags=integration ./test/integration/...
	@echo "✓ Integration tests passed"

test-hardware:
	@echo "Running hardware tests (requires physical devices)..."
	@echo "Note: Ensure USB HID scanner is connected and configured"
	@echo "See test/hardware/README.md for setup instructions"
	@echo ""
	$(GO) test -v -tags=hardware ./test/hardware/...
	@echo "✓ Hardware tests passed"

bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem -run=^$$ ./...
	@echo "✓ Benchmarks complete"

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "✓ Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found" && exit 1)
	golangci-lint run --timeout=5m
	@echo "✓ Linting passed"

fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "✓ Code formatted"

mod-tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy
	@echo "✓ Modules tidied"

proto:
	@echo "Generating protobuf code..."
	@./scripts/generate-proto.sh
	@echo "✓ Protobuf generated"

docker:
	@echo "Building Docker image..."
	docker build -t device-bridge:latest .
	@echo "✓ Docker image built"

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html
	@echo "✓ Clean complete"

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -config configs/config.example.yaml

ci: lint test-unit test-coverage build

.DEFAULT_GOAL := help
