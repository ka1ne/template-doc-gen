# Harness Templates Documentation

This directory contains the generated documentation for Harness Templates.

## Including Template Code in Documentation

The documentation now supports displaying the raw template code in a nicely formatted, syntax-highlighted code block. This provides users with an easy way to copy and download the template.

### How to Enable Template Code Display

1. Modify the `extract_template_metadata` function in `src/utils/template_html.py` to include the raw template data:

```python
def extract_template_metadata(template_data):
    """Extract metadata from Harness template including variables, descriptions, and raw template."""
    
    # Your existing metadata extraction logic
    metadata = {
        'name': template_data.get('name', 'Unnamed Template'),
        'type': template_data.get('type', 'unknown'),
        # ... other metadata
        
        # Add this line to include the raw template data
        'raw_template': template_data
    }
    
    return metadata
```

2. Update the `generate_template_html` function to include the template code section:

```python
def generate_template_html(template_metadata):
    # Get the raw template code
    template_code = template_metadata.get('raw_template', '')
    
    # Format template code nicely for YAML
    if template_code and isinstance(template_code, dict):
        try:
            import yaml
            template_code = yaml.dump(template_code, default_flow_style=False, sort_keys=False)
        except Exception as e:
            import json
            template_code = json.dumps(template_code, indent=2)
    
    # ... existing HTML generation code ...
    
    # Add template code section after variables section
    html += """
                <div class="template-code-section">
                    <h2>Template Code</h2>
    """
    
    if template_code:
        # Generate a safe filename for download based on template name
        import re
        safe_filename = re.sub(r'[^a-zA-Z0-9_\-]', '_', template_metadata['name'].lower()) + '.yaml'
        
        html += f"""
                    <div class="code-container">
                        <div class="code-header">
                            <div class="code-title">
                                <i class="fas fa-file-code"></i> {safe_filename}
                            </div>
                            <div class="code-actions">
                                <button class="code-action-btn" id="copy-btn" data-clipboard-target="#template-code">
                                    <i class="fas fa-copy"></i> <span>Copy</span>
                                </button>
                                <button class="code-action-btn" id="download-btn" data-filename="{safe_filename}">
                                    <i class="fas fa-download"></i> <span>Download</span>
                                </button>
                            </div>
                        </div>
                        <div class="code-block line-numbers">
                            <pre><code class="language-yaml" id="template-code">{template_code}</code></pre>
                        </div>
                    </div>
        """
    else:
        html += '<p class="empty-state"><i class="fas fa-info-circle"></i> No template code available.</p>'
    
    html += """
                </div>
    """
    
    # Continue with the rest of your HTML generation...
```

3. Make sure to include the required JavaScript libraries in the HTML head:

```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/styles/atom-one-dark.min.css">
<script src="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/highlight.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/languages/yaml.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/clipboard@2.0.11/dist/clipboard.min.js"></script>
```

4. Add the clipboard and download functionalities in your JavaScript:

```javascript
// Initialize syntax highlighting
document.addEventListener('DOMContentLoaded', (event) => {
    hljs.highlightAll();
});

// Copy to clipboard functionality
const clipboard = new ClipboardJS('#copy-btn');

clipboard.on('success', function(e) {
    const copyBtn = document.getElementById('copy-btn');
    const originalText = copyBtn.innerHTML;
    
    copyBtn.innerHTML = '<i class="fas fa-check"></i> <span>Copied!</span>';
    setTimeout(() => {
        copyBtn.innerHTML = originalText;
    }, 2000);
    
    e.clearSelection();
});

// Download functionality
document.getElementById('download-btn').addEventListener('click', function() {
    const codeContent = document.getElementById('template-code').textContent;
    const filename = this.getAttribute('data-filename');
    const blob = new Blob([codeContent], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    
    setTimeout(() => {
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }, 0);
});
```

## Benefits of Template Code Display

1. **Transparency**: Users can see the exact template code that will be used
2. **Easy Access**: One-click copy and download functionality
3. **Documentation Completeness**: Templates are self-documented with their implementation
4. **Better User Experience**: Syntax highlighting makes the code more readable
5. **Reduced Friction**: Users can quickly grab and use templates without navigating to another resource 