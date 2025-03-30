import re

def generate_template_html(template_metadata):
    """Generate HTML documentation for a Harness template."""
    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{template_metadata['name']} - Harness Template Documentation</title>
    <link rel="stylesheet" href="../styles.css">
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/styles/atom-one-dark.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">
            <h3><i class="fas fa-layer-group"></i> Harness Templates</h3>
        </div>
        <div class="sidebar-menu">
            <a href="../index.html"><i class="fas fa-home"></i> All Templates</a>
            <a href="../pipeline/"><i class="fas fa-sitemap"></i> Pipelines</a>
            <a href="../stage/"><i class="fas fa-cube"></i> Stages</a>
            <a href="../stepgroup/"><i class="fas fa-cubes"></i> Step Groups</a>
        </div>
    </div>
    
    <div class="content">
        <header>
            <div class="template-header">
                <h1>{template_metadata['name']}</h1>
                <span class="template-badge type-{template_metadata['type']}">{template_metadata['type']}</span>
            </div>
            <p class="breadcrumbs">
                <a href="../index.html">Templates</a> <span>/</span> 
                <a href="../{template_metadata['type']}/">{template_metadata['type'].capitalize()}</a> <span>/</span> 
                <span>{template_metadata['name']}</span>
            </p>
        </header>

        <main>
            <section class="template-section" data-type="{template_metadata['type']}">
                <div class="template-metadata">
                    <div class="metadata-item">
                        <span class="metadata-label">Version</span>
                        <span class="metadata-value">{template_metadata['version']}</span>
                    </div>
                    <div class="metadata-item">
                        <span class="metadata-label">Author</span>
                        <span class="metadata-value">{template_metadata['author']}</span>
                    </div>
                </div>
                
                <div class="template-description">
                    <h2>Description</h2>
                    <p>{template_metadata['description']}</p>
                </div>
                
                <div class="template-tags">
                    <h2>Tags</h2>
