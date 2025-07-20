.PHONY: build test lint fmt clean install-tools

# Binary name
BINARY_NAME=ekslogs
# Version information
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Build flags
LDFLAGS=-ldflags "-X github.com/kzcat/ekslogs/cmd.version=$(VERSION) -X github.com/kzcat/ekslogs/cmd.commit=$(COMMIT) -X github.com/kzcat/ekslogs/cmd.date=$(DATE)"

# Default target
all: lint test build

# Build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Test coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -l -w .

# Static analysis
vet:
	@echo "Running go vet..."
	@go vet ./...

# Lint with golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint &> /dev/null; then \
		echo "Using golangci-lint version: $$(golangci-lint --version)"; \
		golangci-lint run -E errcheck,govet,ineffassign,staticcheck,unused; \
	else \
		echo "golangci-lint not found. Please install it first."; \
		echo "See: https://golangci-lint.run/usage/install/"; \
		echo "You can install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2"; \
		exit 1; \
	fi

# Cleanup
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f bin/$(BINARY_NAME)
	@rm -f coverage.out coverage.html

# Install required tools
install-tools:
	@echo "Installing required tools..."
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	@echo "golangci-lint installed: $$(golangci-lint --version)"
	@echo "Installing pre-commit..."
	@if command -v pip3 &> /dev/null; then \
		pip3 install pre-commit; \
	else \
		echo "pip3 not found. Please install Python3 and pip3 first."; \
	fi
	@echo "Installation complete."

# Install pre-commit hooks
install-hooks:
	@echo "Installing pre-commit hooks..."
	@cp .git/hooks/pre-commit .git/hooks/pre-commit.bak 2>/dev/null || true
	@cp ./.git/hooks/pre-commit ./.git/hooks/pre-commit
	@chmod +x ./.git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully."

# Help
help:
	@echo "Available targets:"
	@echo "  all          : Run lint, test, and build"
	@echo "  build        : Build the binary"
	@echo "  test         : Run tests"
	@echo "  coverage     : Generate test coverage report"
	@echo "  fmt          : Format code"
	@echo "  vet          : Run go vet"
	@echo "  lint         : Run golangci-lint"
	@echo "  clean        : Clean up build artifacts"
	@echo "  install-tools: Install required tools"
	@echo "  install-hooks: Install pre-commit hooks"
	@echo "  help         : Show this help message"
