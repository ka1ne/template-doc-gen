import sys
import os
import yaml
import glob
import argparse
import logging
from datetime import datetime
from src.utils.template_html import generate_template_html, generate_index_html
import time
import requests
import jsonschema
import json
import re

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler()
    ]
)
logger = logging.getLogger('harness-docs')

# Cache for schema files
schema_cache = {}

def setup_argparse():
    """Set up command-line argument parsing."""
    parser = argparse.ArgumentParser(
        description='Generate documentation from Harness templates.',
        formatter_class=argparse.ArgumentDefaultsHelpFormatter
    )
    
    parser.add_argument(
        '--source', '-s',
        default='templates',
        help='Source directory or file containing templates'
    )
    
    parser.add_argument(
        '--output', '-o',
        default='docs/templates',
        help='Output directory for generated documentation'
    )
    
    parser.add_argument(
        '--format', '-f',
        choices=['html', 'markdown', 'json'],
        default='html',
        help='Output format for documentation'
    )
    
    parser.add_argument(
        '--validate', '-v',
        action='store_true',
        help='Validate templates without generating documentation'
    )
    
    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Enable verbose logging'
    )
    
    return parser

def get_harness_schema(schema_type="pipeline"):
    """Fetch the Harness schema from GitHub or use cached version."""
    if schema_type not in schema_cache:
        try:
            # Map schema types to their corresponding files
            schema_mapping = {
                "pipeline": "pipeline.json",
                "stage": "template.json",  # Stages are defined in template.json
                "step": "template.json",   # Steps are defined in template.json
                "stepgroup": "template.json",  # StepGroups are defined in template.json
                "trigger": "trigger.json"
            }
            
            schema_file = schema_mapping.get(schema_type.lower(), "pipeline.json")
            logger.debug(f"Using schema file {schema_file} for type {schema_type}")
                
            # Use v1 schema
            schema_url = f"https://raw.githubusercontent.com/harness/harness-schema/main/v1/{schema_file}"
            logger.debug(f"Fetching schema from {schema_url}")
            
            response = requests.get(schema_url)
            if response.status_code == 200:
                raw_schema = response.json()
                # The v1 schema has a different structure - it's directly the schema we need
                schema_cache[schema_type] = raw_schema
                logger.debug(f"Successfully fetched {schema_type} schema using {schema_file}")
            else:
                logger.error(f"Failed to fetch schema: {response.status_code} from URL {schema_url}")
                schema_cache[schema_type] = {}
        except Exception as e:
            logger.error(f"Error fetching schema: {e}")
            schema_cache[schema_type] = {}
    
    return schema_cache[schema_type]

def validate_template(template_data):
    """Validate a template using the official Harness schema."""
    try:
        # For Harness template format
        if 'template' in template_data and isinstance(template_data['template'], dict):
            template_obj = template_data['template']
            
            # Determine schema type based on template type
            schema_type = "pipeline"  # Default
            if 'type' in template_obj:
                if template_obj['type'] == "Stage":
                    schema_type = "stage"
                elif template_obj['type'] == "Pipeline":
                    schema_type = "pipeline"
                elif template_obj['type'] == "StepGroup":
                    schema_type = "step"
            
            # Get the appropriate schema
            try:
                schema = get_harness_schema(schema_type)
                
                # Validate against schema
                jsonschema.validate(template_data, schema)
                return True, f"Template is valid according to Harness {schema_type} schema"
            except jsonschema.exceptions.ValidationError as e:
                logger.debug(f"Schema validation failed, falling back to basic validation: {str(e)}")
                # Fall back to basic validation
                if 'name' not in template_obj:
                    return False, "Missing required field: name in template object"
                return True, "Basic validation passed (schema validation failed)"
            except Exception as e:
                logger.debug(f"Error during schema validation: {str(e)}")
                # If schema validation fails for any reason, fall back to basic validation
                if 'name' not in template_obj:
                    return False, "Missing required field: name in template object"
                return True, "Basic validation passed (schema validation error)"
        
        # Original validation for our custom template format
        required_fields = ['name', 'description', 'type']
        
        for field in required_fields:
            if field not in template_data:
                return False, f"Missing required field: {field}"
        
        valid_types = ['pipeline', 'stage', 'stepgroup']
        if template_data.get('type') not in valid_types:
            return False, f"Invalid template type: {template_data.get('type')}. Must be one of {valid_types}"
        
        return True, "Template is valid"
    except Exception as e:
        logger.error(f"Validation error: {str(e)}")
        return False, f"Validation error: {str(e)}"

