# Template Documentation Generator Makefile

# Binary name and build directory
BINARY_NAME=tempdocs
FILESERVER_BINARY=fileserver
BUILD_DIR=bin
GO_MAIN=./cmd/tempdocs
FILESERVER_MAIN=./cmd/fileserver

# Configuration (with defaults that can be overridden by .env file)
SOURCE_DIR ?= templates
OUTPUT_DIR ?= docs/output
FORMAT ?= html
VALIDATE_ONLY ?= false
VERBOSE ?= false

# Docker image settings
DOCKER_IMAGE=ka1ne/template-doc-gen
DOCKER_TAG=0.1.0-go

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
include .env
export
endif

############################
# Core Application Targets #
############################

# Default target
.PHONY: all
all: build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)/*

# Create necessary directories
.PHONY: dirs
dirs:
	@echo "Creating necessary directories..."
	mkdir -p $(BUILD_DIR)
	mkdir -p $(OUTPUT_DIR)

# Build the core application
.PHONY: build
build: dirs
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_MAIN)
	@echo "Build successful. Binary location: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests for core application
.PHONY: test
test: build
	@echo "Running unit tests..."
	go test -v ./pkg/...

# Run quick validation test
.PHONY: test-validate
test-validate: build
	@echo "Running validation test..."
	./$(BUILD_DIR)/$(BINARY_NAME) --source $(SOURCE_DIR) --output $(OUTPUT_DIR) --validate --verbose

# Generate documentation
.PHONY: generate
generate: build
	@echo "Generating documentation from templates in $(SOURCE_DIR)..."
	./$(BUILD_DIR)/$(BINARY_NAME) --source $(SOURCE_DIR) --output $(OUTPUT_DIR) --format $(FORMAT) $(if $(filter true,$(VERBOSE)),--verbose,) $(if $(filter true,$(VALIDATE_ONLY)),--validate,)

# Build Docker image (production)
.PHONY: docker-build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run in Docker (production)
.PHONY: docker-run
docker-run:
	@echo "Starting Harness Template Documentation Generator in Docker..."
	@echo "Using Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "Source directory: $(SOURCE_DIR)"
	@echo "Output directory: $(OUTPUT_DIR)"
	
	@# Ensure output directory exists with proper permissions
	mkdir -p $(OUTPUT_DIR)
	chmod 777 $(OUTPUT_DIR)
	
	docker run --rm \
		-v $(abspath $(SOURCE_DIR)):/app/templates \
		-v $(abspath $(OUTPUT_DIR)):/app/docs \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG) \
		$(if $(filter true,$(VERBOSE)),--verbose,) \
		$(if $(filter true,$(VALIDATE_ONLY)),--validate,)

############################
# Development Tools        #
############################

# Build the file server (development only)
.PHONY: build-fileserver
build-fileserver: dirs
	@echo "Creating file server if it doesn't exist..."
	@if [ ! -d ./cmd/fileserver ]; then \
		mkdir -p ./cmd/fileserver; \
		echo 'package main\n\nimport (\n\t"flag"\n\t"log"\n\t"net/http"\n\t"os"\n\t"path/filepath"\n)\n\nfunc main() {\n\tport := flag.String("port", "8000", "Port to serve on")\n\tdir := flag.String("dir", ".", "Directory to serve files from")\n\tflag.Parse()\n\n\tabsDir, err := filepath.Abs(*dir)\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\n\tlog.Printf("Serving files from %s on http://localhost:%s", absDir, *port)\n\tlog.Printf("Press Ctrl+C to stop")\n\t\n\terr = http.ListenAndServe(":"+*port, http.FileServer(http.Dir(absDir)))\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n}' > ./cmd/fileserver/main.go; \
	fi
	@echo "Building $(FILESERVER_BINARY)..."
	go build -o $(BUILD_DIR)/$(FILESERVER_BINARY) $(FILESERVER_MAIN)
	@echo "File server built successfully."

# Run server to preview documentation (development only)
.PHONY: serve
serve: generate build-fileserver
	@echo "Starting local server to preview documentation..."
	@echo "Open your browser and navigate to http://localhost:8000/"
	./$(BUILD_DIR)/$(FILESERVER_BINARY) -dir $(OUTPUT_DIR)

# Build development Docker image (includes dev tools)
.PHONY: docker-build-dev
docker-build-dev:
	@echo "Building development Docker image $(DOCKER_IMAGE):dev..."
	docker build -f Dockerfile.dev -t $(DOCKER_IMAGE):dev .

############################
# Help                     #
############################

# Help command
.PHONY: help
help:
	@echo "Harness Template Documentation Generator"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Core Application Targets:"
	@echo "  build          Build the core application"
	@echo "  clean          Clean build artifacts"
	@echo "  test           Run unit tests"
	@echo "  test-validate  Run quick validation test"
	@echo "  generate       Generate documentation"
	@echo "  docker-build   Build Docker image (production)"
	@echo "  docker-run     Run using Docker (production)"
	@echo ""
	@echo "Development Tools:"
	@echo "  build-fileserver  Build the development file server"
	@echo "  serve             Generate docs and serve locally (development)"
	@echo "  docker-build-dev  Build development Docker image"
	@echo ""
	@echo "Environment variables (can be set in .env file):"
	@echo "  SOURCE_DIR     Source directory (default: $(SOURCE_DIR))"
	@echo "  OUTPUT_DIR     Output directory (default: $(OUTPUT_DIR))"
	@echo "  FORMAT         Output format [html|json|markdown] (default: $(FORMAT))"
	@echo "  VALIDATE_ONLY  Only validate templates (default: $(VALIDATE_ONLY))"
	@echo "  VERBOSE        Enable verbose logging (default: $(VERBOSE))" 