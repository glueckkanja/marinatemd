SHELL := /bin/bash
BINARY_NAME := marinatemd
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/c4a8-azure/marinatemd/cmd/marinatemd.Version=$(VERSION) \
           -X github.com/c4a8-azure/marinatemd/cmd/marinatemd.Commit=$(COMMIT) \
           -X github.com/c4a8-azure/marinatemd/cmd/marinatemd.BuildDate=$(BUILD_DATE)

.PHONY: all
all: build

.PHONY: build
build: ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

.PHONY: install
install: ## Install the CLI binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" .

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

.PHONY: mod
mod: ## Download and tidy dependencies
	@echo "Tidying modules..."
	@go mod download
	@go mod tidy

.PHONY: run
run: ## Run the CLI (use ARGS to pass arguments, e.g., make run ARGS=".")
	@go run -ldflags "$(LDFLAGS)" . $(ARGS)

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
