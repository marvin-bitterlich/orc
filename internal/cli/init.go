package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/looneym/orc/internal/db"
	"github.com/spf13/cobra"
)

// InitCmd returns the init command
func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the ORC database",
		Long:  `Initialize the ORC database at ~/.orc/orc.db with the required schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := db.GetDBPath()
			if err != nil {
				return fmt.Errorf("failed to get database path: %w", err)
			}

			fmt.Printf("Initializing ORC database at %s\n", dbPath)

			// Initialize schema
			if err := db.InitSchema(); err != nil {
				return fmt.Errorf("failed to initialize schema: %w", err)
			}

			fmt.Println("✓ Database initialized successfully")

			// Initialize metadata.json
			if err := initMetadata(); err != nil {
				return fmt.Errorf("failed to initialize metadata: %w", err)
			}

			fmt.Println("✓ Metadata file created at ~/.orc/metadata.json")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  orc expedition create \"My First Expedition\"")
			fmt.Println("  orc status")

			return nil
		},
	}
}

// initMetadata creates the initial metadata.json file
func initMetadata() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	orcDir := fmt.Sprintf("%s/.orc", homeDir)
	metadataPath := fmt.Sprintf("%s/metadata.json", orcDir)

	// Check if file already exists
	if _, err := os.Stat(metadataPath); err == nil {
		return nil // Already exists, skip
	}

	metadata := map[string]any{
		"current_handoff_id":    nil,
		"last_updated":          nil,
		"active_mission_id":     nil,
		"active_operation_id":   nil,
		"active_work_order_id":  nil,
		"active_expedition_id":  nil,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}
