package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/wire"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Create or attach to master ORC TMux session",
	Long: `Create or attach to the master orchestrator TMux session.

This command creates a sophisticated TMux workspace for ORC development:
  - Session name: orc-master
  - Working directory: ~/src/orc
  - Layout:
    ┌─────────────────────┬──────────────┐
    │                     │   vim (top)  │
    │      claude         │──────────────│
    │    (full height)    │  shell (bot) │
    │                     │              │
    └─────────────────────┴──────────────┘

If the session already exists, provides attach instructions.

Examples:
  orc attach`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		sessionName := "orc-master"
		tmuxAdapter := wire.TMuxAdapter()

		// Check if session already exists
		if tmuxAdapter.SessionExists(ctx, sessionName) {
			fmt.Printf("✓ Attaching to existing session: %s\n", sessionName)

			// Find tmux binary
			tmuxPath, err := exec.LookPath("tmux")
			if err != nil {
				return fmt.Errorf("tmux not found in PATH: %w", err)
			}

			// Replace current process with tmux attach
			// This makes the transition seamless - user just runs 'orc attach' and ends up in tmux
			args := []string{"tmux", "attach", "-t", sessionName}
			env := os.Environ()

			if err := syscall.Exec(tmuxPath, args, env); err != nil {
				return fmt.Errorf("failed to exec tmux attach: %w", err)
			}

			// This line never executes if exec succeeds
			return nil
		}

		// TMux lifecycle removed - delegate to `orc tmux apply`
		return fmt.Errorf("'orc attach' session creation is deprecated.\n\nFor ORC development, manually create a session or use 'orc tmux apply <workshop-id>' for workshops.\nTMux lifecycle is now managed by gotmux. See: docs/tmux.md")
	},
}

// AttachCmd returns the attach command
func AttachCmd() *cobra.Command {
	return attachCmd
}
