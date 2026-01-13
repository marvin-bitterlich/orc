package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
)

var handoffCmd = &cobra.Command{
	Use:   "handoff",
	Short: "Manage Claude-to-Claude context handoffs",
	Long:  "Create and view handoff notes for seamless context transfer between Claude sessions",
}

var handoffCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new handoff note",
	Long: `Create a handoff note capturing current context for the next Claude session.

The handoff note should be a narrative written from one Claude to another, explaining:
- What we just accomplished
- Current state of work
- What comes next
- Any important context or gotchas

Examples:
  orc handoff create --note "Just completed the ledger CLI..."
  orc handoff create --file handoff.md
  echo "Context..." | orc handoff create --stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var note string
		var err error

		noteFlag, _ := cmd.Flags().GetString("note")
		fileFlag, _ := cmd.Flags().GetString("file")
		stdinFlag, _ := cmd.Flags().GetBool("stdin")

		// Read note from different sources
		switch {
		case noteFlag != "":
			note = noteFlag
		case fileFlag != "":
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			note = string(data)
		case stdinFlag:
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			note = string(data)
		default:
			return fmt.Errorf("must provide --note, --file, or --stdin")
		}

		note = strings.TrimSpace(note)
		if note == "" {
			return fmt.Errorf("handoff note cannot be empty")
		}

		// Get active context from flags
		missionID, _ := cmd.Flags().GetString("mission")
		operationID, _ := cmd.Flags().GetString("operation")
		workOrderID, _ := cmd.Flags().GetString("work-order")
		expeditionID, _ := cmd.Flags().GetString("expedition")
		todosFile, _ := cmd.Flags().GetString("todos")
		graphitiUUID, _ := cmd.Flags().GetString("graphiti-uuid")

		// Read todos JSON if provided
		var todosJSON string
		if todosFile != "" {
			data, err := os.ReadFile(todosFile)
			if err != nil {
				return fmt.Errorf("failed to read todos file: %w", err)
			}
			todosJSON = string(data)

			// Validate JSON
			var temp interface{}
			if err := json.Unmarshal(data, &temp); err != nil {
				return fmt.Errorf("invalid todos JSON: %w", err)
			}
		}

		handoff, err := models.CreateHandoff(note, missionID, operationID, workOrderID, expeditionID, todosJSON, graphitiUUID)
		if err != nil {
			return fmt.Errorf("failed to create handoff: %w", err)
		}

		fmt.Printf("✓ Created handoff %s\n", handoff.ID)
		fmt.Printf("  Created: %s\n", handoff.CreatedAt.Format("2006-01-02 15:04"))
		if handoff.ActiveMissionID.Valid {
			fmt.Printf("  Mission: %s\n", handoff.ActiveMissionID.String)
		}
		if handoff.ActiveOperationID.Valid {
			fmt.Printf("  Operation: %s\n", handoff.ActiveOperationID.String)
		}
		if handoff.ActiveWorkOrderID.Valid {
			fmt.Printf("  Work Order: %s\n", handoff.ActiveWorkOrderID.String)
		}
		if handoff.ActiveExpeditionID.Valid {
			fmt.Printf("  Expedition: %s\n", handoff.ActiveExpeditionID.String)
		}

		// Update metadata file
		if err := updateMetadata(handoff); err != nil {
			fmt.Printf("  Warning: Failed to update metadata.json: %v\n", err)
		} else {
			fmt.Printf("  Updated: .orc/metadata.json\n")
		}

		return nil
	},
}

var handoffShowCmd = &cobra.Command{
	Use:   "show [handoff-id]",
	Short: "Show handoff details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		handoff, err := models.GetHandoff(id)
		if err != nil {
			return fmt.Errorf("failed to get handoff: %w", err)
		}

		fmt.Printf("\nHandoff: %s\n", handoff.ID)
		fmt.Printf("Created: %s\n", handoff.CreatedAt.Format("2006-01-02 15:04:05"))
		if handoff.ActiveMissionID.Valid {
			fmt.Printf("Mission: %s\n", handoff.ActiveMissionID.String)
		}
		if handoff.ActiveOperationID.Valid {
			fmt.Printf("Operation: %s\n", handoff.ActiveOperationID.String)
		}
		if handoff.ActiveWorkOrderID.Valid {
			fmt.Printf("Work Order: %s\n", handoff.ActiveWorkOrderID.String)
		}
		if handoff.ActiveExpeditionID.Valid {
			fmt.Printf("Expedition: %s\n", handoff.ActiveExpeditionID.String)
		}
		if handoff.GraphitiEpisodeUUID.Valid {
			fmt.Printf("Graphiti: %s\n", handoff.GraphitiEpisodeUUID.String)
		}

		fmt.Printf("\n--- HANDOFF NOTE ---\n\n%s\n\n", handoff.HandoffNote)

		if handoff.TodosSnapshot.Valid {
			fmt.Printf("--- TODOS SNAPSHOT ---\n\n%s\n\n", handoff.TodosSnapshot.String)
		}

		return nil
	},
}

var handoffListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent handoffs",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		handoffs, err := models.ListHandoffs(limit)
		if err != nil {
			return fmt.Errorf("failed to list handoffs: %w", err)
		}

		if len(handoffs) == 0 {
			fmt.Println("No handoffs found")
			return nil
		}

		fmt.Printf("\n%-10s %-20s %-15s %-15s\n", "ID", "CREATED", "MISSION", "OPERATION")
		fmt.Println("─────────────────────────────────────────────────────────────────────")
		for _, h := range handoffs {
			mission := "-"
			if h.ActiveMissionID.Valid {
				mission = h.ActiveMissionID.String
			}
			operation := "-"
			if h.ActiveOperationID.Valid {
				operation = h.ActiveOperationID.String
			}
			fmt.Printf("%-10s %-20s %-15s %-15s\n",
				h.ID,
				h.CreatedAt.Format("2006-01-02 15:04"),
				mission,
				operation,
			)
		}
		fmt.Println()

		return nil
	},
}

// updateMetadata updates the .orc/metadata.json file with the latest handoff
func updateMetadata(handoff *models.Handoff) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	orcDir := fmt.Sprintf("%s/.orc", homeDir)
	metadataPath := fmt.Sprintf("%s/metadata.json", orcDir)

	metadata := map[string]interface{}{
		"current_handoff_id": handoff.ID,
		"last_updated":       handoff.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if handoff.ActiveMissionID.Valid {
		metadata["active_mission_id"] = handoff.ActiveMissionID.String
	}
	if handoff.ActiveOperationID.Valid {
		metadata["active_operation_id"] = handoff.ActiveOperationID.String
	}
	if handoff.ActiveWorkOrderID.Valid {
		metadata["active_work_order_id"] = handoff.ActiveWorkOrderID.String
	}
	if handoff.ActiveExpeditionID.Valid {
		metadata["active_expedition_id"] = handoff.ActiveExpeditionID.String
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// HandoffCmd returns the handoff command
func HandoffCmd() *cobra.Command {
	// Add flags
	handoffCreateCmd.Flags().StringP("note", "n", "", "Handoff note text")
	handoffCreateCmd.Flags().StringP("file", "f", "", "Read handoff note from file")
	handoffCreateCmd.Flags().Bool("stdin", false, "Read handoff note from stdin")
	handoffCreateCmd.Flags().StringP("mission", "m", "", "Active mission ID")
	handoffCreateCmd.Flags().StringP("operation", "o", "", "Active operation ID")
	handoffCreateCmd.Flags().StringP("work-order", "w", "", "Active work order ID")
	handoffCreateCmd.Flags().StringP("expedition", "e", "", "Active expedition ID")
	handoffCreateCmd.Flags().StringP("todos", "t", "", "Path to todos JSON file")
	handoffCreateCmd.Flags().StringP("graphiti-uuid", "g", "", "Graphiti episode UUID")

	handoffListCmd.Flags().IntP("limit", "l", 10, "Maximum number of handoffs to show")

	// Add subcommands
	handoffCmd.AddCommand(handoffCreateCmd)
	handoffCmd.AddCommand(handoffShowCmd)
	handoffCmd.AddCommand(handoffListCmd)

	return handoffCmd
}
