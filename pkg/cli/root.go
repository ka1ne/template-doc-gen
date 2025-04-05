package cli

import (
	"os"

	"github.com/ka1ne/template-doc-gen/internal/version"
	"github.com/ka1ne/template-doc-gen/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Shared configuration - will be accessed by subcommands
	cfg = config.DefaultConfig()

	// Flags that apply to all commands
	verboseFlag bool
)

// NewRootCommand creates the root command for the tempdocs CLI
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tempdocs",
		Short:   "Harness Template Documentation Generator",
		Long:    `Generate and validate documentation for Harness templates.`,
		Version: version.Info(),
		// This will be called for any subcommand
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Update config from global flags
			cfg.Verbose = verboseFlag

			// Configure logger level based on verbose flag
			if cfg.Verbose {
				cfg.Logger.SetLevel(logrus.DebugLevel)
			}
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(newGenerateCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newServeCommand())
	rootCmd.AddCommand(newVersionCommand())

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once.
func Execute() {
	rootCmd := NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		// The error is already printed by cobra
		os.Exit(1)
	}
}
