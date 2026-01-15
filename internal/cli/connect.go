package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// ConnectCmd returns the connect command
func ConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Launch Claude agent with IMP boot instructions",
		Long: `Launch Claude Code with immediate directive to run 'orc prime'.

This command is designed to be the root command for IMP TMux panes, ensuring
that every time a pane spawns or respawns, Claude boots with proper context.

The boot sequence:
  1. orc connect (launches claude)
  2. Claude receives directive: "Run the orc prime shell command IMMEDIATELY"
  3. Claude executes: orc prime
  4. IMP receives full context (identity, assignments, rules)
  5. IMP is ready to work

Usage:
  orc connect                    # Launch Claude in IMP mode

TMux Integration:
  # Set as pane root command
  tmux send-keys -t session:window.pane "orc connect" C-m

  # Respawn pane with orc connect
  tmux respawn-pane -t session:window.pane -k "orc connect"`,
		RunE: runConnect,
	}

	cmd.Flags().Bool("dry-run", false, "Show command that would be executed without running it")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// The prime directive: Claude must exit plan mode (if active) and run orc prime immediately
	// Note: Some enterprise Claude Code configurations (via plugins/hooks) force plan mode on startup
	// We need to explicitly exit plan mode before running commands
	primeDirective := "If you are in plan mode, exit plan mode immediately. Then run the 'orc prime' shell command. Do not greet the user, do not explain what you're doing - just execute the command and show the output."

	// Build claude command
	// Using "claude" assumes it's in PATH (standard Claude Code installation)
	claudeArgs := []string{primeDirective}
	claudeCmd := exec.Command("claude", claudeArgs...)

	// Pass through stdio for interactive session
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Set working directory to current directory
	claudeCmd.Dir, _ = os.Getwd()

	if dryRun {
		fmt.Printf("Would execute: claude %q\n", primeDirective)
		fmt.Printf("Working directory: %s\n", claudeCmd.Dir)
		return nil
	}

	// Launch Claude
	return claudeCmd.Run()
}
