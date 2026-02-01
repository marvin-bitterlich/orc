package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
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
  orc connect --role watchdog --target BENCH-xxx  # Launch as Watchdog

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
	cmd.Flags().String("role", "", "Agent role (imp, goblin, watchdog)")
	cmd.Flags().String("target", "", "Target workbench ID (required for watchdog role)")

	return cmd
}

func runConnect(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	role, _ := cmd.Flags().GetString("role")
	target, _ := cmd.Flags().GetString("target")

	// Load config to get place_id (with Goblin migration if needed)
	cwd, _ := os.Getwd()
	cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)

	// Validate role against place_id
	if err := validateRoleForPlace(role, cfg); err != nil {
		return err
	}

	// Handle watchdog role specially
	if role == "watchdog" {
		return runConnectWatchdog(cmd.Context(), target, cwd, dryRun)
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

// runConnectWatchdog handles the watchdog role connection.
func runConnectWatchdog(ctx context.Context, target, cwd string, dryRun bool) error {
	if target == "" {
		return fmt.Errorf("--target is required for watchdog role")
	}

	// Normalize target: accept BENCH-xxx or IMP-BENCH-xxx
	workbenchID := target
	if strings.HasPrefix(target, "IMP-") {
		workbenchID = strings.TrimPrefix(target, "IMP-")
	}

	// Look up workbench
	workbench, err := wire.WorkbenchService().GetWorkbench(ctx, workbenchID)
	if err != nil {
		return fmt.Errorf("workbench not found: %w", err)
	}

	// Get or create kennel for workbench
	kennel, err := wire.KennelService().GetKennelByWorkbench(ctx, workbenchID)
	if err != nil {
		// Try to create kennel if it doesn't exist
		kennel, err = wire.KennelService().CreateKennel(ctx, workbenchID)
		if err != nil {
			return fmt.Errorf("failed to get or create kennel: %w", err)
		}
	}

	// Update kennel status to occupied
	if kennel.Status != primary.KennelStatusOccupied {
		if err := wire.KennelService().UpdateKennelStatus(ctx, kennel.ID, primary.KennelStatusOccupied); err != nil {
			return fmt.Errorf("failed to update kennel status: %w", err)
		}
	}

	// Watchdog prime directive includes orc prime and patrol start
	impIdentity := fmt.Sprintf("IMP-%s", workbenchID)
	primeDirective := fmt.Sprintf(
		"Run the 'orc prime' shell command IMMEDIATELY. Then start patrol on %s by running 'orc patrol start %s'. Do not greet the user, do not explain what you're doing - just execute these commands.",
		impIdentity, workbenchID,
	)

	if dryRun {
		fmt.Printf("Would execute: claude %q\n", primeDirective)
		fmt.Printf("Working directory: %s\n", cwd)
		fmt.Printf("Role: watchdog\n")
		fmt.Printf("Target workbench: %s (%s)\n", workbench.ID, workbench.Name)
		fmt.Printf("Kennel: %s\n", kennel.ID)
		fmt.Printf("Environment: WATCHDOG_KENNEL_ID=%s\n", kennel.ID)
		return nil
	}

	// Build claude command
	claudeArgs := []string{primeDirective}
	claudeCmd := exec.Command("claude", claudeArgs...)

	// Pass through stdio for interactive session
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Set working directory to workbench path
	claudeCmd.Dir = workbench.Path

	// Set environment variable for kennel identity
	claudeCmd.Env = append(os.Environ(), fmt.Sprintf("WATCHDOG_KENNEL_ID=%s", kennel.ID))

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
