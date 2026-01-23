package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
	"github.com/example/orc/internal/wire"
)

// Color helpers for plan output
var (
	colorExists = color.New(color.FgBlue).SprintFunc()
	colorCreate = color.New(color.FgGreen).SprintFunc()
	colorUpdate = color.New(color.FgYellow).SprintFunc()
	colorDelete = color.New(color.FgRed).SprintFunc()
	colorDim    = color.New(color.Faint).SprintFunc()
)

var commissionCmd = &cobra.Command{
	Use:   "commission",
	Short: "Manage commissions (strategic work streams)",
	Long:  "Create, list, and manage commissions in the ORC ledger",
}

var commissionCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new commission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		return wire.CommissionAdapter().Create(ctx, title, description)
	},
}

var commissionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List commissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		status, _ := cmd.Flags().GetString("status")

		return wire.CommissionAdapter().List(ctx, status)
	},
}

var commissionShowCmd = &cobra.Command{
	Use:   "show [commission-id]",
	Short: "Show commission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]

		// Show commission details via adapter
		_, err := wire.CommissionAdapter().Show(ctx, id)
		if err != nil {
			return err
		}

		// List shipments under this commission via service
		shipments, err := wire.ShipmentService().ListShipments(ctx, primary.ShipmentFilters{CommissionID: id})
		if err == nil && len(shipments) > 0 {
			fmt.Println("Shipments:")
			for _, shipment := range shipments {
				fmt.Printf("  - %s [%s] %s\n", shipment.ID, shipment.Status, shipment.Title)
			}
			fmt.Println()
		}

		return nil
	},
}

var commissionStartCmd = &cobra.Command{
	Use:   "start [commission-id]",
	Short: "Start a commission workspace with TMux session",
	Long: `Create a commission workspace with .orc/config.json and TMux session.

This command:
1. Creates a workspace directory for the commission
2. Writes .orc/config.json for commission context detection
3. Queries database for active workbenches
4. Creates TMux session with ORC pane and workbench panes
5. Materializes git worktrees for workbenches if needed

Examples:
  orc commission start COMM-001
  orc commission start COMM-001 --workspace ~/work/commission-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Check agent identity - only ORC can start commissions
		if err := wire.CommissionOrchestrationService().CheckLaunchPermission(ctx); err != nil {
			return err
		}

		commissionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")

		// Check if we're in ORC source directory
		if orccontext.IsOrcSourceDirectory() {
			return fmt.Errorf("cannot start commission in ORC source directory - please run from another location")
		}

		// Validate Claude workspace trust before creating commission workspace
		if err := validateClaudeWorkspaceTrust(); err != nil {
			return fmt.Errorf("Claude workspace trust validation failed:\n\n%w\n\nRun 'orc doctor' for detailed diagnostics", err)
		}

		// Get commission from service
		commission, err := wire.CommissionService().GetCommission(context.Background(), commissionID)
		if err != nil {
			return fmt.Errorf("failed to get commission: %w", err)
		}

		// Default workspace path: ~/src/commissions/COMM-ID
		if workspacePath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			workspacePath = filepath.Join(home, "src", "commissions", commissionID)
		}

		// Create workspace directory
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Write .orc/config.json for commission context
		if err := orccontext.WriteCommissionContext(workspacePath, commissionID); err != nil {
			return fmt.Errorf("failed to write commission config: %w", err)
		}

		fmt.Printf("✓ Created commission workspace at: %s\n", workspacePath)
		fmt.Printf("  Commission: %s - %s\n", commission.ID, commission.Title)
		fmt.Println()

		// Create TMux session
		sessionName := fmt.Sprintf("orc-%s", commissionID)
		tmuxAdapter := wire.TMuxAdapter()

		// Check if session already exists
		if tmuxAdapter.SessionExists(ctx, sessionName) {
			return fmt.Errorf("TMux session '%s' already exists. Attach with: tmux attach -t %s", sessionName, sessionName)
		}

		fmt.Printf("Creating TMux session: %s\n", sessionName)

		// Create session with base numbering from 1
		if err := tmuxAdapter.CreateSession(ctx, sessionName, workspacePath); err != nil {
			return fmt.Errorf("failed to create TMux session: %w", err)
		}

		// Create ORC window (window 1) with claude
		if err := tmuxAdapter.CreateOrcWindow(ctx, sessionName, workspacePath); err != nil {
			return fmt.Errorf("failed to create ORC window: %w", err)
		}
		fmt.Printf("  ✓ Window 1: orc (claude | vim | shell)\n")

		// Select the ORC window (window 1) as default
		tmuxAdapter.SelectWindow(ctx, sessionName, 1)

		fmt.Println()
		fmt.Printf("Commission workspace ready!\n")
		fmt.Println()
		fmt.Println(tmuxAdapter.AttachInstructions(sessionName))

		return nil
	},
}

var commissionCompleteCmd = &cobra.Command{
	Use:   "complete [commission-id]",
	Short: "Mark a commission as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.CommissionAdapter().Complete(ctx, args[0])
	},
}

var commissionArchiveCmd = &cobra.Command{
	Use:   "archive [commission-id]",
	Short: "Archive a completed commission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.CommissionAdapter().Archive(ctx, args[0])
	},
}

var commissionUpdateCmd = &cobra.Command{
	Use:   "update [commission-id]",
	Short: "Update commission title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		return wire.CommissionAdapter().Update(ctx, id, title, description)
	},
}

var commissionDeleteCmd = &cobra.Command{
	Use:   "delete [commission-id]",
	Short: "Delete a commission from the database",
	Long: `Delete a commission and all associated data from the database.

WARNING: This is a destructive operation. Associated shipments, tasks, and workbenches
will lose their commission reference.

Examples:
  orc commission delete COMM-TEST-001
  orc commission delete COMM-001 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		return wire.CommissionAdapter().Delete(ctx, id, force)
	},
}

