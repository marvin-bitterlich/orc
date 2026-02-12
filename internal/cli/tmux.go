package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// TmuxCmd returns the tmux command
func TmuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmux",
		Short: "TMux session management",
		Long:  `Manage TMux sessions for workshops.`,
	}

	cmd.AddCommand(
		tmuxConnectCmd(),
		tmuxApplyCmd(),
		tmuxEnrichCmd(),
	)

	return cmd
}

func tmuxConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [workshop-id]",
		Short: "Connect to a workshop's TMux session",
		Long: `Attach to an existing TMux session for a workshop.

This command does not create anything - it only connects to an existing session.
If no session exists, run 'orc tmux apply' first to create the session.

Examples:
  orc tmux connect WORK-001
  orc tmux connect WORK-003`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := NewContext()

			// 1. Verify workshop exists
			workshop, err := wire.WorkshopService().GetWorkshop(ctx, workshopID)
			if err != nil {
				return fmt.Errorf("workshop not found: %s", workshopID)
			}

			// 2. Find tmux session by name
			gotmuxAdapter, err := wire.NewGotmuxAdapter()
			if err != nil {
				return fmt.Errorf("failed to create tmux adapter: %w", err)
			}

			if !gotmuxAdapter.SessionExists(workshop.Name) {
				return fmt.Errorf("no tmux session found for %s\nRun: orc tmux apply %s", workshopID, workshopID)
			}

			// 3. Attach to session
			return attachToSession(workshop.Name)
		},
	}

	return cmd
}

// attachToSession replaces the current process with tmux attach.
func attachToSession(sessionName string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Use exec to replace current process
	args := []string{"tmux", "attach-session", "-t", sessionName}
	return syscall.Exec(tmuxPath, args, os.Environ())
}

func tmuxApplyCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "apply [workshop-id]",
		Short: "Reconcile tmux session state for a workshop",
		Long: `Compare desired tmux state (from DB) with actual tmux state and reconcile.

This is the single command for creating, updating, and maintaining workshop
tmux sessions. It replaces the old 'start' and 'refresh' commands.

Actions performed:
- Create session if it doesn't exist
- Add windows for missing workbenches
- Relocate guest panes to -imps windows
- Prune dead panes in -imps windows
- Kill empty -imps windows (all panes dead)
- Reconcile layout (main-vertical, 50% main-pane-width)
- Apply ORC enrichment (bindings, pane titles)

Without --yes, shows a plan and prompts for confirmation.
With --yes, applies immediately (useful for scripts and automation).

Examples:
  orc tmux apply WORK-001          # Show plan, prompt for confirmation
  orc tmux apply WORK-001 --yes    # Apply immediately`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := NewContext()

			// 1. Fetch workshop data
			workshop, err := wire.WorkshopService().GetWorkshop(ctx, workshopID)
			if err != nil {
				return fmt.Errorf("workshop not found: %s", workshopID)
			}

			// 2. Fetch workbenches for this workshop
			workbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
				WorkshopID: workshopID,
			})
			if err != nil {
				return fmt.Errorf("failed to list workbenches: %w", err)
			}

			// 3. Filter active workbenches and validate paths
			var desired []wire.DesiredWorkbench
			for _, wb := range workbenches {
				if wb.Status == "active" {
					if _, err := os.Stat(wb.Path); os.IsNotExist(err) {
						return fmt.Errorf("worktree path does not exist for %s: %s\nRun: orc infra apply %s", wb.ID, wb.Path, workshopID)
					}
					desired = append(desired, wire.DesiredWorkbench{
						Name:       wb.Name,
						Path:       wb.Path,
						ID:         wb.ID,
						WorkshopID: workshopID,
					})
				}
			}

			if len(desired) == 0 {
				return fmt.Errorf("workshop %s has no active workbenches", workshopID)
			}

			// 4. Create gotmux adapter and compute plan
			gotmuxAdapter, err := wire.NewGotmuxAdapter()
			if err != nil {
				return fmt.Errorf("failed to create gotmux adapter: %w", err)
			}

			plan, err := gotmuxAdapter.PlanApply(workshop.Name, desired)
			if err != nil {
				return fmt.Errorf("failed to compute plan: %w", err)
			}

			// 5. Print plan
			printApplyPlan(plan, workshopID)

			if len(plan.Actions) == 0 {
				fmt.Println("\nNothing to do.")
				return nil
			}

			// 6. Confirm or auto-apply
			if !yes {
				fmt.Print("\nApply? [y/n] ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Canceled.")
					return nil
				}
			}

			// 7. Execute plan
			if err := gotmuxAdapter.ExecutePlan(plan); err != nil {
				return fmt.Errorf("apply failed: %w", err)
			}

			fmt.Printf("\n✓ Applied successfully\n")
			fmt.Printf("  Attach with: orc tmux connect %s\n", workshopID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Apply immediately without confirmation")

	return cmd
}

