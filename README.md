# Harness Template Documentation Generator

Automatically extracts and generates documentation from Harness templates, with optional publishing to Confluence. Designed to run within Harness pipelines.

## Overview

This tool analyzes Harness template YAML files (pipeline, stage, and stepgroup) to extract:
- Template metadata (name, type, version, author)
- Variables and parameters with descriptions
- Usage examples
- Steps and configurations

The extracted information is organized into searchable HTML documentation that can be:
- Generated locally for reference
- Published directly to Confluence for team visibility
- Integrated into CI/CD pipelines

## Quick Start

### Local Development

1. **Setup Environment**:
   ```bash
   # Clone repository and install dependencies
   pip install -r requirements.txt
   
   # Create .env file (copy from .env.example)
   cp .env.example .env
   # Edit .env with your settings
   ```

2. **Generate Documentation**:
   ```bash
   # Using helper script
   ./publish-to-confluence.sh
   
   # Or using Python directly
   python process_template.py --source templates --output docs/templates --format html
   ```

3. **Using Docker**:
   ```bash
   # Build image
   docker build -t harness-template-docs .
   
   # Run container
   docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs harness-template-docs --verbose
   ```

## Harness Pipeline Integration

Add this step to your pipeline to automatically update documentation:

```yaml
- step:
    type: Run
    name: Generate Template Documentation
    identifier: generate_template_documentation
    spec:
      connectorRef: <+variables.docker_registry_connector>
      image: <+variables.image_name>
      command: |
        python process_template.py \
          --source /harness/input/codebase/templates \
          --output /harness/output/docs \
          --format html \
          --publish \
          --confluence-url <+secrets.getValue("confluence_url")> \
          --confluence-username <+secrets.getValue("confluence_username")> \
          --confluence-token <+secrets.getValue("confluence_token")> \
          --confluence-space <+secrets.getValue("confluence_space")> \
          --confluence-parent-id <+secrets.getValue("confluence_parent_id")> \
          --verbose
```

## Confluence Setup

1. **Generate API Token**:
   - Go to: https://id.atlassian.net/manage-profile/security/api-tokens
   - Create token named "Harness Template Documentation"

2. **Find Parent Page ID**:
   - Open your Confluence space
   - The ID is in the URL: `https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`
   - ID in this example: `123456789`

3. **Configure Environment**:
   - Set variables in `.env` file or Harness secrets

## Template Requirements

Templates must include these fields:

```yaml
name: Template Name
description: Template description
type: pipeline|stage|stepgroup
author: Author Name
version: 1.0.0
tags:
  - tag1
  - tag2

# Template structure goes here...

# Optional usage examples
examples:
  - |
    # Example code
    # ...
```

## Command Line Options

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

## License

MIT

