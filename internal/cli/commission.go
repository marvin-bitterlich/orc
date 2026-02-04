package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var commissionCmd = &cobra.Command{
	Use:   "commission",
	Short: "Manage commissions (strategic work streams)",
	Long:  "Create, list, and manage commissions in the ORC ledger",
}

var commissionCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new commission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		return wire.CommissionAdapter().Create(ctx, title, description)
	},
}

var commissionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List commissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		status, _ := cmd.Flags().GetString("status")

		return wire.CommissionAdapter().List(ctx, status)
	},
}

var commissionShowCmd = &cobra.Command{
	Use:   "show [commission-id]",
	Short: "Show commission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		id := args[0]

		// Show commission details via adapter
		_, err := wire.CommissionAdapter().Show(ctx, id)
		if err != nil {
			return err
		}

		// List shipments under this commission via service
		shipments, err := wire.ShipmentService().ListShipments(ctx, primary.ShipmentFilters{CommissionID: id})
		if err == nil && len(shipments) > 0 {
			fmt.Println("Shipments:")
			for _, shipment := range shipments {
				fmt.Printf("  - %s [%s] %s\n", shipment.ID, shipment.Status, shipment.Title)
			}
			fmt.Println()
		}

		return nil
	},
}

var commissionStartCmd = &cobra.Command{
	Use:   "start [commission-id]",
	Short: "Start a TMux session for a commission",
	Long: `Start a TMux session for a commission with existing workshop infrastructure.

Prerequisites:
- Workshop infrastructure must exist (run 'orc infra apply <workshop-id>' first)
- Commission must have associated workshops

This command creates a TMux session with windows for the ORC orchestrator.

Examples:
  orc commission start COMM-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()

		// Check agent identity - only ORC can start commissions
		if err := wire.CommissionOrchestrationService().CheckLaunchPermission(ctx); err != nil {
			return err
		}

		commissionID := args[0]

		// Validate Claude workspace trust
		if err := validateClaudeWorkspaceTrust(); err != nil {
			return fmt.Errorf("Claude workspace trust validation failed:\n\n%w\n\nRun 'orc doctor' for detailed diagnostics", err)
		}

		// Get commission from service
		commission, err := wire.CommissionService().GetCommission(context.Background(), commissionID)
		if err != nil {
			return fmt.Errorf("failed to get commission: %w", err)
		}

		// Use home directory as session working directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		fmt.Printf("Starting TMux session for commission: %s - %s\n", commission.ID, commission.Title)
		fmt.Println()

		// Create TMux session
		sessionName := fmt.Sprintf("orc-%s", commissionID)
		tmuxAdapter := wire.TMuxAdapter()

		// Check if session already exists
		if tmuxAdapter.SessionExists(ctx, sessionName) {
			return fmt.Errorf("TMux session '%s' already exists. Attach with: tmux attach -t %s", sessionName, sessionName)
		}

		fmt.Printf("Creating TMux session: %s\n", sessionName)

		// Create session with base numbering from 1
		if err := tmuxAdapter.CreateSession(ctx, sessionName, home); err != nil {
			return fmt.Errorf("failed to create TMux session: %w", err)
		}

		// Create ORC window (window 1) with claude
		if err := tmuxAdapter.CreateOrcWindow(ctx, sessionName, home); err != nil {
			return fmt.Errorf("failed to create ORC window: %w", err)
		}
		fmt.Printf("  ✓ Window 1: orc (claude | vim | shell)\n")

		// Select the ORC window (window 1) as default
		tmuxAdapter.SelectWindow(ctx, sessionName, 1)

		fmt.Println()
		fmt.Printf("Commission session ready!\n")
		fmt.Println()
		fmt.Println(tmuxAdapter.AttachInstructions(sessionName))

		return nil
	},
}

var commissionCompleteCmd = &cobra.Command{
	Use:   "complete [commission-id]",
	Short: "Mark a commission as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		return wire.CommissionAdapter().Complete(ctx, args[0])
	},
}

var commissionArchiveCmd = &cobra.Command{
	Use:   "archive [commission-id]",
	Short: "Archive a completed commission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		return wire.CommissionAdapter().Archive(ctx, args[0])
	},
}

var commissionUpdateCmd = &cobra.Command{
	Use:   "update [commission-id]",
	Short: "Update commission title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		return wire.CommissionAdapter().Update(ctx, id, title, description)
	},
}

var commissionDeleteCmd = &cobra.Command{
	Use:   "delete [commission-id]",
	Short: "Delete a commission from the database",
	Long: `Delete a commission and all associated data from the database.

WARNING: This is a destructive operation. Associated shipments, tasks, and workbenches
will lose their commission reference.

Examples:
  orc commission delete COMM-TEST-001
  orc commission delete COMM-001 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		return wire.CommissionAdapter().Delete(ctx, id, force)
	},
}

var commissionPinCmd = &cobra.Command{
	Use:   "pin [commission-id]",
	Short: "Pin commission to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		return wire.CommissionAdapter().Pin(ctx, args[0])
	},
}

var commissionUnpinCmd = &cobra.Command{
	Use:   "unpin [commission-id]",
	Short: "Unpin commission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		return wire.CommissionAdapter().Unpin(ctx, args[0])
	},
}

// CommissionCmd returns the commission command
func CommissionCmd() *cobra.Command {
	// Add flags
	commissionCreateCmd.Flags().StringP("description", "d", "", "Commission description")
	commissionListCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, complete, archived)")
	commissionUpdateCmd.Flags().StringP("title", "t", "", "New commission title")
	commissionUpdateCmd.Flags().StringP("description", "d", "", "New commission description")
	commissionDeleteCmd.Flags().BoolP("force", "f", false, "Force delete even with associated data")

	// Add subcommands
	commissionCmd.AddCommand(commissionCreateCmd)
	commissionCmd.AddCommand(commissionListCmd)
	commissionCmd.AddCommand(commissionShowCmd)
	commissionCmd.AddCommand(commissionStartCmd)
	commissionCmd.AddCommand(commissionCompleteCmd)
	commissionCmd.AddCommand(commissionArchiveCmd)
	commissionCmd.AddCommand(commissionUpdateCmd)
	commissionCmd.AddCommand(commissionDeleteCmd)
	commissionCmd.AddCommand(commissionPinCmd)
	commissionCmd.AddCommand(commissionUnpinCmd)

	return commissionCmd
}
