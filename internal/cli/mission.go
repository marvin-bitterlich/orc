package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/example/orc/internal/tmux"
	"github.com/spf13/cobra"
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
		// Check agent identity - only ORC can create missions
		identity, err := agent.GetCurrentAgentID()
		if err == nil && identity.Type == agent.AgentTypeIMP {
			return fmt.Errorf("IMPs cannot create missions - only ORC can create missions")
		}

		title := args[0]
		description, _ := cmd.Flags().GetString("description")

		mission, err := models.CreateMission(title, description)
		if err != nil {
			return fmt.Errorf("failed to create mission: %w", err)
		}

		fmt.Printf("✓ Created mission %s: %s\n", mission.ID, mission.Title)
		return nil
	},
}

var missionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List missions",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")

		missions, err := models.ListMissions(status)
		if err != nil {
			return fmt.Errorf("failed to list missions: %w", err)
		}

		if len(missions) == 0 {
			fmt.Println("No missions found")
			return nil
		}

		fmt.Printf("\n%-15s %-10s %s\n", "ID", "STATUS", "TITLE")
		fmt.Println("────────────────────────────────────────────────────────────────")
		for _, m := range missions {
			fmt.Printf("%-15s %-10s %s\n", m.ID, m.Status, m.Title)
		}
		fmt.Println()

		return nil
	},
}

var missionShowCmd = &cobra.Command{
	Use:   "show [mission-id]",
	Short: "Show mission details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		mission, err := models.GetMission(id)
		if err != nil {
			return fmt.Errorf("failed to get mission: %w", err)
		}

		fmt.Printf("\nMission: %s\n", mission.ID)
		fmt.Printf("Title:   %s\n", mission.Title)
		fmt.Printf("Status:  %s\n", mission.Status)
		if mission.Description.Valid {
			fmt.Printf("Description: %s\n", mission.Description.String)
		}
		fmt.Printf("Created: %s\n", mission.CreatedAt.Format("2006-01-02 15:04"))
		if mission.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", mission.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		fmt.Println()

		// List epics under this mission
		epics, err := models.ListEpics(id, "")
		if err == nil && len(epics) > 0 {
			fmt.Println("Epics:")
			for _, epic := range epics {
				fmt.Printf("  - %s [%s] %s\n", epic.ID, epic.Status, epic.Title)
			}
			fmt.Println()
		}

		// List groves for this mission
		groves, err := models.GetGrovesByMission(id)
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
		identity, err := agent.GetCurrentAgentID()
		if err == nil && identity.Type == agent.AgentTypeIMP {
			return fmt.Errorf("IMPs cannot start missions - only ORC can start missions")
		}

		missionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")

		// Check if we're in ORC source directory
		if context.IsOrcSourceDirectory() {
			return fmt.Errorf("cannot start mission in ORC source directory - please run from another location")
		}

		// Validate Claude workspace trust before creating mission workspace
		if err := validateClaudeWorkspaceTrust(); err != nil {
			return fmt.Errorf("Claude workspace trust validation failed:\n\n%w\n\nRun 'orc doctor' for detailed diagnostics", err)
		}

		// Get mission from DB
		mission, err := models.GetMission(missionID)
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
		if err := context.WriteMissionContext(workspacePath, missionID); err != nil {
			return fmt.Errorf("failed to write mission config: %w", err)
		}

		fmt.Printf("✓ Created mission workspace at: %s\n", workspacePath)
		fmt.Printf("  Mission: %s - %s\n", mission.ID, mission.Title)
		fmt.Println()

		// Get active groves for this mission
		groves, err := models.GetGrovesByMission(missionID)
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
		id := args[0]

		err := models.UpdateMissionStatus(id, "complete")
		if err != nil {
			return fmt.Errorf("failed to complete mission: %w", err)
		}

		fmt.Printf("✓ Mission %s marked as complete\n", id)
		return nil
	},
}

