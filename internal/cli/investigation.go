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

var investigationCmd = &cobra.Command{
	Use:     "investigation",
	Aliases: []string{"inv"},
	Short:   "Manage investigations (research containers)",
	Long:    "Create, list, complete, and manage investigations in the ORC ledger",
}

var investigationCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new investigation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		conclaveID, _ := cmd.Flags().GetString("conclave")
		description, _ := cmd.Flags().GetString("description")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		ctx := context.Background()
		resp, err := wire.InvestigationService().CreateInvestigation(ctx, primary.CreateInvestigationRequest{
			CommissionID: commissionID,
			ConclaveID:   conclaveID,
			Title:        title,
			Description:  description,
		})
		if err != nil {
			return fmt.Errorf("failed to create investigation: %w", err)
		}

		investigation := resp.Investigation
		fmt.Printf("✓ Created investigation %s: %s\n", investigation.ID, investigation.Title)
		fmt.Printf("  Commission: %s\n", investigation.CommissionID)
		fmt.Printf("  Status: %s\n", investigation.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc note create \"Note title\" --investigation %s\n", investigation.ID)
		return nil
	},
}

var investigationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List investigations",
	RunE: func(cmd *cobra.Command, args []string) error {
		commissionID, _ := cmd.Flags().GetString("commission")
		conclaveID, _ := cmd.Flags().GetString("conclave")
		status, _ := cmd.Flags().GetString("status")

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
		}

		ctx := context.Background()
		investigations, err := wire.InvestigationService().ListInvestigations(ctx, primary.InvestigationFilters{
			CommissionID: commissionID,
			ConclaveID:   conclaveID,
			Status:       status,
		})
		if err != nil {
			return fmt.Errorf("failed to list investigations: %w", err)
		}

		if len(investigations) == 0 {
			fmt.Println("No investigations found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tCOMMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, inv := range investigations {
			pinnedMark := ""
			if inv.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "🔍"
			if inv.Status == "complete" {
				statusIcon = "✅"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", inv.ID, inv.Title, pinnedMark, statusIcon, inv.Status, inv.CommissionID)
		}
		w.Flush()
		return nil
	},
}

var investigationShowCmd = &cobra.Command{
	Use:   "show [investigation-id]",
	Short: "Show investigation details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		investigation, err := wire.InvestigationService().GetInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("investigation not found: %w", err)
		}

		fmt.Printf("Investigation: %s\n", investigation.ID)
		fmt.Printf("Title: %s\n", investigation.Title)
		if investigation.Description != "" {
			fmt.Printf("Description: %s\n", investigation.Description)
		}
		fmt.Printf("Status: %s\n", investigation.Status)
		fmt.Printf("Commission: %s\n", investigation.CommissionID)
		if investigation.ConclaveID != "" {
			fmt.Printf("Conclave: %s\n", investigation.ConclaveID)
		}
		if investigation.AssignedWorkbenchID != "" {
			fmt.Printf("Assigned Workbench: %s\n", investigation.AssignedWorkbenchID)
		}
		if investigation.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", investigation.CreatedAt)
		if investigation.CompletedAt != "" {
			fmt.Printf("Completed: %s\n", investigation.CompletedAt)
		}

		return nil
	},
}

var investigationCompleteCmd = &cobra.Command{
	Use:   "complete [investigation-id]",
	Short: "Mark investigation as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().CompleteInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to complete investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s marked as complete\n", investigationID)
		return nil
	},
}

var investigationPauseCmd = &cobra.Command{
	Use:   "pause [investigation-id]",
	Short: "Pause an active investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().PauseInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to pause investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s paused\n", investigationID)
		return nil
	},
}

var investigationResumeCmd = &cobra.Command{
	Use:   "resume [investigation-id]",
	Short: "Resume a paused investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().ResumeInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to resume investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s resumed\n", investigationID)
		return nil
	},
}

var investigationUpdateCmd = &cobra.Command{
	Use:   "update [investigation-id]",
	Short: "Update investigation title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		ctx := context.Background()
		err := wire.InvestigationService().UpdateInvestigation(ctx, primary.UpdateInvestigationRequest{
			InvestigationID: investigationID,
			Title:           title,
			Description:     description,
		})
		if err != nil {
			return fmt.Errorf("failed to update investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s updated\n", investigationID)
		return nil
	},
}

var investigationPinCmd = &cobra.Command{
	Use:   "pin [investigation-id]",
	Short: "Pin investigation to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().PinInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to pin investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s pinned 📌\n", investigationID)
		return nil
	},
}

var investigationUnpinCmd = &cobra.Command{
	Use:   "unpin [investigation-id]",
	Short: "Unpin investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().UnpinInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to unpin investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s unpinned\n", investigationID)
		return nil
	},
}

var investigationDeleteCmd = &cobra.Command{
	Use:   "delete [investigation-id]",
	Short: "Delete an investigation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]

		ctx := context.Background()
		err := wire.InvestigationService().DeleteInvestigation(ctx, investigationID)
		if err != nil {
			return fmt.Errorf("failed to delete investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s deleted\n", investigationID)
		return nil
	},
}

var investigationAssignCmd = &cobra.Command{
	Use:   "assign [investigation-id] [workbench-id]",
	Short: "Assign investigation to a workbench",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		investigationID := args[0]
		workbenchID := args[1]

		ctx := context.Background()
		err := wire.InvestigationService().AssignInvestigationToWorkbench(ctx, investigationID, workbenchID)
		if err != nil {
			return fmt.Errorf("failed to assign investigation: %w", err)
		}

		fmt.Printf("✓ Investigation %s assigned to workbench %s\n", investigationID, workbenchID)
		return nil
	},
}

func init() {
	// investigation create flags
	investigationCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	investigationCreateCmd.Flags().String("conclave", "", "Conclave ID (optional)")
	investigationCreateCmd.Flags().StringP("description", "d", "", "Investigation description")

	// investigation list flags
	investigationListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	investigationListCmd.Flags().String("conclave", "", "Filter by conclave")
	investigationListCmd.Flags().StringP("status", "s", "", "Filter by status (active, complete)")

	// investigation update flags
	investigationUpdateCmd.Flags().String("title", "", "New title")
	investigationUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	investigationCmd.AddCommand(investigationCreateCmd)
	investigationCmd.AddCommand(investigationListCmd)
	investigationCmd.AddCommand(investigationShowCmd)
	investigationCmd.AddCommand(investigationCompleteCmd)
	investigationCmd.AddCommand(investigationPauseCmd)
	investigationCmd.AddCommand(investigationResumeCmd)
	investigationCmd.AddCommand(investigationUpdateCmd)
	investigationCmd.AddCommand(investigationPinCmd)
	investigationCmd.AddCommand(investigationUnpinCmd)
	investigationCmd.AddCommand(investigationDeleteCmd)
	investigationCmd.AddCommand(investigationAssignCmd)
}

// InvestigationCmd returns the investigation command
func InvestigationCmd() *cobra.Command {
	return investigationCmd
}
