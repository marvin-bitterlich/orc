package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var cwoCmd = &cobra.Command{
	Use:   "cwo",
	Short: "Manage cycle work orders",
	Long:  "Create, list, update, and manage cycle work orders (CWOs) in the ORC ledger",
}

var cwoCreateCmd = &cobra.Command{
	Use:   "create [outcome]",
	Short: "Create a new cycle work order",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cycleID, _ := cmd.Flags().GetString("cycle-id")
		acceptanceCriteria, _ := cmd.Flags().GetString("acceptance-criteria")

		if cycleID == "" {
			return fmt.Errorf("--cycle-id flag is required")
		}

		req := primary.CreateCycleWorkOrderRequest{
			CycleID:            cycleID,
			Outcome:            args[0],
			AcceptanceCriteria: acceptanceCriteria,
		}

		resp, err := wire.CycleWorkOrderService().CreateCycleWorkOrder(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create CWO: %w", err)
		}

		cwo := resp.CycleWorkOrder
		fmt.Printf("✓ Created cycle work order %s\n", cwo.ID)
		fmt.Printf("  Outcome: %s\n", cwo.Outcome)
		fmt.Printf("  Cycle: %s\n", cwo.CycleID)
		fmt.Printf("  Shipment: %s\n", cwo.ShipmentID)
		fmt.Printf("  Status: %s\n", cwo.Status)
		return nil
	},
}

var cwoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cycle work orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cycleID, _ := cmd.Flags().GetString("cycle-id")
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		status, _ := cmd.Flags().GetString("status")

		cwos, err := wire.CycleWorkOrderService().ListCycleWorkOrders(ctx, primary.CycleWorkOrderFilters{
			CycleID:    cycleID,
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list CWOs: %w", err)
		}

		if len(cwos) == 0 {
			fmt.Println("No cycle work orders found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCYCLE\tSHIPMENT\tOUTCOME\tSTATUS")
		fmt.Fprintln(w, "--\t-----\t--------\t-------\t------")
		for _, item := range cwos {
			// Truncate outcome for display
			outcome := item.Outcome
			if len(outcome) > 30 {
				outcome = outcome[:27] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.CycleID,
				item.ShipmentID,
				outcome,
				item.Status,
			)
		}
		w.Flush()
		return nil
	},
}

var cwoShowCmd = &cobra.Command{
	Use:   "show [cwo-id]",
	Short: "Show cycle work order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID := args[0]

		cwo, err := wire.CycleWorkOrderService().GetCycleWorkOrder(ctx, cwoID)
		if err != nil {
			return fmt.Errorf("CWO not found: %w", err)
		}

		fmt.Printf("Cycle Work Order: %s\n", cwo.ID)
		fmt.Printf("Cycle: %s\n", cwo.CycleID)
		fmt.Printf("Shipment: %s\n", cwo.ShipmentID)
		fmt.Printf("Outcome: %s\n", cwo.Outcome)
		if cwo.AcceptanceCriteria != "" {
			fmt.Printf("Acceptance Criteria: %s\n", cwo.AcceptanceCriteria)
		}
		fmt.Printf("Status: %s\n", cwo.Status)
		fmt.Printf("Created: %s\n", cwo.CreatedAt)
		fmt.Printf("Updated: %s\n", cwo.UpdatedAt)

		return nil
	},
}

var cwoApproveCmd = &cobra.Command{
	Use:   "approve [cwo-id]",
	Short: "Approve a draft cycle work order (cascades: Cycle → approved)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID := args[0]

		err := wire.CycleWorkOrderService().ApproveCycleWorkOrder(ctx, cwoID)
		if err != nil {
			return fmt.Errorf("failed to approve CWO: %w", err)
		}

		fmt.Printf("✓ Cycle work order %s approved (Cycle status → approved)\n", cwoID)
		return nil
	},
}

var cwoCompleteCmd = &cobra.Command{
	Use:   "complete [cwo-id]",
	Short: "Complete an active cycle work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID := args[0]

		err := wire.CycleWorkOrderService().CompleteCycleWorkOrder(ctx, cwoID)
		if err != nil {
			return fmt.Errorf("failed to complete CWO: %w", err)
		}

		fmt.Printf("Cycle work order %s completed\n", cwoID)
		return nil
	},
}

var cwoDeleteCmd = &cobra.Command{
	Use:   "delete [cwo-id]",
	Short: "Delete a cycle work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID := args[0]

		err := wire.CycleWorkOrderService().DeleteCycleWorkOrder(ctx, cwoID)
		if err != nil {
			return fmt.Errorf("failed to delete CWO: %w", err)
		}

		fmt.Printf("✓ Cycle work order %s deleted\n", cwoID)
		return nil
	},
}

func init() {
	// cwo create flags
	cwoCreateCmd.Flags().String("cycle-id", "", "Cycle ID (required)")
	cwoCreateCmd.Flags().String("acceptance-criteria", "", "Acceptance criteria (JSON array)")
	cwoCreateCmd.MarkFlagRequired("cycle-id")

	// cwo list flags
	cwoListCmd.Flags().String("cycle-id", "", "Filter by cycle")
	cwoListCmd.Flags().String("shipment-id", "", "Filter by shipment")
	cwoListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// Register subcommands
	cwoCmd.AddCommand(cwoCreateCmd)
	cwoCmd.AddCommand(cwoListCmd)
	cwoCmd.AddCommand(cwoShowCmd)
	cwoCmd.AddCommand(cwoApproveCmd)
	cwoCmd.AddCommand(cwoCompleteCmd)
	cwoCmd.AddCommand(cwoDeleteCmd)
}

// CycleWorkOrderCmd returns the cwo command
func CycleWorkOrderCmd() *cobra.Command {
	return cwoCmd
}
