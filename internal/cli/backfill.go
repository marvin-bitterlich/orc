package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/db"
)

// BackfillCmd returns the backfill command
func BackfillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backfill",
		Short: "Data migration and backfill commands",
		Long:  `One-time data migration commands. Safe to run multiple times (idempotent).`,
	}

	cmd.AddCommand(backfillLibraryTomesCmd())

	return cmd
}

func backfillLibraryTomesCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "library-tomes",
		Short: "Migrate library tomes to commission root (orphan tomes)",
		Long: `Convert tomes with container_type='library' to orphan tomes at commission root.

This migration:
- Clears container_id and container_type for library tomes
- Library tomes become orphan tomes (visible in ROOT TOMES section)
- Safe to run multiple times (idempotent)

Run this after removing the Library entity from the schema.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.GetDB()
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			// Count library tomes
			var count int
			err = database.QueryRow(`SELECT COUNT(*) FROM tomes WHERE container_type = 'library'`).Scan(&count)
			if err != nil {
				return fmt.Errorf("failed to count library tomes: %w", err)
			}

			if count == 0 {
				fmt.Println("No library tomes found. Migration already complete or no library tomes exist.")
				return nil
			}

			fmt.Printf("Found %d library tomes to migrate...\n", count)

			if dryRun {
				fmt.Println("\n[DRY RUN] No changes made. Run without --dry-run to apply migration.")
				return nil
			}

			// Migrate library tomes to commission root (orphan tomes)
			result, err := database.Exec(`
				UPDATE tomes
				SET container_id = NULL,
				    container_type = NULL,
				    conclave_id = NULL,
				    updated_at = CURRENT_TIMESTAMP
				WHERE container_type = 'library'
			`)
			if err != nil {
				return fmt.Errorf("failed to migrate library tomes: %w", err)
			}

			rowsAffected, _ := result.RowsAffected()
			fmt.Printf("\n✓ Migration complete: %d tomes migrated to commission root\n", rowsAffected)
			fmt.Println("  Library tomes now appear in ROOT TOMES section of orc summary")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without making changes")

	return cmd
}
