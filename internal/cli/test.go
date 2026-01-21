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

- Delete all MISSION-TEST-* missions from database
- Remove all test-canary-* and orc-canary worktrees
- Kill all orc-MISSION-TEST-* TMux sessions

This prevents test state accumulation from repeated runs and ensures
each test starts with a clean environment.

Examples:
  orc test cleanup-pre-flight
  orc test cleanup-pre-flight --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧹 Pre-flight cleanup starting...")
			fmt.Println()

			var totalCleaned int

			// 1. Clean up test missions
			fmt.Println("📦 Cleaning up test missions...")
			ctx := context.Background()
			missions, err := wire.MissionService().ListMissions(ctx, primary.MissionFilters{})
			if err != nil {
				return fmt.Errorf("failed to list missions: %w", err)
			}

			var missionsCleaned int
			for _, mission := range missions {
				if strings.HasPrefix(mission.ID, "MISSION-TEST-") {
					if dryRun {
						fmt.Printf("  [DRY RUN] Would delete: %s (%s)\n", mission.ID, mission.Title)
					} else {
						if err := wire.MissionService().DeleteMission(ctx, primary.DeleteMissionRequest{MissionID: mission.ID, Force: true}); err != nil {
							fmt.Printf("  ⚠️  Failed to delete %s: %v\n", mission.ID, err)
						} else {
							fmt.Printf("  ✓ Deleted: %s\n", mission.ID)
							missionsCleaned++
						}
					}
				}
			}

			if missionsCleaned == 0 && !dryRun {
				fmt.Println("  ℹ️  No test missions found")
			} else if dryRun {
				fmt.Printf("  Found %d test missions\n", missionsCleaned)
			}
			totalCleaned += missionsCleaned
			fmt.Println()

			// 2. Clean up test groves
			fmt.Println("🌳 Cleaning up test groves...")
			groves, err := wire.GroveService().ListGroves(ctx, primary.GroveFilters{})
			if err != nil {
				return fmt.Errorf("failed to list groves: %w", err)
			}

			var grovesCleaned int
			for _, grove := range groves {
				// Match groves associated with test missions
				if strings.HasPrefix(grove.MissionID, "MISSION-TEST-") {
					if dryRun {
						fmt.Printf("  [DRY RUN] Would delete: %s (%s) at %s\n", grove.ID, grove.Name, grove.Path)
					} else {
						// Try to remove worktree
						if _, err := os.Stat(grove.Path); err == nil {
							// Try git worktree remove first
							if err := exec.Command("git", "worktree", "remove", grove.Path, "--force").Run(); err != nil {
								// Fall back to direct removal
								os.RemoveAll(grove.Path)
							}
						}

						// Delete from database
						if err := wire.GroveService().DeleteGrove(ctx, primary.DeleteGroveRequest{GroveID: grove.ID, Force: true}); err != nil {
							fmt.Printf("  ⚠️  Failed to delete %s: %v\n", grove.ID, err)
						} else {
							fmt.Printf("  ✓ Deleted: %s (%s)\n", grove.ID, grove.Name)
							grovesCleaned++
						}
					}
				}
			}

			if grovesCleaned == 0 && !dryRun {
				fmt.Println("  ℹ️  No test groves found")
			} else if dryRun {
				fmt.Printf("  Found %d test groves\n", grovesCleaned)
			}
			totalCleaned += grovesCleaned
			fmt.Println()

			// 3. Clean up orphaned worktrees (not in database)
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
					if strings.HasPrefix(session, "orc-MISSION-TEST-") {
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

			// 5. Clean up mission workspaces
			fmt.Println("🗂️  Cleaning up test mission workspaces...")
			missionsPath := filepath.Join(home, "src", "missions")
			var workspacesCleaned int

			if entries, err := os.ReadDir(missionsPath); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						name := entry.Name()
						if strings.HasPrefix(name, "MISSION-TEST-") {
							fullPath := filepath.Join(missionsPath, name)
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
				fmt.Println("  ℹ️  No test mission workspaces found")
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
