def generate_template_html(template_metadata):
    """Generate HTML documentation for a Harness template."""
    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{template_metadata['name']} - Harness Template Documentation</title>
    <link rel="stylesheet" href="../styles.css">
</head>
<body>
    <header>
        <h1>{template_metadata['name']}</h1>
        <p><a href="../index.html">← Back to All Templates</a></p>
    </header>

    <main>
        <section class="template-section" data-type="{template_metadata['type']}">
            <div class="template-metadata">
                <p class="template-type">Type: <span>{template_metadata['type']}</span></p>
                <p class="template-version">Version: <span>{template_metadata['version']}</span></p>
                <p class="template-author">Author: <span>{template_metadata['author']}</span></p>
            </div>
            
            <div class="template-description">
                <h3>Description</h3>
                <p>{template_metadata['description']}</p>
            </div>
            
            <div class="template-tags">
                <h3>Tags</h3>
                <ul>
    """
    
    for tag in template_metadata.get('tags', []):
        html += f'<li class="tag">{tag}</li>'
    
    html += """
                </ul>
            </div>
            
            <div class="template-parameters">
                <h3>Parameters</h3>
    """
    
    if template_metadata.get('parameters'):
        html += """
                <table class="parameters-table">
                    <tr>
                        <th>Name</th>
                        <th>Description</th>
                        <th>Type</th>
                        <th>Required</th>
                        <th>Default</th>
                        <th>Scope</th>
                    </tr>
        """
        
        for param_name, param_data in template_metadata['parameters'].items():
            html += f"""
                    <tr>
                        <td>{param_name}</td>
                        <td>{param_data['description']}</td>
                        <td>{param_data['type']}</td>
                        <td>{'Yes' if param_data['required'] else 'No'}</td>
                        <td>{param_data['default']}</td>
                        <td>{param_data['scope']}</td>
                    </tr>
            """
        
        html += "</table>"
    else:
        html += "<p>No parameters defined for this template.</p>"
    
    html += """
            </div>
            
            <div class="template-variables">
                <h3>Variables</h3>
    """
    
    if template_metadata.get('variables'):
        html += """
                <table class="variables-table">
                    <tr>
                        <th>Name</th>
                        <th>Description</th>
                        <th>Type</th>
                        <th>Required</th>
                        <th>Scope</th>
                    </tr>
        """
        
        for var_name, var_data in template_metadata['variables'].items():
            html += f"""
                    <tr>
                        <td>{var_name}</td>
                        <td>{var_data['description']}</td>
                        <td>{var_data['type']}</td>
                        <td>{'Yes' if var_data['required'] else 'No'}</td>
                        <td>{var_data['scope']}</td>
                    </tr>
            """
        
        html += "</table>"
    else:
        html += "<p>No variables defined for this template.</p>"
    
    # Add examples section if available
    if template_metadata.get('examples'):
        html += """
            <div class="template-examples">
                <h3>Usage Examples</h3>
                <div class="examples-container">
        """
        
        for i, example in enumerate(template_metadata['examples']):
            html += f"""
                    <div class="example">
                        <h4>Example {i+1}</h4>
                        <pre><code>{example}</code></pre>
                    </div>
            """
        
        html += """
                </div>
            </div>
        """
    
    html += """
            </div>
        </section>
    </main>
    
    <footer>
        <p>Generated on <span id="generation-date"></span></p>
    </footer>
    
    <script>
        document.getElementById('generation-date').textContent = new Date().toLocaleDateString();
    </script>
</body>
</html>
    """
    
    return html

def generate_index_html(templates_metadata):
    """Generate an index HTML page for all templates."""
    html = """<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Harness Templates Documentation</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <header>
        <h1>Harness Templates Documentation</h1>
        <div class="search-container">
            <input type="text" id="searchInput" placeholder="Search templates...">
        </div>
        <div class="filter-container">
            <button class="filter-btn active" data-filter="all">All</button>
            <button class="filter-btn" data-filter="pipeline">Pipelines</button>
            <button class="filter-btn" data-filter="stage">Stages</button>
            <button class="filter-btn" data-filter="stepgroup">Step Groups</button>
        </div>
    </header>
    
    <main>
        <div class="templates-grid">
    """
    
    for metadata in templates_metadata:
        description = metadata['description'][:100] + "..." if len(metadata['description']) > 100 else metadata['description']
        
        html += f"""
            <div class="template-card" data-type="{metadata['type']}">
                <h2>{metadata['name']}</h2>
                <p class="template-type">{metadata['type']}</p>
                <p class="template-description">{description}</p>
                <div class="template-tags">
        """
        
        for tag in metadata.get('tags', [])[:3]:
            html += f'<span class="tag">{tag}</span>'
        
        html += f"""
                </div>
                <a href="{metadata['type']}/{metadata['name'].replace(' ', '_')}.html" class="view-btn">View Details</a>
            </div>
        """
    
    html += """
        </div>
    </main>
    
    <footer>
        <p>Generated on <span id="generation-date"></span></p>
    </footer>
    
    <script>
        document.getElementById('generation-date').textContent = new Date().toLocaleDateString();
        
        // Search functionality
        const searchInput = document.getElementById('searchInput');
        const templateCards = document.querySelectorAll('.template-card');
        
        searchInput.addEventListener('input', function() {
            const query = this.value.toLowerCase();
            
            templateCards.forEach(card => {
                const name = card.querySelector('h2').textContent.toLowerCase();
                const description = card.querySelector('.template-description').textContent.toLowerCase();
                const tags = Array.from(card.querySelectorAll('.tag')).map(tag => tag.textContent.toLowerCase());
                
                if (name.includes(query) || description.includes(query) || tags.some(tag => tag.includes(query))) {
                    card.style.display = 'block';
                } else {
                    card.style.display = 'none';
                }
            });
        });
        
        // Filter functionality
        const filterButtons = document.querySelectorAll('.filter-btn');
        
        filterButtons.forEach(button => {
            button.addEventListener('click', function() {
                const filter = this.getAttribute('data-filter');
                
                filterButtons.forEach(btn => btn.classList.remove('active'));
                this.classList.add('active');
                
                templateCards.forEach(card => {
                    if (filter === 'all' || card.getAttribute('data-type') === filter) {
                        card.style.display = 'block';
                    } else {
                        card.style.display = 'none';
                    }
                });
            });
        });
    </script>
</body>
</html>
    """
    
    return html 