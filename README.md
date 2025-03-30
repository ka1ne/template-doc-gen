# Harness Template Documentation Generator

A tool that automatically extracts and generates documentation from Harness templates with Confluence integration, packaged as a ready-to-use stage template.

## Overview

This documentation generator analyzes Harness YAML templates and creates organized, searchable documentation by:
1. Extracting metadata, variables, parameters, and examples
2. Generating HTML documentation with search and filtering 
3. Publishing to Confluence (optional)

## Using the Stage Template

This project provides a ready-to-use Harness stage template that you can add to any pipeline.

### 1. Add the Template to Your Harness Project

Copy the stage template to your Harness project:
- Source file: `templates/stage/docs_generator_stage.yaml`
- Import it through the Harness UI or CLI

### 2. Add the Stage to Your Pipeline

```yaml
stages:
  - stage:
      template:
        name: Documentation Generator Stage
        identifier: docs_generator_stage
        versionLabel: 1.0.0
      variables:
        docker_registry_connector: your_docker_connector_id
        image_name: harness-template-docs:latest
        source_dir: templates
        output_dir: docs/templates
```

### 3. Enable Confluence Publishing (Optional)

```yaml
stages:
  - stage:
      template:
        name: Documentation Generator Stage
        identifier: docs_generator_stage
        versionLabel: 1.0.0
      variables:
        docker_registry_connector: your_docker_connector_id
        image_name: harness-template-docs:latest
        source_dir: templates
        output_dir: docs/templates
        publish_to_confluence: true
        confluence_url_secret: confluence_url
        confluence_username_secret: confluence_username
        confluence_token_secret: confluence_token
        confluence_space_secret: confluence_space
        confluence_parent_id_secret: confluence_parent_id
```

### 4. Configure Secrets

Create the following secrets in your Harness project if using Confluence:
- `confluence_url`
- `confluence_username`
- `confluence_token`
- `confluence_space`
- `confluence_parent_id`

## Stage Template Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `source_dir` | Directory containing templates | `templates` | No |
| `output_dir` | Output directory for documentation | `docs/templates` | No |
| `format` | Output format (html, markdown, json) | `html` | No |
| `publish_to_confluence` | Whether to publish to Confluence | `false` | No |
| `confluence_url_secret` | Secret for Confluence URL | | No |
| `confluence_username_secret` | Secret for Confluence username | | No |
| `confluence_token_secret` | Secret for Confluence API token | | No |
| `confluence_space_secret` | Secret for Confluence space key | | No |
| `confluence_parent_id_secret` | Secret for Confluence parent page ID | | No |
| `image_name` | Docker image name | `harness-template-docs:latest` | No |
| `docker_registry_connector` | Connector ID for Docker registry | | Yes |

## Local Development

If you need to run or develop the tool outside of Harness:

### Setup Environment

```bash
# Install dependencies
pip install -r requirements.txt

# Configure environment
cp .env.example .env
# Edit .env with your settings
```

### Run Locally

```bash
# Using helper script
./publish-to-confluence.sh

# Or using Python directly
python process_template.py --source templates --output docs/templates --format html
```

### Docker Usage

```bash
# Build container
docker build -t harness-template-docs .

# Run with volume mounts
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs harness-template-docs --verbose
```

## Template Requirements

For optimal documentation, ensure your templates follow this structure:

```yaml
name: Template Name
description: A clear, concise description of what this template does
type: pipeline|stage|stepgroup
author: Author Name
version: 1.0.0
tags:
  - tag1
  - tag2

# Template structure here...

# Optional usage examples
examples:
  - |
    # Example code
    # ...
```

## Confluence Setup

1. **Generate API Token**:
   - Go to: https://id.atlassian.net/manage-profile/security/api-tokens
   - Create token named "Harness Template Documentation"

2. **Find Parent Page ID**:
   - The ID is in the Confluence page URL: `https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`
   - ID in this example: `123456789`

3. **Store in Harness Secrets**:
   Create the following secrets in your Harness project:
   ```
   confluence_url = https://your-domain.atlassian.net
   confluence_username = your-username@example.com
   confluence_token = YOUR_API_TOKEN_HERE
   confluence_space = YOUR_SPACE_KEY
   confluence_parent_id = 123456789
   ```

## Best Practices

- **Schedule Regular Updates**: Run the documentation pipeline after merges to main
- **Validate Templates**: Use the `--validate` flag to check templates without generating docs
- **Include Examples**: Add clear examples to your templates for better documentation
- **Secure Credentials**: Always use Harness secrets for sensitive information
- **Check Logs**: Verify successful publishing by checking pipeline logs

## Command Reference

```
python process_template.py --help

Arguments:
  --source, -s            Source directory containing templates
  --output, -o            Output directory for documentation
  --format, -f            Output format (html, markdown, json)
  --validate, -v          Validate templates without generating docs
  --verbose               Enable detailed logging
  --publish, -p           Publish to Confluence
  --confluence-url        Confluence URL
  --confluence-username   Confluence username
  --confluence-token      Confluence API token
  --confluence-space      Confluence space key
  --confluence-parent-id  Confluence parent page ID
```

