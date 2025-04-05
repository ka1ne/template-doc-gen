package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ka1ne/template-doc-gen/pkg/schema"
	"github.com/sirupsen/logrus"
)

// helper to create test template file
func createTestTemplateFile(t *testing.T, dir string, content string) string {
	path := filepath.Join(dir, "test-template.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template file: %v", err)
	}
	return path
}

func TestNewProcessor(t *testing.T) {
	// test with nil logger
	processor := NewProcessor(nil)
	if processor == nil {
		t.Error("Expected non-nil Processor when passing nil logger")
	}
	if processor.logger == nil {
		t.Error("Expected non-nil logger in Processor")
	}

	// test with provided logger
	logger := logrus.New()
	processor = NewProcessor(logger)
	if processor.logger != logger {
		t.Error("Expected provided logger to be used")
	}
}

func TestValidateTemplate(t *testing.T) {
	processor := NewProcessor(nil)

	tests := []struct {
		name        string
		template    map[string]interface{}
		expectValid bool
	}{
		{
			name: "Valid pipeline template",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Pipeline",
					"type": "Pipeline",
				},
			},
			expectValid: true,
		},
		{
			name: "Valid stage template",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Stage",
					"type": "Stage",
				},
			},
			expectValid: true,
		},
		{
			name: "Missing template key",
			template: map[string]interface{}{
				"name": "Invalid Template",
			},
			expectValid: false,
		},
		{
			name: "Missing name field",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"type": "Pipeline",
				},
			},
			expectValid: false,
		},
		{
			name: "Missing type field",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Pipeline",
				},
			},
			expectValid: false,
		},
		{
			name: "Invalid type field",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Pipeline",
					"type": "Invalid",
				},
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, _ := processor.ValidateTemplate(test.template)
			if valid != test.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", test.expectValid, valid)
			}
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	processor := NewProcessor(nil)

	// // test extraction of a complete template
	template := map[string]interface{}{
		"template": map[string]interface{}{
			"name":         "Complete Template",
			"type":         "Pipeline",
			"description":  "Test template with all fields",
			"author":       "Test Author",
			"versionLabel": "1.2.3",
			"tags": []interface{}{
				"tag1",
				"tag2",
			},
			"variables": map[string]interface{}{
				"var1": map[string]interface{}{
					"description": "First variable",
					"type":        "string",
					"required":    true,
					"scope":       "pipeline",
				},
			},
			"parameters": map[string]interface{}{
				"param1": map[string]interface{}{
					"description": "First parameter",
					"type":        "boolean",
					"required":    false,
					"default":     true,
					"scope":       "pipeline",
				},
			},
			"examples": []interface{}{
				"Example 1",
				"Example 2",
			},
		},
	}

	metadata, err := processor.ExtractMetadata(template)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// // verify extracted fields
	if metadata.Name != "Complete Template" {
		t.Errorf("Expected Name='Complete Template', got '%s'", metadata.Name)
	}
	if metadata.Type != "pipeline" {
		t.Errorf("Expected Type='pipeline', got '%s'", metadata.Type)
	}
	if metadata.Description != "Test template with all fields" {
		t.Errorf("Expected Description='Test template with all fields', got '%s'", metadata.Description)
	}
	if metadata.Author != "Test Author" {
		t.Errorf("Expected Author='Test Author', got '%s'", metadata.Author)
	}
	if metadata.Version != "1.2.3" {
		t.Errorf("Expected Version='1.2.3', got '%s'", metadata.Version)
	}
	if len(metadata.Tags) != 2 || metadata.Tags[0] != "tag1" || metadata.Tags[1] != "tag2" {
		t.Errorf("Tags not extracted correctly: %v", metadata.Tags)
	}

	// // verify variables
	if len(metadata.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(metadata.Variables))
	}
	variable, exists := metadata.Variables["var1"]
	if !exists {
		t.Error("Variable 'var1' not found")
	} else {
		if variable.Description != "First variable" {
			t.Errorf("Expected variable description='First variable', got '%s'", variable.Description)
		}
		if variable.Type != "string" {
			t.Errorf("Expected variable type='string', got '%s'", variable.Type)
		}
		if !variable.Required {
			t.Error("Expected variable required=true")
		}
		if variable.Scope != "pipeline" {
			t.Errorf("Expected variable scope='pipeline', got '%s'", variable.Scope)
		}
	}

	// // verify parameters
	if len(metadata.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(metadata.Parameters))
	}
	parameter, exists := metadata.Parameters["param1"]
	if !exists {
		t.Error("Parameter 'param1' not found")
	} else {
		if parameter.Description != "First parameter" {
			t.Errorf("Expected parameter description='First parameter', got '%s'", parameter.Description)
		}
		if parameter.Type != "boolean" {
			t.Errorf("Expected parameter type='boolean', got '%s'", parameter.Type)
		}
		if parameter.Required {
			t.Error("Expected parameter required=false")
		}
		if parameter.Default != true {
			t.Errorf("Expected parameter default=true, got %v", parameter.Default)
		}
		if parameter.Scope != "pipeline" {
			t.Errorf("Expected parameter scope='pipeline', got '%s'", parameter.Scope)
		}
	}

	// // verify examples
	if len(metadata.Examples) != 2 || metadata.Examples[0] != "Example 1" || metadata.Examples[1] != "Example 2" {
		t.Errorf("Examples not extracted correctly: %v", metadata.Examples)
	}

	// // test extraction with missing template field
	badTemplate := map[string]interface{}{
		"name": "Bad Template",
	}
	_, err = processor.ExtractMetadata(badTemplate)
	if err == nil {
		t.Error("Expected error for template with missing template field")
	}

	// // test extraction with numeric version label
	numericVersionTemplate := map[string]interface{}{
		"template": map[string]interface{}{
			"name":         "Numeric Version Template",
			"type":         "Pipeline",
			"versionLabel": 1.0,
		},
	}
	metadata, err = processor.ExtractMetadata(numericVersionTemplate)
	if err != nil {
		t.Fatalf("ExtractMetadata failed for numeric version: %v", err)
	}
	if metadata.Version != "1.0" {
		t.Errorf("Expected Version='1.0', got '%s'", metadata.Version)
	}

	// // test extraction with empty map tags
	emptyTagsTemplate := map[string]interface{}{
		"template": map[string]interface{}{
			"name": "Empty Tags Template",
			"type": "Pipeline",
			"tags": map[string]interface{}{},
		},
	}
	metadata, err = processor.ExtractMetadata(emptyTagsTemplate)
	if err != nil {
		t.Fatalf("ExtractMetadata failed for empty tags: %v", err)
	}
	if len(metadata.Tags) != 0 {
		t.Errorf("Expected empty tags slice, got %v", metadata.Tags)
	}
}

