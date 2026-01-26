package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// TestCmd returns the test command
func TestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Testing and cleanup utilities for ORC orchestration tests",
		Long:  `Commands for running and managing orchestration tests.`,
	}

	cmd.AddCommand(testCleanupPreFlightCmd())

	return cmd
}

func testCleanupPreFlightCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "cleanup-pre-flight",
		Short: "Clean up all test state before running orchestration tests",
		Long: `Remove all test-related resources to ensure a clean slate:

- Delete all COMM-TEST-* commissions from database
- Remove all test-canary-* and orc-canary worktrees
- Kill all orc-COMM-TEST-* TMux sessions

This prevents test state accumulation from repeated runs and ensures
each test starts with a clean environment.

Examples:
  orc test cleanup-pre-flight
  orc test cleanup-pre-flight --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Pre-flight cleanup starting...")
			fmt.Println()

			var totalCleaned int

			// 1. Clean up test commissions
			fmt.Println("📦 Cleaning up test commissions...")
			ctx := context.Background()
			commissions, err := wire.CommissionService().ListCommissions(ctx, primary.CommissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list commissions: %w", err)
			}

			var commissionsCleaned int
			for _, commission := range commissions {
				if strings.HasPrefix(commission.ID, "COMM-TEST-") {
					if dryRun {
						fmt.Printf("  [DRY RUN] Would delete: %s (%s)\n", commission.ID, commission.Title)
					} else {
						if err := wire.CommissionService().DeleteCommission(ctx, primary.DeleteCommissionRequest{CommissionID: commission.ID, Force: true}); err != nil {
							fmt.Printf("  ⚠️  Failed to delete %s: %v\n", commission.ID, err)
						} else {
							fmt.Printf("  ✓ Test data %s deleted\n", commission.ID)
							commissionsCleaned++
						}
					}
				}
			}

			if commissionsCleaned == 0 && !dryRun {
				fmt.Println("  ℹ️  No test commissions found")
			} else if dryRun {
				fmt.Printf("  Found %d test commissions\n", commissionsCleaned)
			}
			totalCleaned += commissionsCleaned
			fmt.Println()

			// 2. Clean up orphaned worktrees (not in database)
			fmt.Println("📁 Scanning for orphaned test worktrees...")
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}

			worktreesPath := filepath.Join(home, "src", "worktrees")
			var worktreesCleaned int

			if entries, err := os.ReadDir(worktreesPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						name := entry.Name()
						// Match test-canary-* or orc-canary patterns
						if strings.HasPrefix(name, "test-canary-") || name == "orc-canary" {
							fullPath := filepath.Join(worktreesPath, name)
							if dryRun {
								fmt.Printf("  [DRY RUN] Would remove: %s\n", fullPath)
								worktreesCleaned++
							} else {
								// Try git worktree remove first
								if err := exec.Command("git", "worktree", "remove", fullPath, "--force").Run(); err != nil {
									// Fall back to direct removal
									if err := os.RemoveAll(fullPath); err != nil {
										fmt.Printf("  ⚠️  Failed to remove %s: %v\n", fullPath, err)
										continue
									}
								}
								fmt.Printf("  ✓ Removed: %s\n", name)
								worktreesCleaned++
							}
						}
					}
				}
			}

			if worktreesCleaned == 0 && !dryRun {
				fmt.Println("  ℹ️  No orphaned test worktrees found")
			} else if dryRun {
				fmt.Printf("  Found %d orphaned worktrees\n", worktreesCleaned)
			}
			totalCleaned += worktreesCleaned
			fmt.Println()

			// 4. Clean up TMux sessions
			fmt.Println("🪟  Cleaning up test TMux sessions...")
			output, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
			var tmuxCleaned int

			if err == nil {
				sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, session := range sessions {
					session = strings.TrimSpace(session)
					if strings.HasPrefix(session, "orc-COMM-TEST-") {
						if dryRun {
							fmt.Printf("  [DRY RUN] Would kill: %s\n", session)
							tmuxCleaned++
						} else {
							if err := exec.Command("tmux", "kill-session", "-t", session).Run(); err != nil {
								fmt.Printf("  ⚠️  Failed to kill %s: %v\n", session, err)
							} else {
								fmt.Printf("  ✓ Killed: %s\n", session)
								tmuxCleaned++
							}
						}
					}
				}
			} else {
				fmt.Println("  ℹ️  No TMux server running or no sessions found")
			}

			if tmuxCleaned == 0 && !dryRun && err == nil {
				fmt.Println("  ℹ️  No test TMux sessions found")
			} else if dryRun {
				fmt.Printf("  Found %d test sessions\n", tmuxCleaned)
			}
			totalCleaned += tmuxCleaned
			fmt.Println()

			// 5. Clean up commission workspaces
			fmt.Println("🗂️  Cleaning up test commission workspaces...")
			factoriesPath := filepath.Join(home, "src", "factories")
			var workspacesCleaned int

			if entries, err := os.ReadDir(factoriesPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						name := entry.Name()
						if strings.HasPrefix(name, "COMM-TEST-") {
							fullPath := filepath.Join(factoriesPath, name)
							if dryRun {
								fmt.Printf("  [DRY RUN] Would remove: %s\n", fullPath)
								workspacesCleaned++
							} else {
								if err := os.RemoveAll(fullPath); err != nil {
									fmt.Printf("  ⚠️  Failed to remove %s: %v\n", fullPath, err)
								} else {
									fmt.Printf("  ✓ Removed: %s\n", name)
									workspacesCleaned++
								}
							}
						}
					}
				}
			}

			if workspacesCleaned == 0 && !dryRun {
				fmt.Println("  ℹ️  No test commission workspaces found")
			} else if dryRun {
				fmt.Printf("  Found %d test workspaces\n", workspacesCleaned)
			}
			totalCleaned += workspacesCleaned
			fmt.Println()

			// Summary
			fmt.Println("═══════════════════════════════════════")
			if dryRun {
				fmt.Printf("✓ Pre-flight scan complete: %d resources found\n", totalCleaned)
				fmt.Println()
				fmt.Println("Run without --dry-run to perform cleanup")
			} else {
				fmt.Printf("✓ Pre-flight cleanup complete: %d resources cleaned\n", totalCleaned)
				fmt.Println()
				fmt.Println("Environment ready for orchestration test")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleaned without actually deleting")

	return cmd
}
