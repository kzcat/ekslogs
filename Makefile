.PHONY: build test lint fmt clean install-tools

# バイナリ名
BINARY_NAME=ekslogs
# バージョン情報
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# ビルドフラグ
LDFLAGS=-ldflags "-X github.com/kzcat/ekslogs/cmd.version=$(VERSION) -X github.com/kzcat/ekslogs/cmd.commit=$(COMMIT) -X github.com/kzcat/ekslogs/cmd.date=$(DATE)"

# デフォルトターゲット
all: lint test build

# ビルド
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME)

# テスト実行
test:
	@echo "Running tests..."
	@go test -v ./...

# テストカバレッジ
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# コード整形
fmt:
	@echo "Formatting code..."
	@gofmt -l -w .

# 静的解析
vet:
	@echo "Running go vet..."
	@go vet ./...

# golangci-lintによるリント
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it first."; \
		echo "See: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# クリーンアップ
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html

# 必要なツールのインストール
install-tools:
	@echo "Installing required tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/pre-commit/pre-commit@latest

# pre-commit hookのインストール
install-hooks:
	@echo "Installing pre-commit hooks..."
	@cp .git/hooks/pre-commit .git/hooks/pre-commit.bak 2>/dev/null || true
	@cp ./.git/hooks/pre-commit ./.git/hooks/pre-commit
	@chmod +x ./.git/hooks/pre-commit
	@echo "Pre-commit hook installed successfully."

# ヘルプ
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
