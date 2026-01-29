package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/wire"
)

// NudgeCmd returns the nudge command
func NudgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nudge <agent-id> <message>",
		Short: "Send real-time message to running agent",
		Long: `Send a message directly to a running Claude session via tmux injection.

This sends the message immediately to the agent's input, as if they typed it.
The agent must be running in a tmux session for this to work.

Examples:
  orc nudge ORC "Check your mail - urgent task"
  orc nudge IMP-WB-001 "Tests are failing, need fix ASAP"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			message := args[1]

			// Parse agent ID
			identity, err := agent.ParseAgentID(agentID)
			if err != nil {
				return fmt.Errorf("invalid agent ID: %w", err)
			}

			// For IMP agents, we need to look up workbench info and commission
			var target string
			var commissionID string
			ctx := context.Background()

			if identity.Type == agent.AgentTypeIMP {
				// Extract workbench ID and resolve commission via workshop chain
				workbench, err := wire.WorkbenchService().GetWorkbench(ctx, identity.ID)
				if err != nil {
					return fmt.Errorf("failed to get workbench info: %w", err)
				}

				// Get workshop to find factory/commission context
				workshop, err := wire.WorkshopService().GetWorkshop(ctx, workbench.WorkshopID)
				if err != nil {
					return fmt.Errorf("failed to get workshop info: %w", err)
				}

				// For now, use workshop name as part of session name
				// TODO: Resolve commission through factory if needed
				commissionID = workshop.Name

				// Resolve tmux target
				target, err = agent.ResolveTMuxTarget(agentID, workbench.Name, commissionID)
				if err != nil {
					return fmt.Errorf("failed to resolve target: %w", err)
				}
			} else {
				// Goblin
				target, err = agent.ResolveTMuxTarget(agentID, "", "")
				if err != nil {
					return fmt.Errorf("failed to resolve target: %w", err)
				}
			}

			// Check if session exists
			tmuxAdapter := wire.TMuxAdapter()
			sessionName := fmt.Sprintf("orc-%s", commissionID)
			if commissionID == "" {
				sessionName = "ORC" // Goblin session
			}
			if !tmuxAdapter.SessionExists(ctx, sessionName) {
				return fmt.Errorf("tmux session %s not running - agent may not be active", sessionName)
			}

			// Send nudge
			if err := tmuxAdapter.NudgeSession(ctx, target, message); err != nil {
				return fmt.Errorf("failed to send nudge: %w", err)
			}

			fmt.Printf("✓ Nudge sent to %s\n", agentID)
			fmt.Printf("  Target: %s\n", target)
			fmt.Printf("  Message: %s\n", message)

			return nil
		},
	}

	return cmd
}
