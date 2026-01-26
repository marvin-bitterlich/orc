package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/config"
	ctx "github.com/example/orc/internal/context"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/templates"
	"github.com/example/orc/internal/wire"
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

	// Load config to check role
	cfg, _ := config.LoadConfig(cwd)

	// Determine role from config
	var role string
	if cfg != nil {
		role = cfg.Role
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
	case config.IsGoblinRole(role):
		// Explicit Goblin role (or backwards-compat "ORC")
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
func buildGoblinPrimeOutput(cwd string, cfg *config.Config) string {
	var output strings.Builder

	output.WriteString("# Goblin Context (Session Prime)\n\n")

	// Identity
	output.WriteString("## Identity\n\n")
	output.WriteString("**Role**: Goblin (Orchestrator)\n")
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Current focus (if set)
	if cfg != nil && cfg.CurrentFocus != "" {
		containerType, title, status := GetFocusInfo(cfg.CurrentFocus)
		if containerType != "" {
			output.WriteString("## Current Focus\n\n")
			output.WriteString(fmt.Sprintf("**%s**: %s - %s [%s]\n\n", containerType, cfg.CurrentFocus, title, status))
		}
	}

	// Commission context (if set)
	if cfg != nil && cfg.CommissionID != "" {
		output.WriteString(fmt.Sprintf("**Commission**: %s\n\n", cfg.CommissionID))

		// Get commission details
		commission, err := wire.CommissionService().GetCommission(context.Background(), cfg.CommissionID)
		if err == nil {
			output.WriteString(fmt.Sprintf("## %s\n\n", commission.Title))
			if commission.Description != "" {
				descLines := strings.Split(commission.Description, "\n")
				if len(descLines) > 3 {
					output.WriteString(strings.Join(descLines[:3], "\n"))
					output.WriteString("\n...\n\n")
				} else {
					output.WriteString(commission.Description)
					output.WriteString("\n\n")
				}
			}
		}
	} else {
		output.WriteString("**Context**: Master orchestrator (global)\n\n")
		output.WriteString("Run `orc commission list` to see available commissions.\n\n")
	}

	// Cross-workshop summary
	output.WriteString(getCrossWorkshopSummary())

	// Spec-Kit workflow (when shipment focused)
	if cfg != nil && strings.HasPrefix(cfg.CurrentFocus, "SHIP-") {
		output.WriteString(getSpecKitWorkflow())
	}

	// Refinement mode (when CWO is draft)
	if cfg != nil && strings.HasPrefix(cfg.CurrentFocus, "SHIP-") {
		output.WriteString(getRefinementModeContext(cfg.CurrentFocus))
	}

	// Core rules (shared)
	output.WriteString(getCoreRules())

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeGoblin()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are a Goblin - Orchestrator coordinating commissions and IMPs.\n")
	}

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
	output.WriteString(fmt.Sprintf("**Commission**: `%s`\n", workbenchCtx.CommissionID))
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Section 2: Assignment - All containers assigned to this workbench
	output.WriteString("## Assignment\n\n")

	hasAssignments := false

	// Shipments
	shipments, _ := wire.ShipmentService().GetShipmentsByWorkbench(context.Background(), workbenchCtx.WorkbenchID)
	for i, shipment := range shipments {
		hasAssignments = true
		output.WriteString(fmt.Sprintf("### Shipment %d: %s\n\n", i+1, shipment.ID))
		output.WriteString(fmt.Sprintf("**%s** [%s]\n\n", shipment.Title, shipment.Status))

		if shipment.Description != "" {
			descLines := strings.Split(shipment.Description, "\n")
			if len(descLines) > 5 {
				output.WriteString(strings.Join(descLines[:5], "\n"))
				output.WriteString("\n\n*(Description truncated)*\n\n")
			} else {
				output.WriteString(shipment.Description)
				output.WriteString("\n\n")
			}
		}

		// Show ready tasks for this shipment
		tasks, err := wire.ShipmentService().GetShipmentTasks(context.Background(), shipment.ID)
		if err == nil {
			var readyTasks []*primary.Task
			for _, t := range tasks {
				if t.Status == "ready" {
					readyTasks = append(readyTasks, t)
				}
			}
			if len(readyTasks) > 0 {
				output.WriteString("**Ready Tasks**:\n")
				for _, task := range readyTasks {
					output.WriteString(fmt.Sprintf("- %s - %s\n", task.ID, task.Title))
				}
				output.WriteString("\n")
			}
		}
	}

	// Investigations
	investigations, _ := wire.InvestigationService().GetInvestigationsByWorkbench(context.Background(), workbenchCtx.WorkbenchID)
	for i, inv := range investigations {
		hasAssignments = true
		output.WriteString(fmt.Sprintf("### Investigation %d: %s\n\n", i+1, inv.ID))
		output.WriteString(fmt.Sprintf("**%s** [%s] (Research)\n\n", inv.Title, inv.Status))

		if inv.Description != "" {
			output.WriteString(inv.Description)
			output.WriteString("\n\n")
		}
	}

	if !hasAssignments {
		output.WriteString("*No containers currently assigned to this workbench.*\n\n")
		output.WriteString("Run `orc summary` to see the full commission tree.\n\n")
	}

	// Section 3: ORC CLI Primer
	output.WriteString("## ORC CLI Primer\n\n")
	output.WriteString("**Core Commands**:\n")
	output.WriteString("- `orc summary` - View commission tree with all containers\n")
	output.WriteString("- `orc focus ID` - Set focus to a container (SHIP-*, CON-*, INV-*, TOME-*)\n")
	output.WriteString("- `orc task list --shipment SHIP-ID` - List tasks for a shipment\n")
	output.WriteString("- `orc note list --investigation INV-ID` - List notes for an investigation\n")
	output.WriteString("- `orc task complete TASK-ID` - Mark task as completed\n\n")

	// Section 4: Core Rules (shared across all session types)
	output.WriteString(getCoreRules())
	output.WriteString("- **Stay in workbench territory** - Work within assigned containers only\n\n")

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeIMP()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are an IMP - Implementation agent working within a workbench on assigned shipments.\n")
	}

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

