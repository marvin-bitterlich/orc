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

var cycleCmd = &cobra.Command{
	Use:   "cycle",
	Short: "Manage cycles",
	Long:  "Create, list, and manage cycles in the ORC ledger",
}

var cycleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cycle (sequence number auto-assigned)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")

		if shipmentID == "" {
			return fmt.Errorf("--shipment-id flag is required")
		}

		req := primary.CreateCycleRequest{
			ShipmentID: shipmentID,
		}

		resp, err := wire.CycleService().CreateCycle(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create cycle: %w", err)
		}

		cycle := resp.Cycle
		fmt.Printf("✓ Created cycle %s\n", cycle.ID)
		fmt.Printf("  Sequence: %d\n", cycle.SequenceNumber)
		fmt.Printf("  Shipment: %s\n", cycle.ShipmentID)
		fmt.Printf("  Status: %s\n", cycle.Status)
		return nil
	},
}

var cycleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cycles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		status, _ := cmd.Flags().GetString("status")

		cycles, err := wire.CycleService().ListCycles(ctx, primary.CycleFilters{
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list cycles: %w", err)
		}

		if len(cycles) == 0 {
			fmt.Println("No cycles found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSHIPMENT\tSEQ\tSTATUS\tSTARTED\tCOMPLETED")
		fmt.Fprintln(w, "--\t--------\t---\t------\t-------\t---------")
		for _, item := range cycles {
			started := "-"
			if item.StartedAt != "" {
				started = item.StartedAt[:10] // Just date
			}
			completed := "-"
			if item.CompletedAt != "" {
				completed = item.CompletedAt[:10] // Just date
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
				item.ID,
				item.ShipmentID,
				item.SequenceNumber,
				item.Status,
				started,
				completed,
			)
		}
		w.Flush()
		return nil
	},
}

var cycleShowCmd = &cobra.Command{
	Use:   "show [cycle-id]",
	Short: "Show cycle details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cycleID := args[0]

		cycle, err := wire.CycleService().GetCycle(ctx, cycleID)
		if err != nil {
			return fmt.Errorf("cycle not found: %w", err)
		}

		fmt.Printf("Cycle: %s\n", cycle.ID)
		fmt.Printf("Shipment: %s\n", cycle.ShipmentID)
		fmt.Printf("Sequence: %d\n", cycle.SequenceNumber)
		fmt.Printf("Status: %s\n", cycle.Status)
		fmt.Printf("Created: %s\n", cycle.CreatedAt)
		fmt.Printf("Updated: %s\n", cycle.UpdatedAt)
		if cycle.StartedAt != "" {
			fmt.Printf("Started: %s\n", cycle.StartedAt)
		}
		if cycle.CompletedAt != "" {
			fmt.Printf("Completed: %s\n", cycle.CompletedAt)
		}

		return nil
	},
}

var cycleCompleteCmd = &cobra.Command{
	Use:   "complete [cycle-id]",
	Short: "Complete an active cycle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cycleID := args[0]

		err := wire.CycleService().CompleteCycle(ctx, cycleID)
		if err != nil {
			return fmt.Errorf("failed to complete cycle: %w", err)
		}

		fmt.Printf("✓ Cycle %s completed\n", cycleID)
		return nil
	},
}

var cycleDeleteCmd = &cobra.Command{
	Use:   "delete [cycle-id]",
	Short: "Delete a cycle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cycleID := args[0]

		err := wire.CycleService().DeleteCycle(ctx, cycleID)
		if err != nil {
			return fmt.Errorf("failed to delete cycle: %w", err)
		}

		fmt.Printf("✓ Cycle %s deleted\n", cycleID)
		return nil
	},
}

func init() {
	// cycle create flags
	cycleCreateCmd.Flags().String("shipment-id", "", "Shipment ID (required)")
	cycleCreateCmd.MarkFlagRequired("shipment-id")

	// cycle list flags
	cycleListCmd.Flags().String("shipment-id", "", "Filter by shipment")
	cycleListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// Register subcommands
	cycleCmd.AddCommand(cycleCreateCmd)
	cycleCmd.AddCommand(cycleListCmd)
	cycleCmd.AddCommand(cycleShowCmd)
	cycleCmd.AddCommand(cycleCompleteCmd)
	cycleCmd.AddCommand(cycleDeleteCmd)
}

// CycleCmd returns the cycle command
func CycleCmd() *cobra.Command {
	return cycleCmd
}
