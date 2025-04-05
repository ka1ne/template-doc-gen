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
DOCKER_TAG=1.0.0

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
	@echo "Building Docker image $(DOCKER_IMAGE):$(if $(TAG),$(TAG),$(DOCKER_TAG))..."
	docker build -t $(DOCKER_IMAGE):$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.

# run in docker
.PHONY: docker-run
docker-run:
	@echo "Starting Harness Template Documentation Generator in Docker..."
	@echo "Using Docker image: $(DOCKER_IMAGE):$(if $(TAG),$(TAG),$(DOCKER_TAG))"
	@echo "Source directory: $(SOURCE_DIR)"
	@echo "Output directory: $(OUTPUT_DIR)"
	
	@# ensure output dir exists with permissions
	mkdir -p $(OUTPUT_DIR)
	chmod 777 $(OUTPUT_DIR)
	
	docker run --rm \
		-v $(abspath $(SOURCE_DIR)):/app/templates \
		-v $(abspath $(OUTPUT_DIR)):/app/docs \
		--env-file .env \
		$(DOCKER_IMAGE):$(or $(TAG),$(DOCKER_TAG)) \
		$(if $(filter true,$(VERBOSE)),--verbose,) \
		$(if $(filter true,$(VALIDATE_ONLY)),--validate,)

# build multi-architecture docker images
.PHONY: docker-buildx
docker-buildx:
	@echo "Building multi-architecture Docker images for $(DOCKER_IMAGE):$(if $(TAG),$(TAG),$(DOCKER_TAG))..."
	docker buildx create --name tempdocs-builder --use --bootstrap || true
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
		-f Dockerfile.multiarch \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):$(or $(TAG),$(DOCKER_TAG)) \
		--push .
	@echo "Multi-architecture build complete and pushed to Docker Hub"

# build multi-architecture docker images without pushing
.PHONY: docker-buildx-local
docker-buildx-local:
	@echo "Building multi-architecture Docker images for local use..."
	docker buildx create --name tempdocs-builder --use --bootstrap || true
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
		-f Dockerfile.multiarch \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):$(or $(TAG),$(DOCKER_TAG)) \
		--load .
	@echo "Multi-architecture build complete and loaded locally"

# build and push all image variants
.PHONY: docker-release-all
docker-release-all: docker-buildx
	@echo "Building and pushing all Docker image variants..."
	
	# Build and push pipeline image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f Dockerfile.pipeline \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):pipeline \
		--push .
	
	# Build and push development image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f Dockerfile.dev \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):dev \
		--push .
	
	# Tag the main version as latest as well
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
		-f Dockerfile.multiarch \
		--build-arg VERSION=$(or $(TAG),$(DOCKER_TAG)) \
		--build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		--build-arg BUILD_DATE=$$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-t $(DOCKER_IMAGE):latest \
		--push .
	
	@echo "All Docker images built and pushed to Docker Hub"

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
serve: build build-fileserver generate
	@echo "Starting local server to preview documentation..."
	@echo "Open your browser and navigate to http://localhost:8000/"
	./$(BUILD_DIR)/$(BINARY_NAME) serve -d $(OUTPUT_DIR) -p 8000 --no-generate

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
	@echo "  docker-buildx  Build multi-architecture Docker images and push to Docker Hub"
	@echo "  docker-buildx-local Build multi-architecture Docker images locally"
	@echo "  docker-release-all Build and push all Docker image variants"
	@echo "  release        Release a new version (includes build, test, generate, and docker publishing)"
	@echo "  set-version    Update version across all files (Usage: make set-version TAG=x.y.z-suffix)"
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
	@echo ""
	@echo "Parameters:"
	@echo "  TAG            Override the default tag ($(DOCKER_TAG)) for docker commands"
	@echo "                 Example: make docker-build TAG=custom-tag"
	@echo "                 Works with: docker-build, docker-run, docker-buildx, etc."

