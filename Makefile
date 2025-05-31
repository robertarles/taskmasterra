.PHONY: all build clean test lint vet fmt help cross-build release release-patch release-minor release-major

# Binary name
BINARY_NAME=taskmasterra
# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=gofmt -s -w

# Main package path
MAIN_PACKAGE=./cmd/taskmasterra

# Version information
CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION ?= $(CURRENT_VERSION)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(subst v,,$(VERSION)) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Cross compilation settings
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64

# Default target
all: clean lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Cross-compile for all platforms
cross-build:
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach p,$(PLATFORMS), \
		echo "Building for $(p)..." && \
		GOOS=$(word 1,$(subst /, ,$p)) \
		GOARCH=$(word 2,$(subst /, ,$p)) \
		$(GOBUILD) $(LDFLAGS) \
		-o "$(BUILD_DIR)/$(BINARY_NAME)-$(word 1,$(subst /, ,$p))-$(word 2,$(subst /, ,$p))" \
		$(MAIN_PACKAGE) && \
		echo "Done building for $(p)" && \
	) true

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run linting
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) .

# Install binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@echo "To install, run one of these commands:"
	@echo "  GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) go install $(LDFLAGS) github.com/robertarles/taskmasterra/v2/cmd/taskmasterra@$(VERSION)"
	@echo "  # Or for latest:"
	@echo "  go install github.com/robertarles/taskmasterra/v2/cmd/taskmasterra@latest"

# Version bumping targets
MAJOR=$(word 1,$(subst ., ,$(subst v,,$(CURRENT_VERSION))))
MINOR=$(word 2,$(subst ., ,$(subst v,,$(CURRENT_VERSION))))
PATCH=$(word 3,$(subst ., ,$(subst v,,$(CURRENT_VERSION))))

release-patch: VERSION=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))
release-patch: release

release-minor: VERSION=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0
release-minor: release

release-major: VERSION=v$(shell echo $$(($(MAJOR)+1))).0.0
release-major: release

# Create a new release
release:
	@if [ "$(VERSION)" = "$(CURRENT_VERSION)" ]; then \
		echo "Error: No version specified. Use make release-patch, release-minor, or release-major"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@echo "Building release artifacts..."
	@make clean
	@make cross-build
	@echo "Build artifacts are ready in the build directory"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin HEAD
	@git push origin $(VERSION)
	@echo "Created and pushed tag $(VERSION)"
	@echo "Don't forget to create the release on GitHub with the build artifacts"

# Show help
help:
	@echo "Available targets:"
	@echo "  make                - Build the application (same as 'make all')"
	@echo "  make build          - Build the application"
	@echo "  make cross-build    - Build for all platforms (darwin/amd64, darwin/arm64, linux/amd64)"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter"
	@echo "  make vet            - Run go vet"
	@echo "  make fmt            - Format code"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo "  make release-patch  - Release a new patch version (x.y.Z)"
	@echo "  make release-minor  - Release a new minor version (x.Y.0)"
	@echo "  make release-major  - Release a new major version (X.0.0)"
	@echo "  make help           - Show this help" 