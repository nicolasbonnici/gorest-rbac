.PHONY: help test test-coverage lint lint-fix build build-cli clean install

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install development tools
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
	@echo "✓ Tools installed"

test: ## Run tests with coverage
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total:
	@rm -f coverage.out

test-coverage: ## Run tests with HTML coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out

lint: ## Run linter
	@echo "Running golangci-lint..."
	@$$(go env GOPATH)/bin/golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@$$(go env GOPATH)/bin/golangci-lint run --fix ./...

build: ## Build plugin
	@echo "Building plugin..."
	@go build -v ./...
	@echo "✓ Plugin build successful"

build-cli: ## Build CLI tool
	@echo "Building CLI tool..."
	@go build -o rbac-cli ./cmd/rbac-cli
	@echo "✓ CLI build successful: ./rbac-cli"

clean: ## Clean build artifacts and caches
	@echo "Cleaning..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out coverage.html
	@rm -f rbac-cli
	@echo "✓ Cleaned"

all: lint test build build-cli ## Run all checks and builds
