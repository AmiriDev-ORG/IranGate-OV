# IranGate Makefile
# Version: 1.0.0

.PHONY: all build clean install test help

# Variables
BINARY_NAME=irangate
BUILD_DIR=build
GO_VERSION=1.24.0

# Default target
all: build

# Build the application
build:
	@echo "🔨 Building IranGate..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f main
	@rm -f test_build
	@echo "✅ Clean completed"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Format code
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@go vet ./...
	@echo "✅ Linting completed"

# Build for multiple platforms
build-all: clean
	@echo "🌍 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "✅ Multi-platform build completed"

# Development build with race detection
dev-build:
	@echo "🚀 Building development version..."
	@go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev .
	@echo "✅ Development build completed"

# Show help
help:
	@echo "IranGate Build System"
	@echo "===================="
	@echo ""
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install dependencies"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  dev-build  - Build development version with race detection"
	@echo "  help       - Show this help"