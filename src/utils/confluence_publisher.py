import os
import logging
import base64
from typing import List, Dict, Any, Optional
from atlassian import Confluence
from datetime import datetime

logger = logging.getLogger('confluence-publisher')

def publish_templates_to_confluence(
    templates_metadata: List[Dict[str, Any]],
    docs_dir: str,
    confluence_url: str,
    confluence_username: str,
    confluence_api_token: str,
    space_key: str,
    parent_page_id: str
) -> None:
    """
    Publish template documentation to Confluence
    
    Args:
        templates_metadata: List of template metadata dictionaries
        docs_dir: Directory containing generated HTML documentation
        confluence_url: Confluence instance URL
        confluence_username: Confluence username
        confluence_api_token: Confluence API token
        space_key: Confluence space key
        parent_page_id: Confluence parent page ID
    """
    # Initialize Confluence client
    try:
        confluence = Confluence(
            url=confluence_url,
            username=confluence_username,
            password=confluence_api_token,
            cloud=True  # Set to False for server instances
        )
        logger.info(f"Connected to Confluence at {confluence_url}")
    except Exception as e:
        logger.error(f"Failed to connect to Confluence: {e}")
        raise
    
    # Create or update overview page
    overview_page_title = "Harness Templates Documentation"
    overview_page = _find_page(confluence, space_key, overview_page_title, parent_page_id)
    
    if overview_page:
        logger.info(f"Found existing overview page: {overview_page['id']}")
        overview_page_id = overview_page['id']
    else:
        logger.info(f"Creating new overview page under parent {parent_page_id}")
        # Create overview page
        overview_content = _generate_overview_content(templates_metadata)
        overview_page_id = confluence.create_page(
            space=space_key,
            title=overview_page_title,
            body=overview_content,
            parent_id=parent_page_id,
            representation="storage"
        )['id']
        logger.info(f"Created new overview page: {overview_page_id}")
    
    # Create or update template category pages
    categories = {
        "pipeline": "Pipeline Templates",
        "stage": "Stage Templates",
        "stepgroup": "Step Group Templates"
    }
    
    category_page_ids = {}
    
    for category_key, category_title in categories.items():
        # Check if category page exists
        category_page = _find_page(confluence, space_key, category_title, overview_page_id)
        
        if category_page:
            logger.info(f"Found existing category page for {category_key}: {category_page['id']}")
            category_page_id = category_page['id']
        else:
            # Create category page
            logger.info(f"Creating new category page for {category_key}")
            category_content = f"<p>Documentation for {category_title}.</p>"
            category_page_id = confluence.create_page(
                space=space_key,
                title=category_title,
                body=category_content,
                parent_id=overview_page_id,
                representation="storage"
            )['id']
            logger.info(f"Created new category page: {category_page_id}")
        
        category_page_ids[category_key] = category_page_id
    
    # Process each template
    for metadata in templates_metadata:
        template_type = metadata['type']
        template_name = metadata['name']
        
        if template_type not in category_page_ids:
            logger.warning(f"Unknown template type {template_type} for {template_name}, skipping")
            continue
        
        parent_id = category_page_ids[template_type]
        
        # Check if page already exists
        page = _find_page(confluence, space_key, template_name, parent_id)
        
        # Generate content for the template
        template_content = _generate_template_content(metadata, docs_dir)
        
        if page:
            # Update existing page
            logger.info(f"Updating existing page for {template_name}: {page['id']}")
            confluence.update_page(
                page_id=page['id'],
                title=template_name,
                body=template_content,
                parent_id=parent_id,
                representation="storage"
            )
        else:
            # Create new page
            logger.info(f"Creating new page for {template_name}")
            confluence.create_page(
                space=space_key,
                title=template_name,
                body=template_content,
                parent_id=parent_id,
                representation="storage"
            )
    
    # Update overview page with latest template count
    overview_content = _generate_overview_content(templates_metadata)
    confluence.update_page(
        page_id=overview_page_id,
        title=overview_page_title,
        body=overview_content,
        parent_id=parent_page_id,
        representation="storage"
    )
    
    logger.info(f"Successfully published {len(templates_metadata)} templates to Confluence")

def _find_page(confluence: Confluence, space_key: str, title: str, parent_id: Optional[str] = None) -> Optional[Dict[str, Any]]:
    """Find a Confluence page by title and parent ID"""
    try:
        # Try to find the page
        cql = f'type=page AND space="{space_key}" AND title="{title}"'
        if parent_id:
            cql += f' AND parent="{parent_id}"'
        
        results = confluence.cql(cql, limit=1)
        
        if results and results['results'] and len(results['results']) > 0:
            page_id = results['results'][0]['content']['id']
            return confluence.get_page_by_id(page_id)
        return None
    except Exception as e:
        logger.error(f"Error finding page {title}: {e}")
        return None