var missionArchiveCmd = &cobra.Command{
	Use:   "archive [mission-id]",
	Short: "Archive a completed mission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.UpdateMissionStatus(id, "archived")
		if err != nil {
			return fmt.Errorf("failed to archive mission: %w", err)
		}

		fmt.Printf("✓ Mission %s archived\n", id)
		return nil
	},
}

var missionUpdateCmd = &cobra.Command{
	Use:   "update [mission-id]",
	Short: "Update mission title and/or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")

		if title == "" && description == "" {
			return fmt.Errorf("must specify at least --title or --description")
		}

		err := models.UpdateMission(id, title, description)
		if err != nil {
			return fmt.Errorf("failed to update mission: %w", err)
		}

		fmt.Printf("✓ Mission %s updated\n", id)
		return nil
	},
}

var missionDeleteCmd = &cobra.Command{
	Use:   "delete [mission-id]",
	Short: "Delete a mission from the database",
	Long: `Delete a mission and all associated data from the database.

WARNING: This is a destructive operation. Associated epics, tasks, and groves
will lose their mission reference.

Examples:
  orc mission delete MISSION-TEST-001
  orc mission delete MISSION-001 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

		// Get mission details before deleting
		mission, err := models.GetMission(id)
		if err != nil {
			return fmt.Errorf("failed to get mission: %w", err)
		}

		// Check for associated epics and groves
		epics, _ := models.ListEpics(id, "")
		groves, _ := models.GetGrovesByMission(id)

		if !force && (len(epics) > 0 || len(groves) > 0) {
			fmt.Printf("⚠️  Mission %s has associated data:\n", id)
			if len(epics) > 0 {
				fmt.Printf("  - %d epics\n", len(epics))
			}
			if len(groves) > 0 {
				fmt.Printf("  - %d groves\n", len(groves))
			}
			fmt.Println()
			fmt.Println("Use --force to delete anyway")
			return fmt.Errorf("mission has associated data")
		}

		// Delete the mission
		err = models.DeleteMission(id)
		if err != nil {
			return fmt.Errorf("failed to delete mission: %w", err)
		}

		fmt.Printf("✓ Deleted mission %s: %s\n", mission.ID, mission.Title)
		return nil
	},
}

var missionLaunchCmd = &cobra.Command{
	Use:   "launch [mission-id]",
	Short: "Launch mission infrastructure (plan/apply)",
	Long: `Launch or update mission infrastructure using plan/apply pattern.

This command:
1. Reads desired state from database (missions, epics, groves)
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
		// Check agent identity - only ORC can launch missions
		identity, err := agent.GetCurrentAgentID()
		if err == nil && identity.Type == agent.AgentTypeIMP {
			return fmt.Errorf("IMPs cannot launch missions - only ORC can launch missions")
		}

		missionID := args[0]
		workspacePath, _ := cmd.Flags().GetString("workspace")
		createTmux, _ := cmd.Flags().GetBool("tmux")

		// Default workspace path
		if workspacePath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			workspacePath = filepath.Join(homeDir, "src", "missions", missionID)
		}

		// Phase 1: Load desired state from database
		fmt.Printf("🔍 Analyzing mission: %s\n\n", missionID)

		mission, err := models.GetMission(missionID)
		if err != nil {
			return fmt.Errorf("mission not found in database: %w\nCreate it first: orc mission create", err)
		}

		groves, err := models.GetGrovesByMission(missionID)
		if err != nil {
			return fmt.Errorf("failed to load groves: %w", err)
		}

		fmt.Printf("Mission: %s - %s\n", mission.ID, mission.Title)
		fmt.Printf("Groves in DB: %d\n\n", len(groves))

		// Phase 2: Analyze current state and generate plan
		var plan []string

		// Check mission workspace
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			plan = append(plan, fmt.Sprintf("CREATE mission workspace: %s", workspacePath))
		} else {
			plan = append(plan, fmt.Sprintf("EXISTS mission workspace: %s", workspacePath))
		}

		// Check mission config
		missionConfigPath := filepath.Join(workspacePath, ".orc", "config.json")
		if _, err := os.Stat(missionConfigPath); os.IsNotExist(err) {
			plan = append(plan, fmt.Sprintf("CREATE mission config: %s", missionConfigPath))
		} else {
			plan = append(plan, fmt.Sprintf("EXISTS mission config: %s", missionConfigPath))
		}

		// Check groves directory
		grovesDir := filepath.Join(workspacePath, "groves")
		if _, err := os.Stat(grovesDir); os.IsNotExist(err) {
			plan = append(plan, fmt.Sprintf("CREATE groves directory: %s", grovesDir))
		} else {
			plan = append(plan, fmt.Sprintf("EXISTS groves directory: %s", grovesDir))
		}

		// Plan for each grove
		for _, grove := range groves {
			desiredPath := filepath.Join(grovesDir, grove.Name)
			currentPath := grove.Path

			// Check if grove exists at current path
			currentExists := false
			if _, err := os.Stat(currentPath); err == nil {
				currentExists = true
			}

			// Check if grove exists at desired path
			desiredExists := false
			if _, err := os.Stat(desiredPath); err == nil {
				desiredExists = true
			}

			if currentPath != desiredPath {
				if currentExists && !desiredExists {
					plan = append(plan, fmt.Sprintf("MOVE grove %s: %s → %s", grove.ID, currentPath, desiredPath))
				} else if !currentExists && !desiredExists {
					plan = append(plan, fmt.Sprintf("MISSING grove %s: %s (needs materialization)", grove.ID, desiredPath))
				} else if desiredExists {
					plan = append(plan, fmt.Sprintf("EXISTS grove %s: %s", grove.ID, desiredPath))
					plan = append(plan, fmt.Sprintf("UPDATE DB path for %s: %s → %s", grove.ID, currentPath, desiredPath))
				}
			} else {
				if currentExists {
					plan = append(plan, fmt.Sprintf("EXISTS grove %s: %s", grove.ID, desiredPath))
				} else {
					plan = append(plan, fmt.Sprintf("MISSING grove %s: %s (needs materialization)", grove.ID, desiredPath))
				}
			}

			// Check grove config
			groveConfigPath := filepath.Join(desiredPath, ".orc", "config.json")
			if _, err := os.Stat(groveConfigPath); os.IsNotExist(err) {
				plan = append(plan, fmt.Sprintf("CREATE grove config: %s", groveConfigPath))
			} else {
				plan = append(plan, fmt.Sprintf("EXISTS grove config: %s", groveConfigPath))
			}
		}

		// Check for old .orc-mission files to clean up
		oldMissionFile := filepath.Join(workspacePath, ".orc-mission")
		if _, err := os.Stat(oldMissionFile); err == nil {
			plan = append(plan, fmt.Sprintf("DELETE old .orc-mission: %s", oldMissionFile))
		}

		// Plan TMux session if requested
		if createTmux {
			plan = append(plan, "")
			plan = append(plan, "TMux Session:")
			sessionName := fmt.Sprintf("orc-%s", missionID)
			if tmux.SessionExists(sessionName) {
				plan = append(plan, fmt.Sprintf("  EXISTS session: %s", sessionName))
			} else {
				plan = append(plan, fmt.Sprintf("  CREATE session: %s", sessionName))
				for i, grove := range groves {
					grovePath := filepath.Join(grovesDir, grove.Name)
					if _, err := os.Stat(grovePath); err == nil {
						plan = append(plan, fmt.Sprintf("  CREATE window %d (%s): 3 panes - Grove %s IMP", i+1, grove.Name, grove.ID))
					}
				}
			}
		}

		// Phase 3: Show plan
		fmt.Println("📋 Plan:\n")
		for _, step := range plan {
			fmt.Printf("  %s\n", step)
		}
		fmt.Println()

		// Phase 4: Ask for confirmation
		fmt.Print("Apply changes? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}

		// Phase 5: Apply changes
		fmt.Println("\n🚀 Applying changes...\n")

		// Create mission workspace
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return fmt.Errorf("failed to create mission workspace: %w", err)
		}
		fmt.Printf("✓ Mission workspace ready\n")

		// Create .orc directory
		orcDir := filepath.Join(workspacePath, ".orc")
		if err := os.MkdirAll(orcDir, 0755); err != nil {
			return fmt.Errorf("failed to create .orc directory: %w", err)
		}

		// Write mission config
		if err := context.WriteMissionContext(workspacePath, missionID); err != nil {
			return fmt.Errorf("failed to write mission config: %w", err)
		}
		fmt.Printf("✓ Mission config written\n")

		// Create groves directory
		if err := os.MkdirAll(grovesDir, 0755); err != nil {
			return fmt.Errorf("failed to create groves directory: %w", err)
		}
		fmt.Printf("✓ Groves directory ready\n")

		// Process each grove
		for _, grove := range groves {
			desiredPath := filepath.Join(grovesDir, grove.Name)
			currentPath := grove.Path

			// Move grove if it exists elsewhere
			if currentPath != desiredPath {
				if _, err := os.Stat(currentPath); err == nil {
					// Grove exists at old location - move it
					if err := os.Rename(currentPath, desiredPath); err != nil {
						fmt.Printf("  ⚠️  Could not move grove %s: %v\n", grove.ID, err)
						fmt.Printf("      Manual move required: %s → %s\n", currentPath, desiredPath)
					} else {
						fmt.Printf("✓ Moved grove %s\n", grove.ID)
					}
				}
			}

			// Create grove directory if it doesn't exist
			if _, err := os.Stat(desiredPath); os.IsNotExist(err) {
				fmt.Printf("  ℹ️  Grove %s worktree missing: %s\n", grove.ID, desiredPath)
				fmt.Printf("      Materialize with: orc grove create %s --repos <repo> --mission %s\n", grove.Name, missionID)
			}

			// Create .orc directory in grove
			groveOrcDir := filepath.Join(desiredPath, ".orc")
			os.MkdirAll(groveOrcDir, 0755)

			// Write grove config if grove directory exists
			if _, err := os.Stat(desiredPath); err == nil {
				if err := writeGroveConfig(desiredPath, grove); err != nil {
					fmt.Printf("  ⚠️  Could not write config for grove %s: %v\n", grove.ID, err)
				} else {
					fmt.Printf("✓ Grove %s config written\n", grove.ID)
				}
			}

			// Update DB path if changed
			if currentPath != desiredPath {
				if err := models.UpdateGrovePath(grove.ID, desiredPath); err != nil {
					return fmt.Errorf("failed to update grove path in DB: %w", err)
				}
				fmt.Printf("✓ Updated DB path for grove %s\n", grove.ID)
			}
		}

		// Clean up old .orc-mission file
		oldMissionFile = filepath.Join(workspacePath, ".orc-mission")
		if _, err := os.Stat(oldMissionFile); err == nil {
			if err := os.Remove(oldMissionFile); err != nil {
				fmt.Printf("  ⚠️  Could not remove old .orc-mission: %v\n", err)
			} else {
				fmt.Printf("✓ Removed old .orc-mission file\n")
			}
		}

		fmt.Println()
		fmt.Printf("✅ Mission infrastructure ready at: %s\n", workspacePath)

		// Create TMux session if requested
		if createTmux {
			fmt.Println()
			fmt.Println("🖥️  Creating TMux session...")

			sessionName := fmt.Sprintf("orc-%s", missionID)

			// Check if session already exists
			if tmux.SessionExists(sessionName) {
				fmt.Printf("  ℹ️  Session %s already exists\n", sessionName)
				fmt.Printf("     Attach with: tmux attach -t %s\n", sessionName)
			} else {
				// Create session (first window created automatically by tmux)
				// We'll rename it to the first grove
				session, err := tmux.NewSession(sessionName, workspacePath)
				if err != nil {
					return fmt.Errorf("failed to create TMux session: %w", err)
				}

				// Create window for each grove (starting at window 1)
				for i, grove := range groves {
					windowIndex := i + 1 // Windows start at 1, groves start at 1
					grovePath := filepath.Join(grovesDir, grove.Name)

					// Check if grove path exists
					if _, err := os.Stat(grovePath); err == nil {
						if i == 0 {
							// First grove: rename the default first window
							target := fmt.Sprintf("%s:1", sessionName)
							if err := exec.Command("tmux", "rename-window", "-t", target, grove.Name).Run(); err != nil {
								return fmt.Errorf("failed to rename first window: %w", err)
							}
							// Set working directory for first window
							if err := exec.Command("tmux", "send-keys", "-t", target, fmt.Sprintf("cd %s", grovePath), "C-m").Run(); err != nil {
								fmt.Printf("  ⚠️  Could not set working directory: %v\n", err)
							}
							// Create the 3-pane layout
							target = fmt.Sprintf("%s:%s", sessionName, grove.Name)
							if err := session.SplitVertical(target, grovePath); err != nil {
								return fmt.Errorf("failed to split vertical: %w", err)
							}
							rightPane := fmt.Sprintf("%s.2", target)
							if err := session.SplitHorizontal(rightPane, grovePath); err != nil {
								return fmt.Errorf("failed to split horizontal: %w", err)
							}
							fmt.Printf("✓ Window %d: %s (3 panes - empty shells) [%s IMP]\n", windowIndex, grove.Name, grove.ID)
						} else {
							// Other groves: create new window
							if _, err := session.CreateGroveWindowShell(windowIndex, grove.Name, grovePath); err != nil {
								fmt.Printf("  ⚠️  Could not create window for grove %s: %v\n", grove.ID, err)
								continue
							}
							fmt.Printf("✓ Window %d: %s (3 panes - empty shells) [%s IMP]\n", windowIndex, grove.Name, grove.ID)
						}
					} else {
						fmt.Printf("  ℹ️  Grove %s worktree missing, skipping window\n", grove.ID)
					}
				}

				// Select first grove window
				if len(groves) > 0 {
					session.SelectWindow(1)
				}

				fmt.Println()
				fmt.Printf("✓ TMux session created: %s\n", sessionName)
				fmt.Printf("  Attach with: tmux attach -t %s\n", sessionName)
				fmt.Println()
				fmt.Println("Window Layout:")
				for i, grove := range groves {
					if _, err := os.Stat(filepath.Join(grovesDir, grove.Name)); err == nil {
						fmt.Printf("  Window %d (%s): Grove %s IMP - 3 panes (empty shells)\n", i+1, grove.Name, grove.ID)
					}
				}
				fmt.Println()
				fmt.Println("Each window has layout: Left: (vim) | Right Top: (claude) | Right Bottom: (shell)")
			}
		}

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

// writeGroveConfig writes .orc/config.json for a grove
func writeGroveConfig(grovePath string, grove *models.Grove) error {
	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeGrove,
		Grove: &config.GroveConfig{
			GroveID:   grove.ID,
			MissionID: grove.MissionID,
			Name:      grove.Name,
			Repos:     []string{}, // TODO: parse from grove.Repos field
			CreatedAt: grove.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	return config.SaveConfig(grovePath, cfg)
}

var missionPinCmd = &cobra.Command{
	Use:   "pin [mission-id]",
	Short: "Pin mission to keep it visible",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.PinMission(id)
		if err != nil {
			return fmt.Errorf("failed to pin mission: %w", err)
		}

		fmt.Printf("✓ Mission %s pinned 📌\n", id)
		return nil
	},
}

var missionUnpinCmd = &cobra.Command{
	Use:   "unpin [mission-id]",
	Short: "Unpin mission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		err := models.UnpinMission(id)
		if err != nil {
			return fmt.Errorf("failed to unpin mission: %w", err)
		}

		fmt.Printf("✓ Mission %s unpinned\n", id)
		return nil
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
