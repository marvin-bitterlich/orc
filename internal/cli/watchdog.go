package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var watchdogCmd = &cobra.Command{
	Use:   "watchdog",
	Short: "Manage watchdog panes",
	Long:  "Summon and dispatch watchdog panes between kennels and gatehouses",
}

var watchdogSummonCmd = &cobra.Command{
	Use:   "summon [kennel-id]",
	Short: "Move watchdog pane from kennel to gatehouse dogbed",
	Long: `Move a watchdog pane from its kennel (workbench) to the gatehouse dogbed.

This joins the watchdog pane vertically into the gatehouse window,
allowing the Goblin to monitor and assist the IMP.

The kennel status changes from 'occupied' to 'away'.

Examples:
  orc watchdog summon KENNEL-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		kennelID := args[0]

		// Get kennel
		kennel, err := wire.KennelService().GetKennel(ctx, kennelID)
		if err != nil {
			return fmt.Errorf("kennel not found: %w", err)
		}

		if kennel.Status != primary.KennelStatusOccupied {
			return fmt.Errorf("kennel %s is not occupied (status: %s)", kennelID, kennel.Status)
		}

		// Get workbench for this kennel
		workbench, err := wire.WorkbenchService().GetWorkbench(ctx, kennel.WorkbenchID)
		if err != nil {
			return fmt.Errorf("workbench not found: %w", err)
		}

		// Get workshop for this workbench
		workshop, err := wire.WorkshopService().GetWorkshop(ctx, workbench.WorkshopID)
		if err != nil {
			return fmt.Errorf("workshop not found: %w", err)
		}

		// Find the TMux session for this workshop
		sessionName := wire.TMuxAdapter().FindSessionByWorkshopID(ctx, workshop.ID)
		if sessionName == "" {
			return fmt.Errorf("no TMux session found for workshop %s", workshop.ID)
		}

		// Build source and target pane targets
		// Source: workbench window, pane 2 (IMP pane)
		source := fmt.Sprintf("%s:%s.2", sessionName, workbench.Name)
		// Target: ORC window (window 1), pane 1 (Goblin pane)
		target := fmt.Sprintf("%s:1.1", sessionName)

		// Execute join-pane (move pane from workbench to gatehouse)
		if err := wire.TMuxAdapter().JoinPane(ctx, source, target, true, 8); err != nil {
			return fmt.Errorf("failed to join pane: %w", err)
		}

		// Update kennel status to away
		if err := wire.KennelService().UpdateKennelStatus(ctx, kennelID, primary.KennelStatusAway); err != nil {
			return fmt.Errorf("failed to update kennel status: %w", err)
		}

		fmt.Printf("✓ Watchdog summoned from %s to gatehouse\n", kennelID)
		fmt.Printf("  Kennel status: %s → %s\n", primary.KennelStatusOccupied, primary.KennelStatusAway)
		return nil
	},
}

var watchdogDispatchCmd = &cobra.Command{
	Use:   "dispatch [kennel-id]",
	Short: "Return watchdog pane from gatehouse dogbed to kennel",
	Long: `Return a watchdog pane from the gatehouse dogbed back to its kennel (workbench).

This joins the watchdog pane vertically back into the workbench window.

The kennel status changes from 'away' to 'occupied'.

Examples:
  orc watchdog dispatch KENNEL-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		kennelID := args[0]

		// Get kennel
		kennel, err := wire.KennelService().GetKennel(ctx, kennelID)
		if err != nil {
			return fmt.Errorf("kennel not found: %w", err)
		}

		if kennel.Status != primary.KennelStatusAway {
			return fmt.Errorf("kennel %s is not away (status: %s)", kennelID, kennel.Status)
		}

		// Get workbench for this kennel
		workbench, err := wire.WorkbenchService().GetWorkbench(ctx, kennel.WorkbenchID)
		if err != nil {
			return fmt.Errorf("workbench not found: %w", err)
		}

		// Get workshop for this workbench
		workshop, err := wire.WorkshopService().GetWorkshop(ctx, workbench.WorkshopID)
		if err != nil {
			return fmt.Errorf("workshop not found: %w", err)
		}

		// Find the TMux session for this workshop
		sessionName := wire.TMuxAdapter().FindSessionByWorkshopID(ctx, workshop.ID)
		if sessionName == "" {
			return fmt.Errorf("no TMux session found for workshop %s", workshop.ID)
		}

		// The dogbed pane is now in the ORC window - we need to find which pane it is.
		// After summon, it was joined to pane 1, so it becomes a new pane (likely 2 or higher)
		// For now, assume it's the last pane in the ORC window. This may need refinement.
		// Source: ORC window, last pane (we use pane 2 as approximation)
		source := fmt.Sprintf("%s:1.2", sessionName)
		// Target: workbench window, pane 1 (vim pane area, will be placed below)
		target := fmt.Sprintf("%s:%s.1", sessionName, workbench.Name)

		// Execute join-pane (move pane from gatehouse back to workbench)
		if err := wire.TMuxAdapter().JoinPane(ctx, source, target, true, 8); err != nil {
			return fmt.Errorf("failed to join pane: %w", err)
		}

		// Update kennel status to occupied
		if err := wire.KennelService().UpdateKennelStatus(ctx, kennelID, primary.KennelStatusOccupied); err != nil {
			return fmt.Errorf("failed to update kennel status: %w", err)
		}

		fmt.Printf("✓ Watchdog dispatched to %s\n", kennelID)
		fmt.Printf("  Kennel status: %s → %s\n", primary.KennelStatusAway, primary.KennelStatusOccupied)
		return nil
	},
}

func init() {
	watchdogCmd.AddCommand(watchdogSummonCmd)
	watchdogCmd.AddCommand(watchdogDispatchCmd)
}

// WatchdogCmd returns the watchdog command
func WatchdogCmd() *cobra.Command {
	return watchdogCmd
}
