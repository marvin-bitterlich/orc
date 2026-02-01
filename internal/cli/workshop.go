package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
	"github.com/example/orc/internal/wire"
)

// WorkshopCmd returns the workshop command
func WorkshopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workshop",
		Short: "Manage workshops (persistent places)",
		Long:  `Create and manage workshops - persistent places within factories that host workbenches.`,
	}

	cmd.AddCommand(workshopCreateCmd())
	cmd.AddCommand(workshopListCmd())
	cmd.AddCommand(workshopShowCmd())
	cmd.AddCommand(workshopDeleteCmd())
	cmd.AddCommand(workshopCloseCmd())
	cmd.AddCommand(workshopSetCommissionCmd())

	return cmd
}

func workshopCreateCmd() *cobra.Command {
	var factoryID string
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workshop",
		Long: `Create a new workshop within a factory.

A Workshop is a persistent named place within a Factory. Workshops
have atmospheric names from a pool (e.g., "Ironmoss Forge", "Blackpine Foundry").
If no name is provided, one will be assigned from the pool.

If no factory is specified, the "default" factory is used (and created if needed).

Examples:
  orc workshop create                              # uses default factory
  orc workshop create --factory FACT-001           # uses specific factory
  orc workshop create --name "Custom Workshop"     # uses default factory with custom name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			resp, err := wire.WorkshopService().CreateWorkshop(ctx, primary.CreateWorkshopRequest{
				FactoryID: factoryID,
				Name:      name,
			})
			if err != nil {
				return fmt.Errorf("failed to create workshop: %w", err)
			}

			fmt.Printf("✓ Created workshop %s: %s\n", resp.WorkshopID, resp.Workshop.Name)
			fmt.Printf("  Factory: %s\n", resp.Workshop.FactoryID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&factoryID, "factory", "f", "", "Factory ID (uses 'default' if not specified)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Workshop name (optional - uses name pool if empty)")

	return cmd
}

func workshopListCmd() *cobra.Command {
	var factoryID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workshops",
		Long:  `List all workshops with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workshops, err := wire.WorkshopService().ListWorkshops(ctx, primary.WorkshopFilters{
				FactoryID: factoryID,
			})
			if err != nil {
				return fmt.Errorf("failed to list workshops: %w", err)
			}

			if len(workshops) == 0 {
				fmt.Println("No workshops found.")
				fmt.Println()
				fmt.Println("Create your first workshop:")
				fmt.Println("  orc workshop create --factory FACT-001")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tFACTORY\tSTATUS")
			fmt.Fprintln(w, "--\t----\t-------\t------")

			for _, ws := range workshops {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					ws.ID,
					ws.Name,
					ws.FactoryID,
					ws.Status,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&factoryID, "factory", "f", "", "Filter by factory ID")

	return cmd
}

func workshopShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [workshop-id]",
		Short: "Show workshop details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workshop, err := wire.WorkshopService().GetWorkshop(ctx, args[0])
			if err != nil {
				return fmt.Errorf("workshop not found: %w", err)
			}

			fmt.Printf("Workshop: %s\n", workshop.ID)
			fmt.Printf("Name: %s\n", workshop.Name)
			fmt.Printf("Factory: %s\n", workshop.FactoryID)
			fmt.Printf("Status: %s\n", workshop.Status)
			fmt.Printf("Created: %s\n", workshop.CreatedAt)

			return nil
		},
	}
}

func workshopDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [workshop-id]",
		Short: "Delete a workshop",
		Long: `Delete a workshop from the database.

WARNING: This is a destructive operation. Workshops with workbenches
require the --force flag.

Examples:
  orc workshop delete WORK-001
  orc workshop delete WORK-001 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workshopID := args[0]

			err := wire.WorkshopService().DeleteWorkshop(ctx, primary.DeleteWorkshopRequest{
				WorkshopID: workshopID,
				Force:      force,
			})
			if err != nil {
				return err
			}

			fmt.Printf("✓ Workshop %s deleted\n", workshopID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with workbenches")

	return cmd
}

func workshopCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close [workshop-id]",
		Short: "Close a workshop TMux session",
		Long: `Close the TMux session for a workshop.

Examples:
  orc workshop close WORK-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := context.Background()

			if err := wire.WorkshopService().CloseWorkshop(ctx, workshopID); err != nil {
				return fmt.Errorf("failed to close session: %w", err)
			}

			fmt.Printf("✓ Workshop %s session closed\n", workshopID)
			return nil
		},
	}
}

