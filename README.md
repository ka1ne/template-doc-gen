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

## 🚀 Quick Start

### Import the Stage Template

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

### Setup in 3 Steps

1. Import `templates/harness/template-doc-gen.yaml` to your Harness project
2. Reference it in your pipeline (see above)
3. For Confluence publishing, add these secrets:
   - `confluence_url`, `confluence_username`, `confluence_token`
   - `confluence_space`, `confluence_parent_id`

## 📋 Template Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `docker_registry_connector` | 🔗 Docker registry connector | **Required** |
| `image_name` | 🐳 Image to use | `ka1ne/template-doc-gen:0.0.1-alpha` |
| `source_dir` | 📁 Where to find templates | `templates` |
| `output_dir` | 📂 Where to write docs | `docs/templates` |
| `format` | 📄 Output format | `html` |
| `publish_to_confluence` | 🚀 Auto-publish to Confluence | `false` |

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

## 🔗 Confluence Setup

1. **Create API Token**: at https://id.atlassian.net/manage-profile/security/api-tokens
2. **Get Page ID**: From URL - `https://your-domain.atlassian.net/wiki/spaces/SPACE/pages/123456789/Page+Name`
3. **Add as Secrets**: Set up the required secrets in Harness

## 💡 Best Practices

- **Template Structure**: Include name, description, type, author, version and tags
- **Run on Merge**: Setup webhooks to update docs automatically
- **Use Secrets**: Never hardcode Confluence credentials
- **Include Examples**: Make templates self-documenting with examples

## 📖 Full Example

For a complete implementation, see `templates/harness/example-usage.yaml`

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

