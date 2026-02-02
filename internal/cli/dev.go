package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/db"
)

// DevCmd returns the dev command group for development utilities.
func DevCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Development utilities (use via orc-dev shim)",
		Long: `Development utilities for working with the ORC dev database.

These commands are intended to be run via the orc-dev shim, which sets
ORC_DB_PATH to ~/.orc/dev.db. Running without the shim will error to
prevent accidental modification of the production database.`,
	}

	cmd.AddCommand(devResetCmd())
	cmd.AddCommand(devDoctorCmd())
	return cmd
}

func devResetCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset dev database with fresh fixtures",
		Long: `Delete the dev database and recreate it with comprehensive fixture data.

This command:
1. Deletes the existing dev database file
2. Creates a fresh database with the current schema
3. Seeds comprehensive fixture data for development

Safety: This command requires ORC_DB_PATH to be set (via orc-dev shim)
to prevent accidental reset of the production database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Safety check: require ORC_DB_PATH to be set
			dbPath := os.Getenv("ORC_DB_PATH")
			if dbPath == "" {
				return fmt.Errorf("ORC_DB_PATH not set - use 'orc-dev reset' instead of 'orc reset'\n\nThis safety check prevents accidental reset of your production database")
			}

			// Confirmation unless --force
			if !force {
				fmt.Printf("This will delete and recreate: %s\n", dbPath)
				fmt.Print("Continue? [y/N] ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			// Close any existing DB connection
			db.Close()

			// Delete existing database
			if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete database: %w", err)
			}
			fmt.Printf("✓ Deleted %s\n", dbPath)

			// Create fresh database with schema
			database, err := db.GetDB()
			if err != nil {
				return fmt.Errorf("failed to create database: %w", err)
			}
			fmt.Println("✓ Created fresh database with schema")

			// Seed fixtures
			if err := db.SeedFixtures(database); err != nil {
				return fmt.Errorf("failed to seed fixtures: %w", err)
			}
			fmt.Println("✓ Seeded fixture data")

			fmt.Println("\nDev database reset complete!")
			fmt.Println("\nSeeded entities:")
			fmt.Println("  - 3 tags")
			fmt.Println("  - 2 repos")
			fmt.Println("  - 2 factories, 3 workshops, 2 workbenches")
			fmt.Println("  - 3 commissions")
			fmt.Println("  - 5 shipments")
			fmt.Println("  - 10 tasks")
			fmt.Println("  - 2 conclaves, 2 tomes, 4 notes")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func devDoctorCmd() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check dev environment health",
		Long: `Check the health of your ORC development environment.

Verifies:
- ORC_DB_PATH environment variable is set
- Dev database exists
- Atlas is installed
- Schema is in sync with schema.sql

Use with orc-dev shim: orc-dev dev doctor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			issues := 0

			if !quiet {
				fmt.Println("=== ORC Dev Environment Health Check ===")
				fmt.Println()
			}

			// 1. Check ORC_DB_PATH
			dbPath := os.Getenv("ORC_DB_PATH")
			if !quiet {
				fmt.Println("1. Environment Configuration")
			}
			if dbPath == "" {
				if !quiet {
					fmt.Println("   ✗ ORC_DB_PATH not set")
					fmt.Println()
					fmt.Println("   FIX: Use 'orc-dev' instead of 'orc' for development")
					fmt.Println("        The orc-dev shim sets ORC_DB_PATH automatically")
				}
				if quiet {
					os.Exit(1)
				}
				return nil
			}
			if !quiet {
				fmt.Printf("   ✓ ORC_DB_PATH=%s\n", dbPath)
			}

			// 2. Check dev database exists
			if !quiet {
				fmt.Println()
				fmt.Println("2. Development Database")
			}
			if info, err := os.Stat(dbPath); err != nil {
				issues++
				if !quiet {
					fmt.Printf("   ✗ Database not found: %s\n", dbPath)
					fmt.Println()
					fmt.Println("   FIX: Run 'orc-dev dev reset' to create dev database")
				}
			} else {
				if !quiet {
					fmt.Printf("   ✓ Database exists (%d KB)\n", info.Size()/1024)
				}
			}

			// 3. Check Atlas installation
			if !quiet {
				fmt.Println()
				fmt.Println("3. Atlas Installation")
			}
			atlasPath, err := exec.LookPath("atlas")
			if err != nil {
				issues++
				if !quiet {
					fmt.Println("   ✗ Atlas not found in PATH")
					fmt.Println()
					fmt.Println("   FIX: Run 'brew install ariga/tap/atlas'")
				}
			} else {
				if !quiet {
					fmt.Printf("   ✓ Atlas installed: %s\n", atlasPath)
				}
			}

			// 4. Check schema sync (only if atlas is installed and DB exists)
			if !quiet {
				fmt.Println()
				fmt.Println("4. Schema Synchronization")
			}
			if atlasPath != "" {
				if _, err := os.Stat(dbPath); err == nil {
					// Run atlas schema diff
					diffCmd := exec.Command("atlas", "schema", "diff", "--env", "local")
					diffCmd.Env = append(os.Environ(), "ORC_DB_PATH="+dbPath)
					output, err := diffCmd.CombinedOutput()

					if err != nil {
						// Non-zero exit could mean diff found or error
						outStr := strings.TrimSpace(string(output))
						if outStr != "" && !strings.Contains(outStr, "Error") {
							issues++
							if !quiet {
								fmt.Println("   ✗ Schema out of sync")
								fmt.Println()
								fmt.Println("   Differences found:")
								// Show first few lines of diff
								lines := strings.Split(outStr, "\n")
								for i, line := range lines {
									if i >= 5 {
										fmt.Printf("        ... (%d more lines)\n", len(lines)-5)
										break
									}
									fmt.Printf("        %s\n", line)
								}
								fmt.Println()
								fmt.Println("   FIX: Run 'make schema-apply' to sync schema")
							}
						} else if strings.Contains(outStr, "Error") {
							if !quiet {
								fmt.Println("   ⚠️  Could not check schema sync")
								fmt.Printf("        %s\n", outStr)
							}
						}
					} else {
						if !quiet {
							fmt.Println("   ✓ Schema in sync with schema.sql")
						}
					}
				} else {
					if !quiet {
						fmt.Println("   ⚠️  Skipped (database doesn't exist)")
					}
				}
			} else {
				if !quiet {
					fmt.Println("   ⚠️  Skipped (atlas not installed)")
				}
			}

			// Summary
			if !quiet {
				fmt.Println()
				if issues == 0 {
					fmt.Println("=== All checks passed! ===")
				} else {
					fmt.Printf("=== %d issue(s) found ===\n", issues)
				}
			}

			if issues > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only output exit code (0=healthy, 1=issues)")
	return cmd
}

// CheckDevEnvironment returns warnings about dev environment issues.
// Can be called from startup to show inline warnings.
func CheckDevEnvironment() []string {
	var warnings []string

	dbPath := os.Getenv("ORC_DB_PATH")
	if dbPath == "" {
		// Not in dev mode, no warnings needed
		return warnings
	}

	// Check if DB exists
	if _, err := os.Stat(dbPath); err != nil {
		warnings = append(warnings, fmt.Sprintf("⚠️  Dev DB missing: %s (run 'orc-dev dev reset')", dbPath))
	}

	return warnings
}
