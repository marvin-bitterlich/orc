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

var recCmd = &cobra.Command{
	Use:   "rec",
	Short: "Manage receipts",
	Long:  "Create, list, update, and manage receipts (RECs) in the ORC ledger",
}

var recCreateCmd = &cobra.Command{
	Use:   "create [delivered-outcome]",
	Short: "Create a new receipt",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		evidence, _ := cmd.Flags().GetString("evidence")
		notes, _ := cmd.Flags().GetString("notes")

		if shipmentID == "" {
			return fmt.Errorf("--shipment-id flag is required")
		}

		req := primary.CreateReceiptRequest{
			ShipmentID:        shipmentID,
			DeliveredOutcome:  args[0],
			Evidence:          evidence,
			VerificationNotes: notes,
		}

		resp, err := wire.ReceiptService().CreateReceipt(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create REC: %w", err)
		}

		rec := resp.Receipt
		fmt.Printf("✓ Created receipt %s\n", rec.ID)
		fmt.Printf("  Delivered Outcome: %s\n", rec.DeliveredOutcome)
		fmt.Printf("  Shipment: %s\n", rec.ShipmentID)
		fmt.Printf("  Status: %s\n", rec.Status)
		return nil
	},
}

var recListCmd = &cobra.Command{
	Use:   "list",
	Short: "List receipts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		shipmentID, _ := cmd.Flags().GetString("shipment-id")
		status, _ := cmd.Flags().GetString("status")

		recs, err := wire.ReceiptService().ListReceipts(ctx, primary.ReceiptFilters{
			ShipmentID: shipmentID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list RECs: %w", err)
		}

		if len(recs) == 0 {
			fmt.Println("No receipts found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSHIPMENT\tDELIVERED OUTCOME\tSTATUS")
		fmt.Fprintln(w, "--\t--------\t-----------------\t------")
		for _, item := range recs {
			// Truncate outcome for display
			outcome := item.DeliveredOutcome
			if len(outcome) > 40 {
				outcome = outcome[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.ID,
				item.ShipmentID,
				outcome,
				item.Status,
			)
		}
		w.Flush()
		return nil
	},
}

var recShowCmd = &cobra.Command{
	Use:   "show [rec-id]",
	Short: "Show receipt details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		recID := args[0]

		rec, err := wire.ReceiptService().GetReceipt(ctx, recID)
		if err != nil {
			return fmt.Errorf("REC not found: %w", err)
		}

		fmt.Printf("Receipt: %s\n", rec.ID)
		fmt.Printf("Shipment: %s\n", rec.ShipmentID)
		fmt.Printf("Delivered Outcome: %s\n", rec.DeliveredOutcome)
		if rec.Evidence != "" {
			fmt.Printf("Evidence: %s\n", rec.Evidence)
		}
		if rec.VerificationNotes != "" {
			fmt.Printf("Verification Notes: %s\n", rec.VerificationNotes)
		}
		fmt.Printf("Status: %s\n", rec.Status)
		fmt.Printf("Created: %s\n", rec.CreatedAt)
		fmt.Printf("Updated: %s\n", rec.UpdatedAt)

		return nil
	},
}

var recUpdateCmd = &cobra.Command{
	Use:   "update [rec-id]",
	Short: "Update a receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		recID := args[0]

		outcome, _ := cmd.Flags().GetString("outcome")
		evidence, _ := cmd.Flags().GetString("evidence")
		notes, _ := cmd.Flags().GetString("notes")

		req := primary.UpdateReceiptRequest{
			ReceiptID:         recID,
			DeliveredOutcome:  outcome,
			Evidence:          evidence,
			VerificationNotes: notes,
		}

		err := wire.ReceiptService().UpdateReceipt(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to update REC: %w", err)
		}

		fmt.Printf("Receipt %s updated\n", recID)
		return nil
	},
}

var recSubmitCmd = &cobra.Command{
	Use:   "submit [rec-id]",
	Short: "Submit a draft receipt for verification",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		recID := args[0]

		err := wire.ReceiptService().SubmitReceipt(ctx, recID)
		if err != nil {
			return fmt.Errorf("failed to submit REC: %w", err)
		}

		fmt.Printf("Receipt %s submitted for verification\n", recID)
		return nil
	},
}

var recVerifyCmd = &cobra.Command{
	Use:   "verify [rec-id]",
	Short: "Verify a submitted receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		recID := args[0]

		err := wire.ReceiptService().VerifyReceipt(ctx, recID)
		if err != nil {
			return fmt.Errorf("failed to verify REC: %w", err)
		}

		fmt.Printf("Receipt %s verified\n", recID)
		return nil
	},
}

var recDeleteCmd = &cobra.Command{
	Use:   "delete [rec-id]",
	Short: "Delete a receipt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		recID := args[0]

		err := wire.ReceiptService().DeleteReceipt(ctx, recID)
		if err != nil {
			return fmt.Errorf("failed to delete REC: %w", err)
		}

		fmt.Printf("✓ Receipt %s deleted\n", recID)
		return nil
	},
}

func init() {
	// rec create flags
	recCreateCmd.Flags().String("shipment-id", "", "Shipment ID (required)")
	recCreateCmd.Flags().String("evidence", "", "Evidence supporting the delivery")
	recCreateCmd.Flags().String("notes", "", "Verification notes")
	recCreateCmd.MarkFlagRequired("shipment-id")

	// rec list flags
	recListCmd.Flags().String("shipment-id", "", "Filter by shipment")
	recListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// rec update flags
	recUpdateCmd.Flags().String("outcome", "", "Update delivered outcome")
	recUpdateCmd.Flags().String("evidence", "", "Update evidence")
	recUpdateCmd.Flags().String("notes", "", "Update verification notes")

	// Register subcommands
	recCmd.AddCommand(recCreateCmd)
	recCmd.AddCommand(recListCmd)
	recCmd.AddCommand(recShowCmd)
	recCmd.AddCommand(recUpdateCmd)
	recCmd.AddCommand(recSubmitCmd)
	recCmd.AddCommand(recVerifyCmd)
	recCmd.AddCommand(recDeleteCmd)
}

// ReceiptCmd returns the rec command
func ReceiptCmd() *cobra.Command {
	return recCmd
}
