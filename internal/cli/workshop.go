package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
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
	cmd.AddCommand(workshopOpenCmd())
	cmd.AddCommand(workshopCloseCmd())

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

func workshopOpenCmd() *cobra.Command {
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "open [workshop-id]",
		Short: "Open a workshop TMux session",
		Long: `Launch a TMux session for the workshop with:
- Window 1: Gatehouse (Goblin orchestration)
- Window 2+: One per workbench

Shows a plan of what will be created and asks for confirmation.
Use --yes to skip confirmation.

Examples:
  orc workshop open WORK-001
  orc workshop open WORK-001 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := context.Background()

			// 1. Generate plan
			plan, err := wire.WorkshopService().PlanOpenWorkshop(ctx, primary.OpenWorkshopRequest{
				WorkshopID: workshopID,
			})
			if err != nil {
				return err
			}

			// 2. Display plan
			displayOpenPlan(plan)

			// 3. If nothing to do, show attach instructions and return
			if plan.NothingToDo {
				fmt.Println("Nothing to create.")
				fmt.Println()
				fmt.Println(wire.TMuxAdapter().AttachInstructions(plan.SessionName))
				return nil
			}

			// 4. Confirm (unless --yes)
			if !skipConfirm {
				if !confirmPrompt("Proceed?") {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// 5. Apply
			resp, err := wire.WorkshopService().ApplyOpenWorkshop(ctx, plan)
			if err != nil {
				return err
			}

			fmt.Printf("✓ Workshop %s opened: %s\n", workshopID, resp.Workshop.Name)
			fmt.Println(resp.AttachInstructions)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func displayOpenPlan(plan *primary.OpenWorkshopPlan) {
	fmt.Printf("🔍 Analyzing workshop: %s\n\n", plan.WorkshopID)
	fmt.Println("📋 Plan:")
	fmt.Println()

	displayDBState(plan)
	displayInfrastructure(plan)
	displayTMuxPlan(plan)
}

func displayDBState(plan *primary.OpenWorkshopPlan) {
	fmt.Println("Database State:")
	fmt.Printf("  Workshop: %s - %s\n", plan.WorkshopID, plan.WorkshopName)
	fmt.Printf("  Factory: %s - %s\n", plan.FactoryID, plan.FactoryName)

	if plan.DBState != nil {
		fmt.Printf("  Workbenches in DB: %d\n", plan.DBState.WorkbenchCount)
		for _, wb := range plan.DBState.Workbenches {
			fmt.Printf("    %s - %s\n", wb.ID, wb.Name)
			fmt.Printf("      Path: %s\n", wb.Path)
			fmt.Printf("      Status: %s\n", wb.Status)
		}
	} else {
		fmt.Println("  Workbenches in DB: 0")
	}
	fmt.Println()
}

func displayInfrastructure(plan *primary.OpenWorkshopPlan) {
	fmt.Println("Infrastructure:")
	dim := color.New(color.Faint).SprintFunc()

	if plan.GatehouseOp != nil {
		fmt.Printf("  %s gatehouse: %s\n", statusColor(plan.GatehouseOp.Status), plan.GatehouseOp.Path)
		fmt.Printf("  %s gatehouse config: %s/.orc/config.json\n",
			statusColor(plan.GatehouseOp.ConfigStatus), plan.GatehouseOp.Path)
		if plan.GatehouseOp.ConfigStatus == primary.OpCreate {
			fmt.Printf("          %s\n", dim(`{ "role": "GOBLIN", ... }`))
		}
	}

	for _, wb := range plan.WorkbenchOps {
		label := fmt.Sprintf("workbench %s", wb.ID)
		fmt.Printf("  %s %s: %s\n", statusColor(wb.Status), label, wb.Path)

		if wb.Status == primary.OpCreate && wb.RepoName != "" {
			fmt.Printf("          %s\n", dim(fmt.Sprintf("(worktree: %s@%s)", wb.RepoName, wb.Branch)))
		}

		fmt.Printf("  %s %s config: %s/.orc/config.json\n",
			statusColor(wb.ConfigStatus), label, wb.Path)
		if wb.ConfigStatus == primary.OpCreate {
			fmt.Printf("          %s\n", dim(`{ "role": "IMP", ... }`))
		}
	}
	fmt.Println()
}

func displayTMuxPlan(plan *primary.OpenWorkshopPlan) {
	if plan.TMuxOp == nil {
		fmt.Println("TMux Session:")
		fmt.Printf("  %s session: %s\n", statusColor(primary.OpExists), plan.SessionName)
		fmt.Println()
		return
	}

	fmt.Println("TMux Session:")
	fmt.Printf("  %s session: %s\n", statusColor(plan.TMuxOp.SessionStatus), plan.SessionName)
	for _, w := range plan.TMuxOp.Windows {
		fmt.Printf("  %s window %d (%s): %s\n",
			statusColor(w.Status), w.Index, w.Name, w.Path)
	}
	fmt.Println()
}

// statusColor returns a color-formatted status string.
func statusColor(status primary.OpStatus) string {
	switch status {
	case primary.OpExists:
		return color.New(color.FgBlue).Sprint("EXISTS ")
	case primary.OpCreate:
		return color.New(color.FgGreen).Sprint("CREATE ")
	case primary.OpUpdate:
		return color.New(color.FgYellow).Sprint("UPDATE ")
	case primary.OpMissing:
		return color.New(color.FgRed).Sprint("MISSING")
	default:
		return string(status)
	}
}

func confirmPrompt(msg string) bool {
	fmt.Printf("%s [y/N]: ", msg)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
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
