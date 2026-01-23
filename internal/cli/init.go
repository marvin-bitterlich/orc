package cli

import (
	"fmt"

	"github.com/spf13/cobra"

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

			fmt.Println("Database initialized successfully")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("  orc commission create \"My First Commission\"")
			fmt.Println("  orc status")

			return nil
		},
	}
}
