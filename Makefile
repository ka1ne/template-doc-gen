# template docs generator makefile

# binaries and directories
BINARY_NAME=tempdocs
FILESERVER_BINARY=fileserver
BUILD_DIR=bin
GO_MAIN=./cmd/tempdocs
FILESERVER_MAIN=./cmd/fileserver

# config defaults
SOURCE_DIR ?= templates
OUTPUT_DIR ?= docs/output
FORMAT ?= html
VALIDATE_ONLY ?= false
VERBOSE ?= false

# docker settings
DOCKER_IMAGE=ka1ne/template-doc-gen
DOCKER_TAG=0.1.0-go

# load env file if exists
ifneq (,$(wildcard .env))
include .env
export
endif

############################
# core targets             #
############################

# default target
.PHONY: all
all: build

# clean artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)/*

# create dirs
.PHONY: dirs
dirs:
	@echo "Creating necessary directories..."
	mkdir -p $(BUILD_DIR)
	mkdir -p $(OUTPUT_DIR)

# build app
.PHONY: build
build: dirs
	@echo "Building $(BINARY_NAME)..."
	go build \
		-ldflags="-s -w \
		-X github.com/ka1ne/template-doc-gen/internal/version.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") \
		-X github.com/ka1ne/template-doc-gen/internal/version.Commit=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		-X github.com/ka1ne/template-doc-gen/internal/version.BuildDate=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-X github.com/ka1ne/template-doc-gen/internal/version.GoVersion=$$(go version | cut -d ' ' -f 3)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(GO_MAIN)
	@echo "Build successful. Binary location: $(BUILD_DIR)/$(BINARY_NAME)"

# run tests
.PHONY: test
test: build
	@echo "Running unit tests..."
	go test -v ./pkg/...

# run validation test
.PHONY: test-validate
test-validate: build
	@echo "Running validation test..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate --source $(SOURCE_DIR) $(if $(filter true,$(VERBOSE)),--verbose,)

# generate docs
.PHONY: generate
generate: build
	@echo "Generating documentation from templates in $(SOURCE_DIR)..."
	./$(BUILD_DIR)/$(BINARY_NAME) generate --source $(SOURCE_DIR) --output $(OUTPUT_DIR) --format $(FORMAT) $(if $(filter true,$(VERBOSE)),--verbose,)

# build docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# run in docker
.PHONY: docker-run
docker-run:
	@echo "Starting Harness Template Documentation Generator in Docker..."
	@echo "Using Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "Source directory: $(SOURCE_DIR)"
	@echo "Output directory: $(OUTPUT_DIR)"
	
	@# ensure output dir exists with permissions
	mkdir -p $(OUTPUT_DIR)
	chmod 777 $(OUTPUT_DIR)
	
	docker run --rm \
		-v $(abspath $(SOURCE_DIR)):/app/templates \
		-v $(abspath $(OUTPUT_DIR)):/app/docs \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG) \
		$(if $(filter true,$(VERBOSE)),--verbose,) \
		$(if $(filter true,$(VALIDATE_ONLY)),--validate,)

# build pipeline image
.PHONY: docker-build-pipeline
docker-build-pipeline:
	@echo "Building pipeline-optimized Docker image $(DOCKER_IMAGE):pipeline..."
	docker build -f Dockerfile.pipeline \
		--build-arg VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev") \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):pipeline .

# test pipeline image
.PHONY: test-pipeline-image
test-pipeline-image: docker-build-pipeline
	@echo "Testing pipeline Docker image..."
	@echo "1. Testing help command:"
	docker run --rm $(DOCKER_IMAGE):pipeline
	@echo "\n2. Testing version command:"
	docker run --rm $(DOCKER_IMAGE):pipeline version
	@echo "\n3. Testing validate command with the example templates:"
	docker run --rm -v $(PWD)/templates:/templates $(DOCKER_IMAGE):pipeline validate --source /templates
	@echo "\n4. Testing generate command:"
	mkdir -p $(OUTPUT_DIR)
	docker run --rm \
		-v $(PWD)/templates:/templates \
		-v $(PWD)/$(OUTPUT_DIR):/output \
		$(DOCKER_IMAGE):pipeline generate --source /templates --output /output

############################
# dev tools               #
############################

# build fileserver
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

# serve docs locally
.PHONY: serve
serve: build build-fileserver
	@echo "Starting local server to preview documentation..."
	@echo "Open your browser and navigate to http://localhost:8000/"
	./$(BUILD_DIR)/$(BINARY_NAME) serve --dir=$(OUTPUT_DIR) --port=8000

# build dev image
.PHONY: docker-build-dev
docker-build-dev:
	@echo "Building development Docker image $(DOCKER_IMAGE):dev..."
	docker build -f Dockerfile.dev -t $(DOCKER_IMAGE):dev .

############################
# help                     #
############################

# show help
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