# release a new version
.PHONY: release
release:
	@echo "==========================================="
	@echo "Releasing version $(if $(TAG),$(TAG),$(DOCKER_TAG))"
	@echo "==========================================="
	@echo ""
	
	# Build and run tests
	@echo "Building and testing..."
	$(MAKE) clean
	$(MAKE) build
	$(MAKE) test
	
	# Generate documentation
	@echo "Generating documentation..."
	$(MAKE) generate
	
	# Build and push Docker images
	@echo "Building and pushing Docker images for all platforms..."
	$(MAKE) docker-release-all TAG=$(or $(TAG),$(DOCKER_TAG))
	
	# Ask about git tag
	@echo ""
	@echo "Would you like to create and push a git tag for v$(if $(TAG),$(TAG),$(DOCKER_TAG))? [y/N]"
	@read -r REPLY; \
	if [ "$$REPLY" = "y" ] || [ "$$REPLY" = "Y" ]; then \
		echo "Creating git tag v$(if $(TAG),$(TAG),$(DOCKER_TAG))..."; \
		git tag -a "v$(if $(TAG),$(TAG),$(DOCKER_TAG))" -m "Release $(if $(TAG),$(TAG),$(DOCKER_TAG))"; \
		echo "Pushing git tag..."; \
		git push origin "v$(if $(TAG),$(TAG),$(DOCKER_TAG))"; \
		echo "Git tag v$(if $(TAG),$(TAG),$(DOCKER_TAG)) created and pushed."; \
	else \
		echo "Skipping git tag creation."; \
		echo "You can create it later with:"; \
		echo "  git tag -a v$(if $(TAG),$(TAG),$(DOCKER_TAG)) -m \"Release $(if $(TAG),$(TAG),$(DOCKER_TAG))\""; \
		echo "  git push origin v$(if $(TAG),$(TAG),$(DOCKER_TAG))"; \
	fi
	
	# Success message
	@echo ""
	@echo "==========================================="
	@echo "Release $(if $(TAG),$(TAG),$(DOCKER_TAG)) completed!"
	@echo "==========================================="

# update version across all files
.PHONY: set-version
set-version:
	@if [ -z "$(TAG)" ]; then \
		echo "Error: TAG parameter is required."; \
		echo "Usage: make set-version TAG=x.y.z-suffix"; \
		exit 1; \
	fi
	@echo "Updating version to $(TAG) across all files..."
	
	# Update version in internal/version/version.go
	@sed -i 's/Version = ".*"/Version = "$(TAG)"/' internal/version/version.go
	
	# Update DOCKER_TAG in Makefile
	@sed -i 's/DOCKER_TAG=1.0.0
	
	# Update version badge in README.md
	@sed -i 's/version-.*-blue/version-$(subst .,\\.,$(TAG))-blue/' README.md
	@sed -i 's/ka1ne\/template-doc-gen:.*\"/ka1ne\/template-doc-gen:$(TAG)\"/' README.md
	@sed -i 's/versionLabel: v.*"/versionLabel: v$(TAG)"/' README.md
	
	# Update git tag commands in README.md
	@sed -i 's/git tag -a v[0-9].*-m "Release [0-9].*"/git tag -a v$(TAG) -m "Release $(TAG)"/' README.md
	@sed -i 's/git push origin v[0-9].*/git push origin v$(TAG)/' README.md
	
	# Update version in Dockerfiles
	@sed -i 's/VERSION=.*/VERSION=$(TAG)/' Dockerfile
	@sed -i 's/VERSION=.*/VERSION=$(TAG)/' Dockerfile.dev
	@sed -i 's/VERSION=.*/VERSION=$(TAG)/' Dockerfile.pipeline
	@sed -i 's/VERSION=.*/VERSION=$(TAG)/' Dockerfile.multiarch
	@sed -i 's/org.opencontainers.image.version=".*/org.opencontainers.image.version="$(TAG)"/' Dockerfile.pipeline
	@sed -i 's/org.opencontainers.image.version=".*/org.opencontainers.image.version="$(TAG)"/' Dockerfile.multiarch
	
	# Update examples directory if present
	@if [ -d "examples" ]; then \
		echo "Updating examples directory..."; \
		find examples -type f -name "*.yaml" -exec sed -i 's/versionLabel: v[0-9].*"/versionLabel: v$(TAG)"/' {} \; ; \
		find examples -type f -name "*.yaml" -exec sed -i 's/ka1ne\/template-doc-gen:[0-9].*/ka1ne\/template-doc-gen:$(TAG)/' {} \; ; \
	fi
	
	@echo "Version updated to $(TAG) in all files."
	@echo "Don't forget to commit these changes:"
	@echo "  git add internal/version/version.go Makefile README.md Dockerfile* examples/"
	@echo "  git commit -m \"Bump version to $(TAG)\"" 