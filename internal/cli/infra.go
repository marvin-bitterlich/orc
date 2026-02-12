package cli

import (
	"bufio"
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
	cmd.AddCommand(infraArchiveWorkbenchCmd())
	cmd.AddCommand(infraCleanupCmd())

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
			ctx := NewContext()

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

	// Workshop Directory
	fmt.Println("Workshop Directory:")
	if plan.WorkshopDir != nil {
		fmt.Printf("  %s directory: %s\n", infraStatusColor(plan.WorkshopDir.Status), plan.WorkshopDir.Path)
		fmt.Printf("  %s config: %s/.orc/config.json\n", infraStatusColor(plan.WorkshopDir.ConfigStatus), plan.WorkshopDir.Path)
		if plan.WorkshopDir.ID != "" {
			fmt.Printf("  DB record: %s\n", plan.WorkshopDir.ID)
		} else {
			fmt.Printf("  DB record: %s\n", color.New(color.FgYellow).Sprint("(not created)"))
		}
	} else {
		fmt.Printf("  %s (no workshop directory)\n", color.New(color.FgRed).Sprint("MISSING"))
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

	// Show orphaned resources (exist on disk but not in DB)
	if len(plan.OrphanWorkbenches) > 0 || len(plan.OrphanWorkshopDirs) > 0 {
		fmt.Println("Orphaned Resources (exist on disk, not in DB):")

		if len(plan.OrphanWorkshopDirs) > 0 {
			fmt.Println("  Workshop Dirs:")
			for _, od := range plan.OrphanWorkshopDirs {
				fmt.Printf("    %s %s: %s\n",
					infraStatusColor(od.Status),
					od.ID,
					od.Path,
				)
			}
		}

		if len(plan.OrphanWorkbenches) > 0 {
			fmt.Println("  Workbenches:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "    STATUS\tID\tNAME\tPATH")
			fmt.Fprintln(w, "    ------\t--\t----\t----")
			for _, wb := range plan.OrphanWorkbenches {
				fmt.Fprintf(w, "    %s\t%s\t%s\t%s\n",
					infraStatusColor(wb.Status),
					wb.ID,
					wb.Name,
					wb.Path,
				)
			}
			w.Flush()
		}
		fmt.Println()
	}

	// TMux Session
	if plan.TMuxSession != nil {
		fmt.Println("TMux Session:")
		fmt.Printf("  %s session: %s\n", infraStatusColor(plan.TMuxSession.Status), plan.TMuxSession.SessionName)

		if len(plan.TMuxSession.Windows) > 0 {
			fmt.Println("  Windows:")
			for _, w := range plan.TMuxSession.Windows {
				fmt.Printf("    %s %s\n", infraStatusColor(w.Status), w.Name)
				// Show pane tree if window exists and has panes
				if w.Status == primary.OpExists && len(w.Panes) > 0 {
					displayPaneTree(w.Panes)
				}
			}
		}

		if len(plan.TMuxSession.OrphanWindows) > 0 {
			fmt.Println("  Orphan Windows:")
			for _, w := range plan.TMuxSession.OrphanWindows {
				fmt.Printf("    %s %s\n", infraStatusColor(w.Status), w.Name)
			}
		}
		fmt.Println()
	}
}

// displayPaneTree shows the pane verification tree for a window.
func displayPaneTree(panes []primary.InfraTMuxPaneOp) {
	paneNames := []string{"vim", "IMP", "shell"}
	for i, p := range panes {
		paneName := "pane"
		if i < len(paneNames) {
			paneName = paneNames[i]
		}

		// Determine overall status
		allOK := p.PathOK && p.CommandOK

		// Build status indicator
		var statusIcon string
		if allOK {
			statusIcon = color.New(color.FgGreen).Sprint("✓")
		} else {
			statusIcon = color.New(color.FgYellow).Sprint("!")
		}

		fmt.Printf("      %s pane %d (%s)\n", statusIcon, p.Index, paneName)

		// Show path verification
		if p.PathOK {
			fmt.Printf("        path: %s\n", color.New(color.FgGreen).Sprint("OK"))
		} else {
			fmt.Printf("        path: %s\n", color.New(color.FgYellow).Sprint("MISMATCH"))
			fmt.Printf("          expected: %s\n", p.ExpectedPath)
			fmt.Printf("          actual:   %s\n", p.ActualPath)
		}

		// Show command verification (if expected)
		if p.ExpectedCommand != "" {
			if p.CommandOK {
				fmt.Printf("        cmd:  %s (%s)\n", color.New(color.FgGreen).Sprint("OK"), p.ExpectedCommand)
			} else {
				fmt.Printf("        cmd:  %s\n", color.New(color.FgYellow).Sprint("MISMATCH"))
				fmt.Printf("          expected: %s\n", p.ExpectedCommand)
				fmt.Printf("          actual:   %s\n", p.ActualCommand)
			}
		}
	}
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
	case primary.OpDelete:
		return color.New(color.FgRed).Sprint("DELETE ")
	default:
		return string(status)
	}
}

func infraApplyCmd() *cobra.Command {
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "apply [workshop-id]",
		Short: "Apply infrastructure for a workshop",
		Long: `Create infrastructure for a workshop and clean up orphan tmux windows.

Shows the plan first and asks for confirmation (unless --yes).
Creates:
  - Workshop coordination directory
  - Workbench worktrees (via git worktree add)
  - ORC config files in each workbench
  - TMux session and windows

Cleans up:
  - Orphan tmux windows (for archived workbenches)

Note: This command does NOT delete directories. Use 'orc infra cleanup' for
orphan directory removal.

Examples:
  orc infra apply WORK-001
  orc infra apply WORK-001 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := NewContext()

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
			if checkNothingToDo(plan) {
				fmt.Println("Nothing to do. All infrastructure exists.")
				return nil
			}

			// 4. Confirm tmux window cleanup (CREATE operations proceed without confirmation)
			if plan.TMuxSession != nil && len(plan.TMuxSession.OrphanWindows) > 0 && !skipConfirm {
				windowCount := len(plan.TMuxSession.OrphanWindows)
				if !infraConfirmPrompt(fmt.Sprintf("Kill %d orphan tmux window(s)?", windowCount)) {
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
			if resp.WorkshopDirCreated {
				fmt.Println("  - Workshop directory created")
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
	if plan.WorkshopDir != nil && (plan.WorkshopDir.Status == primary.OpCreate || plan.WorkshopDir.ConfigStatus == primary.OpCreate) {
		return false
	}
	for _, wb := range plan.Workbenches {
		if wb.Status == primary.OpCreate || wb.Status == primary.OpMissing || wb.ConfigStatus == primary.OpCreate {
			return false
		}
	}
	// Check TMux state
	if plan.TMuxSession != nil {
		if plan.TMuxSession.Status == primary.OpCreate {
			return false
		}
		for _, w := range plan.TMuxSession.Windows {
			if w.Status == primary.OpCreate {
				return false
			}
		}
		if len(plan.TMuxSession.OrphanWindows) > 0 {
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

func infraArchiveWorkbenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive-workbench",
		Short: "Archive current workbench and clean up tmux window",
		Long: `Archive the workbench at the current directory and clean up its tmux window.

This is a convenience command for the statusline menu that:
1. Detects the workbench from the current directory
2. Archives the workbench (soft-delete in DB)
3. Applies infrastructure to remove the tmux window

Note: The workbench directory is NOT deleted. Use 'rm -rf' manually or
'orc infra cleanup' if you want to remove it later.

Examples:
  cd ~/wb/my-workbench && orc infra archive-workbench`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			// 1. Get current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// 2. Detect workbench from path
			workbench, err := wire.WorkbenchService().GetWorkbenchByPath(ctx, cwd)
			if err != nil {
				return fmt.Errorf("not in a workbench directory: %w", err)
			}

			fmt.Printf("Workbench: %s (%s)\n", workbench.ID, workbench.Name)
			fmt.Printf("Workshop: %s\n\n", workbench.WorkshopID)

			// 3. Archive the workbench
			fmt.Println("Archiving workbench...")
			if err := wire.WorkbenchService().ArchiveWorkbench(ctx, workbench.ID); err != nil {
				return fmt.Errorf("failed to archive workbench: %w", err)
			}
			fmt.Printf("✓ Workbench %s archived\n\n", workbench.ID)

			// 4. Apply infrastructure
			fmt.Println("Applying infrastructure changes...")
			plan, err := wire.InfraService().PlanInfra(ctx, primary.InfraPlanRequest{
				WorkshopID: workbench.WorkshopID,
			})
			if err != nil {
				return fmt.Errorf("failed to plan infrastructure: %w", err)
			}

			_, err = wire.InfraService().ApplyInfra(ctx, plan)
			if err != nil {
				return fmt.Errorf("failed to apply infrastructure: %w", err)
			}

			fmt.Println("✓ Infrastructure applied (tmux window removed)")

			fmt.Println("\nPress Enter to close...")
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')

			return nil
		},
	}

	return cmd
}

func infraCleanupCmd() *cobra.Command {
	var forceDelete bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up orphaned infrastructure",
		Long: `Scan for and remove orphaned infrastructure.

Orphaned infrastructure includes:
  - Workbench directories with no matching DB record
  - Workshop directories with no matching DB record

This is useful for manual recovery when the system is in an inconsistent state.

Examples:
  orc infra cleanup          # Show orphans and confirm before deleting
  orc infra cleanup --force  # Delete orphans with uncommitted changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			resp, err := wire.InfraService().CleanupOrphans(ctx, primary.CleanupOrphansRequest{
				Force: forceDelete,
			})
			if err != nil {
				return err
			}

			if resp.WorkbenchesDeleted == 0 && resp.WorkshopDirsDeleted == 0 {
				fmt.Println("No orphans found")
				return nil
			}

			fmt.Println("✓ Cleanup complete")
			if resp.WorkbenchesDeleted > 0 {
				fmt.Printf("  - %d orphan workbench(es) deleted\n", resp.WorkbenchesDeleted)
			}
			if resp.WorkshopDirsDeleted > 0 {
				fmt.Printf("  - %d orphan workshop dir(s) deleted\n", resp.WorkshopDirsDeleted)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force deletion of dirty worktrees with uncommitted changes")

	return cmd
}
