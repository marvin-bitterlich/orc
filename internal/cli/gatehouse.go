package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var gatehouseCmd = &cobra.Command{
	Use:   "gatehouse",
	Short: "Manage gatehouses",
	Long:  "List and view gatehouses (Goblin seats) in the ORC ledger",
}

var gatehouseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List gatehouses",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		workshopID, _ := cmd.Flags().GetString("workshop")
		status, _ := cmd.Flags().GetString("status")

		gatehouses, err := wire.GatehouseService().ListGatehouses(ctx, primary.GatehouseFilters{
			WorkshopID: workshopID,
			Status:     status,
		})
		if err != nil {
			return fmt.Errorf("failed to list gatehouses: %w", err)
		}

		if len(gatehouses) == 0 {
			fmt.Println("No gatehouses found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKSHOP\tSTATUS\tCREATED")
		fmt.Fprintln(w, "--\t--------\t------\t-------")
		for _, item := range gatehouses {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.ID,
				item.WorkshopID,
				item.Status,
				item.CreatedAt,
			)
		}
		w.Flush()
		return nil
	},
}

var gatehouseShowCmd = &cobra.Command{
	Use:   "show [gatehouse-id]",
	Short: "Show gatehouse details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		gatehouseID := args[0]

		gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, gatehouseID)
		if err != nil {
			return fmt.Errorf("gatehouse not found: %w", err)
		}

		fmt.Printf("Gatehouse: %s\n", gatehouse.ID)
		fmt.Printf("Workshop: %s\n", gatehouse.WorkshopID)
		fmt.Printf("Status: %s\n", gatehouse.Status)
		fmt.Printf("Created: %s\n", gatehouse.CreatedAt)
		fmt.Printf("Updated: %s\n", gatehouse.UpdatedAt)

		return nil
	},
}

var gatehouseEnsureAllCmd = &cobra.Command{
	Use:   "ensure-all",
	Short: "Create gatehouses for all workshops missing them",
	Long: `Create a gatehouse for each workshop that doesn't have one.

This is a data migration command used when introducing the gatehouse entity.
Each workshop should have exactly one gatehouse (1:1 relationship).

This command is idempotent - running it multiple times is safe.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()

		created, err := wire.GatehouseService().EnsureAllWorkshopsHaveGatehouses(ctx)
		if err != nil {
			return fmt.Errorf("failed to ensure gatehouses: %w", err)
		}

		if len(created) == 0 {
			fmt.Println("All workshops already have gatehouses.")
			return nil
		}

		fmt.Printf("Created %d gatehouse(s):\n", len(created))
		for _, id := range created {
			fmt.Printf("  - %s\n", id)
		}
		return nil
	},
}

func init() {
	// gatehouse list flags
	gatehouseListCmd.Flags().String("workshop", "", "Filter by workshop ID")
	gatehouseListCmd.Flags().StringP("status", "s", "", "Filter by status")

	// Register subcommands
	gatehouseCmd.AddCommand(gatehouseListCmd)
	gatehouseCmd.AddCommand(gatehouseShowCmd)
	gatehouseCmd.AddCommand(gatehouseEnsureAllCmd)
}

// GatehouseCmd returns the gatehouse command
func GatehouseCmd() *cobra.Command {
	return gatehouseCmd
}