def _generate_overview_content(templates_metadata: List[Dict[str, Any]]) -> str:
    """Generate content for the overview page"""
    # Count templates by type
    type_counts = {}
    for metadata in templates_metadata:
        template_type = metadata['type']
        type_counts[template_type] = type_counts.get(template_type, 0) + 1
    
    # Generate overview content
    content = f"""
    <p>This space contains automatically generated documentation for Harness templates.</p>
    
    <h2>Template Statistics</h2>
    <ul>
        <li>Total Templates: {len(templates_metadata)}</li>
    """
    
    for template_type, count in type_counts.items():
        content += f"<li>{template_type.capitalize()} Templates: {count}</li>"
    
    content += "</ul>"
    
    content += f"""
    <h2>Last Updated</h2>
    <p>This documentation was last updated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}.</p>
    
    <h2>Template Categories</h2>
    <ul>
        <li><a href="Pipeline Templates">Pipeline Templates</a></li>
        <li><a href="Stage Templates">Stage Templates</a></li>
        <li><a href="Step Group Templates">Step Group Templates</a></li>
    </ul>
    """
    
    return content

def _generate_template_content(metadata: Dict[str, Any], docs_dir: str) -> str:
    """Generate Confluence content for a template"""
    template_name = metadata['name']
    template_type = metadata['type']
    
    # Load HTML content from generated docs
    html_file = os.path.join(docs_dir, template_type, f"{template_name.replace(' ', '_')}.html")
    
    try:
        with open(html_file, 'r') as f:
            html_content = f.read()
    except FileNotFoundError:
        logger.warning(f"HTML file not found for {template_name}: {html_file}")
        html_content = ""
    
    # Convert HTML to Confluence storage format
    # For simplicity, we'll just extract the main content sections
    # In a production environment, you might want to use a proper HTML-to-storage converter
    
    content = f"""
    <h1>{metadata['name']}</h1>
    
    <div class="template-metadata">
        <p><strong>Type:</strong> {metadata['type']}</p>
        <p><strong>Version:</strong> {metadata['version']}</p>
        <p><strong>Author:</strong> {metadata['author']}</p>
    </div>
    
    <h2>Description</h2>
    <p>{metadata['description']}</p>
    """
    
    # Add tags
    if metadata.get('tags'):
        content += "<h2>Tags</h2><ul>"
        for tag in metadata['tags']:
            content += f"<li>{tag}</li>"
        content += "</ul>"
    
    # Add parameters table
    if metadata.get('parameters'):
        content += """
        <h2>Parameters</h2>
        <table>
            <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Type</th>
                <th>Required</th>
                <th>Default</th>
                <th>Scope</th>
            </tr>
        """
        
        for param_name, param_data in metadata['parameters'].items():
            content += f"""
            <tr>
                <td>{param_name}</td>
                <td>{param_data['description']}</td>
                <td>{param_data['type']}</td>
                <td>{'Yes' if param_data['required'] else 'No'}</td>
                <td>{param_data['default']}</td>
                <td>{param_data['scope']}</td>
            </tr>
            """
        
        content += "</table>"
    
    # Add variables table
    if metadata.get('variables'):
        content += """
        <h2>Variables</h2>
        <table>
            <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Type</th>
                <th>Required</th>
                <th>Scope</th>
            </tr>
        """
        
        for var_name, var_data in metadata['variables'].items():
            content += f"""
            <tr>
                <td>{var_name}</td>
                <td>{var_data['description']}</td>
                <td>{var_data['type']}</td>
                <td>{'Yes' if var_data['required'] else 'No'}</td>
                <td>{var_data['scope']}</td>
            </tr>
            """
        
        content += "</table>"
    
    # Add examples
    if metadata.get('examples'):
        content += "<h2>Examples</h2>"
        
        for i, example in enumerate(metadata['examples']):
            content += f"""
            <h3>Example {i+1}</h3>
            <div class="code panel pdl">
                <div class="codeContent panelContent pdl">
                    <pre class="brush: yaml; gutter: false; theme: Confluence">
                        {example}
                    </pre>
                </div>
            </div>
            """
    
    # Add footer
    content += f"""
    <hr />
    <p><em>This documentation was automatically generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}.</em></p>
    """
    
    return content 