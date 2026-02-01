package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var patrolCmd = &cobra.Command{
	Use:   "patrol",
	Short: "Manage patrols",
	Long:  "Start, end, and monitor patrols (Watchdog monitoring sessions) in the ORC ledger",
}

var patrolStartCmd = &cobra.Command{
	Use:   "start [workbench-id]",
	Short: "Start a new patrol for a workbench",
	Long: `Start a new patrol (monitoring session) for the specified workbench.

The patrol will monitor the TMux pane associated with the workbench's kennel.
Only one patrol can be active per kennel at a time.

Examples:
  orc patrol start BENCH-001
  orc patrol start BENCH-014`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		workbenchID := args[0]

		patrol, err := wire.PatrolService().StartPatrol(ctx, workbenchID)
		if err != nil {
			return fmt.Errorf("failed to start patrol: %w", err)
		}

		fmt.Printf("✓ Started patrol %s\n", patrol.ID)
		fmt.Printf("  Kennel: %s\n", patrol.KennelID)
		fmt.Printf("  Target: %s\n", patrol.Target)
		fmt.Printf("  Status: %s\n", patrol.Status)

		return nil
	},
}

var patrolEndCmd = &cobra.Command{
	Use:   "end [patrol-id]",
	Short: "End an active patrol",
	Long: `End an active patrol (monitoring session).

The patrol's status will be set to 'completed' and the ended_at timestamp recorded.

Examples:
  orc patrol end PATROL-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		patrolID := args[0]

		if err := wire.PatrolService().EndPatrol(ctx, patrolID); err != nil {
			return fmt.Errorf("failed to end patrol: %w", err)
		}

		fmt.Printf("✓ Patrol %s ended\n", patrolID)
		return nil
	},
}

var patrolStatusCmd = &cobra.Command{
	Use:   "status [patrol-id]",
	Short: "Show patrol status and stats",
	Long: `Show detailed status for a patrol including check count and duration.

Examples:
  orc patrol status PATROL-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		patrolID := args[0]

		patrol, err := wire.PatrolService().GetPatrol(ctx, patrolID)
		if err != nil {
			return fmt.Errorf("patrol not found: %w", err)
		}

		fmt.Printf("Patrol: %s\n", patrol.ID)
		fmt.Printf("Kennel: %s\n", patrol.KennelID)
		fmt.Printf("Target: %s\n", patrol.Target)
		fmt.Printf("Status: %s\n", patrol.Status)
		fmt.Printf("Started: %s\n", patrol.StartedAt)
		if patrol.EndedAt != "" {
			fmt.Printf("Ended: %s\n", patrol.EndedAt)
		}
		fmt.Printf("Created: %s\n", patrol.CreatedAt)
		fmt.Printf("Updated: %s\n", patrol.UpdatedAt)

		return nil
	},
}

var patrolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List patrols",
	Long: `List patrols with optional filtering by kennel or status.

Examples:
  orc patrol list
  orc patrol list --status active
  orc patrol list --kennel KENNEL-001`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		kennelID, _ := cmd.Flags().GetString("kennel")
		status, _ := cmd.Flags().GetString("status")

		patrols, err := wire.PatrolService().ListPatrols(ctx, primary.PatrolFilters{
			KennelID: kennelID,
			Status:   status,
		})
		if err != nil {
			return fmt.Errorf("failed to list patrols: %w", err)
		}

		if len(patrols) == 0 {
			fmt.Println("No patrols found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tKENNEL\tTARGET\tSTATUS\tSTARTED")
		fmt.Fprintln(w, "--\t------\t------\t------\t-------")
		for _, p := range patrols {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				p.ID,
				p.KennelID,
				p.Target,
				p.Status,
				p.StartedAt,
			)
		}
		w.Flush()
		return nil
	},
}

func init() {
	// patrol list flags
	patrolListCmd.Flags().String("kennel", "", "Filter by kennel ID")
	patrolListCmd.Flags().StringP("status", "s", "", "Filter by status (active|completed|escalated)")

	// Register subcommands
	patrolCmd.AddCommand(patrolStartCmd)
	patrolCmd.AddCommand(patrolEndCmd)
	patrolCmd.AddCommand(patrolStatusCmd)
	patrolCmd.AddCommand(patrolListCmd)
}

// PatrolCmd returns the patrol command
func PatrolCmd() *cobra.Command {
	return patrolCmd
}
