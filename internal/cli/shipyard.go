package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/wire"
)

var shipyardCmd = &cobra.Command{
	Use:   "shipyard",
	Short: "Manage shipyard queue",
	Long:  "View and manage the shipyard queue of ready-to-assign shipments",
}

var shipyardPushCmd = &cobra.Command{
	Use:   "push [SHIP-xxx]",
	Short: "Add shipment to queue",
	Long:  "Move a shipment to the shipyard queue (replaces 'orc shipment park')",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		// Get current state to check if already in shipyard
		shipment, err := wire.ShipmentService().GetShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("shipment not found: %w", err)
		}

		if shipment.ContainerType == "shipyard" {
			fmt.Printf("%s is already in shipyard\n", shipmentID)
			return nil
		}

		// Move to shipyard
		if err := wire.ShipmentService().ParkShipment(ctx, shipmentID); err != nil {
			return fmt.Errorf("failed to push shipment to shipyard: %w", err)
		}

		fmt.Printf("✓ Pushed %s to shipyard queue\n", shipmentID)
		fmt.Printf("  Title: %s\n", shipment.Title)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc shipyard list         # View queue\n")
		fmt.Printf("   orc shipyard prioritize %s 1  # Set as urgent\n", shipmentID)
		return nil
	},
}

var shipyardClaimCmd = &cobra.Command{
	Use:   "claim",
	Short: "Claim top shipment from queue",
	Long:  "IMP claims the top priority shipment from the shipyard queue and assigns it to the current workbench",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Auto-detect workbench from context
		workbenchID := orccontext.GetContextWorkbenchID()
		if workbenchID == "" {
			return fmt.Errorf("not in workbench context\nHint: Run from a workbench directory")
		}

		// Get commission context
		commissionID := orccontext.GetContextCommissionID()

		// Get shipyard queue
		entries, err := wire.ShipmentService().ListShipyardQueue(ctx, commissionID)
		if err != nil {
			return fmt.Errorf("failed to list shipyard queue: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("Shipyard queue is empty - no shipments to claim")
			return nil
		}

		// Claim the top shipment
		topShipment := entries[0]

		// Assign to workbench
		if err := wire.ShipmentService().AssignShipmentToWorkbench(ctx, topShipment.ID, workbenchID); err != nil {
			return fmt.Errorf("failed to claim shipment: %w", err)
		}

		fmt.Printf("✓ Claimed %s: %s\n", topShipment.ID, topShipment.Title)
		fmt.Printf("  Assigned to workbench: %s\n", workbenchID)
		if topShipment.TaskCount > 0 {
			fmt.Printf("  Tasks: %d (%d done)\n", topShipment.TaskCount, topShipment.DoneCount)
		}
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc focus %s         # Focus on this shipment\n", topShipment.ID)
		fmt.Printf("   orc task list --shipment %s  # View tasks\n", topShipment.ID)
		return nil
	},
}

var shipyardPrioritizeCmd = &cobra.Command{
	Use:   "prioritize [SHIP-xxx] [priority]",
	Short: "Set shipment priority",
	Long: `Set priority for a shipment in the queue.

Priority 1 is highest (urgent). Higher numbers = lower priority.
Omit priority argument to clear (return to default FIFO ordering).

Examples:
  orc shipyard prioritize SHIP-001 1     # Set as urgent
  orc shipyard prioritize SHIP-001 2     # Set as high priority
  orc shipyard prioritize SHIP-001       # Clear priority (FIFO)`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		var priority *int
		if len(args) > 1 {
			var p int
			if _, err := fmt.Sscanf(args[1], "%d", &p); err != nil {
				return fmt.Errorf("invalid priority: %s (must be a positive number)", args[1])
			}
			if p < 1 {
				return fmt.Errorf("priority must be at least 1, got %d", p)
			}
			priority = &p
		}

		// Verify shipment exists
		shipment, err := wire.ShipmentService().GetShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("shipment not found: %w", err)
		}

		// Set priority
		if err := wire.ShipmentService().SetShipmentPriority(ctx, shipmentID, priority); err != nil {
			return fmt.Errorf("failed to set priority: %w", err)
		}

		if priority == nil {
			fmt.Printf("✓ Cleared priority for %s: %s\n", shipmentID, shipment.Title)
			fmt.Println("  Position: default FIFO ordering")
		} else {
			fmt.Printf("✓ Set priority %d for %s: %s\n", *priority, shipmentID, shipment.Title)
		}
		return nil
	},
}

var shipyardListCmd = &cobra.Command{
	Use:   "list [COMM-xxx]",
	Short: "List queued shipments",
	Long:  "Show shipments in the shipyard queue, ordered by priority then creation time",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get commission ID from arg or context
		var commissionID string
		if len(args) > 0 {
			commissionID = args[0]
		} else {
			commissionID = orccontext.GetContextCommissionID()
		}

		// List shipyard queue
		entries, err := wire.ShipmentService().ListShipyardQueue(ctx, commissionID)
		if err != nil {
			return fmt.Errorf("failed to list shipyard queue: %w", err)
		}

		if len(entries) == 0 {
			if commissionID != "" {
				fmt.Printf("No shipments queued in shipyard for %s\n", commissionID)
			} else {
				fmt.Println("No shipments queued in shipyard")
			}
			return nil
		}

		// Get shipyard info for display
		var yardID string
		if commissionID != "" {
			yard, err := wire.ShipyardRepository().GetByCommissionID(ctx, commissionID)
			if err == nil {
				yardID = yard.ID
			}
		}

		// Header
		if yardID != "" && commissionID != "" {
			fmt.Printf("%s (%s) - %d shipment(s) queued\n\n", yardID, commissionID, len(entries))
		} else {
			fmt.Printf("Shipyard Queue - %d shipment(s) queued\n\n", len(entries))
		}

		// Table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, " #\tSHIP\tTITLE\tPRI\tTASKS\tQUEUED")
		fmt.Fprintln(w, "--\t----\t-----\t---\t-----\t------")

		for i, entry := range entries {
			// Format priority
			priStr := "-"
			if entry.Priority != nil {
				priStr = fmt.Sprintf("%d", *entry.Priority)
			}

			// Format task count
			taskStr := fmt.Sprintf("%d/%d", entry.DoneCount, entry.TaskCount)

			// Format queued time (relative)
			queuedStr := formatShipyardRelativeTime(entry.CreatedAt)

			fmt.Fprintf(w, "%2d\t%s\t%s\t%s\t%s\t%s\n",
				i+1, entry.ID, truncateTitle(entry.Title, 30), priStr, taskStr, queuedStr)
		}
		w.Flush()

		return nil
	},
}

// formatShipyardRelativeTime formats a timestamp as a relative time string (e.g., "2h ago", "1d ago")
func formatShipyardRelativeTime(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}

	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "now"
	case duration < time.Hour:
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	default:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

// truncateTitle truncates a string to maxLen, adding "..." if truncated
func truncateTitle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	// Register subcommands
	shipyardCmd.AddCommand(shipyardListCmd)
	shipyardCmd.AddCommand(shipyardPushCmd)
	shipyardCmd.AddCommand(shipyardClaimCmd)
	shipyardCmd.AddCommand(shipyardPrioritizeCmd)
}

// ShipyardCmd returns the shipyard command
func ShipyardCmd() *cobra.Command {
	return shipyardCmd
}
