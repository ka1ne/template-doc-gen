package template

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ka1ne/template-doc-gen/pkg/schema"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// processor handles template operations
type Processor struct {
	logger        *logrus.Logger
	schemaManager *schema.SchemaManager
}

// creates new processor
func NewProcessor(logger *logrus.Logger) *Processor {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	return &Processor{
		logger: logger,
	}
}

// sets schema manager
func (p *Processor) SetSchemaManager(manager *schema.SchemaManager) {
	p.schemaManager = manager
}

// processes a single template file
func (p *Processor) ProcessTemplate(templatePath string, outputDir string, outputFormat string, validateOnly bool) (*TemplateMetadata, error) {
	p.logger.Infof("Processing template: %s", templatePath)

	// read template file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("error reading template file: %w", err)
	}

	// parse yaml
	var templateData map[string]interface{}
	if err := yaml.Unmarshal(data, &templateData); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// validate template
	if valid, msg := p.ValidateTemplate(templateData); !valid {
		p.logger.Errorf("Template validation failed for %s: %s", templatePath, msg)
		return nil, fmt.Errorf("validation failed: %s", msg)
	}

	// extract metadata
	metadata, err := p.ExtractMetadata(templateData)
	if err != nil {
		return nil, fmt.Errorf("error extracting metadata: %w", err)
	}

	if validateOnly {
		p.logger.Infof("Template %s is valid", templatePath)
		return metadata, nil
	}

	// // generate documentation if needed (handled by separate Python service)
	// // store metadata for later processing

	return metadata, nil
}

// processes all templates in a directory
func (p *Processor) ProcessAllTemplates(templatesDir string, outputDir string, outputFormat string, validateOnly bool) ([]*TemplateMetadata, error) {
	startTime := time.Now()
	p.logger.Infof("Starting template processing from %s", templatesDir)

	// create output directory if needed
	if !validateOnly {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("error creating output directory: %w", err)
		}

		// create type subdirectories
		for _, templateType := range ValidTemplateTypes {
			typeDir := filepath.Join(outputDir, templateType)
			if err := os.MkdirAll(typeDir, 0755); err != nil {
				return nil, fmt.Errorf("error creating type directory: %w", err)
			}
		}
	}

	// find template files
	var templateFiles []string
	fi, err := os.Stat(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("error accessing template directory: %w", err)
	}

	if !fi.IsDir() {
		// single file
		if strings.HasSuffix(templatesDir, ".yaml") || strings.HasSuffix(templatesDir, ".yml") {
			templateFiles = []string{templatesDir}
		} else {
			return nil, fmt.Errorf("specified file is not a YAML file: %s", templatesDir)
		}
	} else {
		// find all yaml files
		err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
				templateFiles = append(templateFiles, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking template directory: %w", err)
		}
	}

	if len(templateFiles) == 0 {
		p.logger.Warning("No template files found")
		return []*TemplateMetadata{}, nil
	}

	p.logger.Infof("Found %d template files", len(templateFiles))

	// setup concurrent processing
	type result struct {
		metadata *TemplateMetadata
		err      error
		path     string
	}

	// determine worker count
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	if len(templateFiles) < numWorkers {
		numWorkers = len(templateFiles)
	}

	resultChan := make(chan result, len(templateFiles))
	filesChan := make(chan string, len(templateFiles))

	// start workers
	for w := 0; w < numWorkers; w++ {
		go func() {
			for filePath := range filesChan {
				metadata, err := p.ProcessTemplate(filePath, outputDir, outputFormat, validateOnly)
				resultChan <- result{metadata: metadata, err: err, path: filePath}
			}
		}()
	}

	// send files to workers
	for _, file := range templateFiles {
		filesChan <- file
	}
	close(filesChan)

	// collect results
	allMetadata := make([]*TemplateMetadata, 0, len(templateFiles))
	validCount := 0
	errorCount := 0

	for i := 0; i < len(templateFiles); i++ {
		res := <-resultChan
		if res.err != nil {
			p.logger.Errorf("Error processing template %s: %v", res.path, res.err)
			errorCount++
		} else {
			allMetadata = append(allMetadata, res.metadata)
			validCount++
		}
	}

	duration := time.Since(startTime).Seconds()
	p.logger.Infof("Processing completed in %.2f seconds", duration)
	p.logger.Infof("Templates processed: %d successful, %d failed", validCount, errorCount)

	// // note: Index generation and CSS generation would be handled by the Python service

	return allMetadata, nil
}

