package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

type Metadata struct {
	CurrentHandoffID   *string `json:"current_handoff_id"`
	LastUpdated        *string `json:"last_updated"`
	ActiveMissionID    *string `json:"active_mission_id"`
	ActiveOperationID  *string `json:"active_operation_id"`
	ActiveWorkOrderID  *string `json:"active_work_order_id"`
	ActiveExpeditionID *string `json:"active_expedition_id"`
}

// StatusCmd returns the status command
func StatusCmd() *cobra.Command {
	var showHandoff bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current work context from metadata.json",
		Long: `Display the current work context based on metadata.json:
- Active mission, operation, work order, expedition
- Latest handoff ID and timestamp (use --handoff to see full note)

This provides a focused view of "where am I right now?"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read metadata.json
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			metadataPath := fmt.Sprintf("%s/.orc/metadata.json", homeDir)
			data, err := os.ReadFile(metadataPath)
			if err != nil {
				return fmt.Errorf("failed to read metadata.json: %w\nHint: Run 'orc init' if you haven't initialized ORC yet", err)
			}

			var metadata Metadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				return fmt.Errorf("failed to parse metadata.json: %w", err)
			}

			fmt.Println("🎯 ORC Status - Current Context")
			fmt.Println()

			// Display active mission
			if metadata.ActiveMissionID != nil && *metadata.ActiveMissionID != "" {
				mission, err := models.GetMission(*metadata.ActiveMissionID)
				if err != nil {
					fmt.Printf("❌ Mission: %s (error loading: %v)\n", *metadata.ActiveMissionID, err)
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

			// Display active operation
			if metadata.ActiveOperationID != nil && *metadata.ActiveOperationID != "" {
				operation, err := models.GetOperation(*metadata.ActiveOperationID)
				if err != nil {
					fmt.Printf("❌ Operation: %s (error loading: %v)\n", *metadata.ActiveOperationID, err)
				} else {
					fmt.Printf("⚙️  Operation: %s - %s [%s]\n", operation.ID, operation.Title, operation.Status)
					if operation.Description.Valid && operation.Description.String != "" {
						fmt.Printf("   %s\n", operation.Description.String)
					}
				}
			} else {
				fmt.Println("⚙️  Operation: (none active)")
			}
			fmt.Println()

			// Display active work order
			if metadata.ActiveWorkOrderID != nil && *metadata.ActiveWorkOrderID != "" {
				workOrder, err := models.GetWorkOrder(*metadata.ActiveWorkOrderID)
				if err != nil {
					fmt.Printf("❌ Work Order: %s (error loading: %v)\n", *metadata.ActiveWorkOrderID, err)
				} else {
					fmt.Printf("📋 Work Order: %s - %s [%s]\n", workOrder.ID, workOrder.Title, workOrder.Status)
					if workOrder.Description.Valid && workOrder.Description.String != "" {
						fmt.Printf("   %s\n", workOrder.Description.String)
					}
					if workOrder.AssignedGroveID.Valid && workOrder.AssignedGroveID.String != "" {
						fmt.Printf("   Grove: %s\n", workOrder.AssignedGroveID.String)
					}
				}
			} else {
				fmt.Println("📋 Work Order: (none active)")
			}
			fmt.Println()

			// Display active expedition
			if metadata.ActiveExpeditionID != nil && *metadata.ActiveExpeditionID != "" {
				expedition, err := models.GetExpedition(*metadata.ActiveExpeditionID)
				if err != nil {
					fmt.Printf("❌ Expedition: %s (error loading: %v)\n", *metadata.ActiveExpeditionID, err)
				} else {
					fmt.Printf("🌲 Expedition: %s - %s [%s]\n", expedition.ID, expedition.Name, expedition.Status)
					if expedition.AssignedIMP.Valid && expedition.AssignedIMP.String != "" {
						fmt.Printf("   IMP: %s\n", expedition.AssignedIMP.String)
					}
					if expedition.WorkOrderID.Valid && expedition.WorkOrderID.String != "" {
						fmt.Printf("   Work Order: %s\n", expedition.WorkOrderID.String)
					}
				}
			} else {
				fmt.Println("🌲 Expedition: (none active)")
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
