package template_test

import (
	"testing"

	"github.com/ka1ne/template-doc-gen/pkg/template"
)

func TestExtractMetadata(t *testing.T) {
	processor := template.NewProcessor(nil)

	// test complete template
	completeTemplate := map[string]interface{}{
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

	metadata, err := processor.ExtractMetadata(completeTemplate)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// check fields
	if metadata.Name != "Complete Template" {
		t.Errorf("expected Name='Complete Template', got '%s'", metadata.Name)
	}
	if metadata.Type != "pipeline" {
		t.Errorf("expected Type='pipeline', got '%s'", metadata.Type)
	}
	if metadata.Description != "Test template with all fields" {
		t.Errorf("expected Description='Test template with all fields', got '%s'", metadata.Description)
	}
	if metadata.Author != "Test Author" {
		t.Errorf("expected Author='Test Author', got '%s'", metadata.Author)
	}
	if metadata.Version != "1.2.3" {
		t.Errorf("expected Version='1.2.3', got '%s'", metadata.Version)
	}
	if len(metadata.Tags) != 2 || metadata.Tags[0] != "tag1" || metadata.Tags[1] != "tag2" {
		t.Errorf("tags not extracted correctly: %v", metadata.Tags)
	}

	// check variables
	if len(metadata.Variables) != 1 {
		t.Errorf("expected 1 variable, got %d", len(metadata.Variables))
	}
	variable, exists := metadata.Variables["var1"]
	if !exists {
		t.Error("variable 'var1' not found")
	} else {
		if variable.Description != "First variable" {
			t.Errorf("expected variable description='First variable', got '%s'", variable.Description)
		}
		if variable.Type != "string" {
			t.Errorf("expected variable type='string', got '%s'", variable.Type)
		}
		if !variable.Required {
			t.Error("expected variable required=true")
		}
		if variable.Scope != "pipeline" {
			t.Errorf("expected variable scope='pipeline', got '%s'", variable.Scope)
		}
	}

	// check parameters
	if len(metadata.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(metadata.Parameters))
	}
	parameter, exists := metadata.Parameters["param1"]
	if !exists {
		t.Error("parameter 'param1' not found")
	} else {
		if parameter.Description != "First parameter" {
			t.Errorf("expected parameter description='First parameter', got '%s'", parameter.Description)
		}
		if parameter.Type != "boolean" {
			t.Errorf("expected parameter type='boolean', got '%s'", parameter.Type)
		}
		if parameter.Required {
			t.Error("expected parameter required=false")
		}
		if parameter.Default != true {
			t.Errorf("expected parameter default=true, got %v", parameter.Default)
		}
		if parameter.Scope != "pipeline" {
			t.Errorf("expected parameter scope='pipeline', got '%s'", parameter.Scope)
		}
	}

	// check examples
	if len(metadata.Examples) != 2 || metadata.Examples[0] != "Example 1" || metadata.Examples[1] != "Example 2" {
		t.Errorf("examples not extracted correctly: %v", metadata.Examples)
	}

	// test missing template field
	badTemplate := map[string]interface{}{
		"name": "Bad Template",
	}
	_, err = processor.ExtractMetadata(badTemplate)
	if err == nil {
		t.Error("expected error for template with missing template field")
	}

	// test numeric version label
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
		t.Errorf("expected Version='1.0', got '%s'", metadata.Version)
	}

	// test empty map tags
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
		t.Errorf("expected empty tags slice, got %v", metadata.Tags)
	}
}
