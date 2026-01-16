package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
)

type Metadata struct {
	CurrentHandoffID  *string  `json:"current_handoff_id"`
	LastUpdated       *string  `json:"last_updated"`
	ActiveMissionID   *string  `json:"active_mission_id"`
	ActiveWorkOrderID *string  `json:"active_work_order_id"` // Legacy field
	ActiveWorkOrders  []string `json:"active_work_orders"`   // New field (slice of task IDs)
}

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	var showHandoff bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work context from metadata.json",
		Long: `Display the current work context based on metadata.json:
- Active mission, epics, and tasks
- Latest handoff ID and timestamp (use --handoff to see full note)

This provides a focused view of "where am I right now?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we're in a mission context first
			missionCtx, _ := context.DetectMissionContext()
			var metadata Metadata
			var activeMissionID *string

			if missionCtx != nil {
				// Mission context - check workspace .orc/metadata.json first (has active context),
				// then current directory .orc/metadata.json (might be grove metadata)
				workspaceMetadataPath := fmt.Sprintf("%s/.orc/metadata.json", missionCtx.WorkspacePath)
				data, err := os.ReadFile(workspaceMetadataPath)
				if err != nil {
					// No workspace metadata, try current directory
					cwd, _ := os.Getwd()
					localMetadataPath := fmt.Sprintf("%s/.orc/metadata.json", cwd)
					data, err = os.ReadFile(localMetadataPath)
				}

				if err == nil {
					// Try to parse as context metadata (has active_mission_id)
					if err := json.Unmarshal(data, &metadata); err == nil {
						activeMissionID = metadata.ActiveMissionID
					}
				}

				// If still no active mission, use mission from .orc-mission file
				if activeMissionID == nil || *activeMissionID == "" {
					activeMissionID = &missionCtx.MissionID
				}
				fmt.Println("🎯 ORC Status - Mission Context")
			} else {
				// Master context - read from global metadata.json
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}

				metadataPath := fmt.Sprintf("%s/.orc/metadata.json", homeDir)
				data, err := os.ReadFile(metadataPath)
				if err != nil {
					return fmt.Errorf("failed to read metadata.json: %w\nHint: Run 'orc init' if you haven't initialized ORC yet", err)
				}

				if err := json.Unmarshal(data, &metadata); err != nil {
					return fmt.Errorf("failed to parse metadata.json: %w", err)
				}

				activeMissionID = metadata.ActiveMissionID
				fmt.Println("🎯 ORC Status - Current Context")
			}
			fmt.Println()

			// Display active mission
			if activeMissionID != nil && *activeMissionID != "" {
				mission, err := models.GetMission(*activeMissionID)
				if err != nil {
					fmt.Printf("❌ Mission: %s (error loading: %v)\n", *activeMissionID, err)
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

			// Display active tasks (from metadata.ActiveWorkOrders)
			if metadata.ActiveWorkOrders != nil && len(metadata.ActiveWorkOrders) > 0 {
				fmt.Printf("📋 Active Tasks:\n")
				for _, taskID := range metadata.ActiveWorkOrders {
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
			if metadata.CurrentHandoffID != nil && *metadata.CurrentHandoffID != "" {
				handoff, err := models.GetHandoff(*metadata.CurrentHandoffID)
				if err != nil {
					fmt.Printf("❌ Latest Handoff: %s (error loading: %v)\n", *metadata.CurrentHandoffID, err)
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

			if metadata.LastUpdated != nil && *metadata.LastUpdated != "" {
				fmt.Printf("Last updated: %s\n", *metadata.LastUpdated)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showHandoff, "handoff", "n", false, "Show full handoff note")

	return cmd
}