def process_harness_template(template_path, output_dir='docs/templates', output_format='html', validate_only=False):
    """Process a Harness template file and generate documentation."""
    try:
        # Load the template
        logger.info(f"Processing template: {template_path}")
        with open(template_path, 'r') as file:
            template_data = yaml.safe_load(file)
        
        # Validate template
        is_valid, message = validate_template(template_data)
        if not is_valid:
            logger.error(f"Template validation failed for {template_path}: {message}")
            return None
        
        logger.debug(f"Template validation passed: {message}")
        
        # Extract metadata
        metadata = extract_template_metadata(template_data)
        
        if validate_only:
            logger.info(f"Template {template_path} is valid")
            return metadata
        
        # Generate documentation based on format
        if output_format == 'html':
            html = generate_template_html(metadata)
            
            # Determine output path
            type_dir = os.path.join(output_dir, metadata['type'])
            os.makedirs(type_dir, exist_ok=True)
            
            # Sanitize filename - replace any character not alphanumeric, underscore, or hyphen with underscore
            safe_name = re.sub(r'[^a-zA-Z0-9_\-]', '_', metadata['name'])
            output_filename = safe_name + '.html'
            output_path = os.path.join(type_dir, output_filename)
            
            # Write HTML to file
            with open(output_path, 'w') as file:
                file.write(html)
                
            logger.info(f"Generated documentation for {template_path} at {output_path}")
        
        # Add support for other formats here (markdown, json)
        elif output_format == 'markdown':
            logger.warning("Markdown output format not yet implemented")
        elif output_format == 'json':
            logger.warning("JSON output format not yet implemented")
            
        return metadata
        
    except yaml.YAMLError as e:
        logger.error(f"YAML parsing error in {template_path}: {e}")
        return None
    except Exception as e:
        logger.error(f"Error processing template {template_path}: {e}", exc_info=True)
        return None

def process_all_templates(templates_dir, output_dir='docs/templates', output_format='html', validate_only=False):
    """Process all template files in the given directory."""
    start_time = datetime.now()
    logger.info(f"Starting template processing from {templates_dir}")
    
    # Find all template files
    template_files = glob.glob(os.path.join(templates_dir, '**/*.yaml'), recursive=True)
    template_files.extend(glob.glob(os.path.join(templates_dir, '**/*.yml'), recursive=True))
    
    if not template_files:
        logger.warning(f"No template files found in {templates_dir}")
        return []
    
    logger.info(f"Found {len(template_files)} template files")
    
    all_metadata = []
    valid_count = 0
    error_count = 0
    
    for template_file in template_files:
        metadata = process_harness_template(
            template_file, 
            output_dir=output_dir,
            output_format=output_format,
            validate_only=validate_only
        )
        
        if metadata:
            all_metadata.append(metadata)
            valid_count += 1
        else:
            error_count += 1
    
    # Generate index page if not in validate-only mode
    if not validate_only and all_metadata and output_format == 'html':
        # Create output directory if it doesn't exist
        os.makedirs(output_dir, exist_ok=True)
        
        index_html = generate_index_html(all_metadata)
        with open(os.path.join(output_dir, 'index.html'), 'w') as file:
            file.write(index_html)
        
        # Generate CSS file
        generate_css(output_dir)
        
        logger.info(f"Generated index page with {len(all_metadata)} templates")
    
    end_time = datetime.now()
    duration = (end_time - start_time).total_seconds()
    
    logger.info(f"Processing completed in {duration:.2f} seconds")
    logger.info(f"Templates processed: {valid_count} successful, {error_count} failed")
    
    return all_metadata

