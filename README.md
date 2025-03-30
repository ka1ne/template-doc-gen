# Harness Template Documentation Generator

A tool that automatically extracts and generates documentation from Harness templates with Confluence integration.

## Overview

This documentation generator analyzes Harness YAML templates (pipeline, stage, and stepgroup) and:
1. Extracts metadata, variables, parameters, and examples
2. Generates searchable HTML documentation 
3. Publishes to Confluence (optional)
4. Integrates seamlessly with Harness CI/CD pipelines

## Best Practices

### Template Structure

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

### Documentation Generation

* **Validate Templates**: Use the `--validate` flag to check templates before generating docs
* **Format Options**: Choose between HTML (default), Markdown, or JSON outputs
* **Include Examples**: Add clear examples to your templates for better documentation
* **Organize Output**: Structure output by template type (pipeline/stage/stepgroup)

### Confluence Integration

* **Secure Credentials**: Store API tokens in Harness secrets, not in code
* **Parent Pages**: Create a dedicated parent page in Confluence for template docs
* **Regular Updates**: Schedule documentation generation after merge to main
* **Error Handling**: Check logs after publishing to verify successful uploads

## Quick Start Guide

### Local Development

1. **Setup Environment**:
   ```bash
   # Install dependencies
   pip install -r requirements.txt
   
   # Configure environment
   cp .env.example .env
   # Edit .env with your settings
   ```

2. **Generate Documentation**:
   ```bash
   # Quick generation with helper script
   ./publish-to-confluence.sh
   
   # Or customize with Python
   python process_template.py --source templates --output docs/templates --format html
   ```

3. **Docker Deployment**:
   ```bash
   # Build container
   docker build -t harness-template-docs .
   
   # Run with volume mounts
   docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs harness-template-docs --verbose
   ```

## Harness Pipeline Integration

### Complete Example Pipeline

The following example demonstrates a complete pipeline setup with best practices:

```yaml
pipeline:
  name: Update Template Documentation
  identifier: update_template_documentation
  projectIdentifier: <+project.identifier>
  orgIdentifier: <+org.identifier>
  tags: {}
  stages:
    - stage:
        name: Generate and Publish Documentation
        identifier: generate_and_publish_documentation
        type: CI
        spec:
          cloneCodebase: true
          execution:
            steps:
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
                    privileged: false
                    shell: Sh
                    envVariables:
                      PYTHONUNBUFFERED: "1"
                    resources:
                      limits:
                        memory: 512Mi
                        cpu: 500m
          platform:
            os: Linux
            arch: Amd64
          runtime:
            type: Cloud
            spec: {}
          caching:
            enabled: true
            paths:
              - /harness/output/docs
  properties:
    ci:
      codebase:
        connectorRef: <+variables.git_connector>
        repoName: <+variables.repo_name>
        build: <+input>

variables:
  - name: docker_registry_connector
    type: String
    description: Connector ID for Docker registry
    required: true
  - name: git_connector
    type: String
    description: Connector ID for Git repository
    required: true
  - name: repo_name
    type: String
    description: Name of the repository containing Harness templates
    required: true
  - name: image_name
    type: String
    description: Docker image name for documentation generator
    required: true
    value: harness-template-docs:latest
```

### Key Pipeline Features

1. **Resource Limits**: Sets memory (512Mi) and CPU (500m) limits for efficient execution
2. **Caching**: Enables caching of documentation output for faster builds
3. **Secrets Management**: Uses Harness secrets for all sensitive credentials
4. **Environment Variables**: Sets PYTHONUNBUFFERED for improved logging
5. **Variables**: Defines required connectors and values with descriptive comments

## Confluence Setup

1. **API Token Generation**:
   - Navigate to: https://id.atlassian.net/manage-profile/security/api-tokens
   - Create a token named "Harness Template Documentation"
   - Store in Harness secrets management

2. **Page Structure**:
   - Create a parent page for all template documentation
   - Find the page ID in the URL: `https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`
   - Use this ID as the parent page ID in configuration

3. **Environment Configuration**:
   ```
   CONFLUENCE_URL=https://your-domain.atlassian.net
   CONFLUENCE_USERNAME=your-username@example.com
   CONFLUENCE_API_TOKEN="YOUR_API_TOKEN_HERE"
   CONFLUENCE_SPACE_KEY=YOUR_SPACE_KEY
   CONFLUENCE_PARENT_PAGE_ID=123456789
   ```

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

