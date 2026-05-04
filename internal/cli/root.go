// Package cli wires together the Cobra command tree for the overpatch binary.
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "overpatch",
	Short:        "AI-assisted deterministic patching toolkit for codebases",
	Long:         "Overpatch validates, previews, and applies deterministic patch operations described as structured JSON.",
	SilenceUsage: true,
}

// Execute runs the root Cobra command and returns the appropriate exit code.
// It returns 0 on success and 1 on any error.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
