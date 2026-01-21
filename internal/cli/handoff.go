package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
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
		ctx := context.Background()
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
		groveID, _ := cmd.Flags().GetString("grove")
		todosFile, _ := cmd.Flags().GetString("todos")

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

		resp, err := wire.HandoffService().CreateHandoff(ctx, primary.CreateHandoffRequest{
			HandoffNote:     note,
			ActiveMissionID: missionID,
			ActiveGroveID:   groveID,
			TodosSnapshot:   todosJSON,
		})
		if err != nil {
			return fmt.Errorf("failed to create handoff: %w", err)
		}

		handoff := resp.Handoff
		fmt.Printf("✓ Created handoff %s\n", handoff.ID)
		fmt.Printf("  Created: %s\n", handoff.CreatedAt)
		if handoff.ActiveMissionID != "" {
			fmt.Printf("  Mission: %s\n", handoff.ActiveMissionID)
		}
		if handoff.ActiveGroveID != "" {
			fmt.Printf("  Grove: %s\n", handoff.ActiveGroveID)
		}

		// Update global state config
		if err := updateMetadata(handoff); err != nil {
			fmt.Printf("  Warning: Failed to update .orc/config.json: %v\n", err)
		} else {
			fmt.Printf("  Updated: .orc/config.json\n")
		}

		return nil
	},
}

var handoffShowCmd = &cobra.Command{
	Use:   "show [handoff-id]",
	Short: "Show handoff details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		id := args[0]

		handoff, err := wire.HandoffService().GetHandoff(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get handoff: %w", err)
		}

		fmt.Printf("\nHandoff: %s\n", handoff.ID)
		fmt.Printf("Created: %s\n", handoff.CreatedAt)
		if handoff.ActiveMissionID != "" {
			fmt.Printf("Mission: %s\n", handoff.ActiveMissionID)
		}
		if handoff.ActiveGroveID != "" {
			fmt.Printf("Grove: %s\n", handoff.ActiveGroveID)
		}

		fmt.Printf("\n--- HANDOFF NOTE ---\n\n%s\n\n", handoff.HandoffNote)

		if handoff.TodosSnapshot != "" {
			fmt.Printf("--- TODOS SNAPSHOT ---\n\n%s\n\n", handoff.TodosSnapshot)
		}

		return nil
	},
}

var handoffListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent handoffs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		limit, _ := cmd.Flags().GetInt("limit")

		handoffs, err := wire.HandoffService().ListHandoffs(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to list handoffs: %w", err)
		}

		if len(handoffs) == 0 {
			fmt.Println("No handoffs found")
			return nil
		}

		fmt.Printf("\n%-10s %-20s %-15s %-15s\n", "ID", "CREATED", "MISSION", "GROVE")
		fmt.Println("────────────────────────────────────────────────────────────────")
		for _, h := range handoffs {
			mission := "-"
			if h.ActiveMissionID != "" {
				mission = h.ActiveMissionID
			}
			grove := "-"
			if h.ActiveGroveID != "" {
				grove = h.ActiveGroveID
			}
			fmt.Printf("%-10s %-20s %-15s %-15s\n",
				h.ID,
				h.CreatedAt,
				mission,
				grove,
			)
		}
		fmt.Println()

		return nil
	},
}

// updateMetadata updates the .orc/config.json file with global state
func updateMetadata(handoff *primary.Handoff) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	activeMissionID := handoff.ActiveMissionID

	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeGlobal,
		State: &config.StateConfig{
			ActiveMissionID:  activeMissionID,
			CurrentHandoffID: handoff.ID,
			LastUpdated:      time.Now().Format(time.RFC3339),
		},
	}

	return config.SaveConfig(homeDir, cfg)
}

// HandoffCmd returns the handoff command
func HandoffCmd() *cobra.Command {
	// Add flags
	handoffCreateCmd.Flags().StringP("note", "n", "", "Handoff note text")
	handoffCreateCmd.Flags().StringP("file", "f", "", "Read handoff note from file")
	handoffCreateCmd.Flags().Bool("stdin", false, "Read handoff note from stdin")
	handoffCreateCmd.Flags().StringP("mission", "m", "", "Active mission ID")
	handoffCreateCmd.Flags().String("grove", "", "Active grove ID (for IMP handoffs)")
	handoffCreateCmd.Flags().StringP("todos", "t", "", "Path to todos JSON file")

	handoffListCmd.Flags().IntP("limit", "l", 10, "Maximum number of handoffs to show")

	// Add subcommands
	handoffCmd.AddCommand(handoffCreateCmd)
	handoffCmd.AddCommand(handoffShowCmd)
	handoffCmd.AddCommand(handoffListCmd)

	return handoffCmd
}
