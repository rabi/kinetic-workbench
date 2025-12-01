.PHONY: build run clean test fmt vet lint help

# Binary name
BINARY_NAME=kinetic

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.DEFAULT_GOAL := help

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/kinetic
	@echo "Build complete: bin/$(BINARY_NAME)"

## run: Run the application
run:
	@go run ./cmd/kinetic

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete"

## test: Run tests
test:
	@go test -v ./...

## fmt: Format code
fmt:
	@go fmt ./...

## vet: Run go vet
vet:
	@go vet ./...

## lint: Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## tidy: Tidy go.mod
tidy:
	@go mod tidy

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed 's/## //'


