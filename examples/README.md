# Harness Template Documentation Generator Pipeline

Harness pipeline and step templates for generating documentation from Harness templates.

## Features

- Automated HTML documentation generation from templates
- Template hash-based caching system
- Runs on any Kubernetes cluster
- Conditional step execution
- Multiple trigger types: Git, manual, and scheduled

## Files

- `template-docs-pipeline.yaml` - Main pipeline with Kubernetes execution
- `harness-step.yaml` - Basic step template 
- `harness-step-cached.yaml` - Step template with caching
- `local-validation-pipeline.yaml` - Local validation without Kubernetes

## Setup

### Main Pipeline

1. Import `template-docs-pipeline.yaml` to project
2. Configure connectors:
   - Git connector
   - Docker registry connector
   - Kubernetes cluster connector
   - Deployment connector (if needed)

### Local Validation Pipeline

For environments without Kubernetes:
1. Import `local-validation-pipeline.yaml`
2. Configure variables:
   - `SOURCE_DIR`: Templates location
   - `OUTPUT_DIR`: Documentation output
   - `FORCE_REGEN`: Cache override flag

### Kubernetes Requirements

- Generate Documentation Stage:
  - CPU: 0.5 (request) / 1 (limit)
  - Memory: 1Gi (request) / 2Gi (limit)
- Deploy Documentation Stage:
  - CPU: 0.2 (request) / 0.5 (limit)
  - Memory: 512Mi (request) / 1Gi (limit)

### Repository Structure

```
/
├── templates/           # Template YAML files
│   ├── pipeline/        # Pipeline templates
│   ├── step/            # Step templates
│   └── ...              # Other template types
├── docs/                # Generated output
```

## Configuration

### Triggers

1. **Git Push**: Executes on template updates
2. **Manual**: On-demand execution with force option
3. **Weekly**: Sunday midnight UTC execution

### Variables

- `DOCS_TARGET_DIR`: Documentation output location (default: `docs/generated`)
- `FORCE_REGEN`: Override cache validation (default: `false`)
- `DEPLOYMENT_TYPE`: Deployment method (`git`, `artifact`, or `none`)

### Deployment Options

1. **Git (`git`)**: 
   - Pushes to Git repository
   - Compatible with GitHub/GitLab Pages
   - Configurable branch and destination

2. **Artifact (`artifact`)**: 
   - Packages as tarball
   - For custom deployment workflows
   - Repository-agnostic

3. **None (`none`)**: 
   - Workspace-only availability
   - For verification or manual deployment

## Advanced Configuration

### Caching

- Template hash generation for change detection
- Output preservation between runs
- Configurable invalidation with `FORCE_REGEN`

### Kubernetes Customization

- Adjustable CPU/memory limits
- Configurable namespace (`harness-docs` default)
- Node selector and tolerance support

## Troubleshooting

1. Verify template directory exists
2. Check template syntax with `tempdocs validate`
3. Review pipeline logs
4. Check Kubernetes pod events
5. Use `FORCE_REGEN=true` to bypass cache 