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

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// WorkbenchCmd returns the workbench command
func WorkbenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workbench",
		Short: "Manage workbenches (git worktrees)",
		Long:  `Create and manage workbenches - git worktrees within workshops.`,
	}

	cmd.AddCommand(workbenchCreateCmd())
	cmd.AddCommand(workbenchLikeCmd())
	cmd.AddCommand(workbenchListCmd())
	cmd.AddCommand(workbenchShowCmd())
	cmd.AddCommand(workbenchRenameCmd())
	cmd.AddCommand(workbenchOpenCmd())
	cmd.AddCommand(workbenchDeleteCmd())
	cmd.AddCommand(workbenchCheckoutCmd())
	cmd.AddCommand(workbenchStatusCmd())

	return cmd
}

func workbenchCreateCmd() *cobra.Command {
	var workshopID string
	var repos []string
	var basePath string
	var repoID string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new workbench (worktree) in a workshop",
		Long: `Create a new workbench with git worktree integration.

This command:
1. Creates a workbench record in the database
2. Creates git worktree(s) for specified repos
3. Writes .orc/config.json (workbench config)

Examples:
  orc workbench create auth-backend --workshop WORK-001 --repos main-app
  orc workbench create frontend --workshop WORK-001 --repos main-app
  orc workbench create multi --workshop WORK-002 --repos main-app,api-service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			if workshopID == "" {
				return fmt.Errorf("--workshop flag is required")
			}

			// Create workbench via service
			resp, err := wire.WorkbenchService().CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
				Name:       name,
				WorkshopID: workshopID,
				RepoID:     repoID,
				Repos:      repos,
				BasePath:   basePath,
			})
			if err != nil {
				return fmt.Errorf("failed to create workbench: %w", err)
			}

			workbench := resp.Workbench
			fmt.Printf("✓ Created workbench %s: %s\n", workbench.ID, workbench.Name)
			fmt.Printf("  Workshop: %s\n", workbench.WorkshopID)
			fmt.Printf("  Path: %s\n", workbench.Path)

			// Create worktrees for each repo
			if len(repos) > 0 {
				fmt.Println()
				fmt.Println("Creating git worktrees...")
				for _, repo := range repos {
					if err := createWorkbenchWorktree(repo, name, workbench.Path); err != nil {
						fmt.Printf("  Warning: Could not create worktree for %s: %v\n", repo, err)
						fmt.Printf("     You may need to create it manually\n")
					} else {
						fmt.Printf("  Created worktree for %s\n", repo)
					}
				}
			}

			fmt.Println()
			fmt.Printf("Workbench ready at: %s\n", workbench.Path)
			fmt.Printf("Start working: cd %s\n", workbench.Path)

			return nil
		},
	}

	cmd.Flags().StringVarP(&workshopID, "workshop", "w", "", "Workshop ID (required)")
	cmd.Flags().StringSliceVarP(&repos, "repos", "r", nil, "Comma-separated list of repo names")
	cmd.Flags().StringVarP(&basePath, "path", "p", "", "Base path for worktrees (default: ~/src/worktrees)")
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Link to a repo entity (optional)")

	return cmd
}

func workbenchLikeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like [name]",
		Short: "Create a new workbench based on the current one",
		Long: `Create a new workbench with the same workshop as the current workbench.

Detects the current workbench from the working directory and creates a sibling
workbench in the same workshop.

Examples:
  orc workbench like                    # Auto-generate name
  orc workbench like auth-refactor-v2   # Specify name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// Detect source workbench from path
			source, err := wire.WorkbenchService().GetWorkbenchByPath(ctx, cwd)
			if err != nil {
				return fmt.Errorf("not in a workbench directory: %w", err)
			}

			// Determine new name
			newName := ""
			if len(args) > 0 {
				newName = args[0]
			} else {
				newName = generateSiblingName(source.Name)
			}

			// Create new workbench with same workshop
			resp, err := wire.WorkbenchService().CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
				Name:       newName,
				WorkshopID: source.WorkshopID,
				RepoID:     source.RepoID,
			})
			if err != nil {
				return fmt.Errorf("failed to create workbench: %w", err)
			}

			fmt.Printf("✓ Created workbench %s: %s\n", resp.WorkbenchID, resp.Workbench.Name)
			fmt.Printf("  Based on: %s (%s)\n", source.ID, source.Name)
			fmt.Printf("  Workshop: %s\n", resp.Workbench.WorkshopID)
			fmt.Printf("  Path: %s\n", resp.Workbench.Path)

			// Open in new tmux window if in tmux
			if os.Getenv("TMUX") != "" {
				// Get current TMux session name
				sessionNameBytes, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
				if err == nil {
					sessionName := strings.TrimSpace(string(sessionNameBytes))

					// Get next window index
					windowsOutput, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}").Output()
					if err == nil {
						var maxIndex int
						lines := strings.Split(strings.TrimSpace(string(windowsOutput)), "\n")
						for _, line := range lines {
							if idx, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && idx > maxIndex {
								maxIndex = idx
							}
						}
						nextIndex := maxIndex + 1

						// Create workbench window
						tmuxAdapter := wire.TMuxAdapter()
						if err := tmuxAdapter.CreateWorkbenchWindow(ctx, sessionName, nextIndex, resp.Workbench.Name, resp.Workbench.Path); err == nil {
							fmt.Printf("\n  Opened in tmux window: %s:%s\n", sessionName, resp.Workbench.Name)
						}
					}
				}
			}

			return nil
		},
	}

	return cmd
}

