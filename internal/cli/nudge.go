package cli

import (
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

			ctx := NewContext()
			tmuxAdapter := wire.TMuxAdapter()

			var sessionName string
			var target string

			if identity.Type == agent.AgentTypeGoblin {
				if identity.ID == "GOBLIN" {
					// Legacy generic GOBLIN address - check for ORC session
					sessionName = "ORC"
					if !tmuxAdapter.SessionExists(ctx, sessionName) {
						return fmt.Errorf("tmux session %s not running - agent may not be active", sessionName)
					}
					target = "ORC:1.1"
				} else {
					// GOBLIN-GATE-XXX: lookup gatehouse → workshop → session
					gatehouse, err := wire.GatehouseService().GetGatehouse(ctx, identity.ID)
					if err != nil {
						return fmt.Errorf("failed to get gatehouse info: %w", err)
					}

					sessionName = tmuxAdapter.FindSessionByWorkshopID(ctx, gatehouse.WorkshopID)
					if sessionName == "" {
						return fmt.Errorf("no tmux session found for workshop %s - agent may not be active", gatehouse.WorkshopID)
					}

					// Gatehouse is window 1, pane 1 (Claude)
					target = fmt.Sprintf("%s:1.1", sessionName)
				}
			} else {
				// IMP: lookup workbench and find session by workshop ID
				workbench, err := wire.WorkbenchService().GetWorkbench(ctx, identity.ID)
				if err != nil {
					return fmt.Errorf("failed to get workbench info: %w", err)
				}

				// Find session by workshop ID (runtime lookup)
				sessionName = tmuxAdapter.FindSessionByWorkshopID(ctx, workbench.WorkshopID)
				if sessionName == "" {
					return fmt.Errorf("no tmux session found for workshop %s - agent may not be active", workbench.WorkshopID)
				}

				// Window named by workbench, pane 2 is Claude
				target = fmt.Sprintf("%s:%s.2", sessionName, workbench.Name)
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
