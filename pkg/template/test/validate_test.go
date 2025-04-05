package template_test

import (
	"testing"

	"github.com/ka1ne/template-doc-gen/pkg/template"
)

func TestValidateTemplate(t *testing.T) {
	processor := template.NewProcessor(nil)

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
				t.Errorf("expected valid=%v, got valid=%v", test.expectValid, valid)
			}
		})
	}
}