func TestProcessTemplate(t *testing.T) {
	processor := NewProcessor(nil)

	// // create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "template_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// // create output directory
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// // create a valid test template
	validTemplateContent := `
template:
  name: Valid Test Template
  type: Pipeline
  description: A test template
  versionLabel: 1.0.0
  author: Test Author
  tags:
    - test
    - pipeline
  variables:
    var1:
      description: A test variable
      type: string
      required: true
      scope: pipeline
  parameters:
    param1:
      description: A test parameter
      type: boolean
      required: false
      default: true
      scope: pipeline
`
	validTemplatePath := createTestTemplateFile(t, tempDir, validTemplateContent)

	// // create an invalid test template
	invalidTemplateContent := `
template:
  name: Invalid Test Template
  # Missing type field
  description: An invalid test template
`
	invalidTemplatePath := filepath.Join(tempDir, "invalid-template.yaml")
	if err := os.WriteFile(invalidTemplatePath, []byte(invalidTemplateContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid template file: %v", err)
	}

	// // create a malformed YAML template
	malformedTemplateContent := `
template:
  name: Malformed YAML
  type: [Pipeline
  description: Malformed YAML template
`
	malformedTemplatePath := filepath.Join(tempDir, "malformed-template.yaml")
	if err := os.WriteFile(malformedTemplatePath, []byte(malformedTemplateContent), 0644); err != nil {
		t.Fatalf("Failed to write malformed template file: %v", err)
	}

	// // test processing valid template with validate only
	metadata, err := processor.ProcessTemplate(validTemplatePath, outputDir, "html", true)
	if err != nil {
		t.Errorf("Expected no error processing valid template (validate only), got: %v", err)
	}
	if metadata == nil {
		t.Fatal("Expected non-nil metadata for valid template")
	}
	if metadata.Name != "Valid Test Template" {
		t.Errorf("Expected Name='Valid Test Template', got '%s'", metadata.Name)
	}

	// // test processing invalid template
	metadata, err = processor.ProcessTemplate(invalidTemplatePath, outputDir, "html", false)
	if err == nil {
		t.Error("Expected error processing invalid template, got nil")
	}
	if metadata != nil {
		t.Errorf("Expected nil metadata for invalid template, got: %v", metadata)
	}

	// // test processing malformed YAML template
	metadata, err = processor.ProcessTemplate(malformedTemplatePath, outputDir, "html", false)
	if err == nil {
		t.Error("Expected error processing malformed YAML template, got nil")
	}
	if metadata != nil {
		t.Errorf("Expected nil metadata for malformed template, got: %v", metadata)
	}
}

func TestHelperFunctions(t *testing.T) {
	// // test getStringValue
	stringMap := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	if val := getStringValue(stringMap, "key1", "default"); val != "value1" {
		t.Errorf("Expected getStringValue to return 'value1', got '%s'", val)
	}
	if val := getStringValue(stringMap, "key2", "default"); val != "default" {
		t.Errorf("Expected getStringValue to return 'default', got '%s'", val)
	}
	if val := getStringValue(stringMap, "notfound", "default"); val != "default" {
		t.Errorf("Expected getStringValue to return 'default', got '%s'", val)
	}

	// // test getBoolValue
	boolMap := map[string]interface{}{
		"key1": true,
		"key2": false,
		"key3": "true",
		"key4": 1,
	}

	if val := getBoolValue(boolMap, "key1", false); !val {
		t.Error("Expected getBoolValue to return true")
	}
	if val := getBoolValue(boolMap, "key2", true); val {
		t.Error("Expected getBoolValue to return false")
	}
	if val := getBoolValue(boolMap, "key3", false); val {
		t.Error("Expected getBoolValue to return false (default) for non-bool 'true' string")
	}
	if val := getBoolValue(boolMap, "notfound", true); !val {
		t.Error("Expected getBoolValue to return true (default)")
	}
}

func TestProcessAllTemplates(t *testing.T) {
	processor := NewProcessor(nil)

	// // create temporary directory structure
	tempDir, err := os.MkdirTemp("", "all_templates_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// // create subdirectories for different template types
	pipelineDir := filepath.Join(tempDir, "pipeline")
	stageDir := filepath.Join(tempDir, "stage")
	if err := os.MkdirAll(pipelineDir, 0755); err != nil {
		t.Fatalf("Failed to create pipeline directory: %v", err)
	}
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		t.Fatalf("Failed to create stage directory: %v", err)
	}

	// // create output directory
	outputDir := filepath.Join(tempDir, "output")

	// // create test templates
	pipeline1Content := `
template:
  name: Pipeline 1
  type: Pipeline
  description: Test pipeline 1
`
	pipeline1Path := filepath.Join(pipelineDir, "pipeline1.yaml")
	if err := os.WriteFile(pipeline1Path, []byte(pipeline1Content), 0644); err != nil {
		t.Fatalf("Failed to write pipeline1 file: %v", err)
	}

	pipeline2Content := `
template:
  name: Pipeline 2
  type: Pipeline
  description: Test pipeline 2
`
	pipeline2Path := filepath.Join(pipelineDir, "pipeline2.yaml")
	if err := os.WriteFile(pipeline2Path, []byte(pipeline2Content), 0644); err != nil {
		t.Fatalf("Failed to write pipeline2 file: %v", err)
	}

	stageContent := `
template:
  name: Stage 1
  type: Stage
  description: Test stage
`
	stagePath := filepath.Join(stageDir, "stage1.yaml")
	if err := os.WriteFile(stagePath, []byte(stageContent), 0644); err != nil {
		t.Fatalf("Failed to write stage file: %v", err)
	}

	// // create an invalid template to test error handling
	invalidContent := `
template:
  name: Invalid
  # Missing type
`
	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	// // test processing all templates (validate only)
	metadataList, err := processor.ProcessAllTemplates(tempDir, outputDir, "html", true)
	if err != nil {
		t.Fatalf("Failed to process all templates: %v", err)
	}

	// // we should have 3 valid templates (pipeline1, pipeline2, stage1)
	// // the invalid.yaml should be skipped with an error
	if len(metadataList) != 3 {
		t.Errorf("Expected 3 valid templates, got %d", len(metadataList))
	}

	// // check if output directories are created in non-validate mode
	metadataList, err = processor.ProcessAllTemplates(tempDir, outputDir, "json", false)
	if err != nil {
		t.Fatalf("Failed to process all templates in JSON mode: %v", err)
	}

	// // check if output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// // check if type subdirectories were created
	pipelineOutputDir := filepath.Join(outputDir, "pipeline")
	stageOutputDir := filepath.Join(outputDir, "stage")
	if _, err := os.Stat(pipelineOutputDir); os.IsNotExist(err) {
		t.Error("Pipeline output directory was not created")
	}
	if _, err := os.Stat(stageOutputDir); os.IsNotExist(err) {
		t.Error("Stage output directory was not created")
	}

	// // test with a single file
	singleFileMetadata, err := processor.ProcessAllTemplates(pipeline1Path, outputDir, "html", true)
	if err != nil {
		t.Fatalf("Failed to process single template file: %v", err)
	}
	if len(singleFileMetadata) != 1 {
		t.Errorf("Expected 1 metadata for single file, got %d", len(singleFileMetadata))
	}

	// // test with non-existent path
	_, err = processor.ProcessAllTemplates("/path/does/not/exist", outputDir, "html", true)
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	// // test with non-YAML file
	nonYamlPath := filepath.Join(tempDir, "not-yaml.txt")
	if err := os.WriteFile(nonYamlPath, []byte("This is not YAML"), 0644); err != nil {
		t.Fatalf("Failed to write non-YAML file: %v", err)
	}
	_, err = processor.ProcessAllTemplates(nonYamlPath, outputDir, "html", true)
	if err == nil {
		t.Error("Expected error for non-YAML file, got nil")
	}
}

// // testConcurrentProcessing tests that concurrent template processing works correctly
func TestConcurrentProcessing(t *testing.T) {
	// // create a test processor
	logger := logrus.New()
	logger.SetOutput(io.Discard) // // silence logging for tests
	processor := NewProcessor(logger)

	// // create a temporary directory for templates
	tempDir, err := os.MkdirTemp("", "concurrent_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// // create multiple test templates
	numTemplates := 10
	for i := 0; i < numTemplates; i++ {
		templateContent := fmt.Sprintf(`template:
  name: "Test Template %d"
  type: "Pipeline"
  description: "Test template for concurrent processing"
  versionLabel: "1.0.0"
`, i)

		templatePath := filepath.Join(tempDir, fmt.Sprintf("template_%d.yaml", i))
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}
	}

	// // process templates concurrently
	start := time.Now()
	results, err := processor.ProcessAllTemplates(tempDir, "/tmp/output", "json", true)
	duration := time.Since(start)

	// // verify results
	if err != nil {
		t.Fatalf("ProcessAllTemplates failed: %v", err)
	}
	if len(results) != numTemplates {
		t.Errorf("Expected %d results, got %d", numTemplates, len(results))
	}

	// // ensure all templates have unique names
	names := make(map[string]bool)
	for _, result := range results {
		if _, exists := names[result.Name]; exists {
			t.Errorf("Duplicate template name found: %s", result.Name)
		}
		names[result.Name] = true
	}

	t.Logf("Processed %d templates in %v", numTemplates, duration)
}

// // testSchemaValidation tests that schema validation works correctly
func TestSchemaValidation(t *testing.T) {
	// // skip detailed schema validation tests in CI environments
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping schema validation tests in CI environment")
	}

	// // create a test processor with schema manager
	logger := logrus.New()
	logger.SetOutput(io.Discard) // // silence logging for tests
	processor := NewProcessor(logger)

	// // create a real schema manager - will actually use the Harness schema repo
	schemaManager := schema.NewSchemaManager(logger)
	processor.SetSchemaManager(schemaManager)

	// // test cases - only test basic functionality, not detailed schema validation
	tests := []struct {
		name       string
		template   map[string]interface{}
		wantValid  bool
		wantErrMsg string
	}{
		{
			name: "Valid pipeline template structure",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Pipeline",
					"type": "Pipeline",
				},
			},
			wantValid: true,
		},
		{
			name: "Valid stage template structure",
			template: map[string]interface{}{
				"template": map[string]interface{}{
					"name": "Test Stage",
					"type": "Stage",
				},
			},
			wantValid: true,
		},
		{
			name: "Invalid template format",
			template: map[string]interface{}{
				"template": "Not an object",
			},
			wantValid:  false,
			wantErrMsg: "not an object",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, msg := processor.ValidateTemplate(tc.template)
			if valid != tc.wantValid {
				t.Errorf("ValidateTemplate() = %v, want %v, msg: %s", valid, tc.wantValid, msg)
			}
			if !tc.wantValid && tc.wantErrMsg != "" && !strings.Contains(msg, tc.wantErrMsg) {
				t.Errorf("Error message '%s' does not contain expected '%s'", msg, tc.wantErrMsg)
			}
		})
	}
}
