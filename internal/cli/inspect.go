package cli

import (
	"fmt"

	"github.com/Felipe-DePaula/overpatch/internal/schema"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:           "inspect <file>",
	Short:         "Inspect an Overpatch document and print a human-readable summary",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := schema.ParseFile(args[0])
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "invalid: %s\n", err)
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "schema: %s\n", doc.SchemaVersion)
		fmt.Fprintf(out, "status: %s\n", doc.Status)

		if doc.Summary != "" {
			fmt.Fprintf(out, "summary: %s\n", doc.Summary)
		}
		if doc.Reason != "" {
			fmt.Fprintf(out, "reason: %s\n", doc.Reason)
		}

		fmt.Fprintf(out, "operations: %d\n", len(doc.Operations))

		if len(doc.Operations) > 0 {
			fmt.Fprintln(out)
			for _, op := range doc.Operations {
				fmt.Fprintf(out, "%s %s %s\n", op.ID, op.Action, op.Path)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
