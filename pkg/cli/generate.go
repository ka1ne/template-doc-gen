package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ka1ne/template-doc-gen/pkg/html"
	"github.com/ka1ne/template-doc-gen/pkg/schema"
	"github.com/ka1ne/template-doc-gen/pkg/template"
	"github.com/spf13/cobra"
)

// Command line flags for the generate command
var (
	generateSourceDir   string
	generateOutputDir   string
	generateFormat      string
	generateConcurrency int
)

// newGenerateCommand creates a command to generate documentation
func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "Generate documentation from templates",
		Long:    `Process templates and generate documentation in the specified format (html, json).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(cmd.Context())
		},
	}

	// Add flags specific to the generate command
	cmd.Flags().StringVarP(&generateSourceDir, "source", "s", cfg.SourceDir, "Source directory containing templates")
	cmd.Flags().StringVarP(&generateOutputDir, "output", "o", cfg.OutputDir, "Output directory for documentation")
	cmd.Flags().StringVarP(&generateFormat, "format", "f", cfg.OutputFormat, "Output format (html, json, markdown)")
	cmd.Flags().IntVarP(&generateConcurrency, "concurrency", "c", cfg.Concurrency, "Number of concurrent workers")

	return cmd
}

// runGenerate implements the generate command logic
func runGenerate(ctx context.Context) error {
	// Update config with command-specific flags
	cfg.SourceDir = generateSourceDir
	cfg.OutputDir = generateOutputDir
	cfg.OutputFormat = generateFormat
	cfg.Concurrency = generateConcurrency

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create schema manager
	schemaManager := schema.NewSchemaManager(cfg.Logger)

	// Create template processor with schema manager
	processor := template.NewProcessor(cfg.Logger)
	processor.SetSchemaManager(schemaManager)

	// Process all templates
	cfg.Logger.Info("Processing templates from ", cfg.SourceDir)
	metadata, err := processor.ProcessAllTemplates(
		cfg.SourceDir,
		cfg.OutputDir,
		cfg.OutputFormat,
		false, // Not validate-only mode
	)
	if err != nil {
		return fmt.Errorf("error processing templates: %w", err)
	}

	// Output summary
	cfg.Logger.Infof("Successfully processed %d templates", len(metadata))

	// Generate documentation
	if cfg.OutputFormat == "html" {
		// Create HTML generator
		htmlGenerator := html.NewGenerator(cfg.Logger)

		// Generate HTML documentation
		if err := htmlGenerator.GenerateDocumentation(metadata, cfg.OutputDir); err != nil {
			return fmt.Errorf("error generating HTML documentation: %w", err)
		}

		cfg.Logger.Infof("HTML documentation generated in %s", cfg.OutputDir)
	}

	// If processing JSON format, output metadata to a file
	if cfg.OutputFormat == "json" {
		metadataFile := fmt.Sprintf("%s/metadata.json", cfg.OutputDir)
		jsonData, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("error generating JSON metadata: %w", err)
		}

		if err := os.WriteFile(metadataFile, jsonData, 0644); err != nil {
			return fmt.Errorf("error writing JSON metadata: %w", err)
		}

		cfg.Logger.Infof("Metadata written to %s", metadataFile)
	}

	return nil
}
