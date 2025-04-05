package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// Command line flags for the serve command
var (
	serveOutputDir string
	servePort      string
	serveNoGen     bool
)

// newServeCommand creates a command to serve documentation
func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"server", "http"},
		Short:   "Serve documentation via HTTP",
		Long:    `Start an HTTP server to view generated documentation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context())
		},
	}

	// Add flags specific to the serve command
	cmd.Flags().StringVarP(&serveOutputDir, "dir", "d", cfg.OutputDir, "Directory containing documentation to serve")
	cmd.Flags().StringVarP(&servePort, "port", "p", "8000", "Port to serve on")
	cmd.Flags().BoolVarP(&serveNoGen, "no-generate", "n", false, "Don't generate documentation before serving")

	return cmd
}

// runServe implements the serve command logic
func runServe(ctx context.Context) error {
	// Update config
	cfg.OutputDir = serveOutputDir

	// Generate documentation first if needed
	if !serveNoGen {
		cfg.Logger.Info("Generating documentation before serving...")

		// Set up the generate command with proper flags
		generateCmd := newGenerateCommand()

		// We need to configure the generate command to use our output directory
		// but we can't pass the -d flag directly because generate uses different flags
		generateCmd.SetArgs([]string{
			"--output", cfg.OutputDir,
			"--format", cfg.OutputFormat,
			"--source", cfg.SourceDir,
		})

		err := generateCmd.ExecuteContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate documentation: %w", err)
		}
	}

	// Validate that the output directory exists
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", cfg.OutputDir)
	}

	// Set up HTTP file server
	absDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fileServer := http.FileServer(http.Dir(absDir))
	http.Handle("/", fileServer)

	// Create server with graceful shutdown
	server := &http.Server{
		Addr:    ":" + servePort,
		Handler: nil, // Use default http.DefaultServeMux
	}

	// Handle signals for graceful shutdown
	go func() {
		// Set up signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			// Context was cancelled
		case sig := <-sigCh:
			// Received termination signal
			cfg.Logger.Infof("Received signal %v, shutting down...", sig)
		}

		// Create a timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			cfg.Logger.Errorf("Error during server shutdown: %v", err)
		}
	}()

	// Start the server
	cfg.Logger.Infof("Serving documentation from %s", absDir)
	cfg.Logger.Infof("Server started at http://localhost:%s", servePort)
	cfg.Logger.Info("Press Ctrl+C to stop")

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	cfg.Logger.Info("Server stopped")
	return nil
}
