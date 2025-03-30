# Harness Template Documentation Generator

A tool that automatically extracts and generates documentation from Harness templates with Confluence integration, packaged as a ready-to-use stage template.

## Overview

This documentation generator analyzes Harness YAML templates and creates organized, searchable documentation by:
1. Extracting metadata, variables, parameters, and examples
2. Generating HTML documentation with search and filtering 
3. Publishing to Confluence (optional)

## Using the Harness Stage Template

This project provides a ready-to-use Harness stage template that you can add to any pipeline.

### 1. Import the Stage Template

Copy the stage template to your Harness project:
- Source file: `templates/harness/template-doc-gen.yaml`
- Import it through the Harness UI or API

### 2. Add the Stage to Your Pipeline

```yaml
stages:
  - stage:
      name: Generate Template Documentation
      identifier: generate_template_documentation
      template:
        name: template-doc-gen
        identifier: templatedocgen
        versionLabel: v0.0.1-alpha
      variables:
        # Required variable
        docker_registry_connector: your_docker_connector_id
        
        # Optional variables with custom values
        image_name: ka1ne/template-doc-gen:0.0.1-alpha
        source_dir: templates
        output_dir: docs/templates
```

### 3. Enable Confluence Publishing (Optional)

```yaml
stages:
  - stage:
      name: Generate Template Documentation
      identifier: generate_template_documentation
      template:
        name: template-doc-gen
        identifier: templatedocgen
        versionLabel: v0.0.1-alpha
      variables:
        # Required variable
        docker_registry_connector: your_docker_connector_id
        
        # Confluence publishing options
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

## Complete Pipeline Example

A complete pipeline example can be found at `templates/harness/example-usage.yaml` which includes:
- The documentation generator stage
- Configurable variables for customization
- Example usage patterns

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
| `image_name` | Docker image name | `ka1ne/template-doc-gen:0.0.1-alpha` | No |
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

# Using Python directly
python process_template.py --source templates --output docs/templates --format html
```

### Docker Usage

```bash
# Build container
docker build -t ka1ne/template-doc-gen:0.0.1-alpha .

# Run with volume mounts
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs ka1ne/template-doc-gen:0.0.1-alpha --verbose
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