def generate_css(output_dir):
    """Generate CSS file for the documentation."""
    css_path = os.path.join(output_dir, 'styles.css')
    
    # Check if CSS already exists and is up to date
    if os.path.exists(css_path):
        logger.debug("CSS file already exists, skipping generation")
        return
    
    logger.info(f"Generating CSS file at {css_path}")
    
    # CSS content with modern, clean styling
    css = """
    /* Base Styles */
    :root {
        --sidebar-width: 260px;
        --primary-color: #5c6bc0;
        --primary-light: #8e99f3;
        --primary-dark: #26418f;
        --accent-color: #26c6da;
        --text-color: #263238;
        --light-gray: #f5f7fa;
        --mid-gray: #e1e4e8;
        --dark-gray: #546e7a;
        --border-color: #dde1e5;
        --code-bg: #f6f8fa;
        --pipeline-color: #42a5f5;
        --stage-color: #66bb6a;
        --stepgroup-color: #ffa726;
        --gradient: linear-gradient(135deg, var(--primary-color), var(--accent-color));
    }

    * {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }

    body {
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', sans-serif;
        line-height: 1.6;
        color: var(--text-color);
        display: flex;
        min-height: 100vh;
        background-color: #ffffff;
    }

    a {
        color: var(--primary-color);
        text-decoration: none;
        transition: color 0.2s ease;
    }

    a:hover {
        color: var(--primary-dark);
        text-decoration: none;
    }

    h1, h2, h3, h4 {
        font-weight: 600;
        line-height: 1.3;
        margin-bottom: 1rem;
        color: var(--text-color);
    }

    h1 {
        font-size: 2.2rem;
        margin-bottom: 1.5rem;
        background: var(--gradient);
        -webkit-background-clip: text;
        background-clip: text;
        color: transparent;
    }

    h2 {
        font-size: 1.5rem;
        margin-top: 2rem;
        margin-bottom: 1rem;
        padding-bottom: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        position: relative;
    }
    
    h2::after {
        content: '';
        position: absolute;
        bottom: -1px;
        left: 0;
        width: 60px;
        height: 3px;
        background: var(--gradient);
        border-radius: 2px;
    }

    h3 {
        font-size: 1.25rem;
        margin-top: 1.5rem;
    }

    p {
        margin-bottom: 1rem;
    }

    code {
        font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
        font-size: 0.9em;
        padding: 0.2em 0.4em;
        background-color: var(--code-bg);
        border-radius: 3px;
    }

    pre {
        background-color: var(--code-bg);
        padding: 1rem;
        border-radius: 8px;
        overflow-x: auto;
        margin-bottom: 1.5rem;
        border: 1px solid var(--border-color);
        box-shadow: 0 2px 6px rgba(0,0,0,0.05);
    }

    pre code {
        padding: 0;
        background-color: transparent;
    }

    /* Layout */
    .sidebar {
        width: var(--sidebar-width);
        background: var(--gradient);
        position: fixed;
        height: 100vh;
        overflow-y: auto;
        box-shadow: 2px 0 10px rgba(0,0,0,0.1);
        z-index: 10;
    }

    .content {
        flex: 1;
        margin-left: var(--sidebar-width);
        padding: 0;
        max-width: 100%;
    }

    header {
        background-color: #ffffff;
        padding: 2rem 2.5rem 1.5rem;
        border-bottom: 1px solid var(--border-color);
        box-shadow: 0 2px 10px rgba(0,0,0,0.03);
    }

    main {
        padding: 2.5rem;
        max-width: 1200px;
        margin: 0 auto;
    }

    footer {
        margin-top: 4rem;
        padding: 1.5rem 2rem;
        text-align: center;
        color: var(--dark-gray);
        border-top: 1px solid var(--border-color);
        font-size: 0.9rem;
    }

    /* Sidebar Styles */
    .sidebar-header {
        padding: 1.8rem 1.5rem;
        border-bottom: 1px solid rgba(255,255,255,0.1);
    }

    .sidebar-header h3 {
        color: white;
        margin: 0;
        font-weight: 600;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .sidebar-header h3 i {
        font-size: 1.2rem;
    }

    .sidebar-menu {
        padding: 1.2rem 0;
    }

    .sidebar-menu a {
        display: flex;
        align-items: center;
        gap: 0.8rem;
        padding: 0.85rem 1.5rem;
        color: rgba(255,255,255,0.9);
        transition: all 0.2s ease;
        border-left: 3px solid transparent;
        font-weight: 500;
    }

    .sidebar-menu a i {
        font-size: 1.1rem;
        width: 1.5rem;
        text-align: center;
    }

    .sidebar-menu a:hover {
        background-color: rgba(255,255,255,0.1);
        color: white;
    }

    .sidebar-menu a.active {
        background-color: rgba(255,255,255,0.15);
        border-left-color: white;
        color: white;
    }

    /* Header Styles */
    .template-header {
        display: flex;
        align-items: center;
        margin-bottom: 0.8rem;
        gap: 1rem;
    }

    .template-header h1 {
        margin: 0;
    }

    .breadcrumbs {
        color: var(--dark-gray);
        font-size: 0.9rem;
        margin-bottom: 1.5rem;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .breadcrumbs a {
        font-weight: 500;
    }

    .breadcrumbs span {
        color: var(--dark-gray);
    }

    .header-actions {
        margin-top: 1.5rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        align-items: center;
    }

    .search-container {
        flex: 1;
        max-width: 500px;
        position: relative;
    }

    .search-container::before {
        content: "\\f002";
        font-family: "Font Awesome 6 Free";
        font-weight: 900;
        position: absolute;
        left: 1rem;
        top: 50%;
        transform: translateY(-50%);
        color: var(--dark-gray);
        font-size: 0.9rem;
    }

    #searchInput {
        width: 100%;
        padding: 0.75rem 1rem 0.75rem 2.5rem;
        font-size: 1rem;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        transition: all 0.2s ease;
        box-shadow: 0 2px 5px rgba(0,0,0,0.02);
    }

    #searchInput:focus {
        outline: none;
        border-color: var(--primary-color);
        box-shadow: 0 0 0 3px rgba(92,107,192,0.2);
    }

    .filter-container {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }

    .filter-btn {
        padding: 0.6rem 1.2rem;
        background-color: var(--light-gray);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        cursor: pointer;
        font-size: 0.9rem;
        transition: all 0.2s;
        font-weight: 500;
    }

    .filter-btn:hover {
        background-color: var(--mid-gray);
    }

    .filter-btn.active {
        background: var(--gradient);
        color: white;
        border-color: var(--primary-color);
        box-shadow: 0 2px 5px rgba(92,107,192,0.3);
    }

    /* Template Type Badges */
    .template-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.4rem 0.9rem;
        border-radius: 20px;
        font-size: 0.8rem;
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.03em;
    }

    .template-badge::before {
        font-family: "Font Awesome 6 Free";
        font-weight: 900;
    }

    .type-pipeline {
        background-color: rgba(66, 165, 245, 0.15);
        color: var(--pipeline-color);
    }
    
    .type-pipeline::before {
        content: "\\f085";
    }

    .type-stage {
        background-color: rgba(102, 187, 106, 0.15);
        color: var(--stage-color);
    }
    
    .type-stage::before {
        content: "\\f5fd";
    }

    .type-stepgroup {
        background-color: rgba(255, 167, 38, 0.15);
        color: var(--stepgroup-color);
    }
    
    .type-stepgroup::before {
        content: "\\f0ae";
    }

    /* Template Count */
    .template-count {
        margin-bottom: 1.8rem;
        font-size: 0.95rem;
        color: var(--dark-gray);
        background-color: var(--light-gray);
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem;
        border-radius: 20px;
        font-weight: 500;
    }

    .template-count i {
        color: var(--primary-color);
    }

    /* Template Grid */
    .templates-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
        gap: 1.8rem;
    }

    .template-card {
        background-color: white;
        border-radius: 12px;
        padding: 1.8rem;
        transition: all 0.3s ease;
        display: flex;
        flex-direction: column;
        height: 100%;
        border: 1px solid var(--border-color);
        box-shadow: 0 4px 12px rgba(0,0,0,0.03);
        position: relative;
        overflow: hidden;
    }
    
    .template-card::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 5px;
        background: var(--gradient);
        opacity: 0;
        transition: opacity 0.3s ease;
    }

    .template-card:hover {
        box-shadow: 0 12px 20px rgba(0,0,0,0.06);
        transform: translateY(-4px);
    }
    
    .template-card:hover::before {
        opacity: 1;
    }

    .card-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 1.2rem;
    }

    .template-card h2 {
        font-size: 1.3rem;
        margin: 0;
        border: none;
        padding: 0;
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }
    
    .template-card h2::after {
        display: none;
    }
    
    .template-card h2 i {
        color: var(--primary-color);
        font-size: 1.1rem;
    }

    .template-description {
        color: var(--dark-gray);
        margin-bottom: 1.5rem;
        flex-grow: 1;
        line-height: 1.5;
    }

    .template-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        margin-bottom: 1.5rem;
    }

    .tag {
        display: inline-flex;
        align-items: center;
        gap: 0.4rem;
        background-color: var(--light-gray);
        color: var(--dark-gray);
        padding: 0.3rem 0.7rem;
        border-radius: 20px;
        font-size: 0.8rem;
        font-weight: 500;
        transition: all 0.2s ease;
        border: 1px solid transparent;
    }
    
    .tag i {
        color: var(--primary-color);
    }
    
    .tag:hover {
        background-color: var(--primary-light);
        color: white;
        border-color: var(--primary-light);
    }
    
    .tag:hover i {
        color: white;
    }

    .view-btn {
        display: inline-flex;
        align-items: center;
        background: var(--gradient);
        color: white;
        padding: 0.7rem 1.2rem;
        border-radius: 8px;
        transition: all 0.2s ease;
        font-weight: 500;
        box-shadow: 0 2px 5px rgba(92,107,192,0.3);
        position: relative;
        overflow: hidden;
        align-self: flex-start;
    }
    
    .view-btn::after {
        content: "\\f054";
        font-family: "Font Awesome 6 Free";
        font-weight: 900;
        margin-left: 0.5rem;
        font-size: 0.8rem;
        transition: transform 0.2s ease;
    }

    .view-btn:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(92,107,192,0.4);
    }
    
    .view-btn:hover::after {
        transform: translateX(3px);
    }

    /* Template Section Styles */
    .template-section {
        background-color: white;
        border-radius: 12px;
        padding: 2rem;
        box-shadow: 0 5px 15px rgba(0,0,0,0.05);
        border: 1px solid var(--border-color);
    }

    .template-metadata {
        display: flex;
        flex-wrap: wrap;
        gap: 2rem;
        margin-bottom: 2.5rem;
    }

    .metadata-item {
        display: flex;
        flex-direction: column;
        background-color: var(--light-gray);
        padding: 1rem 1.5rem;
        border-radius: 8px;
        min-width: 140px;
        transition: transform 0.2s ease;
    }
    
    .metadata-item:hover {
        transform: translateY(-2px);
    }

    .metadata-label {
        font-size: 0.8rem;
        text-transform: uppercase;
        color: var(--dark-gray);
        margin-bottom: 0.4rem;
        letter-spacing: 0.05em;
    }

    .metadata-value {
        font-weight: 600;
        font-size: 1.1rem;
        color: var(--primary-color);
    }

    /* Table Styles */
    table {
        width: 100%;
        border-collapse: separate;
        border-spacing: 0;
        margin: 1.5rem 0;
        font-size: 0.95rem;
        border-radius: 8px;
        overflow: hidden;
        box-shadow: 0 2px 8px rgba(0,0,0,0.05);
        border: 1px solid var(--border-color);
    }

    thead {
        background-color: var(--light-gray);
    }

    th, td {
        padding: 0.9rem 1rem;
        text-align: left;
        border-bottom: 1px solid var(--border-color);
    }
    
    th {
        font-weight: 600;
        color: var(--text-color);
        font-size: 0.9rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }
    
    tr:last-child td {
        border-bottom: none;
    }

    tr:hover td {
        background-color: rgba(92,107,192,0.05);
    }

    .param-name, .var-name {
        font-weight: 600;
        color: var(--primary-color);
        font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
        font-size: 0.9em;
    }

    /* Empty state */
    .empty-state {
        background-color: var(--light-gray);
        padding: 1rem 1.5rem;
        border-radius: 8px;
        color: var(--dark-gray);
        display: flex;
        align-items: center;
        gap: 0.8rem;
        margin: 1rem 0;
        font-size: 0.95rem;
        border-left: 4px solid var(--primary-color);
    }
    
    .empty-state i {
        color: var(--primary-color);
        font-size: 1.2rem;
    }

    /* Required Badge */
    .required {
        background-color: rgba(239, 68, 68, 0.1);
        color: #ef4444;
        padding: 0.2rem 0.5rem;
        border-radius: 4px;
        font-size: 0.75rem;
        font-weight: 600;
    }

    /* Examples Styles */
    .examples-container {
        margin-top: 2rem;
        display: flex;
        flex-direction: column;
        gap: 2rem;
    }

    .example {
        background-color: var(--light-gray);
        border-radius: 8px;
        padding: 1.5rem;
        border-left: 4px solid var(--primary-color);
    }
    
    .example h3 {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--primary-color);
        margin-top: 0;
        margin-bottom: 1rem;
        font-size: 1.1rem;
    }

    /* Responsive Adjustments */
    @media (max-width: 992px) {
        .sidebar {
            width: 220px;
        }
        .content {
            margin-left: 220px;
        }
    }

    @media (max-width: 768px) {
        body {
            flex-direction: column;
        }
        .sidebar {
            width: 100%;
            position: static;
            height: auto;
            max-height: 300px;
        }
        .content {
            margin-left: 0;
        }
        .templates-grid {
            grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        }
        .template-metadata {
            flex-direction: column;
            gap: 1rem;
        }
        header, main {
            padding: 1.5rem;
        }
        h1 {
            font-size: 1.8rem;
        }
    }
    """
    
    with open(css_path, 'w') as file:
        file.write(css)

