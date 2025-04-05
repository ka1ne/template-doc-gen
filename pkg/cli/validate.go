package cli

import (
	"context"
	"fmt"

	"github.com/ka1ne/template-doc-gen/pkg/schema"
	"github.com/ka1ne/template-doc-gen/pkg/template"
	"github.com/spf13/cobra"
)

// Command line flags for the validate command
var (
	validateSourceDir   string
	validateConcurrency int
	validateFailFast    bool
)

// newValidateCommand creates a command to validate templates
func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		Aliases: []string{"val", "check"},
		Short:   "Validate templates",
		Long:    `Validate templates without generating documentation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context())
		},
	}

	// Add flags specific to the validate command
	cmd.Flags().StringVarP(&validateSourceDir, "source", "s", cfg.SourceDir, "Source directory containing templates")
	cmd.Flags().IntVarP(&validateConcurrency, "concurrency", "c", cfg.Concurrency, "Number of concurrent workers")
	cmd.Flags().BoolVarP(&validateFailFast, "fail-fast", "f", false, "Stop on first validation error")

	return cmd
}

// runValidate implements the validate command logic
func runValidate(ctx context.Context) error {
	// Update config with command-specific flags
	cfg.SourceDir = validateSourceDir
	cfg.Concurrency = validateConcurrency
	cfg.ValidateOnly = true

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create schema manager
	schemaManager := schema.NewSchemaManager(cfg.Logger)

	// Create template processor with schema manager
	processor := template.NewProcessor(cfg.Logger)
	processor.SetSchemaManager(schemaManager)

	// Process all templates in validate-only mode
	cfg.Logger.Info("Validating templates from ", cfg.SourceDir)
	metadata, err := processor.ProcessAllTemplates(
		cfg.SourceDir,
		"",   // No output dir needed for validation
		"",   // No output format needed for validation
		true, // Validate-only mode
	)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Output success message
	cfg.Logger.Infof("Successfully validated %d templates", len(metadata))
	return nil
}
