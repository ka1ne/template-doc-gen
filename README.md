# Harness Template Docs Generator

> Generate documentation from Harness templates automatically.

[![Version](https://img.shields.io/badge/version-1\.0\.0-blue)](https://github.com/ka1ne/template-doc-gen)

## What It Does

Transforms Harness templates into searchable HTML documentation, extracting metadata, parameters, variables, and examples while validating against official schemas.

## Quick Start

### Option 1: Using Make (Recommended)

```bash
# Clone the repository
git clone https://github.com/ka1ne/template-doc-gen.git
cd template-doc-gen

# Generate documentation
make generate

# View documentation in browser
make serve
```

### Option 2: Docker

```bash
# Using production Docker image
docker run --rm \
  -v "$(pwd)/templates:/app/templates" \
  -v "$(pwd)/docs/output:/app/docs" \
  ka1ne/template-doc-gen:0.0.4-alpha

# Or using Make
make docker-run
```

### Option 3: Harness Pipeline

```yaml
# In your Harness pipeline
stages:
  - stage:
      template:
        name: template-doc-gen
        identifier: templatedocgen
        versionLabel: v0.0.4-alpha
      variables:
        docker_registry_connector: your_connector_id
```

## Configuration

Set options via `.env` file or environment variables:

```
SOURCE_DIR=templates     # Where to find templates
OUTPUT_DIR=docs/output   # Where to write docs
FORMAT=html              # Output format (html, json)
VERBOSE=true             # Show detailed logs
VALIDATE_ONLY=false      # Only validate, don't generate docs
```

## Development

### Application Structure

```
pkg/
  ├── template/   # Template processing and metadata extraction
  ├── schema/     # JSON schema handling and validation
  └── utils/      # Configuration and utility functions
cmd/
  └── tempdocs/   # Main application entry point
  └── fileserver/ # Simple file server for local viewing (dev only)
```

### Development Environment

For local development, you can use the development Docker image which includes additional tools:

```bash
# Build the development Docker image
make docker-build-dev

# Run the development container
docker run --rm -it \
  -v "$(pwd):/app" \
  -p 8000:8000 \
  ka1ne/template-doc-gen:dev
```

## Available Make Commands

```
# Core Application
make build                  # Build the core application
make clean                  # Clean build artifacts
make test                   # Run unit tests
make test-validate          # Run quick validation test
make generate               # Generate documentation
make docker-build           # Build Docker image (production)
make docker-build TAG=x.y.z # Build Docker image with custom tag
make docker-run             # Run using Docker (production)
make docker-run TAG=x.y.z   # Run using Docker with custom tag
make docker-buildx          # Build multi-architecture Docker images and push
make docker-buildx-local    # Build multi-architecture Docker images locally

# Version Management
make set-version TAG=x.y.z  # Update version across all files
make release                # Release the version in DOCKER_TAG
make release TAG=x.y.z      # Release a custom version

# Development Tools
make build-fileserver       # Build the development file server
make serve                  # Generate docs and serve locally (development)
make docker-build-dev       # Build development Docker image
```

## Releasing New Versions

To release a new version:

### Option 1: Set Version First (Recommended)

1. Update version across all files with one command:
   ```bash
   make set-version TAG=x.y.z
   # Example: make set-version TAG=0.1.0-beta
   ```

2. Commit the version changes:
   ```bash
   git add internal/version/version.go Makefile README.md Dockerfile* examples/
   git commit -m "Bump version to x.y.z"
   ```

3. Run the release process:
   ```bash
   make release
   ```

### Option 2: Custom Release (One-off)

If you want to build and release a specific version without changing the default version in the files:

```bash
make release TAG=x.y.z
# Example: make release TAG=0.1.0-rc.1
```

This approach is useful for testing or creating release candidates without changing the main version.

### What Happens During Release

The release process:
- Builds and tests the application
- Generates documentation
- Builds and pushes multi-architecture Docker images for all variants:
  - `ka1ne/template-doc-gen:[version]` - Main image
  - `ka1ne/template-doc-gen:latest` - Latest tag
  - `ka1ne/template-doc-gen:pipeline` - Pipeline image
  - `ka1ne/template-doc-gen:dev` - Development image
- Optionally creates and pushes a git tag (you'll be prompted during the release)

The Docker images are built for multiple architectures:
- linux/amd64 (Intel/AMD 64-bit)
- linux/arm64 (ARM 64-bit, e.g. Apple Silicon, AWS Graviton)
- linux/arm/v7 (ARM 32-bit, e.g. Raspberry Pi)

If you skip the git tag during release, you can create it later:
```bash
git tag -a v1.0.0 -m "Release 1.0.0"
git push origin v1.0.0
```

## Roadmap
- Improve test coverage
- Add detailed schema validation
- Add a "what's changed" section in generated documentation
- Support additional output formats

## Learn More

For advanced usage, see the [Harness Schema](https://github.com/harness/harness-schema) documentation.