def extract_template_metadata(template_data, schema=None):
    """Extract metadata from Harness template using schema-driven approach."""
    try:
        # Get the appropriate schema if not provided
        if schema is None:
            template_obj = template_data.get('template', template_data)
            schema_type = "pipeline"  # Default
            if 'type' in template_obj:
                if template_obj['type'] == "Stage":
                    schema_type = "stage"
                elif template_obj['type'] == "Pipeline":
                    schema_type = "pipeline"
                elif template_obj['type'] == "StepGroup":
                    schema_type = "step"
            schema = get_harness_schema(schema_type)

        # Initialize metadata structure
        metadata = {
            'name': 'Unnamed Template',
            'type': 'unknown',
            'variables': {},
            'parameters': {},
            'description': '',
            'tags': [],
            'author': 'Harness',
            'version': '1.0.0',
            'examples': []
        }

        # Extract template object (handle both formats)
        template_obj = template_data.get('template', template_data)
        
        # Use jsonschema to validate and extract data
        try:
            # First validate against the schema
            jsonschema.validate(template_obj, schema)
            
            # Get the properties from the schema
            properties = schema.get('properties', {})
            
            # Extract fields based on schema properties
            for field, field_schema in properties.items():
                if field in template_obj:
                    value = template_obj[field]
                    field_type = field_schema.get('type', 'string')
                    
                    # Map schema fields to metadata
                    if field == 'name':
                        metadata['name'] = value
                    elif field == 'type':
                        metadata['type'] = value.lower()
                    elif field == 'description':
                        metadata['description'] = value
                    elif field == 'version':
                        metadata['version'] = str(value)
                    elif field == 'tags':
                        metadata['tags'] = value if isinstance(value, list) else []
                    elif field == 'spec':
                        # Extract spec metadata using the spec schema
                        spec_schema = field_schema
                        spec_metadata = extract_spec_metadata(value, spec_schema)
                        metadata['variables'].update(spec_metadata.get('variables', {}))
                        metadata['parameters'].update(spec_metadata.get('parameters', {}))

        except jsonschema.exceptions.ValidationError as e:
            logger.warning(f"Schema validation warning: {e}")
            # Continue with basic extraction if validation fails
            metadata['name'] = template_obj.get('name', 'Unnamed Template')
            metadata['type'] = template_obj.get('type', 'unknown').lower()
            metadata['description'] = template_obj.get('description', '')
            metadata['version'] = str(template_obj.get('version', '1.0.0'))
            metadata['tags'] = template_obj.get('tags', [])

        return metadata

    except Exception as e:
        logger.error(f"Error extracting metadata: {e}")
        return {
            'name': template_data.get('name', 'Unnamed Template'),
            'type': 'unknown',
            'variables': {},
            'parameters': {},
            'description': 'Error extracting template metadata',
            'tags': [],
            'author': '',
            'version': '1.0.0',
            'examples': []
        }

