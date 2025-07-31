.PHONY: build run test test-quick test-verbose test-verbose-quick test-unit test-unit-verbose clean deps help

BINARY_NAME=e2e_multinode
BUILD_DIR=build
GOBUILD=go build
GOCLEAN=go clean

# Default target
all: build

# Build the binary to build directory
build:
	@echo "🔨 Building $(BINARY_NAME) to $(BUILD_DIR)/..."
	@mkdir -p $(BUILD_DIR)
	gobuild -o $(BUILD_DIR)/$(BINARY_NAME) -v .

# Run basic determinism test scenario using go run
test:
	@echo "🚀 Running basic determinism test..."
	go run . test -validators=2 -non-validators=3 -height=205 -timeout=15m

# Run quick test scenario (lower height for faster execution)
test-quick:
	@echo "⚡ Running quick determinism test..."
	go run . test -validators=2 -non-validators=2 -height=50 -timeout=5m

# Run verbose test with full logging
test-verbose:
	@echo "🔍 Running verbose determinism test..."
	go run . test -validators=2 -non-validators=3 -height=205 -timeout=15m -verbose

# Run verbose quick test
test-verbose-quick:
	@echo "⚡🔍 Running quick verbose determinism test..."
	go run . test -validators=2 -non-validators=2 -height=50 -timeout=5m -verbose

# Run unit tests using testify
test-unit:
	@echo "🧪 Running unit tests with testify..."
	go test -v ./...

# Run unit tests with verbose output
test-unit-verbose:
	@echo "🧪🔍 Running verbose unit tests with testify..."
	go test -v -race ./...

# Run with custom parameters
run:
	@echo "🏃 Running with custom parameters..."
	@echo "Usage: make run ARGS=\"test -validators=2 -non-validators=3 -height=100 -timeout=10m\""
	go run . $(ARGS)

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary to $(BUILD_DIR)/ directory"
	@echo "  test         - Run basic determinism test (2 validators + 3 non-validators, height 205)"
	@echo "  test-quick   - Run quick test (2 validators + 2 non-validators, height 50)"
	@echo "  test-verbose       - Run test with verbose logging"
	@echo "  test-verbose-quick - Run quick test with verbose logging"
	@echo "  test-unit          - Run unit tests with testify"
	@echo "  test-unit-verbose  - Run unit tests with race detection and verbose output"
	@echo "  run          - Run with custom parameters (use ARGS=\"...\")"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install/update dependencies"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make test                                         # Run standard e2e test"
	@echo "  make test-unit                                    # Run unit tests with testify"
	@echo "  make test-verbose                                 # Run with full debug logging"
	@echo "  make run ARGS=\"test -validators=4 -height=100\"    # Run with custom args"
	@echo "  make build                                        # Build to $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "  $(BUILD_DIR)/$(BINARY_NAME) test -help            # Run built binary"
