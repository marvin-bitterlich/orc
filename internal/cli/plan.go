package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
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
		missionID, _ := cmd.Flags().GetString("mission")
		description, _ := cmd.Flags().GetString("description")
		content, _ := cmd.Flags().GetString("content")
		shipmentID, _ := cmd.Flags().GetString("shipment")

		// Get mission from context or require explicit flag
		if missionID == "" {
			missionID = context.GetContextMissionID()
			if missionID == "" {
				return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
			}
		}

		plan, err := models.CreatePlan(shipmentID, missionID, title, description, content)
		if err != nil {
			return fmt.Errorf("failed to create plan: %w", err)
		}

		fmt.Printf("✓ Created plan %s: %s\n", plan.ID, plan.Title)
		if plan.ShipmentID.Valid {
			fmt.Printf("  Shipment: %s\n", plan.ShipmentID.String)
		}
		fmt.Printf("  Mission: %s\n", plan.MissionID)
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
		missionID, _ := cmd.Flags().GetString("mission")
		shipmentID, _ := cmd.Flags().GetString("shipment")
		status, _ := cmd.Flags().GetString("status")

		// Get mission from context if not specified
		if missionID == "" {
			missionID = context.GetContextMissionID()
		}

		plans, err := models.ListPlans(shipmentID, status)
		if err != nil {
			return fmt.Errorf("failed to list plans: %w", err)
		}

		// Filter by mission if specified
		if missionID != "" {
			var filtered []*models.Plan
			for _, p := range plans {
				if p.MissionID == missionID {
					filtered = append(filtered, p)
				}
			}
			plans = filtered
		}

		if len(plans) == 0 {
			fmt.Println("No plans found")
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
			if p.ShipmentID.Valid {
				ship = p.ShipmentID.String
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

		plan, err := models.GetPlan(planID)
		if err != nil {
			return fmt.Errorf("plan not found: %w", err)
		}

		fmt.Printf("Plan: %s\n", plan.ID)
		fmt.Printf("Title: %s\n", plan.Title)
		if plan.Description.Valid {
			fmt.Printf("Description: %s\n", plan.Description.String)
		}
		fmt.Printf("Status: %s\n", plan.Status)
		if plan.Content.Valid {
			fmt.Printf("\nContent:\n%s\n", plan.Content.String)
		}
		fmt.Printf("\nMission: %s\n", plan.MissionID)
		if plan.ShipmentID.Valid {
			fmt.Printf("Shipment: %s\n", plan.ShipmentID.String)
		}
		if plan.ConclaveID.Valid {
			fmt.Printf("Conclave: %s\n", plan.ConclaveID.String)
		}
		if plan.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		if plan.PromotedFromID.Valid {
			fmt.Printf("Promoted from: %s (%s)\n", plan.PromotedFromID.String, plan.PromotedFromType.String)
		}
		fmt.Printf("Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04"))
		if plan.ApprovedAt.Valid {
			fmt.Printf("Approved: %s\n", plan.ApprovedAt.Time.Format("2006-01-02 15:04"))
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

		err := models.ApprovePlan(planID)
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

		err := models.UpdatePlan(planID, title, description, content)
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

		err := models.PinPlan(planID)
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

		err := models.UnpinPlan(planID)
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

		err := models.DeletePlan(planID)
		if err != nil {
			return fmt.Errorf("failed to delete plan: %w", err)
		}

		fmt.Printf("✓ Plan %s deleted\n", planID)
		return nil
	},
}

func init() {
	// plan create flags
	planCreateCmd.Flags().StringP("mission", "m", "", "Mission ID (defaults to context)")
	planCreateCmd.Flags().StringP("description", "d", "", "Plan description")
	planCreateCmd.Flags().StringP("content", "c", "", "Plan content")
	planCreateCmd.Flags().String("shipment", "", "Shipment ID to attach plan to")

	// plan list flags
	planListCmd.Flags().StringP("mission", "m", "", "Filter by mission")
	planListCmd.Flags().String("shipment", "", "Filter by shipment")
	planListCmd.Flags().StringP("status", "s", "", "Filter by status (draft, approved)")

	// plan update flags
	planUpdateCmd.Flags().String("title", "", "New title")
	planUpdateCmd.Flags().StringP("description", "d", "", "New description")
	planUpdateCmd.Flags().StringP("content", "c", "", "New content")

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