def extract_spec_metadata(spec_data, schema):
    """Extract metadata from spec section using schema."""
    metadata = {'variables': {}, 'parameters': {}}
    
    try:
        # Get the spec properties from the schema
        spec_properties = schema.get('properties', {})
        
        # Extract variables if present in schema
        if 'variables' in spec_properties:
            variables_schema = spec_properties['variables']
            if 'items' in variables_schema:
                item_schema = variables_schema['items']
                if 'properties' in item_schema:
                    for var in spec_data.get('variables', []):
                        if isinstance(var, dict) and 'name' in var:
                            try:
                                # Validate variable against schema
                                jsonschema.validate(var, item_schema)
                                metadata['variables'][var['name']] = {
                                    'description': var.get('description', ''),
                                    'type': var.get('type', 'string'),
                                    'required': var.get('required', False),
                                    'scope': 'template'
                                }
                            except jsonschema.exceptions.ValidationError:
                                continue

        # Extract parameters if present in schema
        if 'parameters' in spec_properties:
            params_schema = spec_properties['parameters']
            if 'items' in params_schema:
                item_schema = params_schema['items']
                if 'properties' in item_schema:
                    for param in spec_data.get('parameters', []):
                        if isinstance(param, dict) and 'name' in param:
                            try:
                                # Validate parameter against schema
                                jsonschema.validate(param, item_schema)
                                metadata['parameters'][param['name']] = {
                                    'description': param.get('description', ''),
                                    'type': param.get('type', 'string'),
                                    'required': param.get('required', False),
                                    'default': param.get('default'),
                                    'scope': 'template'
                                }
                            except jsonschema.exceptions.ValidationError:
                                continue

        # Extract service variables if present in schema
        if 'serviceConfig' in spec_properties:
            service_schema = spec_properties['serviceConfig']
            if 'properties' in service_schema:
                service_props = service_schema['properties']
                if 'serviceDefinition' in service_props:
                    service_def_schema = service_props['serviceDefinition']
                    if 'properties' in service_def_schema:
                        spec_schema = service_def_schema['properties'].get('spec', {})
                        if 'properties' in spec_schema:
                            variables_schema = spec_schema['properties'].get('variables', {})
                            if 'items' in variables_schema:
                                item_schema = variables_schema['items']
                                if 'properties' in item_schema:
                                    service_config = spec_data.get('serviceConfig', {})
                                    service_def = service_config.get('serviceDefinition', {})
                                    for var in service_def.get('spec', {}).get('variables', []):
                                        if isinstance(var, dict) and 'name' in var:
                                            try:
                                                # Validate service variable against schema
                                                jsonschema.validate(var, item_schema)
                                                metadata['variables'][var['name']] = {
                                                    'description': var.get('description', ''),
                                                    'type': var.get('type', 'string'),
                                                    'required': var.get('required', False),
                                                    'scope': 'service'
                                                }
                                            except jsonschema.exceptions.ValidationError:
                                                continue

    except Exception as e:
        logger.error(f"Error extracting spec metadata: {e}")
    
    return metadata

def main():
    """Main function to handle command-line invocation."""
    parser = setup_argparse()
    args = parser.parse_args()
    
    # Configure logging level
    if args.verbose:
        logger.setLevel(logging.DEBUG)
        
    logger.debug(f"Arguments: {args}")
    
    # Process templates
    all_metadata = process_all_templates(
        args.source,
        output_dir=args.output, 
        output_format=args.format,
        validate_only=args.validate
    )
    
    logger.info(f"Successfully generated documentation for {len(all_metadata)} templates")
    logger.info(f"Output directory: {args.output}")
    
    return 0

if __name__ == "__main__":
    sys.exit(main()) 