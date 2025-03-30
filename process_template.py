import sys
import os
import yaml
import glob
import argparse
import logging
from datetime import datetime
from src.utils.template_html import generate_template_html, generate_index_html
from src.utils.confluence_publisher import publish_templates_to_confluence
import time
import requests
import jsonschema
import json

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
    
    # Add Confluence publishing options
    parser.add_argument(
        '--publish', '-p',
        action='store_true',
        help='Publish documentation to Confluence'
    )
    
    parser.add_argument(
        '--confluence-url',
        help='Confluence URL (required if --publish is specified)'
    )
    
    parser.add_argument(
        '--confluence-username',
        help='Confluence username (required if --publish is specified)'
    )
    
    parser.add_argument(
        '--confluence-token',
        help='Confluence API token (required if --publish is specified)'
    )
    
    parser.add_argument(
        '--confluence-space',
        help='Confluence space key (required if --publish is specified)'
    )
    
    parser.add_argument(
        '--confluence-parent-id',
        help='Confluence parent page ID (required if --publish is specified)'
    )
    
    return parser

def get_harness_schema(schema_type="pipeline"):
    """Fetch the Harness schema from GitHub or use cached version."""
    if schema_type not in schema_cache:
        try:
            # Map to the correct schema file
            # The Harness repo has pipeline.json, template.json, and trigger.json
            schema_file = "pipeline.json"  # Default
            
            if schema_type == "stage" or schema_type == "step":
                # Stages and steps are defined in the template.json schema
                schema_file = "template.json"
            elif schema_type == "trigger":
                schema_file = "trigger.json"
                
            # Default to v0 schema
            schema_url = f"https://raw.githubusercontent.com/harness/harness-schema/main/v0/{schema_file}"
            logger.debug(f"Fetching schema from {schema_url}")
            
            response = requests.get(schema_url)
            if response.status_code == 200:
                schema_cache[schema_type] = response.json()
                logger.debug(f"Successfully fetched {schema_type} schema using {schema_file}")
            else:
                logger.error(f"Failed to fetch schema: {response.status_code} from URL {schema_url}")
                # Use empty schema as fallback
                schema_cache[schema_type] = {}
        except Exception as e:
            logger.error(f"Error fetching schema: {e}")
            # Use empty schema as fallback
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
            
            output_filename = metadata['name'].replace(' ', '_') + '.html'
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
    
    # CSS content is the same as before
    css = """
    /* General Styles */
    body {
        font-family: Arial, sans-serif;
        line-height: 1.6;
        color: #333;
        max-width: 1200px;
        margin: 0 auto;
        padding: 20px;
    }
    
    h1, h2, h3, h4 {
        color: #2c3e50;
    }
    
    /* Header Styles */
    header {
        background-color: #f8f9fa;
        padding: 20px;
        border-radius: 5px;
        margin-bottom: 30px;
        border-bottom: 3px solid #3498db;
    }
    
    .search-container {
        margin: 20px 0;
    }
    
    #searchInput {
        width: 100%;
        padding: 10px;
        font-size: 16px;
        border: 1px solid #ddd;
        border-radius: 4px;
    }
    
    .filter-container {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        margin-bottom: 20px;
    }
    
    .filter-btn {
        padding: 8px 15px;
        background-color: #f1f1f1;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.3s;
    }
    
    .filter-btn:hover {
        background-color: #ddd;
    }
    
    .filter-btn.active {
        background-color: #3498db;
        color: white;
    }
    
    /* Template Grid */
    .templates-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 20px;
    }
    
    .template-card {
        border: 1px solid #ddd;
        border-radius: 5px;
        padding: 15px;
        transition: transform 0.3s, box-shadow 0.3s;
    }
    
    .template-card:hover {
        transform: translateY(-5px);
        box-shadow: 0 5px 15px rgba(0,0,0,0.1);
    }
    
    .template-card h2 {
        margin-top: 0;
        color: #3498db;
    }
    
    .template-type {
        display: inline-block;
        background-color: #e1f5fe;
        color: #0288d1;
        padding: 3px 8px;
        border-radius: 3px;
        font-size: 14px;
    }
    
    .template-description {
        color: #666;
        margin: 10px 0;
    }
    
    .template-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 5px;
        margin: 10px 0;
    }
    
    .tag {
        background-color: #f1f1f1;
        padding: 3px 8px;
        border-radius: 3px;
        font-size: 12px;
    }
    
    .view-btn {
        display: inline-block;
        background-color: #3498db;
        color: white;
        padding: 8px 15px;
        text-decoration: none;
        border-radius: 4px;
        margin-top: 10px;
        transition: background-color 0.3s;
    }
    
    .view-btn:hover {
        background-color: #2980b9;
    }
    
    /* Template Section Styles */
    .template-section {
        background-color: #fff;
        border: 1px solid #ddd;
        border-radius: 5px;
        padding: 20px;
        margin-bottom: 30px;
    }
    
    .template-metadata {
        display: flex;
        flex-wrap: wrap;
        gap: 20px;
        margin-bottom: 20px;
        padding-bottom: 15px;
        border-bottom: 1px solid #eee;
    }
    
    .template-metadata p {
        margin: 0;
    }
    
    .template-metadata span {
        font-weight: bold;
        color: #3498db;
    }
    
    .template-description h3,
    .template-tags h3,
    .template-parameters h3,
    .template-variables h3 {
        margin-top: 25px;
        border-bottom: 2px solid #f1f1f1;
        padding-bottom: 10px;
    }
    
    /* Table Styles */
    table {
        width: 100%;
        border-collapse: collapse;
        margin: 20px 0;
    }
    
    th, td {
        padding: 12px 15px;
        text-align: left;
        border-bottom: 1px solid #ddd;
    }
    
    th {
        background-color: #f8f9fa;
        font-weight: bold;
    }
    
    tr:hover {
        background-color: #f5f5f5;
    }
    
    /* Footer Styles */
    footer {
        margin-top: 50px;
        padding-top: 20px;
        border-top: 1px solid #eee;
        text-align: center;
        color: #777;
    }
    
    /* Responsive Adjustments */
    @media (max-width: 768px) {
        .templates-grid {
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
        }
        
        .template-metadata {
            flex-direction: column;
            gap: 10px;
        }
        
        table {
            display: block;
            overflow-x: auto;
        }
    }
    """
    
    with open(css_path, 'w') as file:
        file.write(css)

