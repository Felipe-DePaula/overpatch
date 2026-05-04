package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Felipe-DePaula/overpatch/internal/planner"
	"github.com/Felipe-DePaula/overpatch/internal/schema"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:           "plan <file>",
	Short:         "Preview an Overpatch document without writing changes",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		doc, err := schema.ParseFile(args[0])
		if err != nil {
			fmt.Fprintf(out, "invalid: %s\n", err)
			return err
		}

		root, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(out, "plan: failed")
			fmt.Fprintf(out, "error: getting working directory: %s\n", err)
			return err
		}

		result, err := planner.Plan(doc, root)
		if err != nil {
			fmt.Fprintln(out, "plan: failed")
			var operationErr *planner.OperationError
			if errors.As(err, &operationErr) {
				fmt.Fprintf(out, "operation: %s\n", operationErr.OperationID)
				fmt.Fprintf(out, "error: %s\n", operationErr.Message)
			} else {
				fmt.Fprintf(out, "error: %s\n", err)
			}
			return err
		}

		if result.Status == schema.StatusNoChanges {
			fmt.Fprintln(out, "plan: no_changes")
			fmt.Fprintf(out, "reason: %s\n", result.Reason)
			fmt.Fprintln(out, "operations: 0")
			return nil
		}

		fmt.Fprintln(out, "plan: success")
		fmt.Fprintf(out, "operations: %d\n", result.Operations)
		fmt.Fprintf(out, "files changed: %d\n", result.FilesChanged)
		if result.Diff != "" {
			fmt.Fprintln(out)
			fmt.Fprint(out, result.Diff)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
