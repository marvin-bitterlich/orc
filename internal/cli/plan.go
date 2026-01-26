package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orcctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage plans (implementation strategies)",
	Long:  "Create, list, approve, and manage plans in the ORC ledger",
}

var planCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new plan",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		description, _ := cmd.Flags().GetString("description")
		content, _ := cmd.Flags().GetString("content")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		cycleID, _ := cmd.Flags().GetString("cycle-id")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		ctx := context.Background()
		resp, err := wire.PlanService().CreatePlan(ctx, primary.CreatePlanRequest{
			CommissionID: commissionID,
			ShipmentID:   shipmentID,
			CycleID:      cycleID,
			Title:        title,
			Description:  description,
			Content:      content,
		})
		if err != nil {
			return fmt.Errorf("failed to create plan: %w", err)
		}

		plan := resp.Plan
		fmt.Printf("✓ Created plan %s: %s\n", plan.ID, plan.Title)
		if plan.ShipmentID != "" {
			fmt.Printf("  Shipment: %s\n", plan.ShipmentID)
		}
		fmt.Printf("  Commission: %s\n", plan.CommissionID)
		fmt.Printf("  Status: %s\n", plan.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc plan show %s    # View plan details\n", plan.ID)
		fmt.Printf("   orc plan approve %s # Approve when ready\n", plan.ID)
		return nil
	},
}

var planListCmd = &cobra.Command{
	Use:   "list",
	Short: "List plans",
	RunE: func(cmd *cobra.Command, args []string) error {
		commissionID, _ := cmd.Flags().GetString("commission")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		cycleID, _ := cmd.Flags().GetString("cycle-id")
		status, _ := cmd.Flags().GetString("status")

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
		}

		ctx := context.Background()
		plans, err := wire.PlanService().ListPlans(ctx, primary.PlanFilters{
			CommissionID: commissionID,
			ShipmentID:   shipmentID,
			CycleID:      cycleID,
			Status:       status,
		})
		if err != nil {
			return fmt.Errorf("failed to list plans: %w", err)
		}

		if len(plans) == 0 {
			fmt.Println("No plans found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tSHIPMENT")
		fmt.Fprintln(w, "--\t-----\t------\t--------")
		for _, p := range plans {
			pinnedMark := ""
			if p.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "📝"
			if p.Status == "approved" {
				statusIcon = "✅"
			}
			ship := "-"
			if p.ShipmentID != "" {
				ship = p.ShipmentID
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", p.ID, p.Title, pinnedMark, statusIcon, p.Status, ship)
		}
		w.Flush()
		return nil
	},
}

var planShowCmd = &cobra.Command{
	Use:   "show [plan-id]",
	Short: "Show plan details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]

		ctx := context.Background()
		plan, err := wire.PlanService().GetPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("plan not found: %w", err)
		}

		fmt.Printf("Plan: %s\n", plan.ID)
		fmt.Printf("Title: %s\n", plan.Title)
		if plan.Description != "" {
			fmt.Printf("Description: %s\n", plan.Description)
		}
		fmt.Printf("Status: %s\n", plan.Status)
		if plan.Content != "" {
			fmt.Printf("\nContent:\n%s\n", plan.Content)
		}
		fmt.Printf("\nCommission: %s\n", plan.CommissionID)
		if plan.ShipmentID != "" {
			fmt.Printf("Shipment: %s\n", plan.ShipmentID)
		}
		if plan.CycleID != "" {
			fmt.Printf("Cycle: %s\n", plan.CycleID)
		}
		if plan.ConclaveID != "" {
			fmt.Printf("Conclave: %s\n", plan.ConclaveID)
		}
		if plan.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if plan.PromotedFromID != "" {
			fmt.Printf("Promoted from: %s (%s)\n", plan.PromotedFromID, plan.PromotedFromType)
		}
		fmt.Printf("Created: %s\n", plan.CreatedAt)
		if plan.ApprovedAt != "" {
			fmt.Printf("Approved: %s\n", plan.ApprovedAt)
		}

		return nil
	},
}

var planApproveCmd = &cobra.Command{
	Use:   "approve [plan-id]",
	Short: "Approve a plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]

		ctx := context.Background()
		err := wire.PlanService().ApprovePlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to approve plan: %w", err)
		}

		fmt.Printf("✓ Plan %s approved\n", planID)
		return nil
	},
}

var planUpdateCmd = &cobra.Command{
	Use:   "update [plan-id]",
	Short: "Update plan title, description, and/or content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		content, _ := cmd.Flags().GetString("content")

		if title == "" && description == "" && content == "" {
			return fmt.Errorf("must specify --title, --description, and/or --content")
		}

		ctx := context.Background()
		err := wire.PlanService().UpdatePlan(ctx, primary.UpdatePlanRequest{
			PlanID:      planID,
			Title:       title,
			Description: description,
			Content:     content,
		})
		if err != nil {
			return fmt.Errorf("failed to update plan: %w", err)
		}

		fmt.Printf("✓ Plan %s updated\n", planID)
		return nil
	},
}

var planPinCmd = &cobra.Command{
	Use:   "pin [plan-id]",
	Short: "Pin plan to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]

		ctx := context.Background()
		err := wire.PlanService().PinPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to pin plan: %w", err)
		}

		fmt.Printf("✓ Plan %s pinned 📌\n", planID)
		return nil
	},
}

var planUnpinCmd = &cobra.Command{
	Use:   "unpin [plan-id]",
	Short: "Unpin plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]

		ctx := context.Background()
		err := wire.PlanService().UnpinPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to unpin plan: %w", err)
		}

		fmt.Printf("✓ Plan %s unpinned\n", planID)
		return nil
	},
}

var planDeleteCmd = &cobra.Command{
	Use:   "delete [plan-id]",
	Short: "Delete a plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planID := args[0]

		ctx := context.Background()
		err := wire.PlanService().DeletePlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to delete plan: %w", err)
		}

		fmt.Printf("✓ Plan %s deleted\n", planID)
		return nil
	},
}

func init() {
	// plan create flags
	planCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	planCreateCmd.Flags().StringP("description", "d", "", "Plan description")
	planCreateCmd.Flags().String("content", "", "Plan content")
	planCreateCmd.Flags().String("shipment", "", "Shipment ID to attach plan to")
	planCreateCmd.Flags().String("cycle-id", "", "Cycle ID to attach plan to")

	// plan list flags
	planListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	planListCmd.Flags().String("shipment", "", "Filter by shipment")
	planListCmd.Flags().String("cycle-id", "", "Filter by cycle")
	planListCmd.Flags().StringP("status", "s", "", "Filter by status (draft, approved)")

	// plan update flags
	planUpdateCmd.Flags().String("title", "", "New title")
	planUpdateCmd.Flags().StringP("description", "d", "", "New description")
	planUpdateCmd.Flags().String("content", "", "New content")

	// Register subcommands
	planCmd.AddCommand(planCreateCmd)
	planCmd.AddCommand(planListCmd)
	planCmd.AddCommand(planShowCmd)
	planCmd.AddCommand(planApproveCmd)
	planCmd.AddCommand(planUpdateCmd)
	planCmd.AddCommand(planPinCmd)
	planCmd.AddCommand(planUnpinCmd)
	planCmd.AddCommand(planDeleteCmd)
}

// PlanCmd returns the plan command
func PlanCmd() *cobra.Command {
	return planCmd
}
