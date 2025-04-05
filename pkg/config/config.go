package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Config represents application configuration
type Config struct {
	// Input/Output
	SourceDir    string
	OutputDir    string
	OutputFormat string

	// Processing options
	ValidateOnly bool
	Concurrency  int

	// Runtime behavior
	Verbose bool
	Logger  *logrus.Logger
}

// DefaultConfig creates a new configuration with default values
func DefaultConfig() *Config {
	config := &Config{
		// Default values
		SourceDir:    GetEnvOrDefault("SOURCE_DIR", "templates"),
		OutputDir:    GetEnvOrDefault("OUTPUT_DIR", "docs/output"),
		OutputFormat: GetEnvOrDefault("FORMAT", "html"),
		ValidateOnly: strings.ToLower(GetEnvOrDefault("VALIDATE_ONLY", "false")) == "true",
		Verbose:      strings.ToLower(GetEnvOrDefault("VERBOSE", "false")) == "true",
		Concurrency:  4, // Default to 4 workers
		Logger:       logrus.New(),
	}

	// Configure logger
	config.Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	config.Logger.SetOutput(os.Stdout)

	// Set log level based on verbosity
	config.SetLogLevel()

	return config
}

// WithArgs applies command line arguments to the configuration
// This is useful for both the main application and testing
func (c *Config) WithArgs(args []string) *Config {
	// Define command-line flags in a separate FlagSet
	fs := flag.NewFlagSet("tempdocs", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	// Define flags with current config values as defaults
	fs.StringVar(&c.SourceDir, "source", c.SourceDir, "Source directory containing templates")
	fs.StringVar(&c.OutputDir, "output", c.OutputDir, "Output directory for documentation")
	fs.StringVar(&c.OutputFormat, "format", c.OutputFormat, "Output format (html, markdown, json)")
	fs.BoolVar(&c.ValidateOnly, "validate", c.ValidateOnly, "Validate templates without generating documentation")
	fs.BoolVar(&c.Verbose, "verbose", c.Verbose, "Enable verbose logging")
	fs.IntVar(&c.Concurrency, "concurrency", c.Concurrency, "Number of concurrent workers")

	// Parse command-line flags (overrides environment variables)
	// Ignore errors as some flags might not be provided
	_ = fs.Parse(args)

	// Update log level based on verbosity setting
	c.SetLogLevel()

	return c
}

// FromArgs creates a new configuration from command line arguments
// Combines environment variables and command line args
func FromArgs(args []string) *Config {
	return DefaultConfig().WithArgs(args)
}

// NewConfig creates a new configuration from the process arguments
// This is a convenience method for the main application
func NewConfig() *Config {
	return FromArgs(os.Args[1:])
}

// SetLogLevel configures the logger level based on verbosity setting
func (c *Config) SetLogLevel() {
	if c.Verbose {
		c.Logger.SetLevel(logrus.DebugLevel)
	} else {
		c.Logger.SetLevel(logrus.InfoLevel)
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate source directory
	if _, err := os.Stat(c.SourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", c.SourceDir)
	}

	// Validate output format
	validFormats := []string{"html", "markdown", "json"}
	formatValid := false
	for _, format := range validFormats {
		if strings.EqualFold(c.OutputFormat, format) {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("invalid output format: %s. Must be one of %v", c.OutputFormat, validFormats)
	}

	// Validate concurrency
	if c.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1, got %d", c.Concurrency)
	}

	return nil
}

// GetEnvOrDefault gets environment variable with default value
// Exported to allow reuse in other packages
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