var commissionLaunchCmd = &cobra.Command{
	Use:   "launch [commission-id]",
	Short: "Launch commission infrastructure (plan/apply)",
	Long: `Launch or update commission infrastructure using plan/apply pattern.

This command:
1. Reads desired state from database (commissions, shipments, workbenches)
2. Analyzes current filesystem state
3. Generates a plan of changes needed
4. Shows plan and asks for confirmation
5. Applies changes to converge filesystem to desired state

Idempotent: Can be run multiple times safely.

Examples:
  orc commission launch COMM-002
  orc commission launch COMM-001 --workspace ~/custom/path`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Check agent identity - only ORC can launch commissions
		if err := wire.CommissionOrchestrationService().CheckLaunchPermission(ctx); err != nil {
			return err
		}

		commissionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")
		createTmux, _ := cmd.Flags().GetBool("tmux")

		// Default workspace path
		if workspacePath == "" {
			var err error
			workspacePath, err = config.DefaultWorkspacePath(commissionID)
			if err != nil {
				return err
			}
		}

		// Phase 1: Load state using orchestration service
		fmt.Printf("🔍 Analyzing commission: %s\n\n", commissionID)
		orchSvc := wire.CommissionOrchestrationService()

		state, err := orchSvc.LoadCommissionState(ctx, commissionID)
		if err != nil {
			return fmt.Errorf("commission not found in database: %w\nCreate it first: orc commission create", err)
		}

		// Phase 2: Generate infrastructure plan
		infraPlan := orchSvc.AnalyzeInfrastructure(state, workspacePath)

		// Phase 3: Display plan
		displayCommissionState(state, workspacePath)
		displayInfrastructurePlan(infraPlan)

		if createTmux {
			sessionName := fmt.Sprintf("orc-%s", commissionID)
			tmuxAdapter := wire.TMuxAdapter()
			tmuxPlan := orchSvc.PlanTmuxSession(state, workspacePath, sessionName, tmuxAdapter.SessionExists(ctx, sessionName), &tmuxChecker{adapter: tmuxAdapter})
			displayTmuxPlan(tmuxPlan)
		}

		// Phase 4: Confirm
		fmt.Print("Apply changes? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}

		// Phase 5: Apply infrastructure
		fmt.Print("\n🚀 Applying changes...\n\n")
		result := orchSvc.ApplyInfrastructure(ctx, infraPlan)
		displayInfrastructureResult(result, commissionID)

		// Phase 6: Apply TMux if requested
		if createTmux {
			sessionName := fmt.Sprintf("orc-%s", commissionID)
			applyTmuxSession(sessionName, workspacePath)
		}

		// Next steps
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", workspacePath)
		tmuxAdapterForCheck := wire.TMuxAdapter()
		if createTmux && !tmuxAdapterForCheck.SessionExists(ctx, fmt.Sprintf("orc-%s", commissionID)) {
			fmt.Printf("  tmux attach -t orc-%s\n", commissionID)
		}
		fmt.Printf("  orc summary --commission %s\n", commissionID)

		return nil
	},
}

