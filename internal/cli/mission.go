package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/app"
	orccontext "github.com/example/orc/internal/context"
	coremission "github.com/example/orc/internal/core/mission"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/tmux"
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

var missionCmd = &cobra.Command{
	Use:   "mission",
	Short: "Manage missions (strategic work streams)",
	Long:  "Create, list, and manage missions in the ORC ledger",
}

var missionCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new mission",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		return wire.MissionAdapter().Create(ctx, title, description)
	},
}

var missionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List missions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		status, _ := cmd.Flags().GetString("status")

		return wire.MissionAdapter().List(ctx, status)
	},
}

var missionShowCmd = &cobra.Command{
	Use:   "show [mission-id]",
	Short: "Show mission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]

		// Show mission details via adapter
		_, err := wire.MissionAdapter().Show(ctx, id)
		if err != nil {
			return err
		}

		// List shipments under this mission via service
		shipments, err := wire.ShipmentService().ListShipments(ctx, primary.ShipmentFilters{MissionID: id})
		if err == nil && len(shipments) > 0 {
			fmt.Println("Shipments:")
			for _, shipment := range shipments {
				fmt.Printf("  - %s [%s] %s\n", shipment.ID, shipment.Status, shipment.Title)
			}
			fmt.Println()
		}

		// List groves for this mission via service
		groves, err := wire.GroveService().ListGroves(context.Background(), primary.GroveFilters{MissionID: id})
		if err == nil && len(groves) > 0 {
			fmt.Println("Groves:")
			for _, g := range groves {
				fmt.Printf("  - %s: %s [%s]\n", g.ID, g.Name, g.Status)
			}
			fmt.Println()
		}

		return nil
	},
}

