package cli

import (
	"fmt"
	"os"
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
	cmd.AddCommand(workbenchDeleteCmd())
	cmd.AddCommand(workbenchArchiveCmd())
	cmd.AddCommand(workbenchCheckoutCmd())
	cmd.AddCommand(workbenchStatusCmd())

	return cmd
}

func workbenchCreateCmd() *cobra.Command {
	var workshopID string
	var repoID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workbench record in a workshop",
		Long: `Create a new workbench record in the database.

The workbench name is auto-generated as {repo}-{number} based on the
linked repo. The workbench will be located at ~/wb/<name>.

To create the physical infrastructure (git worktrees, config files), run:
  orc infra apply WORK-xxx

Examples:
  orc workbench create --workshop WORK-001 --repo-id REPO-001`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			if workshopID == "" {
				return fmt.Errorf("--workshop flag is required")
			}
			if repoID == "" {
				return fmt.Errorf("--repo-id flag is required")
			}

			// Create workbench via service (DB only, no filesystem operations)
			// Name is auto-generated from repo name + bench number
			resp, err := wire.WorkbenchService().CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
				Name:            "", // Auto-generated
				WorkshopID:      workshopID,
				RepoID:          repoID,
				SkipConfigWrite: true, // Infrastructure handled by orc infra apply
			})
			if err != nil {
				return fmt.Errorf("failed to create workbench: %w", err)
			}

			workbench := resp.Workbench
			fmt.Printf("✓ Created workbench %s: %s\n", workbench.ID, workbench.Name)
			fmt.Printf("  Workshop: %s\n", workbench.WorkshopID)
			fmt.Printf("  Path: %s\n", workbench.Path)
			fmt.Println()
			fmt.Println("To create the physical infrastructure, run:")
			fmt.Printf("  orc infra apply %s\n", workshopID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&workshopID, "workshop", "w", "", "Workshop ID (required)")
	cmd.Flags().StringVar(&repoID, "repo-id", "", "Repo ID for name generation (required)")
	_ = cmd.MarkFlagRequired("workshop")
	_ = cmd.MarkFlagRequired("repo-id")

	return cmd
}

func workbenchLikeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "like [name]",
		Short: "Create a new workbench based on the current one",
		Long: `Create a new workbench with the same workshop as the current workbench (DB record only).

Detects the current workbench from the working directory and creates a sibling
workbench in the same workshop. This creates the database record only.

To create the git worktree and config files, run:
  orc infra apply <workshop-id>

Examples:
  orc workbench like                    # Auto-generate name
  orc workbench like auth-refactor-v2   # Specify name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

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

			// Create new workbench with same workshop (DB only, worktree via orc infra apply)
			resp, err := wire.WorkbenchService().CreateWorkbench(ctx, primary.CreateWorkbenchRequest{
				Name:            newName,
				WorkshopID:      source.WorkshopID,
				RepoID:          source.RepoID,
				SkipConfigWrite: true,
			})
			if err != nil {
				return fmt.Errorf("failed to create workbench: %w", err)
			}

			fmt.Printf("✓ Created workbench %s: %s\n", resp.WorkbenchID, resp.Workbench.Name)
			fmt.Printf("  Based on: %s (%s)\n", source.ID, source.Name)
			fmt.Printf("  Workshop: %s\n", resp.Workbench.WorkshopID)
			fmt.Printf("\nTo create the git worktree, run:\n")
			fmt.Printf("  orc infra apply %s\n", resp.Workbench.WorkshopID)

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
			ctx := NewContext()

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
			ctx := NewContext()

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
			ctx := NewContext()
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
	cmd := &cobra.Command{
		Use:   "delete [workbench-id]",
		Short: "Delete a workbench (DEPRECATED)",
		Long: `DEPRECATED: Use archive + infra apply instead.

To remove a workbench:
  orc workbench archive BENCH-xxx
  orc infra apply WORK-xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("orc workbench delete is deprecated. Use:\n  orc workbench archive %s\n  orc infra apply <workshop-id>", args[0])
		},
	}

	return cmd
}

func workbenchArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive [workbench-id]",
		Short: "Archive a workbench (soft-delete)",
		Long: `Archive a workbench by setting its status to 'archived'.

This is a soft-delete that keeps the record in the database so
infrastructure planning can detect it as a deletion target.

To physically remove the worktree after archiving:
  orc infra apply WORK-xxx

Examples:
  orc workbench archive BENCH-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			workbenchID := args[0]

			// Get workbench info for display
			workbench, err := wire.WorkbenchService().GetWorkbench(ctx, workbenchID)
			if err != nil {
				return err
			}

			if err := wire.WorkbenchService().ArchiveWorkbench(ctx, workbenchID); err != nil {
				return err
			}

			fmt.Printf("✓ Workbench %s archived\n", workbenchID)
			fmt.Printf("  Name: %s\n", workbench.Name)
			fmt.Printf("  Path: %s\n", workbench.Path)
			fmt.Printf("\nTo remove the worktree, run:\n")
			fmt.Printf("  orc infra apply %s\n", workbench.WorkshopID)

			return nil
		},
	}

	return cmd
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
			ctx := NewContext()
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
			ctx := NewContext()
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
