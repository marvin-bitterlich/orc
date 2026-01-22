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

var crecCmd = &cobra.Command{
	Use:   "crec",
	Short: "Manage cycle receipts",
	Long:  "Create, list, update, and manage cycle receipts (CRECs) in the ORC ledger",
}

var crecCreateCmd = &cobra.Command{
	Use:   "create [delivered-outcome]",
	Short: "Create a new cycle receipt",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID, _ := cmd.Flags().GetString("cwo-id")
		evidence, _ := cmd.Flags().GetString("evidence")
		notes, _ := cmd.Flags().GetString("notes")

		if cwoID == "" {
			return fmt.Errorf("--cwo-id flag is required")
		}

		req := primary.CreateCycleReceiptRequest{
			CWOID:             cwoID,
			DeliveredOutcome:  args[0],
			Evidence:          evidence,
			VerificationNotes: notes,
		}

		resp, err := wire.CycleReceiptService().CreateCycleReceipt(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create CREC: %w", err)
		}

		crec := resp.CycleReceipt
		fmt.Printf("Created cycle receipt %s\n", crec.ID)
		fmt.Printf("  Delivered Outcome: %s\n", crec.DeliveredOutcome)
		fmt.Printf("  CWO: %s\n", crec.CWOID)
		fmt.Printf("  Shipment: %s\n", crec.ShipmentID)
		fmt.Printf("  Status: %s\n", crec.Status)
		return nil
	},
}

var crecListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cycle receipts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cwoID, _ := cmd.Flags().GetString("cwo-id")
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		status, _ := cmd.Flags().GetString("status")

		crecs, err := wire.CycleReceiptService().ListCycleReceipts(ctx, primary.CycleReceiptFilters{
			CWOID:      cwoID,
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list CRECs: %w", err)
		}

		if len(crecs) == 0 {
			fmt.Println("No cycle receipts found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCWO\tSHIPMENT\tDELIVERED OUTCOME\tSTATUS")
		fmt.Fprintln(w, "--\t---\t--------\t-----------------\t------")
		for _, item := range crecs {
			// Truncate outcome for display
			outcome := item.DeliveredOutcome
			if len(outcome) > 30 {
				outcome = outcome[:27] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.CWOID,
				item.ShipmentID,
				outcome,
				item.Status,
			)
		}
		w.Flush()
		return nil
	},
}

var crecShowCmd = &cobra.Command{
	Use:   "show [crec-id]",
	Short: "Show cycle receipt details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		crecID := args[0]

		crec, err := wire.CycleReceiptService().GetCycleReceipt(ctx, crecID)
		if err != nil {
			return fmt.Errorf("CREC not found: %w", err)
		}

		fmt.Printf("Cycle Receipt: %s\n", crec.ID)
		fmt.Printf("CWO: %s\n", crec.CWOID)
		fmt.Printf("Shipment: %s\n", crec.ShipmentID)
		fmt.Printf("Delivered Outcome: %s\n", crec.DeliveredOutcome)
		if crec.Evidence != "" {
			fmt.Printf("Evidence: %s\n", crec.Evidence)
		}
		if crec.VerificationNotes != "" {
			fmt.Printf("Verification Notes: %s\n", crec.VerificationNotes)
		}
		fmt.Printf("Status: %s\n", crec.Status)
		fmt.Printf("Created: %s\n", crec.CreatedAt)
		fmt.Printf("Updated: %s\n", crec.UpdatedAt)

		return nil
	},
}

var crecUpdateCmd = &cobra.Command{
	Use:   "update [crec-id]",
	Short: "Update a cycle receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		crecID := args[0]

		outcome, _ := cmd.Flags().GetString("outcome")
		evidence, _ := cmd.Flags().GetString("evidence")
		notes, _ := cmd.Flags().GetString("notes")

		req := primary.UpdateCycleReceiptRequest{
			CycleReceiptID:    crecID,
			DeliveredOutcome:  outcome,
			Evidence:          evidence,
			VerificationNotes: notes,
		}

		err := wire.CycleReceiptService().UpdateCycleReceipt(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update CREC: %w", err)
		}

		fmt.Printf("Cycle receipt %s updated\n", crecID)
		return nil
	},
}

var crecSubmitCmd = &cobra.Command{
	Use:   "submit [crec-id]",
	Short: "Submit a draft cycle receipt for verification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		crecID := args[0]

		err := wire.CycleReceiptService().SubmitCycleReceipt(ctx, crecID)
		if err != nil {
			return fmt.Errorf("failed to submit CREC: %w", err)
		}

		fmt.Printf("Cycle receipt %s submitted for verification\n", crecID)
		return nil
	},
}

var crecVerifyCmd = &cobra.Command{
	Use:   "verify [crec-id]",
	Short: "Verify a submitted cycle receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		crecID := args[0]

		err := wire.CycleReceiptService().VerifyCycleReceipt(ctx, crecID)
		if err != nil {
			return fmt.Errorf("failed to verify CREC: %w", err)
		}

		fmt.Printf("Cycle receipt %s verified\n", crecID)
		return nil
	},
}

var crecDeleteCmd = &cobra.Command{
	Use:   "delete [crec-id]",
	Short: "Delete a cycle receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		crecID := args[0]

		err := wire.CycleReceiptService().DeleteCycleReceipt(ctx, crecID)
		if err != nil {
			return fmt.Errorf("failed to delete CREC: %w", err)
		}

		fmt.Printf("Cycle receipt %s deleted\n", crecID)
		return nil
	},
}

func init() {
	// crec create flags
	crecCreateCmd.Flags().String("cwo-id", "", "CWO ID (required)")
	crecCreateCmd.Flags().String("evidence", "", "Evidence supporting the delivery")
	crecCreateCmd.Flags().String("notes", "", "Verification notes")
	crecCreateCmd.MarkFlagRequired("cwo-id")

	// crec list flags
	crecListCmd.Flags().String("cwo-id", "", "Filter by CWO")
	crecListCmd.Flags().String("shipment-id", "", "Filter by shipment")
	crecListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// crec update flags
	crecUpdateCmd.Flags().String("outcome", "", "Update delivered outcome")
	crecUpdateCmd.Flags().String("evidence", "", "Update evidence")
	crecUpdateCmd.Flags().String("notes", "", "Update verification notes")

	// Register subcommands
	crecCmd.AddCommand(crecCreateCmd)
	crecCmd.AddCommand(crecListCmd)
	crecCmd.AddCommand(crecShowCmd)
	crecCmd.AddCommand(crecUpdateCmd)
	crecCmd.AddCommand(crecSubmitCmd)
	crecCmd.AddCommand(crecVerifyCmd)
	crecCmd.AddCommand(crecDeleteCmd)
}

// CycleReceiptCmd returns the crec command
func CycleReceiptCmd() *cobra.Command {
	return crecCmd
}
