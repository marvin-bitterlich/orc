package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/db"
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

			// Initialize config.json
			if err := initConfig(); err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			fmt.Println("✓ Config file created at ~/.orc/config.json")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  orc expedition create \"My First Expedition\"")
			fmt.Println("  orc status")

			return nil
		},
	}
}

// initConfig creates the initial config.json file
func initConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := fmt.Sprintf("%s/.orc/config.json", homeDir)

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Already exists, skip
	}

	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeGlobal,
		State: &config.StateConfig{
			ActiveMissionID:  "",
			CurrentHandoffID: "",
			CurrentFocus:     "",
			LastUpdated:      time.Now().Format(time.RFC3339),
		},
	}

	return config.SaveConfig(homeDir, cfg)
}
