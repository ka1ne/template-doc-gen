package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ka1ne/template-doc-gen/pkg/html"
	"github.com/ka1ne/template-doc-gen/pkg/schema"
	"github.com/ka1ne/template-doc-gen/pkg/template"
	"github.com/ka1ne/template-doc-gen/pkg/utils"
)

func main() {
	// Parse configuration
	config := utils.NewConfig()

	// Validate configuration
	if err := config.Validate(); err != nil {
		config.Logger.Fatalf("Configuration error: %v", err)
	}

	// Create schema manager
	schemaManager := schema.NewSchemaManager(config.Logger)

	// Create template processor with schema manager
	processor := template.NewProcessor(config.Logger)

	// Set the schema manager in the processor
	processor.SetSchemaManager(schemaManager)

	// Process all templates
	metadata, err := processor.ProcessAllTemplates(
		config.SourceDir,
		config.OutputDir,
		config.OutputFormat,
		config.ValidateOnly,
	)
	if err != nil {
		config.Logger.Fatalf("Error processing templates: %v", err)
	}

	// Output summary
	config.Logger.Infof("Successfully processed %d templates", len(metadata))

	// Generate documentation if not in validate-only mode
	if !config.ValidateOnly {
		if config.OutputFormat == "html" {
			// Create HTML generator
			htmlGenerator := html.NewGenerator(config.Logger)

			// Generate HTML documentation
			if err := htmlGenerator.GenerateDocumentation(metadata, config.OutputDir); err != nil {
				config.Logger.Errorf("Error generating HTML documentation: %v", err)
			} else {
				config.Logger.Infof("HTML documentation generated in %s", config.OutputDir)
			}
		}

		// If processing JSON format, output metadata to a file
		if config.OutputFormat == "json" {
			metadataFile := fmt.Sprintf("%s/metadata.json", config.OutputDir)
			jsonData, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				config.Logger.Errorf("Error generating JSON metadata: %v", err)
			} else {
				if err := os.WriteFile(metadataFile, jsonData, 0644); err != nil {
					config.Logger.Errorf("Error writing JSON metadata: %v", err)
				} else {
					config.Logger.Infof("Metadata written to %s", metadataFile)
				}
			}
		}
	}

	// Exit with success
	os.Exit(0)
}
