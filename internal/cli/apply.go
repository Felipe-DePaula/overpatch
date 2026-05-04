package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Felipe-DePaula/overpatch/internal/executor"
	"github.com/Felipe-DePaula/overpatch/internal/gitguard"
	"github.com/Felipe-DePaula/overpatch/internal/planner"
	"github.com/Felipe-DePaula/overpatch/internal/schema"
	"github.com/spf13/cobra"
)

var applyYes bool

var applyCmd = &cobra.Command{
	Use:           "apply <file>",
	Short:         "Apply an Overpatch document to the working tree",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if !applyYes {
			err := errors.New("--yes is required to apply changes")
			fmt.Fprintln(out, "apply: refused")
			fmt.Fprintf(out, "error: %s\n", err)
			return err
		}

		root, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(out, "apply: failed")
			fmt.Fprintf(out, "error: getting working directory: %s\n", err)
			return err
		}

		if err := gitguard.CheckCleanWorkingTree(root); err != nil {
			fmt.Fprintln(out, "apply: refused")
			var r *gitguard.Refusal
			if errors.As(err, &r) {
				fmt.Fprintf(out, "error: %s\n", r.Error())
				fmt.Fprintf(out, "hint: %s\n", r.Hint)
			} else {
				fmt.Fprintf(out, "error: %s\n", err)
			}
			return err
		}

		doc, err := schema.ParseFile(args[0])
		if err != nil {
			fmt.Fprintf(out, "invalid: %s\n", err)
			return err
		}

		stage, err := planner.Stage(doc, root)
		if err != nil {
			fmt.Fprintln(out, "apply: failed")
			var operationErr *planner.OperationError
			if errors.As(err, &operationErr) {
				fmt.Fprintf(out, "operation: %s\n", operationErr.OperationID)
				fmt.Fprintf(out, "error: %s\n", operationErr.Message)
			} else {
				fmt.Fprintf(out, "error: %s\n", err)
			}
			return err
		}

		if stage.Status == schema.StatusNoChanges {
			fmt.Fprintln(out, "apply: no_changes")
			fmt.Fprintf(out, "reason: %s\n", stage.Reason)
			fmt.Fprintln(out, "operations: 0")
			return nil
		}

		if stage.Status == schema.StatusFailed {
			fmt.Fprintln(out, "apply: failed")
			if stage.Reason != "" {
				fmt.Fprintf(out, "reason: %s\n", stage.Reason)
			}
			return fmt.Errorf("document status is failed")
		}

		if err := executor.Apply(stage, root); err != nil {
			fmt.Fprintln(out, "apply: failed")
			fmt.Fprintf(out, "error: %s\n", err)
			return err
		}

		fmt.Fprintln(out, "apply: success")
		fmt.Fprintf(out, "operations: %d\n", stage.Operations)
		fmt.Fprintf(out, "files changed: %d\n", len(stage.Changes))

		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&applyYes, "yes", false, "confirm applying changes")
	rootCmd.AddCommand(applyCmd)
}
