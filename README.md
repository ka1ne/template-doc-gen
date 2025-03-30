# 📚 Harness Template Docs Generator

> Automatically extract metadata from Harness templates and generate beautiful, searchable HTML documentation.

![Version](https://img.shields.io/badge/version-0.0.3--alpha-blue)
![Docker](https://img.shields.io/badge/docker-ready-brightgreen)

## ✨ What It Does

Transforms your Harness templates into well-organized documentation:

- 🔍 **Extracts** metadata, variables, parameters, and examples
- 🎨 **Generates** beautiful HTML documentation with modern, responsive design
- 🔄 **Integrates** with your CI/CD pipelines via Harness stage templates
- 📝 **Validates** your templates against the official Harness schema

## ✨ Documentation Features

- **Modern, Clean Design**: Inspired by Kubernetes and Go documentation
- **Responsive Layout**: Works on desktop and mobile devices
- **Syntax Highlighting**: Code examples are easy to read
- **Full-Text Search**: Find templates quickly with built-in search
- **Type Filtering**: Filter by pipeline, stage, or step group templates
- **Sidebar Navigation**: Easy navigation between different template types

## 🚀 Quick Start

### Option 1: Local Preview with Built-in Server

```bash
# Clone this repo
git clone https://github.com/your-org/harness-template-docs.git
cd harness-template-docs

# Set up configuration (optional)
cp .env.example .env
# Edit .env with your settings if needed

# Run the generator with local preview server
./generate-docs.sh
```

This will generate HTML documentation and start a local web server at http://localhost:8000 for you to browse.

### Option 2: One-Click Docker Run

```bash
# Clone this repo
git clone https://github.com/your-org/harness-template-docs.git
cd harness-template-docs

# Run the helper script (will create .env from example if needed)
./docker-run.sh
```

That's it! The script handles everything - mounting volumes, setting up environment variables, and running the container.

### Option 3: Manual Docker Run

```bash
# 1. Create and configure your .env file
cp .env.example .env
# Edit .env with your settings

# 2. Create output directory with proper permissions
mkdir -p docs/output
chmod 777 docs/output

# 3. Run with Docker
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs \
  --env-file .env ka1ne/template-doc-gen:0.0.3-alpha
```

### Option 4: Use In Harness Pipeline

Import the stage template (`templates/harness/template-doc-gen.yaml`):

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

## 🔧 Configuration (.env file)

```
# Template Processing Configuration
SOURCE_DIR=templates                 # Where to find templates
OUTPUT_DIR=docs/output               # Where to write docs
FORMAT=html                          # Output format
VERBOSE=true                         # Show detailed logs
VALIDATE_ONLY=false                  # Only validate, don't generate
```

## 📋 Template Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `docker_registry_connector` | 🔗 Docker registry connector | **Required** |
| `image_name` | 🐳 Image to use | `ka1ne/template-doc-gen:0.0.3-alpha` |
| `source_dir` | 📁 Where to find templates | `templates` |
| `output_dir` | 📂 Where to write docs | `docs/templates` |
| `format` | 📄 Output format | `html` |

## 🔄 Integrating with Documentation Systems

The generator creates static HTML documentation that can be integrated with a variety of systems:

### 🌐 Static Web Hosting

Upload the generated HTML files to any static web hosting service:

```bash
# Example: Using AWS S3
aws s3 sync docs/output/ s3://your-bucket/templates-docs/ --acl public-read

# Example: Using GitHub Pages
cp -r docs/output/* docs/
git add docs
git commit -m "Update documentation"
git push
```

### 📝 Confluence Integration

While this tool doesn't directly publish to Confluence, you can integrate the generated HTML:

1. **Copy HTML Content**: Copy the HTML content from generated files
2. **Paste into Confluence**: Use the "Insert" > "HTML" feature in Confluence's editor
3. **Automate with Scripts**: Use Confluence's REST API in your own scripts to automate publication:

```bash
# Example of using Confluence's REST API to publish content
curl -u username:api_token -X POST -H 'Content-Type: application/json' \
  -d '{"type":"page","title":"Template Documentation","space":{"key":"SPACE"},"body":{"storage":{"value":"YOUR_HTML","representation":"storage"}},"ancestors":[{"id":"123456"}]}' \
  https://your-instance.atlassian.net/wiki/rest/api/content
```

### 📄 Convert to Other Formats

Convert the HTML documentation to other formats using tools like Pandoc:

```bash
# Convert HTML to Markdown
pandoc -f html -t markdown docs/output/index.html -o templates-docs.md

# Convert HTML to PDF
pandoc -f html -t pdf docs/output/index.html -o templates-docs.pdf
```

### 🔄 CI/CD Integration

Set up automatic documentation generation on template changes:

```yaml
# Example GitHub Actions workflow
name: Generate Docs
on:
  push:
    paths:
      - 'templates/**'
jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Generate Documentation
        run: docker run -v ./templates:/app/templates -v ./docs:/app/docs ka1ne/template-doc-gen:0.0.3-alpha
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
```

## 🧪 Local Development

### Docker 

```bash
# Run with Docker
docker run -v $(pwd)/templates:/app/templates -v $(pwd)/docs:/app/docs \
  ka1ne/template-doc-gen:0.0.3-alpha --verbose
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
- **Include Examples**: Make templates self-documenting with examples

## 🛠️ Command Reference

```bash
python process_template.py --help

# Key Arguments
  --source DIR          Source directory with templates  
  --output DIR          Output directory for documentation
  --format FMT          Output format (html|markdown|json)
  --verbose             Show detailed logs
  --validate            Validate templates without generating documentation
```

## 📋 Features

- **Official Schema Validation**: Templates are validated against the official Harness schema
- **Beautiful Documentation**: Generates modern, searchable HTML documentation with syntax highlighting
- **Docker Ready**: Run in any environment with Docker support
- **Flexible Output**: Generate documentation in HTML format (with Markdown and JSON coming soon)
- **Local Preview**: Preview documentation with built-in web server

## 🧠 Pro Tips

- **Schema Validation**: Uses the official [Harness Schema](https://github.com/harness/harness-schema) to validate templates
- **Private Repos**: You can host the image in a private Docker registry
- **Run on Merge**: Set up webhooks to run docs generation when templates change
- **Browser Preview**: The `generate-docs.sh` script automatically starts a local web server