"""
    
    if template_metadata.get('tags'):
        for tag in template_metadata.get('tags', []):
            html += f'<span class="tag"><i class="fas fa-tag"></i> {tag}</span>'
    else:
        html += '<p class="empty-state"><i class="fas fa-info-circle"></i> No tags defined for this template.</p>'
    
    html += """
                </div>
                
                <div class="template-parameters">
                    <h2>Parameters</h2>
    """
    
    if template_metadata.get('parameters'):
        html += """
                    <table class="parameters-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Type</th>
                                <th>Required</th>
                                <th>Default</th>
                                <th>Scope</th>
                            </tr>
                        </thead>
                        <tbody>
        """
        
        for param_name, param_data in template_metadata['parameters'].items():
            html += f"""
                            <tr>
                                <td class="param-name">{param_name}</td>
                                <td>{param_data['description']}</td>
                                <td><code>{param_data['type']}</code></td>
                                <td>{'<span class="required">Yes</span>' if param_data['required'] else 'No'}</td>
                                <td><code>{param_data['default']}</code></td>
                                <td>{param_data['scope']}</td>
                            </tr>
            """
        
        html += """
                        </tbody>
                    </table>
        """
    else:
        html += '<p class="empty-state"><i class="fas fa-info-circle"></i> No parameters defined for this template.</p>'
    
    html += """
                </div>
                
                <div class="template-variables">
                    <h2>Variables</h2>
    """
    
    if template_metadata.get('variables'):
        html += """
                    <table class="variables-table">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Description</th>
                                <th>Type</th>
                                <th>Required</th>
                                <th>Scope</th>
                            </tr>
                        </thead>
                        <tbody>
        """
        
        for var_name, var_data in template_metadata['variables'].items():
            html += f"""
                            <tr>
                                <td class="var-name">{var_name}</td>
                                <td>{var_data['description']}</td>
                                <td><code>{var_data['type']}</code></td>
                                <td>{'<span class="required">Yes</span>' if var_data['required'] else 'No'}</td>
                                <td>{var_data['scope']}</td>
                            </tr>
            """
        
        html += """
                        </tbody>
                    </table>
        """
    else:
        html += '<p class="empty-state"><i class="fas fa-info-circle"></i> No variables defined for this template.</p>'
    
    # Add examples section if available
    if template_metadata.get('examples'):
        html += """
                <div class="template-examples">
                    <h2>Usage Examples</h2>
                    <div class="examples-container">
        """
        
        for i, example in enumerate(template_metadata['examples']):
            html += f"""
                        <div class="example">
                            <h3><i class="fas fa-code"></i> Example {i+1}</h3>
                            <pre><code class="language-yaml">{example}</code></pre>
                        </div>
            """
        
        html += """
                    </div>
                </div>
        """
    
    html += """
            </section>
        </main>
        
        <footer>
            <p>Generated on <span id="generation-date"></span> | <a href="https://harness.io" target="_blank">Harness</a> Template Documentation</p>
        </footer>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/highlight.min.js"></script>
    <script>
        document.getElementById('generation-date').textContent = new Date().toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
        });
        
        // Initialize syntax highlighting
        document.addEventListener('DOMContentLoaded', (event) => {
            document.querySelectorAll('pre code').forEach((el) => {
                hljs.highlightElement(el);
            });
        });
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
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</head>
<body>
    <div class="sidebar">
        <div class="sidebar-header">
            <h3><i class="fas fa-layer-group"></i> Harness Templates</h3>
        </div>
        <div class="sidebar-menu">
            <a href="index.html" class="active"><i class="fas fa-home"></i> All Templates</a>
            <a href="pipeline/"><i class="fas fa-sitemap"></i> Pipelines</a>
            <a href="stage/"><i class="fas fa-cube"></i> Stages</a>
            <a href="stepgroup/"><i class="fas fa-cubes"></i> Step Groups</a>
        </div>
    </div>
    
    <div class="content">
        <header>
            <h1>Harness Templates Documentation</h1>
            <div class="header-actions">
                <div class="search-container">
                    <input type="text" id="searchInput" placeholder="Search templates...">
                </div>
                <div class="filter-container">
                    <button class="filter-btn active" data-filter="all">All</button>
                    <button class="filter-btn" data-filter="pipeline">Pipelines</button>
                    <button class="filter-btn" data-filter="stage">Stages</button>
                    <button class="filter-btn" data-filter="stepgroup">Step Groups</button>
                </div>
            </div>
        </header>
        
        <main>
            <p class="template-count"><i class="fas fa-list"></i> <span id="visible-count">0</span> templates found</p>
            <div class="templates-grid">
    """
    
    # Group templates by type
    pipelines = []
    stages = []
    stepgroups = []
    
    for metadata in templates_metadata:
        if metadata['type'] == 'pipeline':
            pipelines.append(metadata)
        elif metadata['type'] == 'stage':
            stages.append(metadata)
        elif metadata['type'] == 'stepgroup':
            stepgroups.append(metadata)
    
    # Sort each group by name
    pipelines.sort(key=lambda x: x['name'])
    stages.sort(key=lambda x: x['name'])
    stepgroups.sort(key=lambda x: x['name'])
    
    # Combine sorted groups
    sorted_metadata = pipelines + stages + stepgroups
    
    for metadata in sorted_metadata:
        description = metadata['description'][:100] + "..." if len(metadata['description']) > 100 else metadata['description']
        
        # Sanitize filename
        safe_name = re.sub(r'[^a-zA-Z0-9_\-]', '_', metadata['name'])
        
        # Choose icon based on type
        icon = 'sitemap' if metadata['type'] == 'pipeline' else 'cube' if metadata['type'] == 'stage' else 'cubes'
        
        html += f"""
                <div class="template-card" data-type="{metadata['type']}">
                    <div class="card-header">
                        <h2><i class="fas fa-{icon}"></i> {metadata['name']}</h2>
                        <span class="template-badge type-{metadata['type']}">{metadata['type']}</span>
                    </div>
                    <p class="template-description">{description}</p>
                    <div class="template-tags">
        """
        
        for tag in metadata.get('tags', [])[:3]:
            html += f'<span class="tag"><i class="fas fa-tag"></i> {tag}</span>'
        
        html += f"""
                    </div>
                    <a href="{metadata['type']}/{safe_name}.html" class="view-btn">View Details</a>
                </div>
        """
    
    html += """
            </div>
        </main>
        
        <footer>
            <p>Generated on <span id="generation-date"></span> | <a href="https://harness.io" target="_blank">Harness</a> Template Documentation</p>
        </footer>
    </div>
    
    <script>
        document.getElementById('generation-date').textContent = new Date().toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
        });
        
        // Update template count
        function updateTemplateCount() {
            const visibleTemplates = document.querySelectorAll('.template-card:not([style*="display: none"])');
            document.getElementById('visible-count').textContent = visibleTemplates.length;
        }
        
        // Search functionality
        const searchInput = document.getElementById('searchInput');
        const templateCards = document.querySelectorAll('.template-card');
        
        updateTemplateCount();
        
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
            
            updateTemplateCount();
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
                
                updateTemplateCount();
            });
        });
    </script>
</body>
</html>
    """
    
    return html 