package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SchemaManager handles fetching, caching, and providing access to JSON schemas
type SchemaManager struct {
	logger      *logrus.Logger
	schemaCache map[string]map[string]interface{}
	cacheMutex  sync.RWMutex
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(logger *logrus.Logger) *SchemaManager {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}
	return &SchemaManager{
		logger:      logger,
		schemaCache: make(map[string]map[string]interface{}),
	}
}

// GetSchema fetches a schema from cache or from GitHub
func (m *SchemaManager) GetSchema(schemaType string) (map[string]interface{}, error) {
	// Convert to lowercase for consistency
	schemaType = schemaTypeNormalize(schemaType)

	// Check cache first
	m.cacheMutex.RLock()
	schema, exists := m.schemaCache[schemaType]
	m.cacheMutex.RUnlock()

	if exists {
		return schema, nil
	}

	// Fetch from GitHub if not in cache
	return m.fetchSchema(schemaType)
}

// fetchSchema fetches a schema from GitHub
func (m *SchemaManager) fetchSchema(schemaType string) (map[string]interface{}, error) {
	// Map schema types to their corresponding files
	schemaMapping := map[string]string{
		"pipeline":  "pipeline.json",
		"stage":     "template.json",
		"step":      "template.json",
		"stepgroup": "template.json", // Stages, Steps, StepGroups are defined in template.json
		"trigger":   "trigger.json",
	}

	schemaFile, ok := schemaMapping[schemaType]
	if !ok {
		// Default to pipeline schema if not found
		schemaFile = "pipeline.json"
	}

	m.logger.Debugf("Using schema file %s for type %s", schemaFile, schemaType)

	// Use v1 schema URL
	schemaURL := fmt.Sprintf("https://raw.githubusercontent.com/harness/harness-schema/main/v1/%s", schemaFile)
	m.logger.Debugf("Fetching schema from %s", schemaURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make HTTP request
	resp, err := client.Get(schemaURL)
	if err != nil {
		m.logger.Errorf("Error fetching schema: %v", err)
		return nil, fmt.Errorf("error fetching schema: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		m.logger.Errorf("Failed to fetch schema: %d from URL %s", resp.StatusCode, schemaURL)
		return nil, fmt.Errorf("failed to fetch schema: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.logger.Errorf("Error reading schema response: %v", err)
		return nil, fmt.Errorf("error reading schema response: %w", err)
	}

	// Parse JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(body, &schema); err != nil {
		m.logger.Errorf("Error parsing schema JSON: %v", err)
		return nil, fmt.Errorf("error parsing schema JSON: %w", err)
	}

	// Cache the schema
	m.cacheMutex.Lock()
	m.schemaCache[schemaType] = schema
	m.cacheMutex.Unlock()

	m.logger.Debugf("Successfully fetched %s schema using %s", schemaType, schemaFile)
	return schema, nil
}

// schemaTypeNormalize normalizes a schema type to match expected values
func schemaTypeNormalize(schemaType string) string {
	switch schemaType {
	case "Pipeline", "PIPELINE":
		return "pipeline"
	case "Stage", "STAGE":
		return "stage"
	case "StepGroup", "STEPGROUP", "Step Group":
		return "stepgroup"
	case "Step", "STEP":
		return "step"
	default:
		return schemaType
	}
}

// SchemaTypeNormalize normalizes a schema type to match expected values - exported version
func SchemaTypeNormalize(schemaType string) string {
	return schemaTypeNormalize(schemaType)
}
