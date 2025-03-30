import sys
import yaml
import os
import logging

# Configure logging
logger = logging.getLogger('harness-docs')

def load_harness_template(file_path):
    """Load and parse a Harness template YAML file."""
    try:
        with open(file_path, 'r') as file:
            return yaml.safe_load(file)
    except yaml.YAMLError as e:
        logger.error(f"YAML parsing error in {file_path}: {e}")
        raise
    except Exception as e:
        logger.error(f"Error loading template {file_path}: {e}")
        raise
        
def extract_template_metadata(template_data):
    """Extract metadata from Harness template including variables and descriptions."""
    try:
        metadata = {
            'name': template_data.get('name', 'Unnamed Template'),
            'type': determine_template_type(template_data),
            'variables': extract_variables(template_data),
            'parameters': extract_parameters(template_data),
            'description': template_data.get('description', ''),
            'tags': template_data.get('tags', []),
            'author': template_data.get('author', ''),
            'version': template_data.get('version', '1.0.0'),
            'examples': extract_usage_examples(template_data),
            'steps': extract_steps(template_data)
        }
        return metadata
    except Exception as e:
        logger.error(f"Error extracting metadata: {e}")
        raise
    
def determine_template_type(data):
    """Determine if this is a pipeline, stage, or stepgroup template."""
    if 'pipeline' in data:
        return 'pipeline'
    elif 'stage' in data:
        return 'stage'
    elif 'steps' in data:
        return 'stepgroup'
    
    # If type is explicitly defined, use that
    if 'type' in data:
        if data['type'] in ['pipeline', 'stage', 'stepgroup']:
            return data['type']
    
    logger.warning("Could not determine template type, defaulting to 'unknown'")
    return 'unknown'
    
def extract_variables(data):
    """Extract variables and their descriptions from template."""
    variables = {}
    
    try:
        # Handle pipeline variables
        if 'pipeline' in data and 'variables' in data['pipeline']:
            for var_name, var_data in data['pipeline']['variables'].items():
                variables[var_name] = {
                    'description': var_data.get('description', ''),
                    'type': var_data.get('type', 'string'),
                    'required': var_data.get('required', False),
                    'scope': 'pipeline'
                }
                
        # Handle stage variables
        if 'stage' in data and 'variables' in data['stage']:
            for var_name, var_data in data['stage']['variables'].items():
                variables[var_name] = {
                    'description': var_data.get('description', ''),
                    'type': var_data.get('type', 'string'),
                    'required': var_data.get('required', False),
                    'scope': 'stage'
                }
        
        # Handle stepgroup variables
        if 'steps' in data and 'variables' in data:
            for var_name, var_data in data['variables'].items():
                variables[var_name] = {
                    'description': var_data.get('description', ''),
                    'type': var_data.get('type', 'string'),
                    'required': var_data.get('required', False),
                    'scope': 'stepgroup'
                }
    except Exception as e:
        logger.error(f"Error extracting variables: {e}")
    
    return variables

def extract_parameters(data):
    """Extract parameters and their descriptions from template."""
    parameters = {}
    
    # Handle pipeline parameters
    if 'pipeline' in data and 'parameters' in data['pipeline']:
        for param_name, param_data in data['pipeline']['parameters'].items():
            parameters[param_name] = {
                'description': param_data.get('description', ''),
                'type': param_data.get('type', 'string'),
                'required': param_data.get('required', False),
                'default': param_data.get('default', ''),
                'scope': 'pipeline'
            }
    
    # Handle stage parameters
    if 'stage' in data and 'parameters' in data['stage']:
        for param_name, param_data in data['stage']['parameters'].items():
            parameters[param_name] = {
                'description': param_data.get('description', ''),
                'type': param_data.get('type', 'string'),
                'required': param_data.get('required', False),
                'default': param_data.get('default', ''),
                'scope': 'stage'
            }
    
    # Handle stepgroup parameters
    if 'steps' in data and 'parameters' in data:
        for param_name, param_data in data['parameters'].items():
            parameters[param_name] = {
                'description': param_data.get('description', ''),
                'type': param_data.get('type', 'string'),
                'required': param_data.get('required', False),
                'default': param_data.get('default', ''),
                'scope': 'stepgroup'
            }
    
    return parameters 

def extract_usage_examples(data):
    """Extract usage examples from template comments or dedicated fields."""
    examples = []
    
    # Check for examples in top-level comments or dedicated field
    if 'examples' in data:
        if isinstance(data['examples'], list):
            examples.extend(data['examples'])
        else:
            examples.append(data['examples'])
    
    # Look for examples in comments (assuming they're stored in a specific format)
    if 'comments' in data:
        for comment in data['comments']:
            if 'example' in comment.lower():
                examples.append(comment)
    
    return examples

def extract_steps(data):
    """Extract steps from template for documentation."""
    steps = []
    
    # Handle pipeline steps
    if 'pipeline' in data and 'stages' in data['pipeline']:
        for stage in data['pipeline']['stages']:
            steps.append({
                'name': stage.get('name', 'Unnamed Stage'),
                'type': 'stage',
                'description': stage.get('description', '')
            })
    
    # Handle stage steps
    if 'stage' in data and 'steps' in data['stage']:
        for step in data['stage']['steps']:
            steps.append({
                'name': step.get('name', 'Unnamed Step'),
                'type': 'step',
                'description': step.get('description', '')
            })
    
    # Handle stepgroup steps
    if 'steps' in data:
        for step in data['steps']:
            steps.append({
                'name': step.get('name', 'Unnamed Step'),
                'type': 'step',
                'description': step.get('description', '')
            })
    
    return steps

def load_yaml(file_path):
    """Load and parse a YAML file."""
    try:
        with open(file_path, 'r') as file:
            return yaml.safe_load(file)
    except yaml.YAMLError as e:
        logger.error(f"YAML parsing error in {file_path}: {e}")
        raise
    except Exception as e:
        logger.error(f"Error loading YAML file {file_path}: {e}")
        raise

def main(yaml_file_path):
    try:
        data = load_yaml(yaml_file_path)
        
        # Determine if this is a template file
        if any(key in data for key in ['pipeline', 'stage', 'steps']):
            from process_template import process_harness_template
            process_harness_template(yaml_file_path)
        else:
            logger.warning(f"Unsupported YAML format: {yaml_file_path}")
    except Exception as e:
        logger.error(f"Error processing file {yaml_file_path}: {e}", exc_info=True)
        return False
    
    return True

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python main.py <yaml_file_path>")
        sys.exit(1)
        
    success = main(sys.argv[1])
    sys.exit(0 if success else 1) 