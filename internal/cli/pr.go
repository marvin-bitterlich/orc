package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	orccontext "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// PRCmd returns the pr command
func PRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Long:  `Create and manage pull requests linked to shipments.`,
	}

	cmd.AddCommand(prCreateCmd())
	cmd.AddCommand(prListCmd())
	cmd.AddCommand(prShowCmd())
	cmd.AddCommand(prOpenCmd())
	cmd.AddCommand(prApproveCmd())
	cmd.AddCommand(prMergeCmd())
	cmd.AddCommand(prCloseCmd())
	cmd.AddCommand(prLinkCmd())

	return cmd
}

func prCreateCmd() *cobra.Command {
	var repoID, branch, targetBranch, description, url string
	var number int
	var draft bool

	cmd := &cobra.Command{
		Use:   "create [shipment-id] [title]",
		Short: "Create a new pull request for a shipment",
		Long: `Create a new pull request linked to a shipment.

Examples:
  orc pr create SHIP-001 "Add authentication feature" --repo REPO-001 --branch feature/auth
  orc pr create SHIP-001 "Fix bug" --repo REPO-001 --branch fix/bug --draft
  orc pr create SHIP-001 "Update docs" --repo REPO-001 --branch docs/update --target main`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			shipmentID := args[0]
			title := args[1]

			resp, err := wire.PRService().CreatePR(ctx, primary.CreatePRRequest{
				ShipmentID:   shipmentID,
				RepoID:       repoID,
				Title:        title,
				Description:  description,
				Branch:       branch,
				TargetBranch: targetBranch,
				Draft:        draft,
				URL:          url,
				Number:       number,
			})
			if err != nil {
				return fmt.Errorf("failed to create PR: %w", err)
			}

			status := "open"
			if draft {
				status = "draft"
			}

			fmt.Printf("✓ Created PR %s: %s\n", resp.PRID, title)
			fmt.Printf("  Shipment: %s\n", shipmentID)
			fmt.Printf("  Repository: %s\n", repoID)
			fmt.Printf("  Branch: %s\n", branch)
			if targetBranch != "" {
				fmt.Printf("  Target: %s\n", targetBranch)
			}
			fmt.Printf("  Status: %s\n", status)
			if url != "" {
				fmt.Printf("  URL: %s\n", url)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&repoID, "repo", "r", "", "Repository ID (required)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch name (required)")
	cmd.Flags().StringVarP(&targetBranch, "target", "t", "", "Target branch (default: repo default)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "PR description")
	cmd.Flags().StringVarP(&url, "url", "u", "", "External PR URL (for linking)")
	cmd.Flags().IntVarP(&number, "number", "n", 0, "GitHub PR number")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft PR")
	cmd.MarkFlagRequired("repo")
	cmd.MarkFlagRequired("branch")

	return cmd
}

func prListCmd() *cobra.Command {
	var shipmentID, repoID, commissionID, status string
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		Long:  `List pull requests with optional filters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()

			// Try to get commission from context if not specified
			if commissionID == "" {
				commissionID = orccontext.GetContextCommissionID()
			}

			filters := primary.PRFilters{
				ShipmentID:   shipmentID,
				RepoID:       repoID,
				CommissionID: commissionID,
			}

			// Default to non-terminal statuses unless --all is specified
			if !all && status == "" {
				// Show all non-terminal PRs
			} else if status != "" {
				filters.Status = status
			}

			prs, err := wire.PRService().ListPRs(ctx, filters)
			if err != nil {
				return fmt.Errorf("failed to list PRs: %w", err)
			}

			if len(prs) == 0 {
				fmt.Println("No pull requests found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tSHIPMENT\tTITLE\tBRANCH\tSTATUS")
			fmt.Fprintln(w, "--\t--------\t-----\t------\t------")

			for _, p := range prs {
				title := p.Title
				if len(title) > 30 {
					title = title[:27] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					p.ID,
					p.ShipmentID,
					title,
					p.Branch,
					p.Status,
				)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&shipmentID, "shipment", "s", "", "Filter by shipment ID")
	cmd.Flags().StringVarP(&repoID, "repo", "r", "", "Filter by repository ID")
	cmd.Flags().StringVarP(&commissionID, "commission", "c", "", "Filter by commission ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (draft, open, approved, merged, closed)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all PRs including merged/closed")

	return cmd
}

func prShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [pr-id]",
		Short: "Show PR details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			prID := args[0]

			pr, err := wire.PRService().GetPR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to get PR: %w", err)
			}

			fmt.Printf("Pull Request: %s\n", pr.ID)
			fmt.Printf("  Title: %s\n", pr.Title)
			fmt.Printf("  Status: %s\n", pr.Status)
			fmt.Printf("  Shipment: %s\n", pr.ShipmentID)
			fmt.Printf("  Repository: %s\n", pr.RepoID)
			fmt.Printf("  Commission: %s\n", pr.CommissionID)
			fmt.Printf("  Branch: %s\n", pr.Branch)
			if pr.TargetBranch != "" {
				fmt.Printf("  Target: %s\n", pr.TargetBranch)
			}
			if pr.URL != "" {
				fmt.Printf("  URL: %s\n", pr.URL)
			}
			if pr.Number > 0 {
				fmt.Printf("  Number: #%d\n", pr.Number)
			}
			if pr.Description != "" {
				fmt.Printf("  Description: %s\n", pr.Description)
			}
			fmt.Printf("  Created: %s\n", pr.CreatedAt)
			fmt.Printf("  Updated: %s\n", pr.UpdatedAt)
			if pr.MergedAt != "" {
				fmt.Printf("  Merged: %s\n", pr.MergedAt)
			}
			if pr.ClosedAt != "" {
				fmt.Printf("  Closed: %s\n", pr.ClosedAt)
			}

			return nil
		},
	}
}

func prOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open [pr-id]",
		Short: "Open a draft PR for review",
		Long:  `Open a draft PR, making it ready for review.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			prID := args[0]

			err := wire.PRService().OpenPR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to open PR: %w", err)
			}

			fmt.Printf("✓ Opened PR %s (now ready for review)\n", prID)

			return nil
		},
	}
}

func prApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve [pr-id]",
		Short: "Mark a PR as approved",
		Long:  `Mark a PR as approved, indicating it's ready to merge.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			prID := args[0]

			err := wire.PRService().ApprovePR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to approve PR: %w", err)
			}

			fmt.Printf("✓ Approved PR %s (ready to merge)\n", prID)

			return nil
		},
	}
}

func prMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge [pr-id]",
		Short: "Merge a PR",
		Long: `Merge a PR and complete its associated shipment.

This command:
1. Updates the PR status to 'merged'
2. Automatically completes the associated shipment

Examples:
  orc pr merge PR-001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			prID := args[0]

			// Get PR to show shipment info
			pr, err := wire.PRService().GetPR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to get PR: %w", err)
			}

			err = wire.PRService().MergePR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to merge PR: %w", err)
			}

			fmt.Printf("✓ Merged PR %s\n", prID)
			fmt.Printf("  ✓ Completed shipment %s\n", pr.ShipmentID)

			return nil
		},
	}
}

func prCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close [pr-id]",
		Short: "Close a PR without merging",
		Long:  `Close a PR without merging. The associated shipment will NOT be completed.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			prID := args[0]

			err := wire.PRService().ClosePR(ctx, prID)
			if err != nil {
				return fmt.Errorf("failed to close PR: %w", err)
			}

			fmt.Printf("✓ Closed PR %s (shipment NOT completed)\n", prID)

			return nil
		},
	}
}

func prLinkCmd() *cobra.Command {
	var number int

	cmd := &cobra.Command{
		Use:   "link [shipment-id] [url]",
		Short: "Link an existing external PR to a shipment",
		Long: `Link an existing GitHub PR to a shipment.

Examples:
  orc pr link SHIP-001 https://github.com/org/repo/pull/123 --number 123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContext()
			shipmentID := args[0]
			url := args[1]

			pr, err := wire.PRService().LinkPR(ctx, shipmentID, url, number)
			if err != nil {
				return fmt.Errorf("failed to link PR: %w", err)
			}

			fmt.Printf("✓ Linked PR %s to shipment %s\n", pr.ID, shipmentID)
			fmt.Printf("  URL: %s\n", url)
			if number > 0 {
				fmt.Printf("  Number: #%d\n", number)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&number, "number", "n", 0, "GitHub PR number")

	return cmd
}
