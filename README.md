# Harness Template Docs Generator

> Generate documentation from Harness templates automatically.

[![Version](https://img.shields.io/badge/version-0.1.0--go-blue)](https://github.com/ka1ne/template-doc-gen)

## What It Does

Transforms Harness templates into searchable HTML documentation, extracting metadata, parameters, variables, and examples while validating against official schemas.

![demo-gif](https://github.com/user-attachments/assets/c96991bf-9846-483b-8fc2-f5271d70926a)

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
  ka1ne/template-doc-gen:0.1.0-go

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
        versionLabel: v0.1.0-go
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
make build          # Build the core application
make clean          # Clean build artifacts
make test           # Run unit tests
make test-validate  # Run quick validation test
make generate       # Generate documentation
make docker-build   # Build Docker image (production)
make docker-run     # Run using Docker (production)

# Development Tools
make build-fileserver  # Build the development file server
make serve             # Generate docs and serve locally (development)
make docker-build-dev  # Build development Docker image
```

## roadmap
- improve test coverage
- add detailed schema validation
- add a "what's changed" section in generated documentation
- support additional output formats

## learn more

For advanced usage, see the [Harness Schema](https://github.com/harness/harness-schema) documentation.
