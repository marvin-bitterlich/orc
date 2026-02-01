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

var kennelCmd = &cobra.Command{
	Use:   "kennel",
	Short: "Manage kennels",
	Long:  "List and view kennels (Watchdog seats) in the ORC ledger",
}

var kennelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List kennels",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workbenchID, _ := cmd.Flags().GetString("workbench")
		status, _ := cmd.Flags().GetString("status")

		kennels, err := wire.KennelService().ListKennels(ctx, primary.KennelFilters{
			WorkbenchID: workbenchID,
			Status:      status,
		})
		if err != nil {
			return fmt.Errorf("failed to list kennels: %w", err)
		}

		if len(kennels) == 0 {
			fmt.Println("No kennels found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKBENCH\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t---------\t------\t-------")
		for _, item := range kennels {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.ID,
				item.WorkbenchID,
				item.Status,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var kennelShowCmd = &cobra.Command{
	Use:   "show [kennel-id]",
	Short: "Show kennel details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		kennelID := args[0]

		kennel, err := wire.KennelService().GetKennel(ctx, kennelID)
		if err != nil {
			return fmt.Errorf("kennel not found: %w", err)
		}

		fmt.Printf("Kennel: %s\n", kennel.ID)
		fmt.Printf("Workbench: %s\n", kennel.WorkbenchID)
		fmt.Printf("Status: %s\n", kennel.Status)
		fmt.Printf("Created: %s\n", kennel.CreatedAt)
		fmt.Printf("Updated: %s\n", kennel.UpdatedAt)

		return nil
	},
}

var kennelEnsureAllCmd = &cobra.Command{
	Use:   "ensure-all",
	Short: "Create kennels for all workbenches missing them",
	Long: `Create a kennel for each workbench that doesn't have one.

This is a data migration command used when introducing the kennel entity.
Each workbench should have exactly one kennel (1:1 relationship).

This command is idempotent - running it multiple times is safe.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		created, err := wire.KennelService().EnsureAllWorkbenchesHaveKennels(ctx)
		if err != nil {
			return fmt.Errorf("failed to ensure kennels: %w", err)
		}

		if len(created) == 0 {
			fmt.Println("All workbenches already have kennels.")
			return nil
		}

		fmt.Printf("Created %d kennel(s):\n", len(created))
		for _, id := range created {
			fmt.Printf("  - %s\n", id)
		}
		return nil
	},
}

func init() {
	// kennel list flags
	kennelListCmd.Flags().String("workbench", "", "Filter by workbench ID")
	kennelListCmd.Flags().StringP("status", "s", "", "Filter by status (vacant|occupied|away)")

	// Register subcommands
	kennelCmd.AddCommand(kennelListCmd)
	kennelCmd.AddCommand(kennelShowCmd)
	kennelCmd.AddCommand(kennelEnsureAllCmd)
}

// KennelCmd returns the kennel command
func KennelCmd() *cobra.Command {
	return kennelCmd
}
