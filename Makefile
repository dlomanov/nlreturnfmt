LOCAL_BIN := $(shell pwd)/bin
GOLANGCI_LINT_VERSION = v2.1.6

bin:
	mkdir -p bin

install-tools: bin
	@if [ ! -x "$(LOCAL_BIN)/golangci-lint" ]; then \
		echo "installing golangci-lint..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@if [ ! -x "$(LOCAL_BIN)/smartimports" ]; then \
		echo "installing smartimports..."; \
		GOBIN=$(LOCAL_BIN) go install github.com/pav5000/smartimports/cmd/smartimports@latest; \
	fi
	@if [ ! -x "$(LOCAL_BIN)/gofumpt" ]; then \
		echo "installing gofumpt..."; \
		GOBIN=$(LOCAL_BIN) go install mvdan.cc/gofumpt@latest; \
	fi

format: install-tools
	pre-commit run --all-files

build:
	go build -o bin/nlreturnfmt ./cmd/nlreturnfmt

run: build
	./bin/nlreturnfmt

