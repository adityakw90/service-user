.PHONY: test test-cover test-clean build run help verbose

# Default target
.DEFAULT_GOAL := help

# Verbose target (used as: make test verbose)
verbose:
	@:

# Test (usage: make test or make test verbose)
test:
	@if echo "$(MAKECMDGOALS)" | grep -q "verbose"; then \
		echo "Running tests with verbose output..."; \
		go test -v ./...; \
	else \
		echo "Running tests..."; \
		go test ./...; \
	fi

# Test with coverage
test-cover:
	@echo "Running tests with coverage..."
	@if echo "$(MAKECMDGOALS)" | grep -q "verbose"; then \
		go test -v -race -covermode=atomic -coverprofile=coverage.txt ./...; \
	else \
		go test -race -covermode=atomic -coverprofile=coverage.txt ./...; \
	fi

# Clean test cache and coverage files
test-clean:
	@echo "Cleaning test cache and coverage files..."
	@go clean -testcache
	@rm -f coverage.txt
	@echo "Clean complete"

build:
	@echo "Building the application..."
	@GOOS=linux GOARCH=amd64 go build -o bin/app cmd/main.go

lint:
	@echo "Linting the application..."
	@golangci-lint run

run:
	@echo "Running the application..."
	@go run cmd/main.go

# Help target
help:
	@echo "Available targets:"
	@echo "  test [verbose]       - Run all tests (add verbose for verbose output)"
	@echo "  test-cover [verbose] - Run tests with coverage (add verbose for verbose output)"
	@echo "  test-clean           - Clean test cache and coverage files"
	@echo "  bench [verbose]      - Run all benchmarks (add verbose for verbose output)"
	@echo "  build                - Build the application"
	@echo "  run                  - Run the application"
	@echo "  help                 - Show this help message"
