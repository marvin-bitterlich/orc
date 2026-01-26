package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var shipmentCmd = &cobra.Command{
	Use:   "shipment",
	Short: "Manage shipments (execution containers)",
	Long:  "Create, list, assign, and manage shipments in the ORC ledger",
}

var shipmentCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new shipment",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		description, _ := cmd.Flags().GetString("description")
		repoID, _ := cmd.Flags().GetString("repo")
		branch, _ := cmd.Flags().GetString("branch")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		resp, err := wire.ShipmentService().CreateShipment(ctx, primary.CreateShipmentRequest{
			CommissionID: commissionID,
			Title:        title,
			Description:  description,
			RepoID:       repoID,
			Branch:       branch,
		})
		if err != nil {
			return fmt.Errorf("failed to create shipment: %w", err)
		}

		fmt.Printf("✓ Created shipment %s: %s\n", resp.Shipment.ID, resp.Shipment.Title)
		fmt.Printf("  Under commission: %s\n", resp.Shipment.CommissionID)
		if resp.Shipment.Branch != "" {
			fmt.Printf("  Branch: %s\n", resp.Shipment.Branch)
		}
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc task create \"Task title\" --shipment %s\n", resp.Shipment.ID)
		return nil
	},
}

var shipmentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shipments",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		commissionID, _ := cmd.Flags().GetString("commission")
		status, _ := cmd.Flags().GetString("status")

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
		}

		shipments, err := wire.ShipmentService().ListShipments(ctx, primary.ShipmentFilters{
			CommissionID: commissionID,
			Status:       status,
		})
		if err != nil {
			return fmt.Errorf("failed to list shipments: %w", err)
		}

		if len(shipments) == 0 {
			fmt.Println("No shipments found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tCOMMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, s := range shipments {
			pinnedMark := ""
			if s.Pinned {
				pinnedMark = " [pinned]"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s\t%s\n", s.ID, s.Title, pinnedMark, s.Status, s.CommissionID)
		}
		w.Flush()
		return nil
	},
}

var shipmentShowCmd = &cobra.Command{
	Use:   "show [shipment-id]",
	Short: "Show shipment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		shipment, err := wire.ShipmentService().GetShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("shipment not found: %w", err)
		}

		fmt.Printf("Shipment: %s\n", shipment.ID)
		fmt.Printf("Title: %s\n", shipment.Title)
		if shipment.Description != "" {
			fmt.Printf("Description: %s\n", shipment.Description)
		}
		fmt.Printf("Status: %s\n", shipment.Status)
		fmt.Printf("Commission: %s\n", shipment.CommissionID)
		if shipment.AssignedWorkbenchID != "" {
			fmt.Printf("Assigned Workbench: %s\n", shipment.AssignedWorkbenchID)
		}
		if shipment.RepoID != "" {
			fmt.Printf("Repository: %s\n", shipment.RepoID)
		}
		if shipment.Branch != "" {
			fmt.Printf("Branch: %s\n", shipment.Branch)
		}
		if shipment.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", shipment.CreatedAt)
		if shipment.CompletedAt != "" {
			fmt.Printf("Completed: %s\n", shipment.CompletedAt)
		}

		// Show tasks
		tasks, err := wire.ShipmentService().GetShipmentTasks(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %w", err)
		}

		if len(tasks) > 0 {
			fmt.Printf("\nTasks (%d):\n", len(tasks))
			for _, task := range tasks {
				statusIcon := getStatusIcon(task.Status)
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, task.ID, task.Title, task.Status)
			}
		}

		return nil
	},
}

var shipmentCompleteCmd = &cobra.Command{
	Use:   "complete [shipment-id]",
	Short: "Mark shipment as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		err := wire.ShipmentService().CompleteShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to complete shipment: %w", err)
		}

		fmt.Printf("Shipment %s marked as complete\n", shipmentID)
		return nil
	},
}

var shipmentPauseCmd = &cobra.Command{
	Use:   "pause [shipment-id]",
	Short: "Pause an active shipment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		err := wire.ShipmentService().PauseShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to pause shipment: %w", err)
		}

		fmt.Printf("Shipment %s paused\n", shipmentID)
		return nil
	},
}

var shipmentResumeCmd = &cobra.Command{
	Use:   "resume [shipment-id]",
	Short: "Resume a paused shipment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		err := wire.ShipmentService().ResumeShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to resume shipment: %w", err)
		}

		fmt.Printf("Shipment %s resumed\n", shipmentID)
		return nil
	},
}

var shipmentUpdateCmd = &cobra.Command{
	Use:   "update [shipment-id]",
	Short: "Update shipment title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := wire.ShipmentService().UpdateShipment(ctx, primary.UpdateShipmentRequest{
			ShipmentID:  shipmentID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to update shipment: %w", err)
		}

		fmt.Printf("Shipment %s updated\n", shipmentID)
		return nil
	},
}

var shipmentPinCmd = &cobra.Command{
	Use:   "pin [shipment-id]",
	Short: "Pin shipment to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		err := wire.ShipmentService().PinShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to pin shipment: %w", err)
		}

		fmt.Printf("Shipment %s pinned\n", shipmentID)
		return nil
	},
}

var shipmentUnpinCmd = &cobra.Command{
	Use:   "unpin [shipment-id]",
	Short: "Unpin shipment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]

		err := wire.ShipmentService().UnpinShipment(ctx, shipmentID)
		if err != nil {
			return fmt.Errorf("failed to unpin shipment: %w", err)
		}

		fmt.Printf("Shipment %s unpinned\n", shipmentID)
		return nil
	},
}

var shipmentAssignCmd = &cobra.Command{
	Use:   "assign [shipment-id] [workbench-id]",
	Short: "Assign shipment to a workbench",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID := args[0]
		workbenchID := args[1]

		err := wire.ShipmentService().AssignShipmentToWorkbench(ctx, shipmentID, workbenchID)
		if err != nil {
			return fmt.Errorf("failed to assign shipment: %w", err)
		}

		fmt.Printf("Shipment %s assigned to workbench %s\n", shipmentID, workbenchID)
		return nil
	},
}

func init() {
	// shipment create flags
	shipmentCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	shipmentCreateCmd.Flags().StringP("description", "d", "", "Shipment description")
	shipmentCreateCmd.Flags().StringP("repo", "r", "", "Repository ID to link for branch ownership")
	shipmentCreateCmd.Flags().String("branch", "", "Override auto-generated branch name")

	// shipment list flags
	shipmentListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	shipmentListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// shipment update flags
	shipmentUpdateCmd.Flags().String("title", "", "New title")
	shipmentUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	shipmentCmd.AddCommand(shipmentCreateCmd)
	shipmentCmd.AddCommand(shipmentListCmd)
	shipmentCmd.AddCommand(shipmentShowCmd)
	shipmentCmd.AddCommand(shipmentCompleteCmd)
	shipmentCmd.AddCommand(shipmentPauseCmd)
	shipmentCmd.AddCommand(shipmentResumeCmd)
	shipmentCmd.AddCommand(shipmentUpdateCmd)
	shipmentCmd.AddCommand(shipmentPinCmd)
	shipmentCmd.AddCommand(shipmentUnpinCmd)
	shipmentCmd.AddCommand(shipmentAssignCmd)
}

// ShipmentCmd returns the shipment command
func ShipmentCmd() *cobra.Command {
	return shipmentCmd
}
