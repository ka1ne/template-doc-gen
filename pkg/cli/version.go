package cli

import (
	"fmt"

	"github.com/ka1ne/template-doc-gen/internal/version"
	"github.com/spf13/cobra"
)

// newVersionCommand creates a command to display version information
func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Display version information",
		Long:    `Display version, commit, build date, and Go version information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Harness Template Documentation Generator")
			fmt.Printf("Version: %s\n", version.Info())
		},
	}

	return cmd
}
