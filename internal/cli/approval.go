package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "Manage approvals",
	Long:  "List and view plan approvals in the ORC ledger",
}

var approvalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List approvals",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		taskID, _ := cmd.Flags().GetString("task")
		outcome, _ := cmd.Flags().GetString("outcome")

		approvals, err := wire.ApprovalService().ListApprovals(ctx, primary.ApprovalFilters{
			TaskID:  taskID,
			Outcome: outcome,
		})
		if err != nil {
			return fmt.Errorf("failed to list approvals: %w", err)
		}

		if len(approvals) == 0 {
			fmt.Println("No approvals found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPLAN\tTASK\tMECHANISM\tOUTCOME\tCREATED")
		fmt.Fprintln(w, "--\t----\t----\t---------\t-------\t-------")
		for _, item := range approvals {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.PlanID,
				item.TaskID,
				item.Mechanism,
				item.Outcome,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var approvalShowCmd = &cobra.Command{
	Use:   "show [approval-id]",
	Short: "Show approval details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		approvalID := args[0]

		approval, err := wire.ApprovalService().GetApproval(ctx, approvalID)
		if err != nil {
			return fmt.Errorf("approval not found: %w", err)
		}

		fmt.Printf("Approval: %s\n", approval.ID)
		fmt.Printf("Plan: %s\n", approval.PlanID)
		fmt.Printf("Task: %s\n", approval.TaskID)
		fmt.Printf("Mechanism: %s\n", approval.Mechanism)
		fmt.Printf("Outcome: %s\n", approval.Outcome)
		if approval.ReviewerInput != "" {
			fmt.Printf("Reviewer Input: %s\n", approval.ReviewerInput)
		}
		if approval.ReviewerOutput != "" {
			fmt.Printf("Reviewer Output: %s\n", approval.ReviewerOutput)
		}
		fmt.Printf("Created: %s\n", approval.CreatedAt)

		return nil
	},
}

func init() {
	// approval list flags
	approvalListCmd.Flags().String("task", "", "Filter by task ID")
	approvalListCmd.Flags().String("outcome", "", "Filter by outcome (approved|escalated)")

	// Register subcommands
	approvalCmd.AddCommand(approvalListCmd)
	approvalCmd.AddCommand(approvalShowCmd)
}

// ApprovalCmd returns the approval command
func ApprovalCmd() *cobra.Command {
	return approvalCmd
}
