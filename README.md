# Harness Template Documentation Generator

Automatically generates documentation from Harness templates and publishes to Confluence. Designed to run within Harness pipelines after merges to main.

## Features

- Processes pipeline, stage, and stepgroup templates
- Extracts metadata, variables, parameters, and examples
- Generates HTML documentation with search and filtering
- Publishes directly to Confluence
- Runs in Harness pipelines

## Usage in Harness Pipeline

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

## Local Development

### Environment Setup

Create a `.env` file with your Confluence credentials:

```
# Confluence Configuration
CONFLUENCE_URL=https://your-domain.atlassian.net
CONFLUENCE_USERNAME=your-username@example.com
CONFLUENCE_API_TOKEN="YOUR_API_TOKEN_HERE"
CONFLUENCE_SPACE_KEY=YOUR_SPACE_KEY
CONFLUENCE_PARENT_PAGE_ID=123456789

# Processing Options
SOURCE_DIR=templates
OUTPUT_DIR=docs/templates
FORMAT=html
VERBOSE=true
PUBLISH=true
VALIDATE_ONLY=false
```

### Using Docker

```bash
# Build
docker build -t harness-template-docs .

# Run locally
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs harness-template-docs --verbose
```

### Using Helper Script

```bash
./publish-to-confluence.sh
```

## Setting Up Confluence Integration

### API Token

1. Go to: https://id.atlassian.com/manage-profile/security/api-tokens
2. Create a new API token named "Harness Template Documentation"
3. Store the token in Harness secrets

### Finding Your Parent Page ID

The page ID appears in the URL when viewing a Confluence page:
`https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`

In this example, the page ID is `123456789`.

## Template Structure

```yaml
name: Template Name
description: Template description
type: pipeline|stage|stepgroup
author: Author Name
version: 1.0.0
tags:
  - tag1
  - tag2

# Template-specific structure
# ...

# Examples of usage
examples:
  - |
    # Example usage
    # ...
```

## Directory Structure

```
harness-template-docs/
├── Dockerfile
├── README.md
├── main.py
├── process_template.py
├── publish-to-confluence.sh
├── harness-step-example.yaml
├── requirements.txt
├── src/
│   └── utils/
│       ├── template_html.py
│       └── confluence_publisher.py
└── templates/
    ├── pipeline/
    ├── stage/
    └── stepgroup/
```

## License

MIT

