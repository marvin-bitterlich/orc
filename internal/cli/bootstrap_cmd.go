package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// BootstrapCmd returns the bootstrap command for first-time setup
func BootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Start interactive first-run experience",
		Long: `Launch Claude Code with the /orc-first-run skill for first-time setup.

This command is designed for new ORC users after running 'make bootstrap'.
It launches Claude with a directive to run the first-run skill, which will:

  1. Check what ORC entities already exist
  2. Create missing entities (commission, workshop, workbench, shipment)
  3. Explain ORC concepts as it goes
  4. Guide through adding repos and configuring templates
  5. End with the tmux connect command

Usage:
  orc bootstrap              # Start first-run experience
  orc bootstrap --dry-run    # Show what would be executed

The skill is adaptive - if you've already set up ORC, it will tour your
existing setup rather than creating duplicates.`,
		RunE: runBootstrap,
	}

	cmd.Flags().Bool("dry-run", false, "Show command that would be executed without running it")

	return cmd
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	cwd, _ := os.Getwd()

	// The bootstrap directive: Claude must run /orc-first-run immediately
	bootstrapDirective := "Run the /orc-first-run skill IMMEDIATELY. This is a first-time ORC setup - follow the skill's guidance to configure the environment. Do not greet the user first - just start the skill."

	// Build claude command
	claudeArgs := []string{bootstrapDirective}
	claudeCmd := exec.Command("claude", claudeArgs...)

	// Pass through stdio for interactive session
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Set working directory to current directory
	claudeCmd.Dir = cwd

	if dryRun {
		fmt.Printf("Would execute: claude %q\n", bootstrapDirective)
		fmt.Printf("Working directory: %s\n", claudeCmd.Dir)
		return nil
	}

	// Launch Claude
	return claudeCmd.Run()
}
