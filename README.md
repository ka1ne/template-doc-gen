# Harness Template Docs Generator

> Generate beautiful documentation from Harness templates automatically.

[![Version](https://img.shields.io/badge/version-0.0.3--alpha-blue)](https://hub.docker.com/repository/docker/ka1ne/template-doc-gen/tags/0.0.3-alpha/sha256:af1c5885d18f1b3e7d758da1427cb890005af62d05cb28a9f218766d39b0ff9e)

## What It Does

Transforms Harness templates into searchable HTML documentation, extracting metadata, parameters, variables, and examples while validating against official schemas.

## Quick Start

### Option 1: Docker (Recommended)

```bash
# Run with helper script
git clone https://github.com/your-org/harness-template-docs.git
cd harness-template-docs
./docker-run.sh
```

### Option 2: Manual Setup

```bash
# Local setup
pip install -r requirements.txt
python process_template.py --source templates --output docs/templates
```

### Option 3: Harness Pipeline

```yaml
# In your Harness pipeline
stages:
  - stage:
      template:
        name: template-doc-gen
        identifier: templatedocgen
        versionLabel: v0.0.3-alpha
      variables:
        docker_registry_connector: your_connector_id
```

## Configuration

Set options via `.env` file or command line arguments:

```
SOURCE_DIR=templates     # Where to find templates
OUTPUT_DIR=docs/output   # Where to write docs
FORMAT=html              # Output format
VERBOSE=true             # Show detailed logs
```

## Documentation Integration

```

## Command Reference

```bash
python process_template.py --help

# Key Arguments
--source DIR    Source directory with templates  
--output DIR    Output directory for documentation
--format FMT    Output format (html|markdown|json)
--verbose       Show detailed logs
--validate      Validate templates without generating
```

## Learn More

For advanced usage, see the [Harness Schema](https://github.com/harness/harness-schema) documentation.