var missionStartCmd = &cobra.Command{
	Use:   "start [mission-id]",
	Short: "Start a mission workspace with TMux session",
	Long: `Create a mission workspace with .orc/config.json and TMux session.

This command:
1. Creates a workspace directory for the mission
2. Writes .orc/config.json for mission context detection
3. Queries database for active groves
4. Creates TMux session with ORC pane and grove panes
5. Materializes git worktrees for groves if needed

Examples:
  orc mission start MISSION-001
  orc mission start MISSION-001 --workspace ~/work/mission-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check agent identity - only ORC can start missions
		identity, _ := agent.GetCurrentAgentID()
		guardCtx := coremission.GuardContext{
			AgentType: coremission.AgentType(identity.Type),
			AgentID:   identity.FullID,
			MissionID: identity.MissionID,
		}
		if result := coremission.CanStartMission(guardCtx); !result.Allowed {
			return result.Error()
		}

		missionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")

		// Check if we're in ORC source directory
		if orccontext.IsOrcSourceDirectory() {
			return fmt.Errorf("cannot start mission in ORC source directory - please run from another location")
		}

		// Validate Claude workspace trust before creating mission workspace
		if err := validateClaudeWorkspaceTrust(); err != nil {
			return fmt.Errorf("Claude workspace trust validation failed:\n\n%w\n\nRun 'orc doctor' for detailed diagnostics", err)
		}

		// Get mission from service
		mission, err := wire.MissionService().GetMission(context.Background(), missionID)
		if err != nil {
			return fmt.Errorf("failed to get mission: %w", err)
		}

		// Default workspace path: ~/src/missions/MISSION-ID
		if workspacePath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			workspacePath = filepath.Join(home, "src", "missions", missionID)
		}

		// Create workspace directory
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Write .orc/config.json for mission context
		if err := orccontext.WriteMissionContext(workspacePath, missionID); err != nil {
			return fmt.Errorf("failed to write mission config: %w", err)
		}

		fmt.Printf("✓ Created mission workspace at: %s\n", workspacePath)
		fmt.Printf("  Mission: %s - %s\n", mission.ID, mission.Title)
		fmt.Println()

		// Get active groves for this mission via service
		groves, err := wire.GroveService().ListGroves(context.Background(), primary.GroveFilters{MissionID: missionID})
		if err != nil {
			return fmt.Errorf("failed to get groves: %w", err)
		}

		// Create TMux session
		sessionName := fmt.Sprintf("orc-%s", missionID)

		// Check if session already exists
		if tmux.SessionExists(sessionName) {
			return fmt.Errorf("TMux session '%s' already exists. Attach with: tmux attach -t %s", sessionName, sessionName)
		}

		fmt.Printf("Creating TMux session: %s\n", sessionName)

		// Create session with base numbering from 1
		session, err := tmux.NewSession(sessionName, workspacePath)
		if err != nil {
			return fmt.Errorf("failed to create TMux session: %w", err)
		}

		// Create ORC window (window 1) with claude
		if err := session.CreateOrcWindow(workspacePath); err != nil {
			return fmt.Errorf("failed to create ORC window: %w", err)
		}
		fmt.Printf("  ✓ Window 1: orc (claude | vim | shell)\n")

		// Create window for each grove with sophisticated layout
		for i, grove := range groves {
			windowIndex := i + 2 // Windows start at 1, ORC is 1, groves start at 2

			// Check if grove path exists
			pathExists := false
			if _, err := os.Stat(grove.Path); err == nil {
				pathExists = true
			}

			if !pathExists {
				fmt.Printf("  ℹ️  Grove %s worktree not found at %s\n", grove.ID, grove.Path)
				fmt.Printf("      Skipping window creation. Run 'orc grove create %s --repos <repo-names>' to materialize\n", grove.Name)
				continue
			}

			// Create grove window with vim, claude IMP, and shell
			if _, err := session.CreateGroveWindow(windowIndex, grove.Name, grove.Path); err != nil {
				fmt.Printf("  ⚠️  Warning: Could not create window for grove %s: %v\n", grove.ID, err)
				continue
			}

			fmt.Printf("  ✓ Window %d: %s (vim | claude IMP | shell) [%s]\n", windowIndex, grove.Name, grove.ID)
		}

		if len(groves) == 0 {
			fmt.Println("  ℹ️  No groves found for this mission")
			fmt.Printf("     Create groves with: orc grove create <name> --mission %s --repos <repo-names>\n", missionID)
		}

		// Select the ORC window (window 1) as default
		session.SelectWindow(1)

		fmt.Println()
		fmt.Printf("Mission workspace ready!\n")
		fmt.Println()
		fmt.Println(tmux.AttachInstructions(sessionName))

		return nil
	},
}

var missionCompleteCmd = &cobra.Command{
	Use:   "complete [mission-id]",
	Short: "Mark a mission as complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.MissionAdapter().Complete(ctx, args[0])
	},
}

var missionArchiveCmd = &cobra.Command{
	Use:   "archive [mission-id]",
	Short: "Archive a completed mission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.MissionAdapter().Archive(ctx, args[0])
	},
}

var missionUpdateCmd = &cobra.Command{
	Use:   "update [mission-id]",
	Short: "Update mission title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		return wire.MissionAdapter().Update(ctx, id, title, description)
	},
}

var missionDeleteCmd = &cobra.Command{
	Use:   "delete [mission-id]",
	Short: "Delete a mission from the database",
	Long: `Delete a mission and all associated data from the database.

WARNING: This is a destructive operation. Associated shipments, tasks, and groves
will lose their mission reference.

Examples:
  orc mission delete MISSION-TEST-001
  orc mission delete MISSION-001 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		return wire.MissionAdapter().Delete(ctx, id, force)
	},
}