// tmuxChecker implements primary.TmuxWindowChecker using the TMuxAdapter
type tmuxChecker struct {
	adapter secondary.TMuxAdapter
}

func (t *tmuxChecker) WindowExists(session, window string) bool {
	return t.adapter.WindowExists(context.Background(), session, window)
}

func (t *tmuxChecker) GetPaneCount(session, window string) int {
	return t.adapter.GetPaneCount(context.Background(), session, window)
}

func (t *tmuxChecker) GetPaneCommand(session, window string, pane int) string {
	return t.adapter.GetPaneCommand(context.Background(), session, window, pane)
}

// displayCommissionState shows the database state section of the plan
func displayCommissionState(state *primary.CommissionState, workspacePath string) {
	fmt.Print("📋 Plan:\n\n")
	fmt.Println(color.New(color.Bold).Sprint("Database State:"))
	fmt.Printf("  Commission: %s - %s\n", colorDim(state.Commission.ID), state.Commission.Title)
	fmt.Printf("    Workspace: %s\n", workspacePath)
	fmt.Printf("    Created: %s\n", state.Commission.CreatedAt)
	fmt.Println()
}

// displayInfrastructurePlan shows the infrastructure plan section
func displayInfrastructurePlan(plan *primary.InfrastructurePlan) {
	fmt.Println(color.New(color.Bold).Sprint("Infrastructure:"))

	if plan.CreateWorkspace {
		fmt.Printf("  %s commission workspace: %s\n", colorCreate("CREATE"), plan.WorkspacePath)
	} else {
		fmt.Printf("  %s commission workspace: %s\n", colorExists("EXISTS"), plan.WorkspacePath)
	}

	if plan.CreateWorkbenchesDir {
		fmt.Printf("  %s workbenches directory: %s\n", colorCreate("CREATE"), plan.WorkbenchesDir)
	} else {
		fmt.Printf("  %s workbenches directory: %s\n", colorExists("EXISTS"), plan.WorkbenchesDir)
	}

	for _, configWrite := range plan.ConfigWrites {
		fmt.Printf("  %s %s config: %s\n", colorCreate("CREATE"), configWrite.Type, configWrite.Path)
	}

	for _, cleanup := range plan.Cleanups {
		fmt.Printf("  %s %s: %s\n", colorDelete("DELETE"), cleanup.Reason, cleanup.Path)
	}
	fmt.Println()
}

// displayTmuxPlan shows the TMux plan section
func displayTmuxPlan(plan *primary.TmuxSessionPlan) {
	fmt.Println(color.New(color.Bold).Sprint("TMux Session:"))
	if plan.SessionExists {
		fmt.Printf("  %s session: %s\n", colorExists("EXISTS"), plan.SessionName)
	} else {
		fmt.Printf("  %s session: %s\n", colorCreate("CREATE"), plan.SessionName)
	}

	for _, wp := range plan.WindowPlans {
		switch wp.Action {
		case "exists":
			fmt.Printf("  %s window %d (%s): 3 panes, IMP running - Workbench %s\n", colorExists("EXISTS"), wp.Index, wp.Name, wp.WorkbenchID)
		case "create":
			fmt.Printf("  %s window %d (%s): 3 panes in %s - Workbench %s IMP\n", colorCreate("CREATE"), wp.Index, wp.Name, wp.WorkbenchPath, wp.WorkbenchID)
		case "update":
			fmt.Printf("  %s window %d (%s): needs update - Workbench %s\n", colorUpdate("UPDATE"), wp.Index, wp.Name, wp.WorkbenchID)
		case "skip":
			fmt.Printf("  SKIP window %d (%s): workbench path missing\n", wp.Index, wp.Name)
		}
	}
	fmt.Println()
}

