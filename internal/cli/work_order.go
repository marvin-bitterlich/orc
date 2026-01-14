package cli

import (
	"fmt"

	"github.com/looneym/orc/internal/context"
	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

var workOrderCmd = &cobra.Command{
	Use:   "work-order",
	Short: "Manage work orders (individual tasks)",
	Long:  "Create, list, claim, and complete work orders in the ORC ledger",
}

var workOrderCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new work order",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		contextRef, _ := cmd.Flags().GetString("context-ref")
		parentID, _ := cmd.Flags().GetString("parent")

		// Smart default: use deputy context if available, otherwise MISSION-001
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
				fmt.Printf("ℹ️  Using mission from context: %s\n", missionID)
			} else {
				missionID = "MISSION-001"
			}
		}

		wo, err := models.CreateWorkOrder(missionID, title, description, contextRef, parentID)
		if err != nil {
			return fmt.Errorf("failed to create work order: %w", err)
		}

		fmt.Printf("✓ Created work order %s: %s\n", wo.ID, wo.Title)
		fmt.Printf("  Under mission: %s\n", wo.MissionID)
		if wo.ParentID.Valid {
			fmt.Printf("  Parent: %s\n", wo.ParentID.String)
		}
		if wo.ContextRef.Valid {
			fmt.Printf("  Context: %s\n", wo.ContextRef.String)
		}
		return nil
	},
}

var workOrderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List work orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		missionID, _ := cmd.Flags().GetString("mission")
		status, _ := cmd.Flags().GetString("status")

		// Smart default: use deputy context if available and no mission specified
		if missionID == "" {
			if ctxMissionID := context.GetContextMissionID(); ctxMissionID != "" {
				missionID = ctxMissionID
			}
		}

		orders, err := models.ListWorkOrders(missionID, status)
		if err != nil {
			return fmt.Errorf("failed to list work orders: %w", err)
		}

		if len(orders) == 0 {
			fmt.Println("No work orders found")
			return nil
		}

		fmt.Printf("\n%-15s %-15s %-10s %-15s %s\n", "ID", "MISSION", "STATUS", "GROVE", "TITLE")
		fmt.Println("─────────────────────────────────────────────────────────────────────────────────")
		for _, wo := range orders {
			grove := "-"
			if wo.AssignedGroveID.Valid {
				grove = wo.AssignedGroveID.String
			}
			fmt.Printf("%-15s %-15s %-10s %-15s %s\n", wo.ID, wo.MissionID, wo.Status, grove, wo.Title)
		}
		fmt.Println()

		return nil
	},
}

var workOrderShowCmd = &cobra.Command{
	Use:   "show [work-order-id]",
	Short: "Show work order details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		wo, err := models.GetWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to get work order: %w", err)
		}

		fmt.Printf("\nWork Order: %s\n", wo.ID)
		fmt.Printf("Mission:    %s\n", wo.MissionID)
		fmt.Printf("Title:      %s\n", wo.Title)
		fmt.Printf("Status:     %s\n", wo.Status)
		if wo.Type.Valid {
			fmt.Printf("Type:       %s\n", wo.Type.String)
		}
		if wo.Priority.Valid {
			fmt.Printf("Priority:   %s\n", wo.Priority.String)
		}
		if wo.ParentID.Valid {
			fmt.Printf("Parent:     %s\n", wo.ParentID.String)
		}
		if wo.Description.Valid {
			fmt.Printf("Description: %s\n", wo.Description.String)
		}
		if wo.AssignedGroveID.Valid {
			fmt.Printf("Grove:      %s\n", wo.AssignedGroveID.String)
		}
		if wo.ContextRef.Valid {
			fmt.Printf("Context:    %s\n", wo.ContextRef.String)
		}
		fmt.Printf("Created:    %s\n", wo.CreatedAt.Format("2006-01-02 15:04"))
		if wo.ClaimedAt.Valid {
			fmt.Printf("Claimed:    %s\n", wo.ClaimedAt.Time.Format("2006-01-02 15:04"))
		}
		if wo.CompletedAt.Valid {
			fmt.Printf("Completed:  %s\n", wo.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		return nil
	},
}

var workOrderClaimCmd = &cobra.Command{
	Use:   "claim [work-order-id]",
	Short: "Claim a work order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		groveID, _ := cmd.Flags().GetString("grove")

		err := models.ClaimWorkOrder(id, groveID)
		if err != nil {
			return fmt.Errorf("failed to claim work order: %w", err)
		}

		if groveID != "" {
			fmt.Printf("✓ Work order %s claimed by %s\n", id, groveID)
		} else {
			fmt.Printf("✓ Work order %s claimed by IMP-UNKNOWN\n", id)
		}
		fmt.Printf("  Status: in_progress\n")
		return nil
	},
}

var workOrderCompleteCmd = &cobra.Command{
	Use:   "complete [work-order-id]",
	Short: "Mark a work order as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.CompleteWorkOrder(id)
		if err != nil {
			return fmt.Errorf("failed to complete work order: %w", err)
		}

		fmt.Printf("✓ Work order %s marked as complete\n", id)
		return nil
	},
}

var workOrderSetParentCmd = &cobra.Command{
	Use:   "set-parent [work-order-id]",
	Short: "Set or update the parent of a work order",
	Long: `Set or update the parent work order to create an epic hierarchy.

Examples:
  orc work-order set-parent WO-015 --parent WO-010  # Make WO-015 a child of WO-010
  orc work-order set-parent WO-015 --parent ""      # Remove parent (make top-level)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		parentID, _ := cmd.Flags().GetString("parent")

		err := models.SetParentWorkOrder(id, parentID)
		if err != nil {
			return fmt.Errorf("failed to set parent: %w", err)
		}

		if parentID != "" {
			fmt.Printf("✓ Work order %s parent set to %s\n", id, parentID)
		} else {
			fmt.Printf("✓ Work order %s parent removed (now top-level)\n", id)
		}
		return nil
	},
}

// WorkOrderCmd returns the work-order command
func WorkOrderCmd() *cobra.Command {
	// Add flags
	workOrderCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to MISSION-001)")
	workOrderCreateCmd.Flags().StringP("description", "d", "", "Work order description")
	workOrderCreateCmd.Flags().StringP("context-ref", "c", "", "Graphiti context reference (e.g., graphiti:episode-uuid)")
	workOrderCreateCmd.Flags().StringP("parent", "p", "", "Parent work order ID (for creating sub-tasks)")

	workOrderListCmd.Flags().StringP("mission", "m", "", "Filter by mission ID")
	workOrderListCmd.Flags().StringP("status", "s", "", "Filter by status (backlog, next, in_progress, complete)")

	workOrderClaimCmd.Flags().StringP("grove", "g", "", "Grove ID claiming the work order")

	workOrderSetParentCmd.Flags().StringP("parent", "p", "", "Parent work order ID (empty string to remove parent)")
	workOrderSetParentCmd.MarkFlagRequired("parent")

	// Add subcommands
	workOrderCmd.AddCommand(workOrderCreateCmd)
	workOrderCmd.AddCommand(workOrderListCmd)
	workOrderCmd.AddCommand(workOrderShowCmd)
	workOrderCmd.AddCommand(workOrderClaimCmd)
	workOrderCmd.AddCommand(workOrderCompleteCmd)
	workOrderCmd.AddCommand(workOrderSetParentCmd)

	return workOrderCmd
}
