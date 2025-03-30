# 📚 Harness Template Docs Generator

> Automagically extract metadata from Harness templates and publish beautiful, searchable docs to Confluence.

![Version](https://img.shields.io/badge/version-0.0.1--alpha-blue)
![Docker](https://img.shields.io/badge/docker-ready-brightgreen)

## ✨ What It Does

Transforms your Harness templates into well-organized documentation:

- 🔍 **Extracts** metadata, variables, parameters, and examples
- 🎨 **Generates** beautiful HTML documentation with search capabilities
- 🚀 **Publishes** directly to Confluence (optional)
- 🔄 **Integrates** with your CI/CD pipelines via Harness stage templates
- 📝 **Validates** your templates against the official Harness schema

## 🚀 Quick Start

### Option 1: One-Click Docker Run (Recommended)

```bash
# Clone this repo
git clone https://github.com/your-org/harness-template-docs.git
cd harness-template-docs

# Run the helper script (will create .env from example if needed)
./docker-run.sh
```

That's it! The script handles everything - mounting volumes, setting up environment variables, and running the container.

### Option 2: Manual Docker Run

```bash
# 1. Create and configure your .env file
cp .env.example .env
# Edit .env with your settings

# 2. Create output directory with proper permissions
mkdir -p docs/output
chmod 777 docs/output

# 3. Run with Docker
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs \
  --env-file .env ka1ne/template-doc-gen:0.0.1-alpha
```

### Option 3: Use In Harness Pipeline

Import the stage template (`templates/harness/template-doc-gen.yaml`):

```yaml
# In your Harness pipeline
stages:
  - stage:
      template:
        name: template-doc-gen
        identifier: templatedocgen
        versionLabel: v0.0.1-alpha
      variables:
        docker_registry_connector: your_connector_id
        # Enable Confluence publishing (optional)
        publish_to_confluence: true
```

## 🔧 Configuration (.env file)

```
# Confluence Configuration
CONFLUENCE_URL=https://your-company.atlassian.net
CONFLUENCE_USERNAME=your-email@example.com
CONFLUENCE_API_TOKEN=your-api-token
CONFLUENCE_SPACE_KEY=TEAM            # or ~123456789 for personal spaces
CONFLUENCE_PARENT_PAGE_ID=123456     # Page ID where docs will be created

# Processing Options
SOURCE_DIR=templates                 # Where to find templates
OUTPUT_DIR=docs/output               # Where to write docs
FORMAT=html                          # Output format
VERBOSE=true                         # Show detailed logs
PUBLISH=true                         # Enable publishing to Confluence
```

## 📋 Template Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `docker_registry_connector` | 🔗 Docker registry connector | **Required** |
| `image_name` | 🐳 Image to use | `ka1ne/template-doc-gen:0.0.1-alpha` |
| `source_dir` | 📁 Where to find templates | `templates` |
| `output_dir` | 📂 Where to write docs | `docs/templates` |
| `format` | 📄 Output format | `html` |
| `publish_to_confluence` | 🚀 Auto-publish to Confluence | `false` |

## 🔗 Confluence Setup

1. **Create API Token**: at https://id.atlassian.net/manage-profile/security/api-tokens
2. **Get Page ID**: From URL - `https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`
3. **Add to .env**: Update the `.env` file with your credentials

## 🧠 Pro Tips

- **Personal Space**: Use `~61dd87bce67ea2006b2c1082` format for personal space keys
- **Check Permissions**: Ensure your API token has permission to create pages
- **Private Repos**: You can host the image in a private Docker registry
- **Run on Merge**: Set up webhooks to run docs generation when templates change
- **Include Examples**: Template examples make for better documentation

## 📖 Full Example

For a complete implementation, see `templates/harness/example-usage.yaml`

## 🛠️ Shell Scripts

This repo includes helper scripts to simplify your workflow:

- **docker-run.sh**: One-click Docker execution with environment handling
- **publish-to-confluence.sh**: Local Python execution with Confluence publishing

## 🧪 Local Development

### Docker 

```bash
# Run with Docker
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs \
  ka1ne/template-doc-gen:0.0.1-alpha --verbose
```

### Manual Setup

```bash
# Local setup
pip install -r requirements.txt
cp .env.example .env  # Edit with your settings

# Run manually
python process_template.py --source templates --output docs/templates
```

## 💡 Best Practices

- **Template Structure**: Include name, description, type, author, version and tags
- **Run on Merge**: Setup webhooks to update docs automatically
- **Use Secrets**: Never hardcode Confluence credentials
- **Include Examples**: Make templates self-documenting with examples

## 🛠️ Command Reference

```bash
python process_template.py --help

# Key Arguments
  --source DIR          Source directory with templates  
  --output DIR          Output directory for documentation
  --format FMT          Output format (html|markdown|json)
  --publish             Publish to Confluence
  --verbose             Show detailed logs
```

## 📋 Features

- **Official Schema Validation**: Templates are validated against the official Harness schema
- **Beautiful Documentation**: Generates searchable HTML documentation with syntax highlighting
- **Confluence Integration**: Publish documentation directly to your Confluence workspace
- **Docker Ready**: Run in any environment with Docker support
- **Flexible Output**: Generate documentation in HTML format (with Markdown and JSON coming soon)

## 🧠 Pro Tips

- **Schema Validation**: Uses the official [Harness Schema](https://github.com/harness/harness-schema) to validate templates
- **Personal Space**: Use `~61dd87bce67ea2006b2c1082` format for personal space keys
- **Check Permissions**: Ensure your API token has permission to create pages
- **Private Repos**: You can host the image in a private Docker registry
- **Run on Merge**: Set up webhooks to run docs generation when templates change

