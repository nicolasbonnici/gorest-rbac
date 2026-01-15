# RBAC CLI Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags "-X 'github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli/commands.version=$(VERSION)' -X 'github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli/commands.commit=$(COMMIT)' -X 'github.com/nicolasbonnici/gorest-rbac/cmd/rbac-cli/commands.date=$(DATE)'"

# Binary name
BINARY := rbac-cli

# Platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o $(BINARY) .
	@echo "Build complete: $(BINARY)"

.PHONY: install
install:
	@echo "Installing $(BINARY) to $(GOPATH)/bin..."
	go install $(LDFLAGS) .
	@echo "Installation complete"

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	rm -f $(BINARY)-*
	rm -f .rbac-cli.yaml
	@echo "Clean complete"

.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied"

.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor
	@echo "Vendoring complete"

.PHONY: release
release: clean
	@echo "Building release binaries..."
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build $(LDFLAGS) -o $(BINARY)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,) . ; \
	done
	@echo "Release builds complete"

.PHONY: release-linux
release-linux:
	@echo "Building Linux binary..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .
	@echo "Linux build complete"

.PHONY: release-darwin
release-darwin:
	@echo "Building macOS binaries..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64 .
	@echo "macOS builds complete"

.PHONY: release-windows
release-windows:
	@echo "Building Windows binary..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe .
	@echo "Windows build complete"

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary (default)"
	@echo "  install         - Install to GOPATH/bin"
	@echo "  clean           - Remove built binaries and config files"
	@echo "  test            - Run tests"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  tidy            - Tidy dependencies"
	@echo "  vendor          - Vendor dependencies"
	@echo "  release         - Build binaries for all platforms"
	@echo "  release-linux   - Build Linux binary"
	@echo "  release-darwin  - Build macOS binaries"
	@echo "  release-windows - Build Windows binary"
	@echo "  version         - Show version information"
	@echo "  help            - Show this help message"