// printApplyPlan displays the reconciliation plan.
func printApplyPlan(plan *wire.ApplyPlan, workshopID string) {
	fmt.Printf("orc tmux apply %s\n", workshopID)

	if plan.SessionExists {
		fmt.Printf("Session: %s (exists)\n", plan.SessionName)
	} else {
		fmt.Printf("Session: %s (will create)\n", plan.SessionName)
	}

	// Show window summaries
	for _, ws := range plan.WindowSummary {
		if ws.IsImps {
			if ws.DeadPanes == ws.PaneCount {
				fmt.Printf("Window: %s (%d dead panes) -> KILL\n", ws.Name, ws.DeadPanes)
			} else if ws.DeadPanes > 0 {
				fmt.Printf("Window: %s (%d panes, %d dead) -> PRUNE\n", ws.Name, ws.PaneCount, ws.DeadPanes)
			} else {
				fmt.Printf("Window: %s (%d panes)\n", ws.Name, ws.PaneCount)
			}
		} else {
			if ws.Healthy {
				fmt.Printf("Window: %s (%d panes, healthy)\n", ws.Name, ws.PaneCount)
			} else {
				fmt.Printf("Window: %s (%d panes)\n", ws.Name, ws.PaneCount)
			}
		}
	}

	// Show planned actions
	if len(plan.Actions) > 0 {
		fmt.Println("\nActions:")
		for i, action := range plan.Actions {
			fmt.Printf("  %d. [%s] %s\n", i+1, action.Type, action.Description)
		}
	}
}

func tmuxEnrichCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enrich [workshop-id]",
		Short: "Apply ORC enrichment to running tmux session",
		Long: `Apply ORC enrichment to a running tmux session:
- Apply global key bindings (status bar, popups, context menus)
- Set pane titles via select-pane -T (based on @pane_role or index heuristic)
- Set window options (@orc_enriched=1)

This command is idempotent and safe to run multiple times. Enrichment applies
ORC's visual and interactive layer on top of any tmux session (gotmux, manual).

If no workshop ID is provided, uses current workshop from context.

Examples:
  orc tmux enrich WORK-001
  orc tmux enrich           # Uses current workshop`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			// Get workshop ID from arg or context
			var workshopID string
			if len(args) > 0 {
				workshopID = args[0]
			} else {
				// Try to get from current workbench context
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}

				workbench, err := wire.WorkbenchService().GetWorkbenchByPath(ctx, cwd)
				if err != nil {
					return fmt.Errorf("no workshop ID provided and not in a workbench directory")
				}
				workshopID = workbench.WorkshopID
			}

			// Verify workshop exists
			workshop, err := wire.WorkshopService().GetWorkshop(ctx, workshopID)
			if err != nil {
				return fmt.Errorf("workshop not found: %s", workshopID)
			}

			// Apply global bindings (idempotent)
			wire.ApplyGlobalTMuxBindings()

			fmt.Printf("✓ Applied global bindings\n")

			// Apply session-wide enrichment (pane titles, window options)
			if err := wire.EnrichSession(workshop.Name); err != nil {
				return fmt.Errorf("failed to enrich session: %w", err)
			}

			fmt.Printf("✓ Applied session enrichment to: %s\n", workshop.Name)
			fmt.Printf("  - Set pane titles (based on @pane_role or index heuristic)\n")
			fmt.Printf("  - Set window options (@orc_enriched)\n")

			return nil
		},
	}

	return cmd
}
