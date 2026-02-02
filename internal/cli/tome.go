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

var tomeCmd = &cobra.Command{
	Use:   "tome",
	Short: "Manage tomes (knowledge organization containers)",
	Long:  "Create, list, close, and manage tomes in the ORC ledger",
}

var tomeCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new tome",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		commissionID, _ := cmd.Flags().GetString("commission")
		conclaveID, _ := cmd.Flags().GetString("conclave")
		description, _ := cmd.Flags().GetString("description")

		// Get commission from context or require explicit flag
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
			if commissionID == "" {
				return fmt.Errorf("no commission context detected\nHint: Use --commission flag or run from a workbench directory")
			}
		}

		// Validate entity IDs
		if err := validateEntityID(conclaveID, "conclave"); err != nil {
			return err
		}

		// Resolve container (optional - tomes can exist at commission root)
		var containerID, containerType string
		if conclaveID != "" {
			containerID = conclaveID
			containerType = "conclave"
		}
		// else: containerID and containerType remain empty (root tome at commission level)

		resp, err := wire.TomeService().CreateTome(ctx, primary.CreateTomeRequest{
			CommissionID:  commissionID,
			ConclaveID:    conclaveID,
			Title:         title,
			Description:   description,
			ContainerID:   containerID,
			ContainerType: containerType,
		})
		if err != nil {
			return fmt.Errorf("failed to create tome: %w", err)
		}

		tome := resp.Tome
		fmt.Printf("✓ Created tome %s: %s\n", tome.ID, tome.Title)
		fmt.Printf("  Commission: %s\n", tome.CommissionID)
		if tome.ContainerType == "conclave" {
			fmt.Printf("  Conclave: %s\n", tome.ContainerID)
		} else {
			fmt.Printf("  Location: commission root\n")
		}
		fmt.Printf("  Status: %s\n", tome.Status)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("   orc note create \"Learning title\" --tome %s --type learning\n", tome.ID)
		return nil
	},
}

var tomeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tomes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		commissionID, _ := cmd.Flags().GetString("commission")
		status, _ := cmd.Flags().GetString("status")

		// Get commission from context if not specified
		if commissionID == "" {
			commissionID = orccontext.GetContextCommissionID()
		}

		tomes, err := wire.TomeService().ListTomes(ctx, primary.TomeFilters{
			CommissionID: commissionID,
			Status:       status,
		})
		if err != nil {
			return fmt.Errorf("failed to list tomes: %w", err)
		}

		if len(tomes) == 0 {
			fmt.Println("No tomes found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tCOMMISSION")
		fmt.Fprintln(w, "--\t-----\t------\t-------")
		for _, t := range tomes {
			pinnedMark := ""
			if t.Pinned {
				pinnedMark = " [pinned]"
			}
			statusIcon := "📚"
			if t.Status == "closed" {
				statusIcon = "✅"
			}
			fmt.Fprintf(w, "%s\t%s%s\t%s %s\t%s\n", t.ID, t.Title, pinnedMark, statusIcon, t.Status, t.CommissionID)
		}
		w.Flush()
		return nil
	},
}

var tomeShowCmd = &cobra.Command{
	Use:   "show [tome-id]",
	Short: "Show tome details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		tome, err := wire.TomeService().GetTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("tome not found: %w", err)
		}

		fmt.Printf("Tome: %s\n", tome.ID)
		fmt.Printf("Title: %s\n", tome.Title)
		if tome.Description != "" {
			fmt.Printf("Description: %s\n", tome.Description)
		}
		fmt.Printf("Status: %s\n", tome.Status)
		fmt.Printf("Commission: %s\n", tome.CommissionID)
		if tome.Pinned {
			fmt.Printf("Pinned: yes\n")
		}
		fmt.Printf("Created: %s\n", tome.CreatedAt)
		if tome.ClosedAt != "" {
			fmt.Printf("Closed: %s\n", tome.ClosedAt)
		}

		// Show notes in this tome
		notes, err := wire.TomeService().GetTomeNotes(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to get notes: %w", err)
		}

		if len(notes) > 0 {
			fmt.Printf("\nNotes (%d):\n", len(notes))
			for _, note := range notes {
				typeStr := ""
				if note.Type != "" {
					typeStr = fmt.Sprintf(" [%s]", note.Type)
				}
				fmt.Printf("  📝 %s: %s%s\n", note.ID, note.Title, typeStr)
			}
		}

		return nil
	},
}

