package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

// TmuxCmd returns the tmux command
func TmuxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmux",
		Short: "TMux session management",
		Long:  `Manage TMux sessions for workshops.`,
	}

	cmd.AddCommand(tmuxConnectCmd())

	return cmd
}

func tmuxConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [workshop-id]",
		Short: "Connect to a workshop's TMux session",
		Long: `Attach to an existing TMux session for a workshop.

This command does not create anything - it only connects to an existing session.
If no session exists, run 'orc infra apply' first to create infrastructure.

Examples:
  orc tmux connect WORK-001
  orc tmux connect WORK-003`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workshopID := args[0]
			ctx := context.Background()

			// 1. Verify workshop exists
			_, err := wire.WorkshopService().GetWorkshop(ctx, workshopID)
			if err != nil {
				return fmt.Errorf("workshop not found: %s", workshopID)
			}

			// 2. Check infrastructure exists
			plan, err := wire.InfraService().PlanInfra(ctx, primary.InfraPlanRequest{
				WorkshopID: workshopID,
			})
			if err != nil {
				return fmt.Errorf("failed to check infrastructure: %w", err)
			}

			// Check if infra is missing
			if plan.Gatehouse != nil && plan.Gatehouse.Status == primary.OpCreate {
				return fmt.Errorf("infrastructure not created for %s\nRun: orc infra apply %s", workshopID, workshopID)
			}

			// 3. Find tmux session
			sessionName := wire.TMuxAdapter().FindSessionByWorkshopID(ctx, workshopID)
			if sessionName == "" {
				return fmt.Errorf("no tmux session found for %s\nRun: orc infra apply %s", workshopID, workshopID)
			}

			// 4. Attach to session
			return attachToSession(sessionName)
		},
	}

	return cmd
}

// attachToSession replaces the current process with tmux attach.
func attachToSession(sessionName string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Use exec to replace current process
	args := []string{"tmux", "attach-session", "-t", sessionName}
	return syscall.Exec(tmuxPath, args, os.Environ())
}
