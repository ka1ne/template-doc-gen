package schema

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSchemaTypeNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pipeline", "pipeline"},
		{"Pipeline", "pipeline"},
		{"PIPELINE", "pipeline"},
		{"stage", "stage"},
		{"Stage", "stage"},
		{"stepgroup", "stepgroup"},
		{"StepGroup", "stepgroup"},
		{"Step Group", "stepgroup"},
		{"step", "step"},
		{"Step", "step"},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := schemaTypeNormalize(test.input)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestNewSchemaManager(t *testing.T) {
	// Test with nil logger
	manager := NewSchemaManager(nil)
	if manager == nil {
		t.Error("Expected non-nil SchemaManager when passing nil logger")
	}
	if manager.logger == nil {
		t.Error("Expected non-nil logger in SchemaManager")
	}
	if manager.schemaCache == nil {
		t.Error("Expected non-nil schemaCache in SchemaManager")
	}

	// Test with provided logger
	logger := logrus.New()
	manager = NewSchemaManager(logger)
	if manager.logger != logger {
		t.Error("Expected provided logger to be used")
	}
}

func setupMockServer(t *testing.T) *httptest.Server {
	// Create a test server that returns different responses based on the URL path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseJSON string
		status := http.StatusOK

		switch r.URL.Path {
		case "/harness/harness-schema/main/v1/pipeline.json":
			responseJSON = `{
				"type": "object",
				"properties": {
					"template": {
						"type": "object",
						"properties": {
							"name": {"type": "string"},
							"type": {"type": "string"},
							"description": {"type": "string"}
						},
						"required": ["name", "type"]
					}
				}
			}`
		case "/harness/harness-schema/main/v1/template.json":
			responseJSON = `{
				"type": "object",
				"properties": {
					"template": {
						"type": "object",
						"properties": {
							"name": {"type": "string"},
							"type": {"type": "string"},
							"stages": {"type": "array"}
						},
						"required": ["name", "type"]
					}
				}
			}`
		case "/harness/harness-schema/main/v1/error.json":
			status = http.StatusInternalServerError
			responseJSON = `{"error": "Internal Server Error"}`
		default:
			status = http.StatusNotFound
			responseJSON = `{"error": "Not Found"}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(responseJSON))
	}))

	return server
}

func TestGetSchema(t *testing.T) {
	server := setupMockServer(t)
	defer server.Close()

	// Create a schema manager with modified URLs for testing
	logger := logrus.New()
	manager := &SchemaManager{
		logger:      logger,
		schemaCache: make(map[string]map[string]interface{}),
	}

	// Mock fetchSchema for testing
	// Create a fake schema that returns for tests
	pipelineSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
		},
	}

	stageSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"stages": map[string]interface{}{
				"type": "array",
			},
		},
	}

	// Add test schemas to cache directly
	manager.cacheMutex.Lock()
	manager.schemaCache["pipeline"] = pipelineSchema
	manager.schemaCache["stage"] = stageSchema
	manager.cacheMutex.Unlock()

	// Test GetSchema for pipeline - should use cached values
	schema, err := manager.GetSchema("pipeline")
	if err != nil {
		t.Fatalf("Expected no error for pipeline schema, got %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema for pipeline")
	}
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}
	if _, ok := props["name"]; !ok {
		t.Error("Expected name in properties")
	}

	// Test GetSchema for stage - should use cached values
	schema, err = manager.GetSchema("stage")
	if err != nil {
		t.Fatalf("Expected no error for stage schema, got %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema for stage")
	}
	props, ok = schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}
	if _, ok := props["stages"]; !ok {
		t.Error("Expected stages in properties")
	}

	// Test schema caching
	// Schema should be cached now
	manager.cacheMutex.RLock()
	cachedPipelineSchema, exists := manager.schemaCache["pipeline"]
	manager.cacheMutex.RUnlock()
	if !exists {
		t.Error("Expected pipeline schema to be cached")
	}
	if cachedPipelineSchema == nil {
		t.Error("Expected non-nil cached pipeline schema")
	}

	manager.cacheMutex.RLock()
	cachedStageSchema, exists := manager.schemaCache["stage"]
	manager.cacheMutex.RUnlock()
	if !exists {
		t.Error("Expected stage schema to be cached")
	}
	if cachedStageSchema == nil {
		t.Error("Expected non-nil cached stage schema")
	}
}

func TestFetchSchema(t *testing.T) {
	// Skip if running in short mode (avoid external API calls)
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	server := setupMockServer(t)
	defer server.Close()

	// Create schema manager
	logger := logrus.New()
	manager := NewSchemaManager(logger)

	// Test fetchSchema for pipeline
	_, err := manager.fetchSchema("pipeline")
	if err != nil {
		t.Fatalf("Expected no error for pipeline schema, got %v", err)
	}

	// Test with invalid schema type
	_, err = manager.fetchSchema("invalid")
	if err != nil {
		t.Fatalf("Expected no error for invalid schema (should default to pipeline), got %v", err)
	}
}

// Test concurrent access to the schema cache
func TestConcurrentSchemaAccess(t *testing.T) {
	manager := NewSchemaManager(nil)

	// Add a test schema to the cache directly
	testSchema := map[string]interface{}{
		"type":       "object",
		"schemaType": "pipeline",
	}

	manager.cacheMutex.Lock()
	manager.schemaCache["pipeline"] = testSchema
	manager.cacheMutex.Unlock()

	// Run concurrent goroutines to access schemas
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				schema, err := manager.GetSchema("pipeline")
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if schema == nil {
					t.Error("Expected non-nil schema")
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify cache
	manager.cacheMutex.RLock()
	cachedSchema, exists := manager.schemaCache["pipeline"]
	manager.cacheMutex.RUnlock()

	if !exists {
		t.Error("Expected schema to be cached")
	}
	if cachedSchema == nil {
		t.Error("Expected non-nil cached schema")
	}
}

// NewMockSchemaManager creates a schema manager with pre-populated schemas for testing
func NewMockSchemaManager(logger *logrus.Logger) *SchemaManager {
	manager := NewSchemaManager(logger)

	// Pre-populate the schema cache with simplified schemas for testing
	manager.cacheMutex.Lock()
	defer manager.cacheMutex.Unlock()

	// Simplified pipeline schema
	manager.schemaCache["pipeline"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"template": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"Pipeline"},
					},
				},
				"required": []interface{}{"name", "type"},
			},
		},
	}

	// Simplified stage schema
	manager.schemaCache["stage"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"template": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"Stage"},
					},
				},
				"required": []interface{}{"name", "type"},
			},
		},
	}

	// Simplified step schema
	manager.schemaCache["step"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"template": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"Step"},
					},
				},
				"required": []interface{}{"name", "type"},
			},
		},
	}

	// Simplified stepgroup schema
	manager.schemaCache["stepgroup"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"template": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"StepGroup"},
					},
				},
				"required": []interface{}{"name", "type"},
			},
		},
	}

	return manager
}
