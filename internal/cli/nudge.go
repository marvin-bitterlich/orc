package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/agent"
	"github.com/example/orc/internal/tmux"
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
  orc nudge IMP-GROVE-001 "Tests are failing, need fix ASAP"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			message := args[1]

			// Parse agent ID
			identity, err := agent.ParseAgentID(agentID)
			if err != nil {
				return fmt.Errorf("invalid agent ID: %w", err)
			}

			// For IMP agents, we need to look up grove info
			var target string
			if identity.Type == agent.AgentTypeIMP {
				// Extract grove ID from IMP-GROVE-001
				grove, err := wire.GroveService().GetGrove(context.Background(), identity.ID)
				if err != nil {
					return fmt.Errorf("failed to get grove info: %w", err)
				}

				// Resolve tmux target
				target, err = agent.ResolveTMuxTarget(agentID, grove.Name)
				if err != nil {
					return fmt.Errorf("failed to resolve target: %w", err)
				}

				// Update identity with mission ID for session check
				identity.MissionID = grove.MissionID
			} else {
				// ORC
				target, err = agent.ResolveTMuxTarget(agentID, "")
				if err != nil {
					return fmt.Errorf("failed to resolve target: %w", err)
				}
			}

			// Check if session exists
			sessionName := fmt.Sprintf("orc-%s", identity.MissionID)
			if !tmux.SessionExists(sessionName) {
				return fmt.Errorf("tmux session %s not running - agent may not be active", sessionName)
			}

			// Send nudge
			if err := tmux.NudgeSession(target, message); err != nil {
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
