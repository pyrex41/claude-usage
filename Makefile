# Claude Usage - Makefile

.PHONY: all build install clean test fmt lint help

# Binary name
BINARY := claude-usage
CMD_PATH := ./cmd/claude-usage

# Build flags
LDFLAGS := -ldflags "-s -w"

all: build

## Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD_PATH)

## Install to GOPATH/bin
install:
	go install $(CMD_PATH)

## Clean build artifacts
clean:
	rm -f $(BINARY)
	go clean

## Run tests
test:
	go test ./...

## Format code
fmt:
	go fmt ./...

## Run linter (if golangci-lint is installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

## Run with sample data
sample:
	./$(BINARY) daily --path testdata --compact

## Run with real data (by project)
real:
	./$(BINARY) daily --instances --compact

## Show help
help:
	./$(BINARY) daily --help

## Development build with debug info
dev:
	go build -o $(BINARY) $(CMD_PATH)

## Update dependencies
deps:
	go mod tidy
	go mod verify

## Show version info
version:
	@echo "Claude Usage - Go rewrite"
	@echo "Build date: $$(date)"
	@go version

# Default target
.DEFAULT_GOAL := help