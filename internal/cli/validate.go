package cli

import (
	"fmt"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:           "validate <file>",
	Short:         "Validate an Overpatch document against the v1 schema",
	Args:          cobra.ExactArgs(1),
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		_, err := schema.ParseFile(args[0])
		if err != nil {
			fmt.Fprintf(out, "invalid: %s\n", err)
			return err
		}
		fmt.Fprintln(out, "valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
