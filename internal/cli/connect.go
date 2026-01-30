package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
)

// ConnectCmd returns the connect command
func ConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Launch Claude agent with boot instructions",
		Long: `Launch Claude Code with immediate directive to run 'orc prime'.

This command is designed to be the root command for agent TMux panes, ensuring
that every time a pane spawns or respawns, Claude boots with proper context.

The boot sequence:
  1. orc connect --role <role> (launches claude)
  2. Claude receives directive: "Run the orc prime shell command IMMEDIATELY"
  3. Claude executes: orc prime
  4. Agent receives full context (identity, assignments, rules)
  5. Agent is ready to work

Usage:
  orc connect                    # Launch Claude (role from place_id)
  orc connect --role imp         # Launch as IMP (workbench only)
  orc connect --role goblin      # Launch as Goblin (gatehouse only)

Role validation:
  - IMP role is only allowed from workbenches (BENCH-XXX)
  - Goblin role is only allowed from gatehouses (GATE-XXX)
  - If no role specified, it's inferred from the place_id

TMux Integration:
  # Set as pane root command
  tmux send-keys -t session:window.pane "orc connect" C-m

  # Respawn pane with orc connect
  tmux respawn-pane -t session:window.pane -k "orc connect"`,
		RunE: runConnect,
	}

	cmd.Flags().Bool("dry-run", false, "Show command that would be executed without running it")
	cmd.Flags().String("role", "", "Agent role (imp, goblin)")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	role, _ := cmd.Flags().GetString("role")

	// Load config to get place_id (with Goblin migration if needed)
	cwd, _ := os.Getwd()
	cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)

	// Validate role against place_id
	if err := validateRoleForPlace(role, cfg); err != nil {
		return err
	}

	// The prime directive: Claude must run orc prime immediately upon boot
	// Note: settings.local.json ensures Claude starts in normal mode (not plan mode)
	primeDirective := "Run the 'orc prime' shell command IMMEDIATELY. Do not greet the user, do not explain what you're doing - just execute the command and show the output."

	// Build claude command
	// Using "claude" assumes it's in PATH (standard Claude Code installation)
	claudeArgs := []string{primeDirective}
	claudeCmd := exec.Command("claude", claudeArgs...)

	// Pass through stdio for interactive session
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Set working directory to current directory
	claudeCmd.Dir = cwd

	if dryRun {
		fmt.Printf("Would execute: claude %q\n", primeDirective)
		fmt.Printf("Working directory: %s\n", claudeCmd.Dir)
		if role != "" {
			fmt.Printf("Role: %s\n", role)
		}
		return nil
	}

	// Launch Claude
	return claudeCmd.Run()
}

// validateRoleForPlace validates that the requested role is allowed for the current place.
// - IMP role is only allowed from workbenches (BENCH-XXX)
// - Goblin role is only allowed from gatehouses (GATE-XXX)
// - If no role specified, validation always passes (role inferred from place_id)
func validateRoleForPlace(role string, cfg *config.Config) error {
	// If no role specified, no validation needed
	if role == "" {
		return nil
	}

	// If no config, we can't validate against place
	if cfg == nil || cfg.PlaceID == "" {
		// Allow any role if we can't determine place
		return nil
	}

	placeType := config.GetPlaceType(cfg.PlaceID)

	switch role {
	case "imp":
		if placeType == config.PlaceTypeGatehouse {
			return fmt.Errorf("IMP role not allowed from gatehouse (%s)", cfg.PlaceID)
		}
	case "goblin":
		if placeType == config.PlaceTypeWorkbench {
			return fmt.Errorf("Goblin role not allowed from workbench (%s)", cfg.PlaceID)
		}
	case "watchdog":
		// Watchdog can run from anywhere (for now)
	default:
		return fmt.Errorf("unknown role: %s (valid roles: imp, goblin, watchdog)", role)
	}

	return nil
}
