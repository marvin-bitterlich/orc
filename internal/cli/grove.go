package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/tmux"
	"github.com/example/orc/internal/wire"
)

// GroveCmd returns the grove command
func GroveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grove",
		Short: "Manage groves (mission worktrees)",
		Long:  `Create and manage groves - isolated development workspaces for missions.`,
	}

	cmd.AddCommand(groveCreateCmd())
	cmd.AddCommand(groveListCmd())
	cmd.AddCommand(groveShowCmd())
	cmd.AddCommand(groveRenameCmd())
	cmd.AddCommand(groveOpenCmd())
	cmd.AddCommand(groveDeleteCmd())

	return cmd
}

func groveCreateCmd() *cobra.Command {
	var missionID string
	var repos []string
	var basePath string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new grove (worktree) for a mission",
		Long: `Create a new grove with git worktree integration.

This command:
1. Creates a grove record in the database
2. Creates git worktree(s) for specified repos
3. Writes .orc/config.json (grove config)

Examples:
  orc grove create auth-backend --repos main-app --mission MISSION-001
  orc grove create frontend --repos main-app --mission MISSION-001
  orc grove create multi --repos main-app,api-service --mission MISSION-002`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Validate Claude workspace trust before creating grove
			if err := validateClaudeWorkspaceTrust(); err != nil {
				return fmt.Errorf("Claude workspace trust validation failed:\n\n%w\n\nRun 'orc doctor' for detailed diagnostics", err)
			}

			name := args[0]

			// Get mission from context or require explicit flag
			if missionID == "" {
				missionID = orccontext.GetContextMissionID()
				if missionID == "" {
					return fmt.Errorf("no mission context detected\nHint: Use --mission flag or run from a grove/mission directory")
				}
			}

			// Create grove via service (handles DB + directory + config)
			resp, err := wire.GroveService().CreateGrove(ctx, primary.CreateGroveRequest{
				Name:      name,
				MissionID: missionID,
				Repos:     repos,
				BasePath:  basePath,
			})
			if err != nil {
				return fmt.Errorf("failed to create grove: %w", err)
			}

			grove := resp.Grove
			fmt.Printf("✓ Created grove %s: %s\n", grove.ID, grove.Name)
			fmt.Printf("  Mission: %s\n", grove.MissionID)
			fmt.Printf("  Path: %s\n", grove.Path)
			if len(repos) > 0 {
				fmt.Printf("  Repos: %v\n", repos)
			}
			fmt.Println()
			fmt.Printf("  ✓ Created directory and .orc/config.json\n")

			// For each repo, try to create git worktree (not handled by service)
			if len(repos) > 0 {
				fmt.Println()
				fmt.Println("Creating git worktrees...")
				for _, repo := range repos {
					if err := createWorktree(repo, name, grove.Path); err != nil {
						fmt.Printf("  ⚠️  Warning: Could not create worktree for %s: %v\n", repo, err)
						fmt.Printf("     You may need to create it manually\n")
					} else {
						fmt.Printf("  ✓ Created worktree for %s\n", repo)
					}
				}
			}

			fmt.Println()
			fmt.Printf("Grove ready at: %s\n", grove.Path)
			fmt.Printf("Start working: cd %s\n", grove.Path)

			return nil
		},
	}

	cmd.Flags().StringVarP(&missionID, "mission", "m", "", "Mission ID (defaults to context)")
	cmd.Flags().StringSliceVarP(&repos, "repos", "r", nil, "Comma-separated list of repo names")
	cmd.Flags().StringVarP(&basePath, "path", "p", "", "Base path for worktrees (default: ~/src/worktrees)")

	return cmd
}

