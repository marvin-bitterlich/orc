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

var conclaveCmd = &cobra.Command{
	Use:   "conclave",
	Short: "Manage conclaves (ideation session containers)",
	Long:  "Create, list, complete, and manage conclaves in the ORC ledger",
}

var conclaveCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new conclave",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		description, _ := cmd.Flags().GetString("description")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		resp, err := wire.ConclaveService().CreateConclave(context.Background(), primary.CreateConclaveRequest{
			CommissionID: commissionID,
			Title:        title,
			Description:  description,
		})
		if err != nil {
			return fmt.Errorf("failed to create conclave: %w", err)
		}

		fmt.Printf("✓ Created conclave %s: %s\n", resp.Conclave.ID, resp.Conclave.Title)
		fmt.Printf("  Commission: %s\n", resp.Conclave.CommissionID)
		fmt.Printf("  Status: %s\n", resp.Conclave.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("   Conclaves collect tasks, questions, and plans generated during ideation")
		return nil
	},
}

var conclaveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List conclaves",
	RunE: func(cmd *cobra.Command, args []string) error {
		commissionID, _ := cmd.Flags().GetString("commission")
		status, _ := cmd.Flags().GetString("status")

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orcctx.GetContextCommissionID()
		}

		conclaves, err := wire.ConclaveService().ListConclaves(context.Background(), primary.ConclaveFilters{
			CommissionID: commissionID,
			Status:       status,
		})
		if err != nil {
			return fmt.Errorf("failed to list conclaves: %w", err)
		}

		if len(conclaves) == 0 {
			fmt.Println("No conclaves found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tCOMMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, c := range conclaves {
			pinnedMark := ""
			if c.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "🧠"
			if c.Status == "closed" {
				statusIcon = "✅"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", c.ID, c.Title, pinnedMark, statusIcon, c.Status, c.CommissionID)
		}
		w.Flush()
		return nil
	},
}

var conclaveShowCmd = &cobra.Command{
	Use:   "show [conclave-id]",
	Short: "Show conclave details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]
		ctx := context.Background()

		conclave, err := wire.ConclaveService().GetConclave(ctx, conclaveID)
		if err != nil {
			return fmt.Errorf("conclave not found: %w", err)
		}

		fmt.Printf("Conclave: %s\n", conclave.ID)
		fmt.Printf("Title: %s\n", conclave.Title)
		if conclave.Description != "" {
			fmt.Printf("Description: %s\n", conclave.Description)
		}
		fmt.Printf("Status: %s\n", conclave.Status)
		fmt.Printf("Commission: %s\n", conclave.CommissionID)
		if conclave.ShipmentID != "" {
			fmt.Printf("Shipment: %s\n", conclave.ShipmentID)
		}
		if conclave.Decision != "" {
			fmt.Printf("Decision: %s\n", conclave.Decision)
		}
		if conclave.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", conclave.CreatedAt)
		if conclave.DecidedAt != "" {
			fmt.Printf("Decided: %s\n", conclave.DecidedAt)
		}

		// Show tasks in this conclave
		tasks, err := wire.ConclaveService().GetConclaveTasks(ctx, conclaveID)
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

		// Show plans in this conclave
		plans, err := wire.ConclaveService().GetConclavePlans(ctx, conclaveID)
		if err != nil {
			return fmt.Errorf("failed to get plans: %w", err)
		}

		if len(plans) > 0 {
			fmt.Printf("\nPlans (%d):\n", len(plans))
			for _, p := range plans {
				statusIcon := "📝"
				if p.Status == "approved" {
					statusIcon = "✅"
				}
				fmt.Printf("  %s %s: %s [%s]\n", statusIcon, p.ID, p.Title, p.Status)
			}
		}

		return nil
	},
}

var conclaveCompleteCmd = &cobra.Command{
	Use:   "complete [conclave-id]",
	Short: "Mark conclave as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().CompleteConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to complete conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s marked as complete\n", conclaveID)
		return nil
	},
}

var conclavePauseCmd = &cobra.Command{
	Use:   "pause [conclave-id]",
	Short: "Pause an active conclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().PauseConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to pause conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s paused\n", conclaveID)
		return nil
	},
}

var conclaveResumeCmd = &cobra.Command{
	Use:   "resume [conclave-id]",
	Short: "Resume a paused conclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().ResumeConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to resume conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s resumed\n", conclaveID)
		return nil
	},
}

var conclaveUpdateCmd = &cobra.Command{
	Use:   "update [conclave-id]",
	Short: "Update conclave title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := wire.ConclaveService().UpdateConclave(context.Background(), primary.UpdateConclaveRequest{
			ConclaveID:  conclaveID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to update conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s updated\n", conclaveID)
		return nil
	},
}

var conclavePinCmd = &cobra.Command{
	Use:   "pin [conclave-id]",
	Short: "Pin conclave to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().PinConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to pin conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s pinned 📌\n", conclaveID)
		return nil
	},
}

var conclaveUnpinCmd = &cobra.Command{
	Use:   "unpin [conclave-id]",
	Short: "Unpin conclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().UnpinConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to unpin conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s unpinned\n", conclaveID)
		return nil
	},
}

var conclaveDeleteCmd = &cobra.Command{
	Use:   "delete [conclave-id]",
	Short: "Delete a conclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conclaveID := args[0]

		err := wire.ConclaveService().DeleteConclave(context.Background(), conclaveID)
		if err != nil {
			return fmt.Errorf("failed to delete conclave: %w", err)
		}

		fmt.Printf("✓ Conclave %s deleted\n", conclaveID)
		return nil
	},
}

func init() {
	// conclave create flags
	conclaveCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	conclaveCreateCmd.Flags().StringP("description", "d", "", "Conclave description")

	// conclave list flags
	conclaveListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	conclaveListCmd.Flags().StringP("status", "s", "", "Filter by status (open, paused, closed)")

	// conclave update flags
	conclaveUpdateCmd.Flags().String("title", "", "New title")
	conclaveUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// Register subcommands
	conclaveCmd.AddCommand(conclaveCreateCmd)
	conclaveCmd.AddCommand(conclaveListCmd)
	conclaveCmd.AddCommand(conclaveShowCmd)
	conclaveCmd.AddCommand(conclaveCompleteCmd)
	conclaveCmd.AddCommand(conclavePauseCmd)
	conclaveCmd.AddCommand(conclaveResumeCmd)
	conclaveCmd.AddCommand(conclaveUpdateCmd)
	conclaveCmd.AddCommand(conclavePinCmd)
	conclaveCmd.AddCommand(conclaveUnpinCmd)
	conclaveCmd.AddCommand(conclaveDeleteCmd)
}

// ConclaveCmd returns the conclave command
func ConclaveCmd() *cobra.Command {
	return conclaveCmd
}
