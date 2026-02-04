package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// RepoCmd returns the repo command
func RepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repository configurations",
		Long:  `Create and manage repository configurations for ORC.`,
	}

	cmd.AddCommand(repoCreateCmd())
	cmd.AddCommand(repoListCmd())
	cmd.AddCommand(repoShowCmd())
	cmd.AddCommand(repoUpdateCmd())
	cmd.AddCommand(repoArchiveCmd())
	cmd.AddCommand(repoRestoreCmd())
	cmd.AddCommand(repoDeleteCmd())

	return cmd
}

func repoCreateCmd() *cobra.Command {
	var url, localPath, defaultBranch string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new repository configuration",
		Long: `Create a new repository configuration.

Examples:
  orc repo create orc --url git@github.com:org/orc.git
  orc repo create intercom --url git@github.com:org/intercom.git --path ~/src/intercom
  orc repo create api --default-branch develop`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			name := args[0]

			resp, err := wire.RepoService().CreateRepo(ctx, primary.CreateRepoRequest{
				Name:          name,
				URL:           url,
				LocalPath:     localPath,
				DefaultBranch: defaultBranch,
			})
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}

			fmt.Printf("✓ Created repository %s: %s\n", resp.RepoID, resp.Repo.Name)
			if resp.Repo.URL != "" {
				fmt.Printf("  URL: %s\n", resp.Repo.URL)
			}
			if resp.Repo.LocalPath != "" {
				fmt.Printf("  Path: %s\n", resp.Repo.LocalPath)
			}
			fmt.Printf("  Default Branch: %s\n", resp.Repo.DefaultBranch)

			return nil
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "Repository URL (e.g., git@github.com:org/repo.git)")
	cmd.Flags().StringVarP(&localPath, "path", "p", "", "Local path to repository")
	cmd.Flags().StringVarP(&defaultBranch, "default-branch", "b", "main", "Default branch name")

	return cmd
}

func repoListCmd() *cobra.Command {
	var status string
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all repositories",
		Long:  `List all repository configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			filters := primary.RepoFilters{}
			if !all && status == "" {
				filters.Status = "active"
			} else if status != "" {
				filters.Status = status
			}

			repos, err := wire.RepoService().ListRepos(ctx, filters)
			if err != nil {
				return fmt.Errorf("failed to list repositories: %w", err)
			}

			if len(repos) == 0 {
				fmt.Println("No repositories found.")
				fmt.Println()
				fmt.Println("Create your first repository:")
				fmt.Println("  orc repo create my-repo --url git@github.com:org/my-repo.git")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tURL\tBRANCH\tSTATUS")
			fmt.Fprintln(w, "--\t----\t---\t------\t------")

			for _, r := range repos {
				url := r.URL
				if len(url) > 40 {
					url = url[:37] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					r.ID,
					r.Name,
					url,
					r.DefaultBranch,
					r.Status,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (active, archived)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all repositories including archived")

	return cmd
}

func repoShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [repo-id]",
		Short: "Show repository details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			repoID := args[0]

			repo, err := wire.RepoService().GetRepo(ctx, repoID)
			if err != nil {
				return fmt.Errorf("failed to get repository: %w", err)
			}

			fmt.Printf("Repository: %s\n", repo.ID)
			fmt.Printf("  Name: %s\n", repo.Name)
			fmt.Printf("  Status: %s\n", repo.Status)
			if repo.URL != "" {
				fmt.Printf("  URL: %s\n", repo.URL)
			}
			if repo.LocalPath != "" {
				fmt.Printf("  Local Path: %s\n", repo.LocalPath)
			}
			fmt.Printf("  Default Branch: %s\n", repo.DefaultBranch)
			fmt.Printf("  Created: %s\n", repo.CreatedAt)
			fmt.Printf("  Updated: %s\n", repo.UpdatedAt)

			return nil
		},
	}
}

func repoUpdateCmd() *cobra.Command {
	var url, localPath, defaultBranch string

	cmd := &cobra.Command{
		Use:   "update [repo-id]",
		Short: "Update repository configuration",
		Long: `Update repository configuration.

Examples:
  orc repo update REPO-001 --url git@github.com:new/url.git
  orc repo update REPO-001 --path ~/src/new-path
  orc repo update REPO-001 --default-branch develop`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			repoID := args[0]

			err := wire.RepoService().UpdateRepo(ctx, primary.UpdateRepoRequest{
				RepoID:        repoID,
				URL:           url,
				LocalPath:     localPath,
				DefaultBranch: defaultBranch,
			})
			if err != nil {
				return fmt.Errorf("failed to update repository: %w", err)
			}

			fmt.Printf("✓ Updated repository %s\n", repoID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&url, "url", "u", "", "New repository URL")
	cmd.Flags().StringVarP(&localPath, "path", "p", "", "New local path")
	cmd.Flags().StringVarP(&defaultBranch, "default-branch", "b", "", "New default branch")

	return cmd
}

func repoArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive [repo-id]",
		Short: "Archive a repository",
		Long:  `Archive a repository (soft delete). Archived repositories cannot be used for new PRs.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			repoID := args[0]

			err := wire.RepoService().ArchiveRepo(ctx, repoID)
			if err != nil {
				return fmt.Errorf("failed to archive repository: %w", err)
			}

			fmt.Printf("✓ Archived repository %s\n", repoID)

			return nil
		},
	}
}

func repoRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [repo-id]",
		Short: "Restore an archived repository",
		Long:  `Restore an archived repository to active status.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			repoID := args[0]

			err := wire.RepoService().RestoreRepo(ctx, repoID)
			if err != nil {
				return fmt.Errorf("failed to restore repository: %w", err)
			}

			fmt.Printf("✓ Restored repository %s\n", repoID)

			return nil
		},
	}
}

func repoDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [repo-id]",
		Short: "Delete a repository",
		Long: `Delete a repository from the database.

WARNING: This is a destructive operation. Repositories with active PRs cannot be deleted.

Examples:
  orc repo delete REPO-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			repoID := args[0]

			// Get repo details before deleting
			repo, err := wire.RepoService().GetRepo(ctx, repoID)
			if err != nil {
				return fmt.Errorf("failed to get repository: %w", err)
			}

			err = wire.RepoService().DeleteRepo(ctx, repoID)
			if err != nil {
				return fmt.Errorf("failed to delete repository: %w", err)
			}

			fmt.Printf("✓ Deleted repository %s (%s)\n", repoID, repo.Name)
			_ = force // Reserved for future use

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete (reserved for future use)")

	return cmd
}