// generateSiblingName creates a sibling name like "auth-2", "auth-3"
func generateSiblingName(baseName string) string {
	// Strip trailing -N suffix if present
	base := baseName
	suffix := 2
	if idx := strings.LastIndex(baseName, "-"); idx > 0 {
		if n, err := strconv.Atoi(baseName[idx+1:]); err == nil {
			base = baseName[:idx]
			suffix = n + 1
		}
	}
	return fmt.Sprintf("%s-%d", base, suffix)
}

func workbenchListCmd() *cobra.Command {
	var workshopID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workbenches",
		Long:  `List all workbenches with their current status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workbenches, err := wire.WorkbenchService().ListWorkbenches(ctx, primary.WorkbenchFilters{
				WorkshopID: workshopID,
			})
			if err != nil {
				return fmt.Errorf("failed to list workbenches: %w", err)
			}

			if len(workbenches) == 0 {
				fmt.Println("No workbenches found.")
				fmt.Println()
				fmt.Println("Create your first workbench:")
				fmt.Println("  orc workbench create my-workbench --workshop WORK-001 --repos main-app")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tWORKSHOP\tSTATUS\tPATH")
			fmt.Fprintln(w, "--\t----\t--------\t------\t----")

			for _, wb := range workbenches {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					wb.ID,
					wb.Name,
					wb.WorkshopID,
					wb.Status,
					wb.Path,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&workshopID, "workshop", "w", "", "Filter by workshop ID")

	return cmd
}

func workbenchShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [workbench-id]",
		Short: "Show workbench details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			workbench, err := wire.WorkbenchService().GetWorkbench(ctx, args[0])
			if err != nil {
				return fmt.Errorf("workbench not found: %w", err)
			}

			fmt.Printf("Workbench: %s\n", workbench.ID)
			fmt.Printf("Name: %s\n", workbench.Name)
			fmt.Printf("Workshop: %s\n", workbench.WorkshopID)
			fmt.Printf("Path: %s\n", workbench.Path)
			fmt.Printf("Status: %s\n", workbench.Status)
			if workbench.HomeBranch != "" {
				fmt.Printf("Home Branch: %s\n", workbench.HomeBranch)
			}
			if workbench.CurrentBranch != "" {
				fmt.Printf("Current Branch: %s\n", workbench.CurrentBranch)
			}
			fmt.Printf("Created: %s\n", workbench.CreatedAt)

			return nil
		},
	}
}

func workbenchRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename [workbench-id] [new-name]",
		Short: "Rename a workbench",
		Long: `Rename a workbench in the database.

Examples:
  orc workbench rename BENCH-001 tooling
  orc workbench rename BENCH-001 backend-refactor`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			id := args[0]
			newName := args[1]

			err := wire.WorkbenchService().RenameWorkbench(ctx, primary.RenameWorkbenchRequest{
				WorkbenchID: id,
				NewName:     newName,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Renamed workbench %s to %s\n", id, newName)
			return nil
		},
	}

	return cmd
}

func workbenchDeleteCmd() *cobra.Command {
	var force bool
	var removeWorktree bool

	cmd := &cobra.Command{
		Use:   "delete [workbench-id]",
		Short: "Delete a workbench from the database",
		Long: `Delete a workbench from the database and optionally remove its worktree.

WARNING: This is a destructive operation. By default, only the database record
is removed. Use --remove-worktree to also delete the git worktree.

Examples:
  orc workbench delete BENCH-001
  orc workbench delete BENCH-001 --remove-worktree
  orc workbench delete BENCH-001 --force --remove-worktree`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workbenchID := args[0]

			// Get workbench path before deleting
			workbench, err := wire.WorkbenchService().GetWorkbench(ctx, workbenchID)
			if err != nil {
				return err
			}
			workbenchPath := workbench.Path

			// Remove worktree if requested
			if removeWorktree {
				if _, err := os.Stat(workbenchPath); err == nil {
					fmt.Printf("Removing worktree at: %s\n", workbenchPath)

					// Try to remove git worktree first
					if err := exec.Command("git", "worktree", "remove", workbenchPath, "--force").Run(); err != nil {
						fmt.Printf("  Warning: git worktree remove failed: %v\n", err)
						fmt.Printf("  Attempting direct directory removal...\n")

						// Fall back to direct directory removal
						if err := os.RemoveAll(workbenchPath); err != nil {
							return fmt.Errorf("failed to remove worktree directory: %w", err)
						}
					}
					fmt.Printf("  Worktree removed\n")
				} else {
					fmt.Printf("  Worktree not found at %s (already removed)\n", workbenchPath)
				}
			}

			// Delete from database
			err = wire.WorkbenchService().DeleteWorkbench(ctx, primary.DeleteWorkbenchRequest{
				WorkbenchID:    workbenchID,
				Force:          force,
				RemoveWorktree: removeWorktree,
			})
			if err != nil {
				return err
			}

			fmt.Printf("✓ Workbench %s deleted\n", workbenchID)

			if !removeWorktree {
				fmt.Printf("  Worktree still exists at: %s\n", workbenchPath)
				fmt.Printf("     Use --remove-worktree to delete it\n")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with errors")
	cmd.Flags().BoolVar(&removeWorktree, "remove-worktree", false, "Also remove the git worktree directory")

	return cmd
}

func workbenchOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open [workbench-id]",
		Short: "Open a workbench in a new TMux window",
		Long: `Open a workbench by creating a new TMux window with IMP workspace layout.

Creates a new window in the current TMux session with:
- Pane 1 (left): vim
- Pane 2 (top right): claude (IMP)
- Pane 3 (bottom right): shell

All panes start in the workbench's working directory.

Examples:
  orc workbench open BENCH-001
  orc workbench open BENCH-002`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workbenchID := args[0]

			// Get workbench from service
			workbench, err := wire.WorkbenchService().GetWorkbench(context.Background(), workbenchID)
			if err != nil {
				return fmt.Errorf("failed to get workbench: %w", err)
			}

			// Check if workbench path exists
			if _, err := os.Stat(workbench.Path); os.IsNotExist(err) {
				return fmt.Errorf("workbench worktree not found at %s\nRun 'orc workbench create %s --repos <repo-names>' to materialize", workbench.Path, workbench.Name)
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

			// Create workbench window with IMP layout via adapter
			ctx := context.Background()
			tmuxAdapter := wire.TMuxAdapter()
			err = tmuxAdapter.CreateWorkbenchWindow(ctx, sessionName, nextIndex, workbench.Name, workbench.Path)
			if err != nil {
				return fmt.Errorf("failed to create workbench window: %w", err)
			}

			fmt.Printf("Opened workbench %s (%s)\n", workbench.ID, workbench.Name)
			fmt.Printf("  Window: %s:%s\n", sessionName, workbench.Name)
			fmt.Printf("  Path: %s\n", workbench.Path)
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

func workbenchCheckoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout [workbench-id] [branch]",
		Short: "Switch to a target branch using stash dance",
		Long: `Switch to a target branch using the stash dance workflow.

The stash dance safely handles uncommitted changes:
1. Stash any uncommitted changes
2. Checkout the target branch
3. Pop the stash (reapply changes)

This allows seamless context switching even with dirty working directories.

Examples:
  orc workbench checkout BENCH-001 main
  orc workbench checkout BENCH-001 ml/SHIP-205-feature`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workbenchID := args[0]
			targetBranch := args[1]

			resp, err := wire.WorkbenchService().CheckoutBranch(ctx, primary.CheckoutBranchRequest{
				WorkbenchID:  workbenchID,
				TargetBranch: targetBranch,
			})
			if err != nil {
				return fmt.Errorf("checkout failed: %w", err)
			}

			fmt.Printf("Switched from %s to %s\n", resp.PreviousBranch, resp.CurrentBranch)
			if resp.StashApplied {
				fmt.Println("  (stashed changes have been reapplied)")
			}

			return nil
		},
	}

	return cmd
}

func workbenchStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [workbench-id]",
		Short: "Show git status of a workbench",
		Long: `Show the current git status of a workbench including:
- Current branch
- Home branch
- Dirty state (uncommitted changes)
- Ahead/behind remote

Examples:
  orc workbench status BENCH-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			workbenchID := args[0]

			status, err := wire.WorkbenchService().GetWorkbenchStatus(ctx, workbenchID)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			fmt.Printf("Workbench: %s\n", status.WorkbenchID)
			fmt.Printf("Current Branch: %s\n", status.CurrentBranch)
			fmt.Printf("Home Branch: %s\n", status.HomeBranch)

			if status.IsDirty {
				fmt.Printf("Status: dirty (%d modified files)\n", status.DirtyFiles)
			} else {
				fmt.Println("Status: clean")
			}

			if status.AheadBy > 0 || status.BehindBy > 0 {
				fmt.Printf("Remote: %d ahead, %d behind\n", status.AheadBy, status.BehindBy)
			}

			return nil
		},
	}

	return cmd
}

// createWorkbenchWorktree attempts to create a git worktree for a repo
func createWorkbenchWorktree(repo, branch, targetPath string) error {
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
	execCmd := exec.Command("git", "worktree", "add", targetPath, "-b", branch)
	execCmd.Dir = repoPath

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	return nil
}
