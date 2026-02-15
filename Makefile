.PHONY: help test test-coverage lint lint-fix build build-cli clean install

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies, dev tools, and git hooks
	@echo "[INFO] Installing development environment..."
	@echo ""
	@echo "[1/3] Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"
	@echo ""
	@echo "[2/3] Installing development tools..."
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "  Installing golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "✓ Development tools installed"
	@echo ""
	@echo "[3/3] Installing git hooks..."
	@bash .githooks/install.sh
	@echo ""
	@echo "✅ Installation complete! Ready to develop."
	@echo ""
	@echo "Next steps:"
	@echo "  • Run 'make test' to verify your setup"
	@echo "  • Run 'make lint' to check code quality"
	@echo "  • See 'make help' for all available commands"
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
	@cd cmd/rbac-cli && go build -o ../../rbac-cli
	@echo "✓ CLI build successful: ./rbac-cli"

clean: ## Clean build artifacts and caches
	@echo "Cleaning..."
	@go clean -cache -testcache -modcache
	@rm -f coverage.out coverage.html
	@rm -f rbac-cli
	@echo "✓ Cleaned"

all: lint test build build-cli ## Run all checks and builds
