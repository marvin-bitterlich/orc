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

// InfraCmd returns the infra command
func InfraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "infra",
		Short: "Infrastructure planning and management",
		Long:  `Plan and apply infrastructure changes for workshops and workbenches.`,
	}

	cmd.AddCommand(infraPlanCmd())
	cmd.AddCommand(infraApplyCmd())

	return cmd
}

func infraPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [workshop-id]",
		Short: "Show infrastructure plan for a workshop",
		Long: `Display the current infrastructure state for a workshop.

Shows what exists and what would need to be created:
  EXISTS  - Resource exists in both DB and filesystem
  CREATE  - Resource needs to be created
  MISSING - Resource in DB but missing from filesystem

Examples:
  orc infra plan WORK-001
  orc infra plan WORK-003`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := context.Background()

			plan, err := wire.InfraService().PlanInfra(ctx, primary.InfraPlanRequest{
				WorkshopID: workshopID,
			})
			if err != nil {
				return err
			}

			displayInfraPlan(plan)
			return nil
		},
	}

	return cmd
}

func displayInfraPlan(plan *primary.InfraPlan) {
	fmt.Printf("🔍 Infrastructure Plan: %s\n\n", plan.WorkshopID)

	// Workshop info
	fmt.Println("Workshop:")
	fmt.Printf("  ID: %s\n", plan.WorkshopID)
	fmt.Printf("  Name: %s\n", plan.WorkshopName)
	fmt.Printf("  Factory: %s (%s)\n", plan.FactoryID, plan.FactoryName)
	fmt.Println()

	// Gatehouse
	fmt.Println("Gatehouse:")
	if plan.Gatehouse != nil {
		fmt.Printf("  %s directory: %s\n", infraStatusColor(plan.Gatehouse.Status), plan.Gatehouse.Path)
		fmt.Printf("  %s config: %s/.orc/config.json\n", infraStatusColor(plan.Gatehouse.ConfigStatus), plan.Gatehouse.Path)
		if plan.Gatehouse.ID != "" {
			fmt.Printf("  DB record: %s\n", plan.Gatehouse.ID)
		} else {
			fmt.Printf("  DB record: %s\n", color.New(color.FgYellow).Sprint("(not created)"))
		}
	} else {
		fmt.Printf("  %s (no gatehouse)\n", color.New(color.FgRed).Sprint("MISSING"))
	}
	fmt.Println()

	// Workbenches
	fmt.Println("Workbenches:")
	if len(plan.Workbenches) == 0 {
		fmt.Println("  (none)")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  STATUS\tID\tNAME\tPATH")
		fmt.Fprintln(w, "  ------\t--\t----\t----")
		for _, wb := range plan.Workbenches {
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
				infraStatusColor(wb.Status),
				wb.ID,
				wb.Name,
				wb.Path,
			)
		}
		w.Flush()

		// Show config status separately
		fmt.Println()
		fmt.Println("  Config Status:")
		for _, wb := range plan.Workbenches {
			fmt.Printf("    %s %s: %s/.orc/config.json\n",
				infraStatusColor(wb.ConfigStatus),
				wb.ID,
				wb.Path,
			)
		}
	}
	fmt.Println()
}

// infraStatusColor returns a color-formatted status string for infra display.
func infraStatusColor(status primary.OpStatus) string {
	switch status {
	case primary.OpExists:
		return color.New(color.FgBlue).Sprint("EXISTS ")
	case primary.OpCreate:
		return color.New(color.FgGreen).Sprint("CREATE ")
	case primary.OpMissing:
		return color.New(color.FgRed).Sprint("MISSING")
	default:
		return string(status)
	}
}

func infraApplyCmd() *cobra.Command {
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "apply [workshop-id]",
		Short: "Apply infrastructure for a workshop",
		Long: `Create infrastructure for a workshop.

Shows the plan first and asks for confirmation (unless --yes).
Creates:
  - Gatehouse directory and config
  - Workbench worktrees (via git worktree add)
  - ORC config files in each workbench

Examples:
  orc infra apply WORK-001
  orc infra apply WORK-001 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := context.Background()

			// 1. Generate plan
			plan, err := wire.InfraService().PlanInfra(ctx, primary.InfraPlanRequest{
				WorkshopID: workshopID,
			})
			if err != nil {
				return err
			}

			// 2. Display plan
			displayInfraPlan(plan)

			// 3. Check if anything to do
			nothingToDo := checkNothingToDo(plan)
			if nothingToDo {
				fmt.Println("Nothing to create. All infrastructure exists.")
				return nil
			}

			// 4. Confirm (unless --yes)
			if !skipConfirm {
				if !infraConfirmPrompt("Proceed with creation?") {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// 5. Apply
			resp, err := wire.InfraService().ApplyInfra(ctx, plan)
			if err != nil {
				return err
			}

			// 6. Display result
			fmt.Println("✓ Infrastructure applied:")
			if resp.GatehouseCreated {
				fmt.Println("  - Gatehouse directory created")
			}
			if resp.WorkbenchesCreated > 0 {
				fmt.Printf("  - %d workbench worktree(s) created\n", resp.WorkbenchesCreated)
			}
			if resp.ConfigsCreated > 0 {
				fmt.Printf("  - %d config file(s) created\n", resp.ConfigsCreated)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func checkNothingToDo(plan *primary.InfraPlan) bool {
	if plan.Gatehouse != nil && (plan.Gatehouse.Status == primary.OpCreate || plan.Gatehouse.ConfigStatus == primary.OpCreate) {
		return false
	}
	for _, wb := range plan.Workbenches {
		if wb.Status == primary.OpCreate || wb.Status == primary.OpMissing || wb.ConfigStatus == primary.OpCreate {
			return false
		}
	}
	return true
}

func infraConfirmPrompt(msg string) bool {
	fmt.Printf("%s [y/N]: ", msg)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