func groveListCmd() *cobra.Command {
	var missionID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all groves",
		Long:  `List all groves with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Use adapter for listing - now supports empty missionID (lists all)
			if missionID != "" {
				_, err := wire.GroveAdapter().List(ctx, missionID)
				return err
			}

			// List all groves via service (now supports empty missionID)
			groveService := wire.GroveService()
			serviceGroves, err := groveService.ListGroves(ctx, primary.GroveFilters{})
			if err != nil {
				return fmt.Errorf("failed to list groves: %w", err)
			}

			if len(serviceGroves) == 0 {
				fmt.Println("No groves found.")
				fmt.Println()
				fmt.Println("Create your first grove:")
				fmt.Println("  orc grove create my-grove --repos main-app --mission MISSION-001")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tMISSION\tSTATUS\tPATH")
			fmt.Fprintln(w, "--\t----\t-------\t------\t----")

			for _, g := range serviceGroves {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					g.ID,
					g.Name,
					g.MissionID,
					g.Status,
					g.Path,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&missionID, "mission", "m", "", "Filter by mission ID")

	return cmd
}

func groveShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [grove-id]",
		Short: "Show grove details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			_, err := wire.GroveAdapter().Show(ctx, args[0])
			return err
		},
	}
}

func groveRenameCmd() *cobra.Command {
	var updateMetadata bool

	cmd := &cobra.Command{
		Use:   "rename [grove-id] [new-name]",
		Short: "Rename a grove",
		Long: `Rename a grove in the database.

Examples:
  orc grove rename GROVE-001 tooling
  orc grove rename GROVE-001 backend-refactor --update-config`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			id := args[0]
			newName := args[1]

			// Rename via adapter (handles get + rename + output)
			_, err := wire.GroveAdapter().Rename(ctx, id, newName)
			if err != nil {
				return err
			}

			// Update .orc/config.json if requested (CLI-specific filesystem logic)
			if updateMetadata {
				// Reload grove with new name via service
				grove, err := wire.GroveService().GetGrove(ctx, id)
				if err != nil {
					fmt.Printf("  ⚠️  Warning: Could not reload grove for config update: %v\n", err)
					return nil
				}

				if err := writeGroveMetadata(grove); err != nil {
					fmt.Printf("  ⚠️  Warning: Could not update .orc/config.json: %v\n", err)
				} else {
					fmt.Printf("  ✓ Updated .orc/config.json\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&updateMetadata, "update-config", false, "Also update .orc/config.json file (optional)")

	return cmd
}

func groveDeleteCmd() *cobra.Command {
	var force bool
	var removeWorktree bool

	cmd := &cobra.Command{
		Use:   "delete [grove-id]",
		Short: "Delete a grove from the database",
		Long: `Delete a grove from the database and optionally remove its worktree.

WARNING: This is a destructive operation. By default, only the database record
is removed. Use --remove-worktree to also delete the git worktree.

Examples:
  orc grove delete GROVE-001
  orc grove delete GROVE-001 --remove-worktree
  orc grove delete GROVE-TEST-001 --force --remove-worktree`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			groveID := args[0]

			// Get grove path before deleting (for worktree removal)
			grove, err := wire.GroveAdapter().GetGrove(ctx, groveID)
			if err != nil {
				return err
			}
			grovePath := grove.Path

			// Remove worktree if requested (CLI-specific filesystem operation)
			if removeWorktree {
				if _, err := os.Stat(grovePath); err == nil {
					fmt.Printf("Removing worktree at: %s\n", grovePath)

					// Try to remove git worktree first
					if err := exec.Command("git", "worktree", "remove", grovePath, "--force").Run(); err != nil {
						fmt.Printf("  ⚠️  Warning: git worktree remove failed: %v\n", err)
						fmt.Printf("  Attempting direct directory removal...\n")

						// Fall back to direct directory removal
						if err := os.RemoveAll(grovePath); err != nil {
							return fmt.Errorf("failed to remove worktree directory: %w", err)
						}
					}
					fmt.Printf("  ✓ Worktree removed\n")
				} else {
					fmt.Printf("  ℹ️  Worktree not found at %s (already removed)\n", grovePath)
				}
			}

			// Delete from database via adapter
			_, err = wire.GroveAdapter().Delete(ctx, groveID, force)
			if err != nil {
				return err
			}

			if !removeWorktree {
				fmt.Printf("  ℹ️  Worktree still exists at: %s\n", grovePath)
				fmt.Printf("     Use --remove-worktree to delete it\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with errors")
	cmd.Flags().BoolVar(&removeWorktree, "remove-worktree", false, "Also remove the git worktree directory")

	return cmd
}

func groveOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open [grove-id]",
		Short: "Open a grove in a new TMux window",
		Long: `Open a grove by creating a new TMux window with IMP workspace layout.

Creates a new window in the current TMux session with:
- Pane 1 (left): vim
- Pane 2 (top right): claude (IMP)
- Pane 3 (bottom right): shell

All panes start in the grove's working directory.

Examples:
  orc grove open GROVE-001
  orc grove open GROVE-002`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groveID := args[0]

			// Get grove from service
			grove, err := wire.GroveService().GetGrove(context.Background(), groveID)
			if err != nil {
				return fmt.Errorf("failed to get grove: %w", err)
			}

			// Check if grove path exists
			if _, err := os.Stat(grove.Path); os.IsNotExist(err) {
				return fmt.Errorf("grove worktree not found at %s\nRun 'orc grove create %s --repos <repo-names>' to materialize", grove.Path, grove.Name)
			}

			// Check if in TMux session
			if os.Getenv("TMUX") == "" {
				return fmt.Errorf("not in a TMux session\nRun this command from within a TMux session")
			}

			// Get current TMux session name
			sessionNameBytes, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
			if err != nil {
				return fmt.Errorf("failed to detect TMux session name: %w", err)
			}
			sessionName := strings.TrimSpace(string(sessionNameBytes))

			// Create session object
			session := &tmux.Session{Name: sessionName}

			// Get next window index by listing current windows
			windowsOutput, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}").Output()
			if err != nil {
				return fmt.Errorf("failed to list windows: %w", err)
			}

			// Parse window indices to find the next available
			var maxIndex int
			lines := strings.Split(strings.TrimSpace(string(windowsOutput)), "\n")
			for _, line := range lines {
				if idx, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && idx > maxIndex {
					maxIndex = idx
				}
			}
			nextIndex := maxIndex + 1

			// Create grove window with IMP layout
			_, err = session.CreateGroveWindow(nextIndex, grove.Name, grove.Path)
			if err != nil {
				return fmt.Errorf("failed to create grove window: %w", err)
			}

			fmt.Printf("✓ Opened grove %s (%s)\n", grove.ID, grove.Name)
			fmt.Printf("  Window: %s:%s\n", sessionName, grove.Name)
			fmt.Printf("  Path: %s\n", grove.Path)
			fmt.Println()
			fmt.Printf("Layout:\n")
			fmt.Printf("  Pane 1 (left): vim\n")
			fmt.Printf("  Pane 2 (top right): claude (IMP)\n")
			fmt.Printf("  Pane 3 (bottom right): shell\n")
			fmt.Println()
			fmt.Printf("Switch to window: Ctrl+b then w (select from list)\n")

			return nil
		},
	}
}

// createWorktree attempts to create a git worktree for a repo
func createWorktree(repo, branch, targetPath string) error {
	// Determine repo path (assume repos are in ~/src/)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	repoPath := filepath.Join(home, "src", repo)

	// Check if repo exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repo not found at %s", repoPath)
	}

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", targetPath, "-b", branch)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}

// writeGroveMetadata writes .orc/config.json with type="grove"
func writeGroveMetadata(grove *primary.Grove) error {
	cfg := &config.Config{
		Version: "1.0",
		Type:    config.TypeGrove,
		Grove: &config.GroveConfig{
			GroveID:   grove.ID,
			MissionID: grove.MissionID,
			Name:      grove.Name,
			Repos:     grove.Repos,
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	}

	return config.SaveConfig(grove.Path, cfg)
}
