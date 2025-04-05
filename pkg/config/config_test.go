package config

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDefaultConfig(t *testing.T) {
	// Save original environment
	origSourceDir := os.Getenv("SOURCE_DIR")
	origOutputDir := os.Getenv("OUTPUT_DIR")
	origFormat := os.Getenv("FORMAT")
	origValidateOnly := os.Getenv("VALIDATE_ONLY")
	origVerbose := os.Getenv("VERBOSE")

	// Restore environment after test
	defer func() {
		os.Setenv("SOURCE_DIR", origSourceDir)
		os.Setenv("OUTPUT_DIR", origOutputDir)
		os.Setenv("FORMAT", origFormat)
		os.Setenv("VALIDATE_ONLY", origValidateOnly)
		os.Setenv("VERBOSE", origVerbose)
	}()

	// Test default values
	os.Unsetenv("SOURCE_DIR")
	os.Unsetenv("OUTPUT_DIR")
	os.Unsetenv("FORMAT")
	os.Unsetenv("VALIDATE_ONLY")
	os.Unsetenv("VERBOSE")

	config := DefaultConfig()
	if config.SourceDir != "templates" {
		t.Errorf("Expected default SourceDir to be 'templates', got '%s'", config.SourceDir)
	}
	if config.OutputDir != "docs/output" {
		t.Errorf("Expected default OutputDir to be 'docs/output', got '%s'", config.OutputDir)
	}
	if config.OutputFormat != "html" {
		t.Errorf("Expected default OutputFormat to be 'html', got '%s'", config.OutputFormat)
	}
	if config.ValidateOnly != false {
		t.Errorf("Expected default ValidateOnly to be false, got %v", config.ValidateOnly)
	}
	if config.Verbose != false {
		t.Errorf("Expected default Verbose to be false, got %v", config.Verbose)
	}
	if config.Concurrency != 4 {
		t.Errorf("Expected default Concurrency to be 4, got %d", config.Concurrency)
	}
	if config.Logger == nil {
		t.Error("Expected Logger to be initialized")
	}

	// Test with environment variables
	os.Setenv("SOURCE_DIR", "custom-templates")
	os.Setenv("OUTPUT_DIR", "custom-output")
	os.Setenv("FORMAT", "json")
	os.Setenv("VALIDATE_ONLY", "true")
	os.Setenv("VERBOSE", "true")

	config = DefaultConfig()
	if config.SourceDir != "custom-templates" {
		t.Errorf("Expected SourceDir to be 'custom-templates', got '%s'", config.SourceDir)
	}
	if config.OutputDir != "custom-output" {
		t.Errorf("Expected OutputDir to be 'custom-output', got '%s'", config.OutputDir)
	}
	if config.OutputFormat != "json" {
		t.Errorf("Expected OutputFormat to be 'json', got '%s'", config.OutputFormat)
	}
	if config.ValidateOnly != true {
		t.Errorf("Expected ValidateOnly to be true, got %v", config.ValidateOnly)
	}
	if config.Verbose != true {
		t.Errorf("Expected Verbose to be true, got %v", config.Verbose)
	}
}

func TestWithArgs(t *testing.T) {
	// Create a default config
	config := DefaultConfig()

	// Test with command-line arguments
	args := []string{
		"-source=arg-templates",
		"-output=arg-output",
		"-format=markdown",
		"-validate=true",
		"-verbose=true",
		"-concurrency=8",
	}

	config = config.WithArgs(args)

	if config.SourceDir != "arg-templates" {
		t.Errorf("Expected SourceDir to be 'arg-templates', got '%s'", config.SourceDir)
	}
	if config.OutputDir != "arg-output" {
		t.Errorf("Expected OutputDir to be 'arg-output', got '%s'", config.OutputDir)
	}
	if config.OutputFormat != "markdown" {
		t.Errorf("Expected OutputFormat to be 'markdown', got '%s'", config.OutputFormat)
	}
	if config.ValidateOnly != true {
		t.Errorf("Expected ValidateOnly to be true, got %v", config.ValidateOnly)
	}
	if config.Verbose != true {
		t.Errorf("Expected Verbose to be true, got %v", config.Verbose)
	}
	if config.Concurrency != 8 {
		t.Errorf("Expected Concurrency to be 8, got %d", config.Concurrency)
	}
}

func TestFromArgs(t *testing.T) {
	// Create config from args directly
	args := []string{"-source=fromargs-templates", "-concurrency=16"}

	config := FromArgs(args)

	if config.SourceDir != "fromargs-templates" {
		t.Errorf("Expected SourceDir to be 'fromargs-templates', got '%s'", config.SourceDir)
	}
	if config.Concurrency != 16 {
		t.Errorf("Expected Concurrency to be 16, got %d", config.Concurrency)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Save original environment
	origValue := os.Getenv("TEST_ENV_VAR")
	defer os.Setenv("TEST_ENV_VAR", origValue)

	// Test with unset variable
	os.Unsetenv("TEST_ENV_VAR")
	value := GetEnvOrDefault("TEST_ENV_VAR", "default_value")
	if value != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%s'", value)
	}

	// Test with set variable
	os.Setenv("TEST_ENV_VAR", "custom_value")
	value = GetEnvOrDefault("TEST_ENV_VAR", "default_value")
	if value != "custom_value" {
		t.Errorf("Expected custom value 'custom_value', got '%s'", value)
	}
}

func TestConfigValidate(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name        string
		sourceDir   string
		outputDir   string
		format      string
		concurrency int
		expectError bool
	}{
		{
			name:        "Valid configuration",
			sourceDir:   tempDir,
			outputDir:   "docs/output",
			format:      "html",
			concurrency: 4,
			expectError: false,
		},
		{
			name:        "Non-existent source directory",
			sourceDir:   "/path/does/not/exist",
			outputDir:   "docs/output",
			format:      "html",
			concurrency: 4,
			expectError: true,
		},
		{
			name:        "Invalid format",
			sourceDir:   tempDir,
			outputDir:   "docs/output",
			format:      "invalid",
			concurrency: 4,
			expectError: true,
		},
		{
			name:        "Invalid concurrency",
			sourceDir:   tempDir,
			outputDir:   "docs/output",
			format:      "html",
			concurrency: 0,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Config{
				SourceDir:    test.sourceDir,
				OutputDir:    test.outputDir,
				OutputFormat: test.format,
				Concurrency:  test.concurrency,
				Logger:       logrus.New(),
			}

			err := config.Validate()
			if test.expectError && err == nil {
				t.Error("Expected error but got nil")
			} else if !test.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
