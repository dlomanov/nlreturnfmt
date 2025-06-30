LOCAL_BIN := $(shell pwd)/bin
GOLANGCI_LINT_VERSION = v2.1.6

.PHONY: help install-tools format build run test test-verbose test-coverage lint clean

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.bin: # Create bin directory (hidden from help)
	mkdir -p bin

.install-pre-commit:
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "installing pre-commit..."; \
		pip install pre-commit; \
	fi

install-tools: .bin .install-pre-commit ## Install all required tools
	@if [ ! -x "$(LOCAL_BIN)/golangci-lint" ]; then \
		echo "installing golangci-lint..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@if [ ! -x "$(LOCAL_BIN)/gofumpt" ]; then \
		echo "installing gofumpt..."; \
		GOBIN=$(LOCAL_BIN) go install mvdan.cc/gofumpt@latest; \
	fi

format: install-tools ## Format code using pre-commit hooks
	pre-commit run --all-files

build: ## Build the application
	go build -o bin/nlreturnfmt ./cmd/nlreturnfmt

run: build ## Build and run the application
	./bin/nlreturnfmt

test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -cover ./...

lint: install-tools ## Run linter
	./bin/golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf bin/