// getCrossWorkshopSummary returns a summary of active workshops for Goblin context
func getCrossWorkshopSummary() string {
	var output strings.Builder

	// Get all active commissions
	commissions, err := wire.CommissionService().ListCommissions(context.Background(), primary.CommissionFilters{})
	if err != nil || len(commissions) == 0 {
		return ""
	}

	// Filter to active commissions
	var activeCommissions []*primary.Commission
	for _, c := range commissions {
		if c.Status == "active" || c.Status == "in_progress" {
			activeCommissions = append(activeCommissions, c)
		}
	}

	if len(activeCommissions) == 0 {
		return ""
	}

	output.WriteString("## Workshop Floor Summary\n\n")

	for _, comm := range activeCommissions {
		// Get shipments for this commission to count work
		shipments, _ := wire.ShipmentService().ListShipments(context.Background(), primary.ShipmentFilters{CommissionID: comm.ID})
		activeCount := 0
		for _, s := range shipments {
			if s.Status == "active" || s.Status == "in_progress" {
				activeCount++
			}
		}

		if activeCount > 0 {
			output.WriteString(fmt.Sprintf("- **%s**: %s (%d active shipments)\n", comm.ID, comm.Title, activeCount))
		} else {
			output.WriteString(fmt.Sprintf("- **%s**: %s (idle)\n", comm.ID, comm.Title))
		}
	}

	output.WriteString("\n")
	return output.String()
}

// getSpecKitWorkflow returns the Spec-Kit ceremony workflow guide
func getSpecKitWorkflow() string {
	return `## Spec-Kit Workflow

When working on a shipment, follow this ceremony:

` + "```" + `
/orc-cycle  → Scope WHAT (create CWO)
/orc-plan   → Plan HOW (design implementation)
/orc-deliver → Close cycle (create CREC)
` + "```" + `

**Current cycle status:** Run ` + "`./orc cycle list --shipment-id SHIP-XXX`" + ` to see.

`
}

// getRefinementModeContext checks for draft CWOs and returns context about refinement mode
func getRefinementModeContext(shipmentID string) string {
	cwos, err := wire.CycleWorkOrderService().ListCycleWorkOrders(
		context.Background(),
		primary.CycleWorkOrderFilters{
			ShipmentID: shipmentID,
			Status:     "draft",
		},
	)
	if err != nil || len(cwos) == 0 {
		return ""
	}

	var output strings.Builder
	output.WriteString("## Refinement Mode\n\n")
	for _, cwo := range cwos {
		output.WriteString(fmt.Sprintf("**%s** is in DRAFT status.\n", cwo.ID))
		output.WriteString("- Review/refine the scope before approving\n")
		output.WriteString(fmt.Sprintf("- Run `./orc cwo approve %s` when ready\n\n", cwo.ID))
	}
	return output.String()
}
