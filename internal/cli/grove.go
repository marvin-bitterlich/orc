package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"

	"github.com/looneym/orc/internal/models"
	"github.com/spf13/cobra"
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
3. Writes metadata.json for reference

Examples:
  orc grove create auth-backend --repos intercom --mission MISSION-001
  orc grove create frontend --repos intercom --mission MISSION-001
  orc grove create multi --repos intercom,api-service --mission MISSION-002`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Default mission to MISSION-001
			if missionID == "" {
				missionID = "MISSION-001"
			}

			// Default base path
			if basePath == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("failed to get home directory: %w", err)
				}
				basePath = filepath.Join(home, "src", "worktrees")
			}

			// Full path for this grove
			grovePath := filepath.Join(basePath, name)

			// Create grove in database
			grove, err := models.CreateGrove(missionID, name, grovePath, repos)
			if err != nil {
				return fmt.Errorf("failed to create grove in database: %w", err)
			}

			fmt.Printf("✓ Created grove %s: %s\n", grove.ID, grove.Name)
			fmt.Printf("  Mission: %s\n", grove.MissionID)
			fmt.Printf("  Path: %s\n", grove.Path)
			if len(repos) > 0 {
				fmt.Printf("  Repos: %v\n", repos)
			}
			fmt.Println()

			// Create worktree directory if it doesn't exist
			if err := os.MkdirAll(grovePath, 0755); err != nil {
				return fmt.Errorf("failed to create grove directory: %w", err)
			}

			// For each repo, try to create git worktree
			if len(repos) > 0 {
				fmt.Println("Creating git worktrees...")
				for _, repo := range repos {
					if err := createWorktree(repo, name, grovePath); err != nil {
						fmt.Printf("  ⚠️  Warning: Could not create worktree for %s: %v\n", repo, err)
						fmt.Printf("     You may need to create it manually\n")
					} else {
						fmt.Printf("  ✓ Created worktree for %s\n", repo)
					}
				}
				fmt.Println()
			}

			// Write metadata.json (reference only)
			if err := writeGroveMetadata(grove); err != nil {
				fmt.Printf("  ⚠️  Warning: Could not write metadata.json: %v\n", err)
			} else {
				fmt.Printf("  ✓ Wrote metadata.json\n")
			}

			fmt.Println()
			fmt.Printf("Grove ready at: %s\n", grovePath)
			fmt.Printf("Start working: cd %s\n", grovePath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&missionID, "mission", "m", "", "Mission ID (defaults to MISSION-001)")
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
			groves, err := models.ListGroves(missionID)
			if err != nil {
				return fmt.Errorf("failed to list groves: %w", err)
			}

			if len(groves) == 0 {
				fmt.Println("No groves found.")
				fmt.Println()
				fmt.Println("Create your first grove:")
				fmt.Println("  orc grove create my-grove --repos intercom --mission MISSION-001")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tMISSION\tSTATUS\tPATH")
			fmt.Fprintln(w, "--\t----\t-------\t------\t----")

			for _, grove := range groves {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					grove.ID,
					grove.Name,
					grove.MissionID,
					grove.Status,
					grove.Path,
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
			id := args[0]

			grove, err := models.GetGrove(id)
			if err != nil {
				return fmt.Errorf("failed to get grove: %w", err)
			}

			fmt.Printf("\nGrove: %s\n", grove.ID)
			fmt.Printf("Name:    %s\n", grove.Name)
			fmt.Printf("Mission: %s\n", grove.MissionID)
			fmt.Printf("Path:    %s\n", grove.Path)
			fmt.Printf("Status:  %s\n", grove.Status)
			if grove.Repos.Valid {
				fmt.Printf("Repos:   %s\n", grove.Repos.String)
			}
			fmt.Printf("Created: %s\n", grove.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Println()

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

// writeGroveMetadata writes metadata.json for reference (database is source of truth)
func writeGroveMetadata(grove *models.Grove) error {
	metadata := map[string]interface{}{
		"grove_id":   grove.ID,
		"mission_id": grove.MissionID,
		"name":       grove.Name,
		"repos":      []string{},
		"created_at": grove.CreatedAt,
	}

	// Parse repos JSON if present
	if grove.Repos.Valid {
		repos := []string{}
		if err := json.Unmarshal([]byte(grove.Repos.String), &repos); err == nil {
			metadata["repos"] = repos
		}
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(grove.Path, "metadata.json")
	return os.WriteFile(metadataPath, data, 0644)
}
