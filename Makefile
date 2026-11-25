BINARY ?= bin/functiond
LDFLAGS ?= -s -w

.DEFAULT_GOAL := help
.PHONY: help build test lint fmt generate clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "%-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the functiond executable
	@mkdir -p $(dir $(BINARY))
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/runner

test: ## Run all Go tests with coverage
	go test -v -cover ./...

lint: ## Run golangci-lint
	go tool golangci-lint run ./...

fmt: ## Format Go source files
	go fmt ./...

generate: ## Generate interface mocks
	go generate ./...

clean: ## Remove build artifacts
	rm -rf bin