var missionLaunchCmd = &cobra.Command{
	Use:   "launch [mission-id]",
	Short: "Launch mission infrastructure (plan/apply)",
	Long: `Launch or update mission infrastructure using plan/apply pattern.

This command:
1. Reads desired state from database (missions, shipments, groves)
2. Analyzes current filesystem state
3. Generates a plan of changes needed
4. Shows plan and asks for confirmation
5. Applies changes to converge filesystem to desired state

Idempotent: Can be run multiple times safely.

Examples:
  orc mission launch MISSION-002
  orc mission launch MISSION-001 --workspace ~/custom/path`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Check agent identity - only ORC can launch missions
		identity, _ := agent.GetCurrentAgentID()
		guardCtx := coremission.GuardContext{
			AgentType: coremission.AgentType(identity.Type),
			AgentID:   identity.FullID,
			MissionID: identity.MissionID,
		}
		if result := coremission.CanLaunchMission(guardCtx); !result.Allowed {
			return result.Error()
		}

		missionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")
		createTmux, _ := cmd.Flags().GetBool("tmux")

		// Default workspace path
		if workspacePath == "" {
			var err error
			workspacePath, err = app.DefaultWorkspacePath(missionID)
			if err != nil {
				return err
			}
		}

		// Phase 1: Load state using orchestration service
		fmt.Printf("🔍 Analyzing mission: %s\n\n", missionID)
		orchSvc := wire.MissionOrchestrationService()

		state, err := orchSvc.LoadMissionState(ctx, missionID)
		if err != nil {
			return fmt.Errorf("mission not found in database: %w\nCreate it first: orc mission create", err)
		}

		// Phase 2: Generate infrastructure plan
		infraPlan := orchSvc.AnalyzeInfrastructure(state, workspacePath)

		// Phase 3: Display plan
		displayMissionState(state, workspacePath)
		displayInfrastructurePlan(infraPlan)

		if createTmux {
			sessionName := fmt.Sprintf("orc-%s", missionID)
			tmuxPlan := orchSvc.PlanTmuxSession(state, workspacePath, sessionName, tmux.SessionExists(sessionName), &tmuxChecker{})
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
		displayInfrastructureResult(result, missionID)

		// Phase 6: Apply TMux if requested
		if createTmux {
			sessionName := fmt.Sprintf("orc-%s", missionID)
			grovesDir := filepath.Join(workspacePath, "groves")
			applyTmuxSession(sessionName, state.Groves, grovesDir, workspacePath)
		}

		// Next steps
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", workspacePath)
		if createTmux && !tmux.SessionExists(fmt.Sprintf("orc-%s", missionID)) {
			fmt.Printf("  tmux attach -t orc-%s\n", missionID)
		}
		fmt.Printf("  orc summary --mission %s\n", missionID)

		return nil
	},
}

// tmuxChecker implements app.TmuxWindowChecker for the tmux package
type tmuxChecker struct{}

func (t *tmuxChecker) WindowExists(session, window string) bool {
	return tmux.WindowExists(session, window)
}

func (t *tmuxChecker) GetPaneCount(session, window string) int {
	return tmux.GetPaneCount(session, window)
}

func (t *tmuxChecker) GetPaneCommand(session, window string, pane int) string {
	return tmux.GetPaneCommand(session, window, pane)
}

// displayMissionState shows the database state section of the plan
func displayMissionState(state *app.MissionState, workspacePath string) {
	fmt.Print("📋 Plan:\n\n")
	fmt.Println(color.New(color.Bold).Sprint("Database State:"))
	fmt.Printf("  Mission: %s - %s\n", colorDim(state.Mission.ID), state.Mission.Title)
	fmt.Printf("    Workspace: %s\n", workspacePath)
	fmt.Printf("    Created: %s\n", state.Mission.CreatedAt)
	fmt.Println()
	fmt.Printf("  Groves in DB: %d\n", len(state.Groves))
	for _, grove := range state.Groves {
		fmt.Printf("    %s - %s\n", colorDim(grove.ID), grove.Name)
		fmt.Printf("      Path: %s\n", grove.Path)
		if len(grove.Repos) > 0 {
			fmt.Printf("      Repos: %v\n", grove.Repos)
		}
		fmt.Printf("      Status: %s\n", grove.Status)
	}
	fmt.Println()
}