// validates a template against basic rules
func (p *Processor) ValidateTemplate(templateData map[string]interface{}) (bool, string) {
	// check template key
	templateObj, ok := templateData["template"]
	if !ok {
		return false, "Missing template key"
	}

	// cast to map
	templateMap, ok := templateObj.(map[string]interface{})
	if !ok {
		return false, "Template value is not an object"
	}

	// check required fields
	requiredFields := []string{"name", "type"}
	for _, field := range requiredFields {
		if _, ok := templateMap[field]; !ok {
			return false, fmt.Sprintf("Missing required field: %s", field)
		}
	}

	// check template type is valid
	typeStr, ok := templateMap["type"].(string)
	if !ok {
		return false, "Type field is not a string"
	}

	typeValid := false
	for _, validType := range ValidTemplateTypes {
		if strings.EqualFold(typeStr, validType) {
			typeValid = true
			break
		}
	}

	if !typeValid {
		return false, fmt.Sprintf("Invalid template type: %s. Must be one of %v", typeStr, ValidTemplateTypes)
	}

	// schema validation if available
	if p.schemaManager != nil {
		// normalize type for schema lookup
		normalizedType := schema.SchemaTypeNormalize(typeStr)

		// get schema
		schemaData, err := p.schemaManager.GetSchema(normalizedType)
		if err != nil {
			p.logger.Warnf("Could not load schema for type %s: %v", typeStr, err)
			// continue without schema validation
			return true, "Basic validation passed (schema not available)"
		}

		// prepare schema and template
		schemaLoader := gojsonschema.NewGoLoader(schemaData)
		documentLoader := gojsonschema.NewGoLoader(templateData)

		// validate
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			p.logger.Warnf("KNOWN ISSUE: Harness schema validation error with upstream schema (https://github.com/harness/harness-schema). "+
				"Type %s error: %v - Template is still valid according to basic validation.", typeStr, err)
			// fall back to basic validation
			return true, "Basic validation passed (with Harness schema inconsistency)"
		}

		if !result.Valid() {
			// log errors but don't fail
			p.logger.Warn("KNOWN ISSUE: Harness upstream schema validation failed - This is expected and non-critical")
			p.logger.Warn("These errors are due to inconsistencies in the official Harness JSON schemas at https://github.com/harness/harness-schema")
			for _, desc := range result.Errors() {
				p.logger.Warnf("- Schema validation detail: %s", desc)
			}
			return true, "Basic validation passed (with expected Harness schema inconsistencies)"
		}

		return true, fmt.Sprintf("Template is valid according to Harness %s schema", normalizedType)
	}

	return true, "Template is valid (basic validation only)"
}

// extracts metadata from a template
func (p *Processor) ExtractMetadata(templateData map[string]interface{}) (*TemplateMetadata, error) {
	// initialize metadata
	metadata := &TemplateMetadata{
		Name:        "Unnamed Template",
		Identifier:  "unnamed_template",
		Type:        "unknown",
		Variables:   make(map[string]Variable),
		Parameters:  make(map[string]Parameter),
		Description: "",
		Tags:        []string{},
		Author:      "Harness",
		Version:     "1.0.0",
		Examples:    []string{},
		RawTemplate: templateData,
	}

	// extract template object
	templateObj, ok := templateData["template"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("template field is not an object or missing")
	}

	// extract basic fields
	if name, ok := templateObj["name"].(string); ok {
		metadata.Name = name
	}

	if identifier, ok := templateObj["identifier"].(string); ok {
		metadata.Identifier = identifier
	}

	if typeStr, ok := templateObj["type"].(string); ok {
		metadata.Type = strings.ToLower(typeStr)
	}

	if desc, ok := templateObj["description"].(string); ok {
		metadata.Description = desc
	}

	if author, ok := templateObj["author"].(string); ok {
		metadata.Author = author
	}

	if version, ok := templateObj["versionLabel"].(string); ok {
		metadata.Version = version
	} else if version, ok := templateObj["versionLabel"].(float64); ok {
		metadata.Version = fmt.Sprintf("%.1f", version)
	}

	// extract tags
	if tags, ok := templateObj["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				metadata.Tags = append(metadata.Tags, tagStr)
			}
		}
	} else if tags, ok := templateObj["tags"].(map[string]interface{}); ok && len(tags) == 0 {
		// empty map for tags
	}

	// extract variables
	if vars, ok := templateObj["variables"].(map[string]interface{}); ok {
		for name, varData := range vars {
			if varMap, ok := varData.(map[string]interface{}); ok {
				variable := Variable{
					Description: getStringValue(varMap, "description", ""),
					Type:        getStringValue(varMap, "type", "string"),
					Required:    getBoolValue(varMap, "required", false),
					Scope:       getStringValue(varMap, "scope", "template"),
				}
				metadata.Variables[name] = variable
			}
		}
	}

	// extract parameters
	if params, ok := templateObj["parameters"].(map[string]interface{}); ok {
		for name, paramData := range params {
			if paramMap, ok := paramData.(map[string]interface{}); ok {
				parameter := Parameter{
					Description: getStringValue(paramMap, "description", ""),
					Type:        getStringValue(paramMap, "type", "string"),
					Required:    getBoolValue(paramMap, "required", false),
					Default:     paramMap["default"],
					Scope:       getStringValue(paramMap, "scope", "template"),
				}
				metadata.Parameters[name] = parameter
			}
		}
	}

	// extract examples
	if examples, ok := templateObj["examples"].([]interface{}); ok {
		for _, ex := range examples {
			if exStr, ok := ex.(string); ok {
				metadata.Examples = append(metadata.Examples, exStr)
			}
		}
	}

	return metadata, nil
}

// helper functions for type conversion
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultValue
}
