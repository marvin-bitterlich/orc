package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/tmux"
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
		sessionName := "orc-master"

		// Get ORC source directory path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		orcPath := filepath.Join(home, "src", "orc")

		// Check if session already exists
		if tmux.SessionExists(sessionName) {
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

		// Verify ORC directory exists
		if _, err := os.Stat(orcPath); os.IsNotExist(err) {
			return fmt.Errorf("ORC source directory not found at %s", orcPath)
		}

		fmt.Printf("Creating master ORC TMux session: %s\n", sessionName)
		fmt.Printf("Working directory: %s\n", orcPath)
		fmt.Println()

		// Create session with base numbering from 1
		session, err := tmux.NewSession(sessionName, orcPath)
		if err != nil {
			return fmt.Errorf("failed to create TMux session: %w", err)
		}

		// Create ORC orchestrator window with sophisticated layout
		if err := session.CreateOrcWindow(orcPath); err != nil {
			return fmt.Errorf("failed to create ORC window: %w", err)
		}

		fmt.Println("✓ ORC session created")
		fmt.Println()
		fmt.Println("Layout:")
		fmt.Println("  - Left pane:  Claude (orchestrator)")
		fmt.Println("  - Top right:  Vim (code editing)")
		fmt.Println("  - Bot right:  Shell (commands)")
		fmt.Println()
		fmt.Println("Attaching to session...")

		// Find tmux binary
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			// Fallback to instructions if tmux not found
			fmt.Println(tmux.AttachInstructions(sessionName))
			return nil
		}

		// Replace current process with tmux attach
		attachArgs := []string{"tmux", "attach", "-t", sessionName}
		env := os.Environ()

		if err := syscall.Exec(tmuxPath, attachArgs, env); err != nil {
			return fmt.Errorf("failed to exec tmux attach: %w", err)
		}

		return nil
	},
}

// AttachCmd returns the attach command
func AttachCmd() *cobra.Command {
	return attachCmd
}
