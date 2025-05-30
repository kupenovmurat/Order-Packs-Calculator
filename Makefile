# Pack Calculator Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop lint fmt vet

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run all tests"
	@echo "  test-edge    - Run the edge case test specifically"
	@echo "  bench        - Run benchmark tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run application in Docker"
	@echo "  docker-stop  - Stop Docker containers"
	@echo "  lint         - Run golangci-lint"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download dependencies"

# Build the application
build:
	@echo "Building pack calculator..."
	go build -o bin/pack-calculator ./cmd/server

# Run the application locally
run:
	@echo "Starting pack calculator server..."
	go run ./cmd/server

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run the edge case test specifically
test-edge:
	@echo "Running edge case test..."
	go test -v ./internal/service -run TestPackCalculatorService_EdgeCase

# Run benchmark tests
bench:
	@echo "Running benchmark tests..."
	go test -bench=. ./internal/service

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t pack-calculator .

# Run application in Docker
docker-run:
	@echo "Starting application with Docker Compose..."
	docker-compose up -d

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Run all checks (format, vet, lint, test)
check: fmt vet test
	@echo "All checks passed!"

# Quick development cycle
dev: fmt vet test run

# Production build
prod-build: clean fmt vet test build
	@echo "Production build complete!" 