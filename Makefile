.PHONY: build clean install test run lint help release

BINARY_NAME=ekslogs
BUILD_DIR=bin

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-s -w -X github.com/kzcat/ekslogs/cmd.version=$(VERSION) -X github.com/kzcat/ekslogs/cmd.commit=$(COMMIT) -X github.com/kzcat/ekslogs/cmd.date=$(BUILD_DATE)"

all: build

help:
	@echo "Available targets:"
	@echo "  build            - Build the binary for the current platform"
	@echo "  build-all        - Build binaries for multiple platforms"
	@echo "  install          - Install the binary to /usr/local/bin"
	@echo "  clean            - Clean build artifacts"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  test-coverage-html - Generate HTML coverage report"
	@echo "  lint             - Run linter"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  run              - Build and run the binary"
	@echo "  release          - Create release archives"
	@echo "  example-basic    - Run basic example"
	@echo "  example-tail     - Run tail mode example"
	@echo "  example-filter   - Run filter example"
	@echo "  example-preset   - Run preset filter example"

build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 .
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 .
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe .

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

test-coverage-html:
	@echo "Generating HTML coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linter..."
	$(GOLINT) run

deps:
	$(GOMOD) download
	$(GOMOD) tidy

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/release
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)_linux_amd64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)_darwin_amd64
	tar -czf $(BUILD_DIR)/release/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)_darwin_arm64
	zip -j $(BUILD_DIR)/release/$(BINARY_NAME)_$(VERSION)_windows_amd64.zip $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe
	@echo "Release archives created in $(BUILD_DIR)/release/"
