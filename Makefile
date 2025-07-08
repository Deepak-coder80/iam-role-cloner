# Makefile for IAM Role Cloner

APP_NAME := iam-role-cloner
VERSION := 1.0.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build directory
BUILD_DIR := build

# Go build flags
LDFLAGS := -s -w
LDFLAGS += -X 'iam-role-cloner/cmd.Version=$(VERSION)'
LDFLAGS += -X 'iam-role-cloner/cmd.GitCommit=$(GIT_COMMIT)'
LDFLAGS += -X 'iam-role-cloner/cmd.BuildDate=$(BUILD_DATE)'

# Default target
.PHONY: all
all: clean test build

# Clean build directory
.PHONY: clean
clean:
	@echo "🧹 Cleaning build directory..."
	rm -rf $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Build for current platform
.PHONY: build
build:
	@echo "🔨 Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) .

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "🚀 Building $(APP_NAME) v$(VERSION) for all platforms..."
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"
	@echo ""
	@mkdir -p $(BUILD_DIR)

	@echo "📦 Building binaries..."
	@echo "  Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

	@echo "  Building for Linux (arm64)..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .

	@echo "  Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

	@echo "  Building for macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

	@echo "  Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .

	@echo "  Building for Windows (arm64)..."
	@GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe .

	@echo ""
	@echo "✅ Build completed! Binaries created in $(BUILD_DIR)/"
	@echo ""
	@echo "📊 Build artifacts:"
	@ls -lh $(BUILD_DIR)/

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Run locally
.PHONY: run
run: build
	@echo "🏃‍♂️ Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Development mode - build and run
.PHONY: dev
dev:
	@echo "🔧 Development mode..."
	go run . $(ARGS)

# Format code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Security check
.PHONY: security
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping..."; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Create release archives
.PHONY: package
package: build-all
	@echo "📦 Creating release packages..."
	@mkdir -p $(BUILD_DIR)/packages

	# Linux packages
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-arm64

	# macOS packages
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64
	@tar -czf $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-arm64

	# Windows packages
	@zip -j $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe
	@zip -j $(BUILD_DIR)/packages/$(APP_NAME)-$(VERSION)-windows-arm64.zip $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe

	@echo "📦 Packages created in $(BUILD_DIR)/packages/"
	@ls -lh $(BUILD_DIR)/packages/

# Install locally
.PHONY: install
install: build
	@echo "📥 Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "✅ $(APP_NAME) installed successfully!"
	@echo "🎯 Test with: $(APP_NAME) version"

# Uninstall
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling $(APP_NAME)..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@echo "✅ $(APP_NAME) uninstalled successfully!"

# Show help
.PHONY: help
help:
	@echo "📚 IAM Role Cloner - Available commands:"
	@echo ""
	@echo "🔨 Build commands:"
	@echo "  make build      - Build for current platform"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make package    - Create release packages"
	@echo ""
	@echo "🧪 Development commands:"
	@echo "  make dev ARGS='clone --help'  - Run in development mode"
	@echo "  make run        - Build and run"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make lint       - Lint code"
	@echo "  make security   - Run security checks"
	@echo ""
	@echo "📦 Dependencies:"
	@echo "  make deps       - Install dependencies"
	@echo ""
	@echo "🏠 Installation:"
	@echo "  make install    - Install to /usr/local/bin"
	@echo "  make uninstall  - Remove from /usr/local/bin"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  make clean      - Clean build directory"
	@echo ""
	@echo "📋 Current build info:"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo "  Go Version: $(GO_VERSION)"