// displayInfrastructurePlan shows the infrastructure plan section
func displayInfrastructurePlan(plan *app.InfrastructurePlan) {
	fmt.Println(color.New(color.Bold).Sprint("Infrastructure:"))

	if plan.CreateWorkspace {
		fmt.Printf("  %s mission workspace: %s\n", colorCreate("CREATE"), plan.WorkspacePath)
	} else {
		fmt.Printf("  %s mission workspace: %s\n", colorExists("EXISTS"), plan.WorkspacePath)
	}

	if plan.CreateGrovesDir {
		fmt.Printf("  %s groves directory: %s\n", colorCreate("CREATE"), plan.GrovesDir)
	} else {
		fmt.Printf("  %s groves directory: %s\n", colorExists("EXISTS"), plan.GrovesDir)
	}

	for _, action := range plan.GroveActions {
		switch action.Action {
		case "exists":
			fmt.Printf("  %s grove %s: %s\n", colorExists("EXISTS"), action.GroveID, action.DesiredPath)
		case "move":
			fmt.Printf("  MOVE grove %s: %s → %s\n", action.GroveID, action.CurrentPath, action.DesiredPath)
		case "missing":
			fmt.Printf("  MISSING grove %s: %s (needs materialization)\n", action.GroveID, action.DesiredPath)
		}
		if action.UpdateDBPath && action.Action != "move" {
			fmt.Printf("  UPDATE DB path for %s: %s → %s\n", action.GroveID, action.CurrentPath, action.DesiredPath)
		}
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
func displayTmuxPlan(plan *app.TmuxSessionPlan) {
	fmt.Println(color.New(color.Bold).Sprint("TMux Session:"))
	if plan.SessionExists {
		fmt.Printf("  %s session: %s\n", colorExists("EXISTS"), plan.SessionName)
	} else {
		fmt.Printf("  %s session: %s\n", colorCreate("CREATE"), plan.SessionName)
	}

	for _, wp := range plan.WindowPlans {
		switch wp.Action {
		case "exists":
			fmt.Printf("  %s window %d (%s): 3 panes, IMP running - Grove %s\n", colorExists("EXISTS"), wp.Index, wp.Name, wp.GroveID)
		case "create":
			fmt.Printf("  %s window %d (%s): 3 panes in %s - Grove %s IMP\n", colorCreate("CREATE"), wp.Index, wp.Name, wp.GrovePath, wp.GroveID)
		case "update":
			fmt.Printf("  %s window %d (%s): needs update - Grove %s\n", colorUpdate("UPDATE"), wp.Index, wp.Name, wp.GroveID)
		case "skip":
			fmt.Printf("  SKIP window %d (%s): grove path missing\n", wp.Index, wp.Name)
		}
	}
	fmt.Println()
}

// displayInfrastructureResult shows the results of applying infrastructure changes
func displayInfrastructureResult(result *app.InfrastructureApplyResult, missionID string) {
	if result.WorkspaceCreated {
		fmt.Println("✓ Mission workspace created")
	} else {
		fmt.Println("✓ Mission workspace ready")
	}

	if result.GrovesDirCreated {
		fmt.Println("✓ Groves directory created")
	} else {
		fmt.Println("✓ Groves directory ready")
	}

	if result.GrovesProcessed > 0 {
		fmt.Printf("✓ Processed %d groves\n", result.GrovesProcessed)
	}

	if result.ConfigsWritten > 0 {
		fmt.Printf("✓ Wrote %d config files\n", result.ConfigsWritten)
	}

	if result.CleanupsDone > 0 {
		fmt.Printf("✓ Cleaned up %d old files\n", result.CleanupsDone)
	}

	for _, grove := range result.GrovesNeedingWork {
		fmt.Printf("  ℹ️  Grove %s worktree missing: %s\n", grove.GroveID, grove.DesiredPath)
		fmt.Printf("      Materialize with: orc grove create %s --repos <repo> --mission %s\n", grove.GroveName, missionID)
	}

	for _, err := range result.Errors {
		fmt.Printf("  ⚠️  %s\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Mission infrastructure ready")
}

// applyTmuxSession creates or updates the TMux session for a mission
func applyTmuxSession(sessionName string, groves []*primary.Grove, grovesDir, workspacePath string) {
	fmt.Println()
	fmt.Println("🖥️  Creating TMux session...")

	if tmux.SessionExists(sessionName) {
		fmt.Printf("  ℹ️  Session %s already exists - checking windows\n", sessionName)

		for i, grove := range groves {
			windowIndex := i + 1
			grovePath := filepath.Join(grovesDir, grove.Name)

			if _, err := os.Stat(grovePath); err == nil {
				if tmux.WindowExists(sessionName, grove.Name) {
					paneCount := tmux.GetPaneCount(sessionName, grove.Name)
					pane2Cmd := tmux.GetPaneCommand(sessionName, grove.Name, 2)

					if paneCount == 3 && pane2Cmd == "orc" {
						fmt.Printf("  ✓ Window %d (%s): IMP already running [%s]\n", windowIndex, grove.Name, grove.ID)
					} else if paneCount == 3 {
						target := fmt.Sprintf("%s:%s.2", sessionName, grove.Name)
						connectCmd := exec.Command("tmux", "respawn-pane", "-t", target, "-k", "orc", "connect")
						if err := connectCmd.Run(); err != nil {
							fmt.Printf("  ⚠️  Could not respawn pane in window %s: %v\n", grove.Name, err)
						} else {
							fmt.Printf("  ✓ Window %d (%s): IMP rebooted [%s]\n", windowIndex, grove.Name, grove.ID)
						}
					} else {
						fmt.Printf("  ⚠️  Window %d (%s): has %d panes (expected 3) - manual fix needed\n", windowIndex, grove.Name, paneCount)
					}
				} else {
					fmt.Printf("  ⚠️  Window %d (%s): missing - attach to session and create manually\n", windowIndex, grove.Name)
				}
			}
		}

		fmt.Println()
		fmt.Printf("✓ Session updated: %s\n", sessionName)
		fmt.Printf("  Attach with: tmux attach -t %s\n", sessionName)
	} else {
		startDir := workspacePath
		if len(groves) > 0 {
			firstGrovePath := filepath.Join(grovesDir, groves[0].Name)
			if _, err := os.Stat(firstGrovePath); err == nil {
				startDir = firstGrovePath
			}
		}

		session, err := tmux.NewSession(sessionName, startDir)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to create TMux session: %v\n", err)
			return
		}

		for i, grove := range groves {
			windowIndex := i + 1
			grovePath := filepath.Join(grovesDir, grove.Name)

			if _, err := os.Stat(grovePath); err == nil {
				if i == 0 {
					target := fmt.Sprintf("%s:1", sessionName)
					exec.Command("tmux", "rename-window", "-t", target, grove.Name).Run()
					target = fmt.Sprintf("%s:%s", sessionName, grove.Name)
					session.SplitVertical(target, grovePath)
					rightPane := fmt.Sprintf("%s.2", target)
					session.SplitHorizontal(rightPane, grovePath)
					topRightPane := fmt.Sprintf("%s.2", target)
					exec.Command("tmux", "respawn-pane", "-t", topRightPane, "-k", "orc", "connect").Run()
					fmt.Printf("✓ Window %d: %s (IMP auto-booting) [%s]\n", windowIndex, grove.Name, grove.ID)
				} else {
					if _, err := session.CreateGroveWindowShell(windowIndex, grove.Name, grovePath); err != nil {
						fmt.Printf("  ⚠️  Could not create window for grove %s: %v\n", grove.ID, err)
						continue
					}
					fmt.Printf("✓ Window %d: %s (IMP auto-booting) [%s]\n", windowIndex, grove.Name, grove.ID)
				}
			} else {
				fmt.Printf("  ℹ️  Grove %s worktree missing, skipping window\n", grove.ID)
			}
		}

		if len(groves) > 0 {
			session.SelectWindow(1)
		}

		fmt.Println()
		fmt.Printf("✓ TMux session created: %s\n", sessionName)
		fmt.Printf("  Attach with: tmux attach -t %s\n", sessionName)
		fmt.Println()
		fmt.Println("Window Layout: Left: (vim) | Right Top: (claude) | Right Bottom: (shell)")
	}
}

var missionPinCmd = &cobra.Command{
	Use:   "pin [mission-id]",
	Short: "Pin mission to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.MissionAdapter().Pin(ctx, args[0])
	},
}

var missionUnpinCmd = &cobra.Command{
	Use:   "unpin [mission-id]",
	Short: "Unpin mission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return wire.MissionAdapter().Unpin(ctx, args[0])
	},
}

