package html

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	tmpl "github.com/ka1ne/template-doc-gen/pkg/template"
	"github.com/sirupsen/logrus"
)

// Generator handles HTML generation for template documentation
type Generator struct {
	logger    *logrus.Logger
	templates *template.Template
}

// NewGenerator creates a new HTML generator
func NewGenerator(logger *logrus.Logger) *Generator {
	if logger == nil {
		logger = logrus.New()
	}

	// Initialize and parse templates
	templates := template.New("").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"add": func(a, b int) int {
			return a + b
		},
	})

	// Parse all templates
	templates = template.Must(templates.Parse(indexTemplate))
	templates = template.Must(templates.Parse(templateDetailsTemplate))

	return &Generator{
		logger:    logger,
		templates: templates,
	}
}

// GenerateDocumentation generates HTML documentation for templates
func (g *Generator) GenerateDocumentation(metadata []*tmpl.TemplateMetadata, outputDir string) error {
	// Create index file with links to all templates
	if err := g.generateIndexFile(metadata, outputDir); err != nil {
		return err
	}

	// Generate CSS file
	if err := g.generateCSSFile(outputDir); err != nil {
		return err
	}

	// Generate individual template documentation files
	for _, m := range metadata {
		if err := g.generateTemplateFile(m, outputDir); err != nil {
			g.logger.Errorf("Error generating documentation for %s: %v", m.Name, err)
		}
	}

	return nil
}

