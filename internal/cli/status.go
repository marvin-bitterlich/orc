package cli

import (
	"fmt"
	"os"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	var showHandoff bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work context from config.json",
		Long: `Display the current work context based on .orc/config.json:
- Active mission, epics, and tasks
- Latest handoff ID and timestamp (use --handoff to see full note)

This provides a focused view of "where am I right now?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we're in a mission context first
			missionCtx, _ := context.DetectMissionContext()
			var activeMissionID string
			var activeWorkOrders []string
			var currentHandoffID string
			var lastUpdated string
			var currentEpic string

			if missionCtx != nil {
				// Mission context - try to load config from workspace or current directory
				cfg, err := config.LoadConfig(missionCtx.WorkspacePath)
				if err != nil {
					// Try current directory
					cwd, _ := os.Getwd()
					cfg, err = config.LoadConfig(cwd)
				}

				if err == nil {
					// Extract fields based on config type
					switch cfg.Type {
					case config.TypeGrove:
						activeMissionID = cfg.Grove.MissionID
						currentEpic = cfg.Grove.CurrentEpic
					case config.TypeMission:
						activeMissionID = cfg.Mission.MissionID
						currentEpic = cfg.Mission.CurrentEpic
					case config.TypeGlobal:
						activeMissionID = cfg.State.ActiveMissionID
						activeWorkOrders = cfg.State.ActiveWorkOrders
						currentHandoffID = cfg.State.CurrentHandoffID
						lastUpdated = cfg.State.LastUpdated
						currentEpic = cfg.State.CurrentEpic
					}
				}

				// If still no active mission, use mission from .orc-mission file
				if activeMissionID == "" {
					activeMissionID = missionCtx.MissionID
				}
				fmt.Println("🎯 ORC Status - Mission Context")
			} else {
				// Master context - read from global config.json
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}

				cfg, err := config.LoadConfig(homeDir)
				if err != nil {
					return fmt.Errorf("failed to read config.json: %w\nHint: Run 'orc init' if you haven't initialized ORC yet", err)
				}

				if cfg.State != nil {
					activeMissionID = cfg.State.ActiveMissionID
					activeWorkOrders = cfg.State.ActiveWorkOrders
					currentHandoffID = cfg.State.CurrentHandoffID
					lastUpdated = cfg.State.LastUpdated
					currentEpic = cfg.State.CurrentEpic
				}

				fmt.Println("🎯 ORC Status - Current Context")
			}
			fmt.Println()

			// Display active mission
			if activeMissionID != "" {
				mission, err := models.GetMission(activeMissionID)
				if err != nil {
					fmt.Printf("❌ Mission: %s (error loading: %v)\n", activeMissionID, err)
				} else {
					fmt.Printf("🎯 Mission: %s - %s [%s]\n", mission.ID, mission.Title, mission.Status)
					if mission.Description.Valid && mission.Description.String != "" {
						fmt.Printf("   %s\n", mission.Description.String)
					}
				}
			} else {
				fmt.Println("🎯 Mission: (none active)")
			}
			fmt.Println()

			// Display current epic if focused
			if currentEpic != "" {
				epic, err := models.GetEpic(currentEpic)
				if err != nil {
					fmt.Printf("🎯 Focused Epic: %s (error loading: %v)\n", currentEpic, err)
				} else {
					fmt.Printf("🎯 Focused Epic: %s - %s [%s]\n", epic.ID, epic.Title, epic.Status)
				}
				fmt.Println()
			}

			// Display active tasks (from config.State.ActiveWorkOrders)
			if len(activeWorkOrders) > 0 {
				fmt.Printf("📋 Active Tasks:\n")
				for _, taskID := range activeWorkOrders {
					task, err := models.GetTask(taskID)
					if err != nil {
						fmt.Printf("   ❌ %s (error loading: %v)\n", taskID, err)
					} else {
						fmt.Printf("   %s - %s [%s]\n", task.ID, task.Title, task.Status)
						if task.AssignedGroveID.Valid && task.AssignedGroveID.String != "" {
							fmt.Printf("      Grove: %s\n", task.AssignedGroveID.String)
						}
					}
				}
			} else {
				fmt.Println("📋 Active Tasks: (none)")
			}
			fmt.Println()

			// Display latest handoff
			if currentHandoffID != "" {
				handoff, err := models.GetHandoff(currentHandoffID)
				if err != nil {
					fmt.Printf("❌ Latest Handoff: %s (error loading: %v)\n", currentHandoffID, err)
				} else {
					fmt.Printf("📝 Latest Handoff: %s\n", handoff.ID)
					fmt.Printf("   Created: %s\n", handoff.CreatedAt.Format("2006-01-02 15:04:05"))

					// Show full note if --handoff flag is set
					if showHandoff {
						fmt.Println()
						fmt.Println("--- HANDOFF NOTE ---")
						fmt.Println(handoff.HandoffNote)
						fmt.Println("--- END HANDOFF ---")
					}
				}
			} else {
				fmt.Println("📝 Latest Handoff: (none)")
			}
			fmt.Println()

			if lastUpdated != "" {
				fmt.Printf("Last updated: %s\n", lastUpdated)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showHandoff, "handoff", "n", false, "Show full handoff note")

	return cmd
}