// MissionCmd returns the mission command
func MissionCmd() *cobra.Command {
	// Add flags
	missionCreateCmd.Flags().StringP("description", "d", "", "Mission description")
	missionListCmd.Flags().StringP("status", "s", "", "Filter by status (active, paused, complete, archived)")
	missionStartCmd.Flags().StringP("workspace", "w", "", "Custom workspace path (default: ~/missions/MISSION-ID)")
	missionLaunchCmd.Flags().StringP("workspace", "w", "", "Custom workspace path (default: ~/src/missions/MISSION-ID)")
	missionLaunchCmd.Flags().Bool("tmux", false, "Create TMux session with window layout (no apps launched)")
	missionUpdateCmd.Flags().StringP("title", "t", "", "New mission title")
	missionUpdateCmd.Flags().StringP("description", "d", "", "New mission description")
	missionDeleteCmd.Flags().BoolP("force", "f", false, "Force delete even with associated data")

	// Add subcommands
	missionCmd.AddCommand(missionCreateCmd)
	missionCmd.AddCommand(missionListCmd)
	missionCmd.AddCommand(missionShowCmd)
	missionCmd.AddCommand(missionStartCmd)
	missionCmd.AddCommand(missionLaunchCmd)
	missionCmd.AddCommand(missionCompleteCmd)
	missionCmd.AddCommand(missionArchiveCmd)
	missionCmd.AddCommand(missionUpdateCmd)
	missionCmd.AddCommand(missionDeleteCmd)
	missionCmd.AddCommand(missionPinCmd)
	missionCmd.AddCommand(missionUnpinCmd)

	return missionCmd
}