var tomeCloseCmd = &cobra.Command{
	Use:   "close [tome-id]",
	Short: "Mark tome as closed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().CloseTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to close tome: %w", err)
		}

		fmt.Printf("✓ Tome %s marked as closed\n", tomeID)
		return nil
	},
}

var tomeUpdateCmd = &cobra.Command{
	Use:   "update [tome-id]",
	Short: "Update tome title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify --title and/or --description")
		}

		err := wire.TomeService().UpdateTome(ctx, primary.UpdateTomeRequest{
			TomeID:      tomeID,
			Title:       title,
			Description: description,
		})
		if err != nil {
			return fmt.Errorf("failed to update tome: %w", err)
		}

		fmt.Printf("✓ Tome %s updated\n", tomeID)
		return nil
	},
}

var tomePinCmd = &cobra.Command{
	Use:   "pin [tome-id]",
	Short: "Pin tome to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().PinTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to pin tome: %w", err)
		}

		fmt.Printf("✓ Tome %s pinned 📌\n", tomeID)
		return nil
	},
}

var tomeUnpinCmd = &cobra.Command{
	Use:   "unpin [tome-id]",
	Short: "Unpin tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().UnpinTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to unpin tome: %w", err)
		}

		fmt.Printf("✓ Tome %s unpinned\n", tomeID)
		return nil
	},
}

var tomeDeleteCmd = &cobra.Command{
	Use:   "delete [tome-id]",
	Short: "Delete a tome",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]

		err := wire.TomeService().DeleteTome(ctx, tomeID)
		if err != nil {
			return fmt.Errorf("failed to delete tome: %w", err)
		}

		fmt.Printf("✓ Tome %s deleted\n", tomeID)
		return nil
	},
}

var tomeParkCmd = &cobra.Command{
	Use:   "park [tome-id]",
	Short: "[DEPRECATED] Library entity has been removed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Library entity has been removed. Tomes now exist at commission root when not in a conclave.\nTo move a tome to commission root, use: orc tome update %s (no container change needed - it's already there)", args[0])
	},
}

var tomeUnparkCmd = &cobra.Command{
	Use:   "unpark [tome-id]",
	Short: "Move tome to Conclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		tomeID := args[0]
		conclaveID, _ := cmd.Flags().GetString("conclave")

		// Validate entity IDs
		if err := validateEntityID(conclaveID, "conclave"); err != nil {
			return err
		}

		if conclaveID == "" {
			return fmt.Errorf("specify --conclave CON-xxx")
		}

		if err := wire.TomeService().UnparkTome(ctx, tomeID, conclaveID); err != nil {
			return fmt.Errorf("failed to unpark tome: %w", err)
		}

		fmt.Printf("Moved %s to %s\n", tomeID, conclaveID)
		return nil
	},
}

func init() {
	// tome create flags
	tomeCreateCmd.Flags().StringP("commission", "c", "", "Commission ID (defaults to context)")
	tomeCreateCmd.Flags().String("conclave", "", "Parent conclave ID (CON-xxx)")
	tomeCreateCmd.Flags().StringP("description", "d", "", "Tome description")

	// tome list flags
	tomeListCmd.Flags().StringP("commission", "c", "", "Filter by commission")
	tomeListCmd.Flags().StringP("status", "s", "", "Filter by status (open, closed)")

	// tome update flags
	tomeUpdateCmd.Flags().String("title", "", "New title")
	tomeUpdateCmd.Flags().StringP("description", "d", "", "New description")

	// tome unpark flags
	tomeUnparkCmd.Flags().String("conclave", "", "Target conclave ID (CON-xxx)")

	// Register subcommands
	tomeCmd.AddCommand(tomeCreateCmd)
	tomeCmd.AddCommand(tomeListCmd)
	tomeCmd.AddCommand(tomeShowCmd)
	tomeCmd.AddCommand(tomeCloseCmd)
	tomeCmd.AddCommand(tomeUpdateCmd)
	tomeCmd.AddCommand(tomePinCmd)
	tomeCmd.AddCommand(tomeUnpinCmd)
	tomeCmd.AddCommand(tomeDeleteCmd)
	tomeCmd.AddCommand(tomeParkCmd)
	tomeCmd.AddCommand(tomeUnparkCmd)
}

// TomeCmd returns the tome command
func TomeCmd() *cobra.Command {
	return tomeCmd
}