func workshopSetCommissionCmd() *cobra.Command {
	var clearFlag bool

	cmd := &cobra.Command{
		Use:   "set-commission [commission-id]",
		Short: "Set the active commission for this workshop",
		Long: `Set which commission the workshop is actively working on.

Only one commission can be active per workshop at a time.
Focus commands are scoped to the active commission.

Examples:
  orc workshop set-commission COMM-001
  orc workshop set-commission --clear`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetCommission(args, clearFlag)
		},
	}

	cmd.Flags().BoolVar(&clearFlag, "clear", false, "Clear the active commission")

	return cmd
}

func runSetCommission(args []string, clearFlag bool) error {
	ctx := context.Background()

	// Get workshop ID from config (Goblin context)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadConfig(cwd)
	if err != nil {
		return fmt.Errorf("no ORC config found in current directory")
	}

	// Must be Goblin context (gatehouse)
	if !config.IsGatehouse(cfg.PlaceID) {
		return fmt.Errorf("set-commission requires Goblin context (gatehouse directory)")
	}

	// Look up workshop from gatehouse
	gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, cfg.PlaceID)
	if err != nil {
		return fmt.Errorf("failed to get gatehouse: %w", err)
	}
	workshopID := gatehouse.WorkshopID

	if clearFlag {
		if err := wire.WorkshopService().SetActiveCommission(ctx, workshopID, ""); err != nil {
			return fmt.Errorf("failed to clear commission: %w", err)
		}

		// Rename session to workshop name only
		if os.Getenv("TMUX") != "" {
			workshop, _ := wire.WorkshopService().GetWorkshop(ctx, workshopID)
			if workshop != nil {
				sessionName := wire.TMuxAdapter().FindSessionByWorkshopID(ctx, workshopID)
				if sessionName != "" {
					_ = wire.TMuxAdapter().RenameSession(ctx, sessionName, workshop.Name)
					_ = wire.TMuxAdapter().ConfigureStatusBar(ctx, workshop.Name, secondary.StatusBarConfig{
						StatusLeft: fmt.Sprintf(" %s ", workshop.Name),
					})
				}
			}
		}

		fmt.Printf("✓ Workshop %s commission cleared\n", workshopID)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("Usage: orc workshop set-commission <COMM-xxx> or orc workshop set-commission --clear")
	}

	commissionID := args[0]
	if !strings.HasPrefix(commissionID, "COMM-") {
		return fmt.Errorf("invalid commission ID: %s (expected COMM-xxx)", commissionID)
	}

	// Validate commission exists
	commission, err := wire.CommissionService().GetCommission(ctx, commissionID)
	if err != nil {
		return fmt.Errorf("commission %s not found", commissionID)
	}

	if err := wire.WorkshopService().SetActiveCommission(ctx, workshopID, commissionID); err != nil {
		return fmt.Errorf("failed to set commission: %w", err)
	}

	// Rename tmux session if inside one
	if os.Getenv("TMUX") != "" {
		if err := renameSessionForCommission(ctx, workshopID, commission); err != nil {
			fmt.Printf("  (tmux session rename skipped: %v)\n", err)
		}
	}

	fmt.Printf("✓ Workshop %s now active on %s: %s\n", workshopID, commissionID, commission.Title)
	return nil
}

// renameSessionForCommission renames the tmux session to reflect the active commission.
// Format: "Workshop Name - COMM-XXX - Commission Title"
// Status bar shows workshop name only.
func renameSessionForCommission(ctx context.Context, workshopID string, commission *primary.Commission) error {
	// Get workshop name
	workshop, err := wire.WorkshopService().GetWorkshop(ctx, workshopID)
	if err != nil {
		return err
	}

	// Find current session by workshop ID
	sessionName := wire.TMuxAdapter().FindSessionByWorkshopID(ctx, workshopID)
	if sessionName == "" {
		return fmt.Errorf("no session found for workshop")
	}

	// Build new name: "Workshop Name - COMM-XXX - Commission Title"
	newName := fmt.Sprintf("%s - %s - %s", workshop.Name, commission.ID, commission.Title)

	// Rename session
	if err := wire.TMuxAdapter().RenameSession(ctx, sessionName, newName); err != nil {
		return err
	}

	// Configure status bar to show workshop name only
	return wire.TMuxAdapter().ConfigureStatusBar(ctx, newName, secondary.StatusBarConfig{
		StatusLeft: fmt.Sprintf(" %s ", workshop.Name),
	})
}
