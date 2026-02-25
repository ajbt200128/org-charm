.PHONY: build test run clean fmt lint

# Default target
all: build

# Build the binary
build:
	devenv shell -- go build -o org-charm .

# Run tests
test:
	devenv shell -- go test -v ./...

# Run tests with coverage
test-coverage:
	devenv shell -- go test -coverprofile=coverage.out ./...
	devenv shell -- go tool cover -html=coverage.out -o coverage.html

# Run the server
run: build
	devenv shell -- ./org-charm

# Run with custom options
run-dev: build
	devenv shell -- ./org-charm -port 2222 -dir ./orgfiles

# Format code
fmt:
	devenv shell -- go fmt ./...

# Lint code
lint:
	devenv shell -- go vet ./...

# Clean build artifacts
clean:
	rm -f org-charm coverage.out coverage.html

# Install dependencies
deps:
	devenv shell -- go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the org-charm binary"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Build and run the server"
	@echo "  run-dev       - Run with development settings"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Run go vet"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Tidy Go modules"
	@echo "  help          - Show this help"
