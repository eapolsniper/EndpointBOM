.PHONY: build clean install test run help

# Binary name
BINARY_NAME=endpointbom

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/endpointbom/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/endpointbom/main.go
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/endpointbom/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/endpointbom/main.go
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/endpointbom/main.go
	@echo "Build complete for all platforms"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/endpointbom/main.go
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) cmd/endpointbom/main.go
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Display help
help:
	@echo "EndpointBOM Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the binary"
	@echo "  make build-all   - Build for all platforms (macOS, Windows, Linux)"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make deps        - Download and tidy dependencies"
	@echo "  make test        - Run tests"
	@echo "  make run         - Build and run the application"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make help        - Display this help message"