// generateIndexFile creates an index.html file with links to all templates
func (g *Generator) generateIndexFile(metadata []*tmpl.TemplateMetadata, outputDir string) error {
	indexPath := filepath.Join(outputDir, "index.html")
	g.logger.Infof("Generating index file: %s", indexPath)

	// Group templates by type
	templatesByType := make(map[string][]*tmpl.TemplateMetadata)
	for _, m := range metadata {
		templatesByType[m.Type] = append(templatesByType[m.Type], m)
	}

	// Create template data
	data := map[string]interface{}{
		"TemplatesByType": templatesByType,
		"ValidTypes":      tmpl.ValidTemplateTypes,
	}

	// Create file
	file, err := os.Create(indexPath)
	if err != nil {
		return fmt.Errorf("error creating index file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := g.templates.ExecuteTemplate(file, "index.html", data); err != nil {
		return fmt.Errorf("error executing index template: %w", err)
	}

	return nil
}

// generateCSSFile creates the CSS file for styling the documentation
func (g *Generator) generateCSSFile(outputDir string) error {
	cssPath := filepath.Join(outputDir, "styles.css")
	g.logger.Infof("Generating CSS file: %s", cssPath)

	return os.WriteFile(cssPath, []byte(cssStyles), 0644)
}

// generateTemplateFile creates an HTML file for a specific template
func (g *Generator) generateTemplateFile(metadata *tmpl.TemplateMetadata, outputDir string) error {
	// Create directory for template type if it doesn't exist
	typeDir := filepath.Join(outputDir, metadata.Type)
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return fmt.Errorf("error creating directory for template type: %w", err)
	}

	// Create HTML file path
	filePath := filepath.Join(typeDir, fmt.Sprintf("%s.html", metadata.Identifier))
	g.logger.Infof("Generating template documentation: %s", filePath)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating template file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := g.templates.ExecuteTemplate(file, "template.html", metadata); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

// HTML Templates
const indexTemplate = `
{{define "index.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Harness Template Documentation</title>
    <link rel="stylesheet" href="styles.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github.min.css">
</head>
<body>
    <div class="index-container">
        <header>
            <h1>Harness Template Documentation</h1>
            <div class="search-container">
                <input type="text" id="searchInput" placeholder="Search templates..." class="search-input">
            </div>
        </header>
        
        <div class="template-categories">
            <div class="category-tabs">
                <button class="category-tab active" data-type="all">All Templates</button>
                {{range $typeIndex, $templateType := .ValidTypes}}
                    {{if index $.TemplatesByType $templateType}}
                        <button class="category-tab" data-type="{{$templateType}}">{{$templateType}}</button>
                    {{end}}
                {{end}}
            </div>
        </div>
        
        <main>
            <div class="template-list" id="templateList">
                {{range $typeIndex, $templateType := .ValidTypes}}
                    {{with index $.TemplatesByType $templateType}}
                        <section class="template-type" data-type="{{$templateType}}">
                            <h2 id="{{$templateType}}">{{$templateType}} Templates</h2>
                            <ul>
                                {{range .}}
                                    <li class="template-item" data-name="{{.Name}}" data-tags="{{range .Tags}}{{.}} {{end}}" data-description="{{.Description}}">
                                        <a href="{{.Type}}/{{.Identifier}}.html">{{.Name}}</a>
                                        <span class="version">v{{.Version}}</span>
                                        <p class="description">{{.Description}}</p>
                                        {{if .Tags}}
                                        <div class="tag-list-small">
                                            {{range .Tags}}
                                            <span class="tag-small">{{.}}</span>
                                            {{end}}
                                        </div>
                                        {{end}}
                                    </li>
                                {{end}}
                            </ul>
                        </section>
                    {{end}}
                {{end}}
            </div>
        </main>
        
        <footer>
            <p>Generated by Harness Template Documentation Generator</p>
            <p><a href="#" id="backToTop">Back to Top</a></p>
        </footer>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Search functionality
            const searchInput = document.getElementById('searchInput');
            const templateItems = document.querySelectorAll('.template-item');
            const templateTypes = document.querySelectorAll('.template-type');
            const categoryTabs = document.querySelectorAll('.category-tab');
            
            // Handle category tabs
            categoryTabs.forEach(tab => {
                tab.addEventListener('click', function() {
                    // Update active tab
                    categoryTabs.forEach(t => t.classList.remove('active'));
                    this.classList.add('active');
                    
                    const selectedType = this.getAttribute('data-type');
                    
                    // Show/hide sections based on selected type
                    templateTypes.forEach(section => {
                        const sectionType = section.getAttribute('data-type');
                        if (selectedType === 'all' || selectedType === sectionType) {
                            section.style.display = '';
                        } else {
                            section.style.display = 'none';
                        }
                    });
                    
                    // Apply search filter again to handle visible items
                    filterBySearch();
                });
            });
            
            // Filter by search term
            function filterBySearch() {
                const searchTerm = searchInput.value.toLowerCase();
                
                // Filter individual templates by search term
                templateItems.forEach(item => {
                    const section = item.closest('.template-type');
                    if (section.style.display === 'none') return;
                    
                    const name = item.getAttribute('data-name').toLowerCase();
                    const description = item.getAttribute('data-description').toLowerCase();
                    const tags = item.getAttribute('data-tags').toLowerCase();
                    
                    if (name.includes(searchTerm) || 
                        description.includes(searchTerm) || 
                        tags.includes(searchTerm)) {
                        item.style.display = '';
                    } else {
                        item.style.display = 'none';
                    }
                });
                
                // Hide empty sections
                templateTypes.forEach(section => {
                    if (section.style.display === 'none') return;
                    
                    const visibleItems = section.querySelectorAll('.template-item:not([style="display: none;"])');
                    if (visibleItems.length === 0) {
                        section.style.display = 'none';
                    }
                });
            }
            
            // Add event listeners
            searchInput.addEventListener('input', filterBySearch);
            
            // Back to top functionality
            document.getElementById('backToTop').addEventListener('click', function(e) {
                e.preventDefault();
                window.scrollTo({ top: 0, behavior: 'smooth' });
            });
        });
    </script>
</body>
</html>
{{end}}
`

const templateDetailsTemplate = `
{{define "template.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Name}} - Harness Template Documentation</title>
    <link rel="stylesheet" href="../styles.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github.min.css">
</head>
<body>
    <div class="page-container">
        <div class="back-navigation">
            <a href="../index.html" class="back-link">← Back to Templates</a>
        </div>
        
        <div class="main-content-wrapper">
            <main class="main-content">
                <header>
                    <div class="breadcrumb">
                        <a href="../index.html">Home</a> &gt; 
                        <a href="../index.html#{{.Type}}">{{.Type}} Templates</a> &gt; 
                        <span>{{.Name}}</span>
                    </div>
                    <h1>{{.Name}}</h1>
                </header>
                
                <div class="content-area">
                    <div class="metadata-section" id="metadata">
                        <h2>Template Metadata</h2>
                        <table class="metadata-table">
                            <tr>
                                <th class="metadata-label">Identifier:</th>
                                <td class="metadata-value">{{.Identifier}}</td>
                            </tr>
                            <tr>
                                <th class="metadata-label">Version:</th>
                                <td class="metadata-value">{{.Version}}</td>
                            </tr>
                            <tr>
                                <th class="metadata-label">Type:</th>
                                <td class="metadata-value">{{.Type}}</td>
                            </tr>
                            {{if .Description}}
                            <tr>
                                <th class="metadata-label">Description:</th>
                                <td class="metadata-value">{{.Description}}</td>
                            </tr>
                            {{end}}
                            {{if .Author}}
                            <tr>
                                <th class="metadata-label">Author:</th>
                                <td class="metadata-value">{{.Author}}</td>
                            </tr>
                            {{end}}
                        </table>
                    </div>

                    {{if .Tags}}
                    <div class="metadata-section" id="tags">
                        <h2>Tags</h2>
                        <div class="tag-list">
                            {{range .Tags}}
                            <span class="tag">{{.}}</span>
                            {{end}}
                        </div>
                    </div>
                    {{end}}

                    {{if .Parameters}}
                    <div class="metadata-section" id="parameters">
                        <h2>Parameters</h2>
                        <table class="parameter-table">
                            <tr>
                                <th>Name</th>
                                <th>Type</th>
                                <th class="required-col">Required</th>
                                <th>Default</th>
                                <th>Description</th>
                            </tr>
                            {{range $name, $param := .Parameters}}
                            <tr>
                                <td class="name-cell">
                                    <span class="field-name">{{$name}}</span>
                                </td>
                                <td>{{$param.Type}}</td>
                                <td class="required-col">
                                    {{if $param.Required}}
                                    <span class="required-badge">Yes</span>
                                    {{else}}
                                    <span class="optional-text">No</span>
                                    {{end}}
                                </td>
                                <td>{{if $param.Default}}{{$param.Default}}{{end}}</td>
                                <td>{{$param.Description}}</td>
                            </tr>
                            {{end}}
                        </table>
                    </div>
                    {{end}}

                    {{if .Variables}}
                    <div class="metadata-section" id="variables">
                        <h2>Variables</h2>
                        <table class="variable-table">
                            <tr>
                                <th>Name</th>
                                <th>Type</th>
                                <th class="required-col">Required</th>
                                <th>Scope</th>
                                <th>Description</th>
                            </tr>
                            {{range $name, $var := .Variables}}
                            <tr>
                                <td class="name-cell">
                                    <span class="field-name">{{$name}}</span>
                                </td>
                                <td>{{$var.Type}}</td>
                                <td class="required-col">
                                    {{if $var.Required}}
                                    <span class="required-badge">Yes</span>
                                    {{else}}
                                    <span class="optional-text">No</span>
                                    {{end}}
                                </td>
                                <td>{{$var.Scope}}</td>
                                <td>{{$var.Description}}</td>
                            </tr>
                            {{end}}
                        </table>
                    </div>
                    {{end}}

                    {{if .Examples}}
                    <div class="metadata-section" id="examples">
                        <h2>Example Usage</h2>
                        {{range $index, $example := .Examples}}
                        <h3>Example {{add $index 1}}</h3>
                        <pre><code class="language-yaml">{{$example}}</code></pre>
                        {{end}}
                    </div>
                    {{end}}
                    
                    <footer>
                        <p>Generated by Harness Template Documentation Generator</p>
                        <p><a href="#" id="backToTop">Back to Top</a></p>
                    </footer>
                </div>
            </main>
            
            <aside class="right-sidebar">
                <div class="sidebar-sticky">
                    <div class="sidebar-header">
                        <h3>On This Page</h3>
                    </div>
                    <nav class="sidebar-nav">
                        <ul>
                            <li><a href="#metadata" class="nav-link">Template Metadata</a></li>
                            {{if .Tags}}<li><a href="#tags" class="nav-link">Tags</a></li>{{end}}
                            {{if .Parameters}}<li><a href="#parameters" class="nav-link">Parameters</a></li>{{end}}
                            {{if .Variables}}<li><a href="#variables" class="nav-link">Variables</a></li>{{end}}
                            {{if .Examples}}<li><a href="#examples" class="nav-link">Example Usage</a></li>{{end}}
                        </ul>
                    </nav>
                </div>
            </aside>
        </div>
    </div>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/languages/yaml.min.js"></script>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize syntax highlighting
            hljs.highlightAll();
            
            // Back to top functionality
            document.getElementById('backToTop').addEventListener('click', function(e) {
                e.preventDefault();
                window.scrollTo({ top: 0, behavior: 'smooth' });
            });
            
            // Highlight active section in sidebar
            const sections = document.querySelectorAll('.metadata-section');
            const navLinks = document.querySelectorAll('.sidebar-nav a');
            
            function highlightNavigation() {
                let scrollPosition = window.scrollY + 100;
                
                sections.forEach(section => {
                    const sectionTop = section.offsetTop;
                    const sectionHeight = section.offsetHeight;
                    
                    if (scrollPosition >= sectionTop && scrollPosition < sectionTop + sectionHeight) {
                        const id = section.getAttribute('id');
                        
                        navLinks.forEach(link => {
                            link.classList.remove('active');
                            if (link.getAttribute('href') === '#' + id) {
                                link.classList.add('active');
                            }
                        });
                    }
                });
            }
            
            window.addEventListener('scroll', highlightNavigation);
            highlightNavigation();
        });
    </script>
</body>
</html>
{{end}}
`

// CSS Styles
const cssStyles = `:root {
    --primary-color: #0078D4;
    --secondary-color: #106EBE;
    --accent-color: #2B88D8;
    --light-bg: #F8F9FA;
    --dark-text: #333333;
    --light-text: #666666;
    --card-border: #E1E1E1;
    --code-bg: #F5F5F5;
    --shadow: 0 2px 4px rgba(0,0,0,0.1);
}

* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    line-height: 1.6;
    color: var(--dark-text);
    background-color: #ffffff;
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    margin-bottom: 30px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--card-border);
}

h1 {
    color: var(--primary-color);
    margin-bottom: 10px;
}

h2 {
    color: var(--secondary-color);
    margin: 20px 0 15px 0;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--card-border);
}

h3 {
    color: var(--accent-color);
    margin: 25px 0 10px 0;
}

a {
    color: var(--primary-color);
    text-decoration: none;
}

a:hover {
    text-decoration: underline;
}

ul {
    list-style-type: none;
}

/* Search and Filtering */
.search-container {
    margin: 20px 0;
}

.search-input {
    width: 100%;
    padding: 10px;
    border: 1px solid var(--card-border);
    border-radius: 4px;
    font-size: 16px;
    margin-bottom: 10px;
}

.filter-options {
    display: flex;
    flex-wrap: wrap;
    gap: 15px;
    margin-bottom: 10px;
}

.filter-options label {
    display: flex;
    align-items: center;
    gap: 5px;
    cursor: pointer;
}

/* Template List */
.template-type {
    margin-bottom: 30px;
}

.template-list li {
    background-color: var(--light-bg);
    padding: 15px;
    margin-bottom: 10px;
    border-radius: 5px;
    border-left: 4px solid var(--primary-color);
    box-shadow: var(--shadow);
    transition: transform 0.2s, box-shadow 0.2s;
}

.template-list li:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
}

.version {
    background-color: var(--primary-color);
    color: white;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.8em;
    margin-left: 10px;
}

.description {
    color: var(--light-text);
    font-size: 0.9em;
    margin-top: 5px;
}

.tag-list-small {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
    margin-top: 10px;
}

.tag-small {
    background-color: rgba(0,0,0,0.05);
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.75em;
    color: var(--light-text);
}

/* Harness-style layout */
.page-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0;
    position: relative;
}

.back-navigation {
    position: absolute;
    top: 20px;
    left: 20px;
    z-index: 10;
}

.back-link {
    display: inline-flex;
    align-items: center;
    font-size: 14px;
    font-weight: 500;
    color: var(--primary-color);
    text-decoration: none;
}

.back-link:hover {
    text-decoration: underline;
}

.main-content-wrapper {
    display: flex;
    padding: 20px;
    padding-top: 60px; /* Space for back button */
}

.main-content {
    flex: 1;
    max-width: calc(100% - 280px);
    padding-right: 30px;
}

.content-area {
    max-width: 100%;
}

.right-sidebar {
    width: 260px;
    flex-shrink: 0;
    position: relative;
}

.sidebar-sticky {
    position: sticky;
    top: 20px;
    background-color: var(--light-bg);
    border-radius: 6px;
    border: 1px solid var(--card-border);
    padding: 5px 0;
    margin-top: 60px; /* Align with content */
}

.sidebar-header {
    padding: 15px 20px;
    border-bottom: 1px solid var(--card-border);
}

.sidebar-header h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--dark-text);
}

.sidebar-nav ul {
    list-style: none;
    padding: 10px 0;
    margin: 0;
}

.sidebar-nav li {
    margin: 0;
}

.sidebar-nav .nav-link {
    display: block;
    padding: 8px 20px;
    color: var(--dark-text);
    font-size: 14px;
    border-left: 3px solid transparent;
    transition: all 0.2s;
}

.sidebar-nav .nav-link:hover,
.sidebar-nav .nav-link.active {
    background-color: rgba(0, 120, 212, 0.08);
    border-left-color: var(--primary-color);
    text-decoration: none;
    color: var(--primary-color);
}

@media (max-width: 992px) {
    .main-content-wrapper {
        flex-direction: column;
        padding-top: 80px;
    }
    
    .main-content {
        max-width: 100%;
        padding-right: 0;
    }
    
    .right-sidebar {
        width: 100%;
        margin-top: 30px;
    }
    
    .sidebar-sticky {
        position: relative;
        top: 0;
        margin-top: 0;
    }
    
    .sidebar-nav ul {
        display: flex;
        flex-wrap: wrap;
        padding: 10px;
    }
    
    .sidebar-nav li {
        margin-right: 5px;
        margin-bottom: 5px;
    }
    
    .sidebar-nav .nav-link {
        padding: 6px 12px;
        border-left: none;
        border-radius: 4px;
        white-space: nowrap;
    }
    
    .sidebar-nav .nav-link:hover,
    .sidebar-nav .nav-link.active {
        border-left-color: transparent;
    }
}

/* Detail Page Layout */
.breadcrumb {
    margin-bottom: 15px;
    color: var(--light-text);
    font-size: 0.9em;
}

.metadata-section {
    margin-bottom: 40px;
    padding: 20px;
    background-color: white;
    border-radius: 5px;
    box-shadow: var(--shadow);
}

.tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 15px;
}

.tag {
    background-color: var(--code-bg);
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 0.85em;
    display: inline-block;
}

pre, code {
    background-color: var(--code-bg);
    padding: 10px;
    border-radius: 5px;
    overflow-x: auto;
    font-family: 'Source Code Pro', monospace;
    margin: 15px 0;
    line-height: 1.5;
}

pre {
    padding: 15px;
    box-shadow: var(--shadow);
}

code {
    display: inline-block;
    padding-bottom: 3px;
}

.metadata-table {
    width: 100%;
    border-collapse: collapse;
    margin: 15px 0;
    border: 1px solid var(--card-border);
    border-radius: 4px;
    overflow: hidden;
}

.metadata-table tr:hover {
    background-color: rgba(0,0,0,0.02);
}

.metadata-table tr {
    border-bottom: 1px solid var(--card-border);
}

.metadata-table tr:last-child {
    border-bottom: none;
}

.metadata-label {
    width: 150px;
    min-width: 150px;
    max-width: 200px;
    text-align: left;
    padding: 12px;
    font-weight: 600;
    background-color: var(--light-bg);
    vertical-align: top;
}

.metadata-value {
    padding: 12px;
    word-break: break-word;
}

.parameter-table, .variable-table {
    width: 100%;
    border-collapse: collapse;
    margin: 20px 0;
    border: 1px solid var(--card-border);
    border-radius: 4px;
    overflow: hidden;
}

.parameter-table th, .variable-table th {
    text-align: left;
    padding: 12px;
    background-color: var(--light-bg);
    font-weight: 600;
}

.parameter-table th:first-child, .variable-table th:first-child {
    width: 150px;
    min-width: 150px;
}

.parameter-table th:nth-child(2), .variable-table th:nth-child(2) {
    width: 100px;
    min-width: 100px;
}

.parameter-table th:last-child, .variable-table th:last-child {
    width: 40%;
}

.parameter-table tr:not(:last-child), 
.variable-table tr:not(:last-child) {
    border-bottom: 1px solid var(--card-border);
}

.parameter-table td, .variable-table td {
    vertical-align: middle;
    padding: 12px;
    word-break: break-word;
}

.parameter-table tr:hover, .variable-table tr:hover {
    background-color: rgba(0,0,0,0.02);
}

.name-cell {
    vertical-align: middle;
    padding: 12px;
    text-align: left;
    width: 150px;
    min-width: 150px;
    max-width: 200px;
    position: relative;
}

.field-name {
    font-weight: 500;
    display: inline-block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
}

.required-col {
    width: 100px;
    text-align: center;
    vertical-align: middle;
}

.required-badge {
    display: inline-block;
    background-color: #e8f5e9;
    color: #2e7d32;
    font-size: 0.85em;
    padding: 3px 10px;
    border-radius: 3px;
    border: 1px solid #c8e6c9;
    font-weight: 600;
    min-width: 40px;
    text-align: center;
}

.optional-text {
    display: inline-block;
    color: #757575;
    font-size: 0.85em;
    min-width: 40px;
    text-align: center;
}

footer {
    margin-top: 50px;
    padding-top: 20px;
    border-top: 1px solid var(--card-border);
    text-align: center;
    color: var(--light-text);
    display: flex;
    justify-content: space-between;
    align-items: center;
}

#backToTop {
    display: inline-block;
    background-color: var(--primary-color);
    color: white;
    padding: 8px 15px;
    border-radius: 4px;
    text-decoration: none;
    transition: background-color 0.2s;
}

#backToTop:hover {
    background-color: var(--secondary-color);
    text-decoration: none;
}

@media (max-width: 768px) {
    body {
        padding: 15px;
    }
    
    .metadata-table,
    .parameter-table, 
    .variable-table {
        display: block;
        overflow-x: auto;
        white-space: nowrap;
        border: 1px solid var(--card-border);
        border-radius: 4px;
    }
    
    .metadata-table th,
    .parameter-table th, 
    .variable-table th,
    .metadata-table td,
    .parameter-table td, 
    .variable-table td {
        white-space: normal;
    }
    
    .metadata-label,
    .name-cell {
        position: sticky;
        left: 0;
        background-color: var(--light-bg);
        z-index: 1;
        border-right: 1px solid var(--card-border);
    }
    
    footer {
        flex-direction: column;
        gap: 15px;
    }
}

/* Index Page Layout */
.index-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

.template-categories {
    margin: 20px 0;
    border-bottom: 1px solid var(--card-border);
}

.category-tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
}

.category-tab {
    background: none;
    border: none;
    padding: 10px 15px;
    font-size: 14px;
    border-bottom: 3px solid transparent;
    cursor: pointer;
    color: var(--dark-text);
    transition: all 0.2s;
}

.category-tab:hover {
    color: var(--primary-color);
}

.category-tab.active {
    color: var(--primary-color);
    font-weight: 500;
    border-bottom-color: var(--primary-color);
}

@media (max-width: 768px) {
    .category-tabs {
        overflow-x: auto;
        white-space: nowrap;
        padding-bottom: 5px;
    }
}
`