def publish_to_confluence(documentation_files, args):
    """Publish generated documentation to Confluence."""
    logger.info("Publishing documentation to Confluence...")
    
    try:
        import requests
        import json
        
        # Initialize direct REST API approach instead of using the atlassian-python-api library
        # which has compatibility issues with the 'get_current_user' method
        confluence_url = args.confluence_url.rstrip('/')
        confluence_api_url = f"{confluence_url}/wiki/rest/api"
        auth = (args.confluence_username, args.confluence_token)
        
        # Test connection with a simple request
        test_response = requests.get(
            f"{confluence_api_url}/user?accountId=current",
            auth=auth
        )
        
        if test_response.status_code != 200:
            logger.error(f"Failed to connect to Confluence: {test_response.status_code} - {test_response.text}")
            return False
        else:
            user_data = test_response.json()
            logger.debug(f"Connected to Confluence as {user_data.get('displayName', 'Unknown user')}")
        
        # Publish index page first
        index_file = os.path.join(args.output, 'index.html')
        if os.path.exists(index_file):
            logger.info(f"Publishing index page to Confluence: {index_file}")
            
            with open(index_file, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # Create page title with timestamp to avoid conflicts
            page_title = f"Harness Template Documentation - {time.strftime('%Y-%m-%d %H:%M:%S')}"
            
            # Prepare the JSON payload
            payload = {
                "type": "page",
                "title": page_title,
                "space": {"key": args.confluence_space},
                "body": {
                    "storage": {
                        "value": content,
                        "representation": "storage"
                    }
                },
                "ancestors": [{"id": args.confluence_parent_id}]
            }
            
            # Make the API call to create the page
            response = requests.post(
                f"{confluence_api_url}/content",
                json=payload,
                auth=auth,
                headers={"Content-Type": "application/json"}
            )
            
            if response.status_code == 200:
                page_data = response.json()
                logger.info(f"Published documentation index to Confluence: {page_data.get('_links', {}).get('webui', '')}")
                return True
            else:
                logger.error(f"Failed to publish to Confluence. Status code: {response.status_code}")
                logger.error(f"Response: {response.text}")
                return False
                
        else:
            logger.error(f"Index file not found: {index_file}")
            return False
            
    except ImportError as e:
        logger.error(f"Failed to import required library: {e}")
        return False
    except Exception as e:
        logger.error(f"Failed to publish to Confluence: {e}")
        return False

def extract_template_metadata(template_data):
    """Extract metadata from Harness template including variables and descriptions."""
    try:
        # For the Harness template format where template is a root key
        if 'template' in template_data and isinstance(template_data['template'], dict):
            template_obj = template_data['template']
            
            # Determine template type
            template_type = None
            if 'type' in template_obj:
                if template_obj['type'] in ['Stage', 'Pipeline', 'StepGroup']:
                    template_type = template_obj['type'].lower()
            
            # Extract variables from spec sections
            variables = {}
            parameters = {}
            
            # Try to extract variables and parameters from the spec if available
            if 'spec' in template_obj:
                spec = template_obj['spec']
                if 'serviceConfig' in spec.get('spec', {}):
                    service_config = spec['spec']['serviceConfig']
                    if 'serviceDefinition' in service_config:
                        service_def = service_config['serviceDefinition']
                        if 'spec' in service_def and 'variables' in service_def['spec']:
                            for var in service_def['spec'].get('variables', []):
                                if isinstance(var, dict) and 'name' in var:
                                    variables[var['name']] = {
                                        'description': var.get('description', ''),
                                        'type': var.get('type', 'string'),
                                        'required': var.get('required', False),
                                        'scope': 'stage'
                                    }
                
                # Extract step parameters if present
                if 'execution' in spec.get('spec', {}):
                    execution = spec['spec']['execution']
                    if 'steps' in execution:
                        for step in execution['steps']:
                            if 'step' in step and 'spec' in step['step']:
                                step_spec = step['step']['spec']
                                for param_name, param_value in step_spec.items():
                                    parameters[param_name] = {
                                        'description': f"Parameter for {step['step'].get('name', 'step')}",
                                        'type': 'boolean' if isinstance(param_value, bool) else 'string',
                                        'required': False,
                                        'default': param_value,
                                        'scope': 'step'
                                    }
            
            # Extract tags
            tags = []
            if 'tags' in template_obj and template_obj['tags']:
                if isinstance(template_obj['tags'], dict):
                    # If tags are key-value pairs
                    for tag_key, tag_value in template_obj['tags'].items():
                        tags.append(tag_key)
                        if tag_value and str(tag_value).strip():
                            tags.append(str(tag_value))
                elif isinstance(template_obj['tags'], list):
                    # If tags are a list
                    tags.extend(template_obj['tags'])
            
            metadata = {
                'name': template_obj.get('name', 'Unnamed Template'),
                'type': template_type or 'stage',  # Default to stage if unknown
                'variables': variables,
                'parameters': parameters,
                'description': template_obj.get('description', f"Harness {template_type or 'stage'} template"),
                'tags': tags,
                'author': 'Harness',
                'version': template_obj.get('versionLabel', '1.0.0'),
                'examples': []
            }
            return metadata
        
        # For older templates with direct metadata
        return {
            'name': template_data.get('name', 'Unnamed Template'),
            'type': template_data.get('type', 'unknown'),
            'variables': template_data.get('variables', {}),
            'parameters': template_data.get('parameters', {}),
            'description': template_data.get('description', ''),
            'tags': template_data.get('tags', []),
            'author': template_data.get('author', ''),
            'version': template_data.get('version', '1.0.0'),
            'examples': template_data.get('examples', [])
        }
    except Exception as e:
        logger.error(f"Error extracting metadata: {e}")
        # Return basic metadata to avoid complete failure
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
    
    # If publish flag is set, publish to Confluence
    if args.publish:
        # Check for required Confluence arguments
        missing_args = []
        if not args.confluence_url:
            missing_args.append("--confluence-url")
        if not args.confluence_username:
            missing_args.append("--confluence-username")
        if not args.confluence_token:
            missing_args.append("--confluence-token")
        if not args.confluence_space:
            missing_args.append("--confluence-space")
        if not args.confluence_parent_id:
            missing_args.append("--confluence-parent-id")
            
        if missing_args:
            logger.error(f"Missing required arguments for Confluence publishing: {', '.join(missing_args)}")
            sys.exit(1)
            
        # Publish to Confluence
        logger.info("Publishing documentation to Confluence")
        try:
            publish_to_confluence(all_metadata, args)
            logger.info("Successfully published documentation to Confluence")
        except Exception as e:
            logger.error(f"Failed to publish documentation to Confluence: {e}", exc_info=True)
            sys.exit(1)
    
    return 0

if __name__ == "__main__":
    sys.exit(main()) 