// displayInfrastructureResult shows the results of applying infrastructure changes
func displayInfrastructureResult(result *primary.InfrastructureApplyResult, commissionID string) {
	if result.WorkspaceCreated {
		fmt.Println("✓ Commission workspace created")
	} else {
		fmt.Println("✓ Commission workspace ready")
	}

	if result.WorkbenchesDirCreated {
		fmt.Println("✓ Workbenches directory created")
	} else {
		fmt.Println("✓ Workbenches directory ready")
	}

	if result.WorkbenchesProcessed > 0 {
		fmt.Printf("✓ Processed %d workbenches\n", result.WorkbenchesProcessed)
	}

	if result.ConfigsWritten > 0 {
		fmt.Printf("✓ Wrote %d config files\n", result.ConfigsWritten)
	}

	if result.CleanupsDone > 0 {
		fmt.Printf("✓ Cleaned up %d old files\n", result.CleanupsDone)
	}

	for _, err := range result.Errors {
		fmt.Printf("  ⚠️  %s\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Commission infrastructure ready")
}

// applyTmuxSession creates or updates the TMux session for a commission
func applyTmuxSession(sessionName, workspacePath string) {
	ctx := context.Background()
	tmuxAdapter := wire.TMuxAdapter()

	fmt.Println()
	fmt.Println("🖥️  Creating TMux session...")

	if tmuxAdapter.SessionExists(ctx, sessionName) {
		fmt.Printf("  ℹ️  Session %s already exists\n", sessionName)
		fmt.Printf("  Attach with: tmux attach -t %s\n", sessionName)
		return
	}

	if err := tmuxAdapter.CreateSession(ctx, sessionName, workspacePath); err != nil {
		fmt.Printf("  ⚠️  Failed to create TMux session: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("✓ TMux session created: %s\n", sessionName)
	fmt.Printf("  Attach with: tmux attach -t %s\n", sessionName)
}

var commissionPinCmd = &cobra.Command{
	Use:   "pin [commission-id]",
	Short: "Pin commission to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.CommissionAdapter().Pin(ctx, args[0])
	},
}

var commissionUnpinCmd = &cobra.Command{
	Use:   "unpin [commission-id]",
	Short: "Unpin commission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.CommissionAdapter().Unpin(ctx, args[0])
	},
}

// CommissionCmd returns the commission command
func CommissionCmd() *cobra.Command {
	// Add flags
	commissionCreateCmd.Flags().StringP("description", "d", "", "Commission description")
	commissionListCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, complete, archived)")
	commissionStartCmd.Flags().StringP("workspace", "w", "", "Custom workspace path (default: ~/commissions/COMM-ID)")
	commissionLaunchCmd.Flags().StringP("workspace", "w", "", "Custom workspace path (default: ~/src/commissions/COMM-ID)")
	commissionLaunchCmd.Flags().Bool("tmux", false, "Create TMux session with window layout (no apps launched)")
	commissionUpdateCmd.Flags().StringP("title", "t", "", "New commission title")
	commissionUpdateCmd.Flags().StringP("description", "d", "", "New commission description")
	commissionDeleteCmd.Flags().BoolP("force", "f", false, "Force delete even with associated data")

	// Add subcommands
	commissionCmd.AddCommand(commissionCreateCmd)
	commissionCmd.AddCommand(commissionListCmd)
	commissionCmd.AddCommand(commissionShowCmd)
	commissionCmd.AddCommand(commissionStartCmd)
	commissionCmd.AddCommand(commissionLaunchCmd)
	commissionCmd.AddCommand(commissionCompleteCmd)
	commissionCmd.AddCommand(commissionArchiveCmd)
	commissionCmd.AddCommand(commissionUpdateCmd)
	commissionCmd.AddCommand(commissionDeleteCmd)
	commissionCmd.AddCommand(commissionPinCmd)
	commissionCmd.AddCommand(commissionUnpinCmd)

	return commissionCmd
}
