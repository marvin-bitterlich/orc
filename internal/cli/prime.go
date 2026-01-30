package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	ctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/templates"
)

// PrimeCmd returns the prime command
func PrimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prime",
		Short: "Output lightweight context for session priming",
		Long: `Generate concise context snapshot for Claude agents on startup.

NOTE: Originally designed for SessionStart hook injection, but hooks are currently
broken in Claude Code v2.1.7. ORC now uses direct prompt pattern when starting agents:
  claude "Run orc prime"

This command detects the agent's location (workbench/global) and provides
appropriate context automatically.

Output includes:
- Current location (cwd)
- Commission context (if any)
- Current focus (if any)
- Core rules for ORC usage

Still useful for:
- Manual context refresh during long sessions
- Debugging and understanding commission context
- Testing what would be injected via hooks (if they worked)

Examples:
  orc prime
  orc prime --format text
  orc prime --max-lines 40`,
		RunE: runPrime,
	}

	cmd.Flags().String("format", "text", "Output format (text or json)")
	cmd.Flags().Int("max-lines", 60, "Maximum lines of output (text format only)")

	return cmd
}

func runPrime(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	maxLines, _ := cmd.Flags().GetInt("max-lines")

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "(unknown)"
	}

	// Detect workbench context (IMP territory)
	workbenchCtx, _ := ctx.DetectWorkbenchContext()

	// Load config to determine role from place_id (with Goblin migration if needed)
	cfg, _ := MigrateGoblinConfigIfNeeded(cmd.Context(), cwd)

	// Determine role from place_id
	var role string
	if cfg != nil && cfg.PlaceID != "" {
		role = config.GetRoleFromPlaceID(cfg.PlaceID)
	}

	// Route based on role
	var fullOutput string
	switch {
	case role == config.RoleIMP:
		if workbenchCtx != nil {
			fullOutput = buildIMPPrimeOutput(workbenchCtx, cwd)
		} else {
			// IMP role but no workbench context - show Goblin fallback
			fullOutput = buildGoblinPrimeOutput(cwd, cfg)
		}
	case role == config.RoleGoblin:
		// Explicit Goblin role from gatehouse
		fullOutput = buildGoblinPrimeOutput(cwd, cfg)
	default:
		// No config = Goblin context (default behavior)
		fullOutput = buildGoblinPrimeOutput(cwd, nil)
	}

	fmt.Println(truncateOutput(fullOutput, format, maxLines))
	return nil
}

// truncateOutput truncates output to max lines if needed
func truncateOutput(output, format string, maxLines int) string {
	if format == "text" {
		lines := strings.Split(output, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, "...", "", "*(Output truncated to max lines)*")
		}
		return strings.Join(lines, "\n")
	}
	return output
}

// buildGoblinPrimeOutput creates Goblin orchestrator context output
func buildGoblinPrimeOutput(cwd string, _ *config.Config) string {
	var output strings.Builder

	output.WriteString("# Goblin Context (Session Prime)\n\n")

	// Identity
	output.WriteString("## Identity\n\n")
	output.WriteString("**Role**: Goblin (Orchestrator)\n")
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Core rules (shared)
	output.WriteString(getCoreRules())

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeGoblin()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are a Goblin - Orchestrator coordinating commissions and IMPs.\n")
	}

	// Call to action - run summary for dynamic context
	output.WriteString("\n---\n\n**Run `orc summary` now to see active commissions and work.**\n")

	return output.String()
}

// buildIMPPrimeOutput creates IMP-focused context prompt when in workbench territory
func buildIMPPrimeOutput(workbenchCtx *ctx.WorkbenchContext, cwd string) string {
	var output strings.Builder

	output.WriteString("# IMP Boot Context\n\n")

	// Section 1: IMP Identity
	output.WriteString("## Identity\n\n")
	output.WriteString("**Role**: Implementation Agent (IMP)\n")
	output.WriteString(fmt.Sprintf("**Workbench**: `%s`\n", workbenchCtx.WorkbenchID))
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Section 2: ORC CLI Primer
	output.WriteString("## ORC CLI Primer\n\n")
	output.WriteString("**Core Commands**:\n")
	output.WriteString("- `orc summary` - View commission tree with all containers\n")
	output.WriteString("- `orc focus ID` - Set focus to a container (SHIP-*, CON-*, TOME-*)\n")
	output.WriteString("- `orc task list --shipment SHIP-ID` - List tasks for a shipment\n")
	output.WriteString("- `orc note list --tome TOME-ID` - List notes for a tome\n")
	output.WriteString("- `orc task complete TASK-ID` - Mark task as completed\n\n")

	// Section 3: Core Rules (shared across all session types)
	output.WriteString(getCoreRules())
	output.WriteString("- **Stay in workbench territory** - Work within assigned containers only\n\n")

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeIMP()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are an IMP - Implementation agent working within a workbench on assigned shipments.\n")
	}

	// Call to action - run summary for dynamic context
	output.WriteString("\n---\n\n**Run `orc summary` now to see your current assignments and context.**\n")

	return output.String()
}

// getGitInstructions returns instructions for Claude to discover git context
// Content loaded from templates/prime/git-discovery.tmpl
func getGitInstructions() string {
	content, err := templates.GetGitDiscovery()
	if err != nil {
		// Fallback to inline if template fails
		return "## Git Context Discovery\n\nRun `git status` to see repository state.\n\n"
	}
	return content
}

// getCoreRules returns the core rules that apply to ALL Claude sessions
// Content loaded from templates/prime/core-rules.tmpl
func getCoreRules() string {
	content, err := templates.GetCoreRules()
	if err != nil {
		// Fallback to inline if template fails
		return "## Core Rules\n\n- Track all work in ORC ledger\n- TodoWrite tool is NOT ALLOWED\n\n"
	}
	return content
}
