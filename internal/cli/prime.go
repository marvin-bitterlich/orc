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

This command detects the agent's location (grove/mission/global) and provides
appropriate context automatically.

Output includes:
- Current location (cwd)
- Mission context (if any)
- Current focus (if any)
- Core rules for ORC usage

Still useful for:
- Manual context refresh during long sessions
- Debugging and understanding mission context
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

	// Detect contexts
	groveCtx, _ := ctx.DetectGroveContext()
	missionCtx, _ := ctx.DetectMissionContext()

	// Load config to check role
	cfg, _ := config.LoadConfig(cwd)

	// Determine role from config
	var role string
	if cfg != nil {
		switch cfg.Type {
		case config.TypeGrove:
			if cfg.Grove != nil {
				role = cfg.Grove.Role
			}
		case config.TypeMission:
			if cfg.Mission != nil {
				role = cfg.Mission.Role
			}
		case config.TypeGlobal:
			if cfg.State != nil {
				role = cfg.State.Role
			}
		}
	}

	// If no role configured, show fallback guidance
	if role == "" {
		output := buildFallbackOutput(cwd, groveCtx, missionCtx)
		fmt.Println(truncateOutput(output, format, maxLines))
		return nil
	}

	// Route based on role
	var fullOutput string
	switch role {
	case config.RoleIMP:
		if groveCtx != nil {
			fullOutput = buildIMPPrimeOutput(groveCtx, cwd)
		} else {
			fullOutput = buildFallbackOutput(cwd, groveCtx, missionCtx)
		}
	case config.RoleORC:
		fullOutput = buildORCPrimeOutput(missionCtx, cwd, cfg)
	default:
		fullOutput = buildFallbackOutput(cwd, groveCtx, missionCtx)
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

// buildORCPrimeOutput creates ORC orchestrator context output
func buildORCPrimeOutput(missionCtx *ctx.MissionContext, cwd string, cfg *config.Config) string {
	var output strings.Builder

	output.WriteString("# ORC Context (Session Prime)\n\n")

	// Identity
	output.WriteString("## Identity\n\n")
	output.WriteString("**Role**: Orchestrator (ORC)\n")
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Current focus (if set)
	if cfg != nil {
		focusID := GetCurrentFocus(cfg)
		if focusID != "" {
			containerType, title, status := GetFocusInfo(focusID)
			if containerType != "" {
				output.WriteString("## Current Focus\n\n")
				output.WriteString(fmt.Sprintf("**%s**: %s - %s [%s]\n\n", containerType, focusID, title, status))
			}
		}
	}

	// Mission context
	if missionCtx != nil {
		output.WriteString(fmt.Sprintf("**Mission**: %s\n", missionCtx.MissionID))
		output.WriteString(fmt.Sprintf("**Workspace**: %s\n\n", missionCtx.WorkspacePath))

		// Get mission details
		mission, err := wire.MissionService().GetMission(context.Background(), missionCtx.MissionID)
		if err == nil {
			output.WriteString(fmt.Sprintf("## %s\n\n", mission.Title))
			if mission.Description != "" {
				descLines := strings.Split(mission.Description, "\n")
				if len(descLines) > 3 {
					output.WriteString(strings.Join(descLines[:3], "\n"))
					output.WriteString("\n...\n\n")
				} else {
					output.WriteString(mission.Description)
					output.WriteString("\n\n")
				}
			}
		}
	} else {
		output.WriteString("**Context**: Master orchestrator (global)\n\n")
		output.WriteString("Run `orc mission list` to see available missions.\n\n")
	}

	// Core rules (shared)
	output.WriteString(getCoreRules())

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeORC()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are the ORC - Orchestrator coordinating missions and IMPs.\n")
	}

	return output.String()
}

// buildIMPPrimeOutput creates IMP-focused context prompt when in grove territory
func buildIMPPrimeOutput(groveCtx *ctx.GroveContext, cwd string) string {
	var output strings.Builder

	output.WriteString("# IMP Boot Context\n\n")

	// Section 1: IMP Identity
	output.WriteString("## Identity\n\n")
	output.WriteString("**Role**: Implementation Agent (IMP)\n")
	output.WriteString(fmt.Sprintf("**Grove**: %s (`%s`)\n", groveCtx.Name, groveCtx.GroveID))
	output.WriteString(fmt.Sprintf("**Mission**: `%s`\n", groveCtx.MissionID))
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitInstructions())

	// Section 2: Assignment - All containers assigned to this grove
	output.WriteString("## Assignment\n\n")

	hasAssignments := false

	// Shipments
	shipments, _ := wire.ShipmentService().GetShipmentsByGrove(context.Background(), groveCtx.GroveID)
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

	// Conclaves
	conclaves, _ := wire.ConclaveService().GetConclavesByGrove(context.Background(), groveCtx.GroveID)
	for i, conclave := range conclaves {
		hasAssignments = true
		output.WriteString(fmt.Sprintf("### Conclave %d: %s\n\n", i+1, conclave.ID))
		output.WriteString(fmt.Sprintf("**%s** [%s] (Ideation Session)\n\n", conclave.Title, conclave.Status))

		if conclave.Description != "" {
			output.WriteString(conclave.Description)
			output.WriteString("\n\n")
		}
	}

	// Investigations
	investigations, _ := wire.InvestigationService().GetInvestigationsByGrove(context.Background(), groveCtx.GroveID)
	for i, inv := range investigations {
		hasAssignments = true
		output.WriteString(fmt.Sprintf("### Investigation %d: %s\n\n", i+1, inv.ID))
		output.WriteString(fmt.Sprintf("**%s** [%s] (Research)\n\n", inv.Title, inv.Status))

		if inv.Description != "" {
			output.WriteString(inv.Description)
			output.WriteString("\n\n")
		}

		// Show open questions
		questions, err := wire.InvestigationService().GetInvestigationQuestions(context.Background(), inv.ID)
		if err == nil {
			var openQuestions []*primary.InvestigationQuestion
			for _, q := range questions {
				if q.Status == "open" {
					openQuestions = append(openQuestions, q)
				}
			}
			if len(openQuestions) > 0 {
				output.WriteString("**Open Questions**:\n")
				for _, q := range openQuestions {
					output.WriteString(fmt.Sprintf("- %s - %s\n", q.ID, q.Title))
				}
				output.WriteString("\n")
			}
		}
	}

	if !hasAssignments {
		output.WriteString("*No containers currently assigned to this grove.*\n\n")
		output.WriteString("Run `orc summary` to see the full mission tree.\n\n")
	}

	// Section 3: ORC CLI Primer
	output.WriteString("## ORC CLI Primer\n\n")
	output.WriteString("**Core Commands**:\n")
	output.WriteString("- `orc summary` - View mission tree with all containers\n")
	output.WriteString("- `orc focus ID` - Set focus to a container (SHIP-*, CON-*, INV-*, TOME-*)\n")
	output.WriteString("- `orc task list --shipment SHIP-ID` - List tasks for a shipment\n")
	output.WriteString("- `orc question list --investigation INV-ID` - List questions for an investigation\n")
	output.WriteString("- `orc task complete TASK-ID` - Mark task as completed\n")
	output.WriteString("- `orc question answer Q-ID` - Record answer to a question\n\n")

	// Section 4: Core Rules (shared across all session types)
	output.WriteString(getCoreRules())
	output.WriteString("- **Stay in grove territory** - Work within assigned containers only\n\n")

	// Footer (loaded from template)
	welcome, err := templates.GetWelcomeIMP()
	if err == nil {
		output.WriteString(welcome)
	} else {
		output.WriteString("---\nYou are an IMP - Implementation agent working within a grove on assigned shipments.\n")
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

// buildFallbackOutput creates output when no role is configured
func buildFallbackOutput(cwd string, groveCtx *ctx.GroveContext, missionCtx *ctx.MissionContext) string {
	var output strings.Builder

	output.WriteString("# ORC Context - Role Not Configured\n\n")
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	output.WriteString("## Configuration Required\n\n")
	output.WriteString("No role is configured for this context. The `role` field must be set in `.orc/config.json`.\n\n")

	if groveCtx != nil {
		output.WriteString(fmt.Sprintf("**Detected grove**: %s (%s)\n", groveCtx.Name, groveCtx.GroveID))
		output.WriteString("**Suggested role**: `IMP` (Implementation Agent)\n\n")
		output.WriteString("To configure, edit `.orc/config.json` and add:\n")
		output.WriteString("```json\n")
		output.WriteString("\"role\": \"IMP\"\n")
		output.WriteString("```\n\n")
	} else if missionCtx != nil {
		output.WriteString(fmt.Sprintf("**Detected mission**: %s\n", missionCtx.MissionID))
		output.WriteString("**Suggested role**: `ORC` (Orchestrator)\n\n")
		output.WriteString("To configure, edit `.orc/config.json` and add:\n")
		output.WriteString("```json\n")
		output.WriteString("\"role\": \"ORC\"\n")
		output.WriteString("```\n\n")
	} else {
		output.WriteString("No mission or grove context detected.\n")
		output.WriteString("You may be in the wrong directory.\n\n")
		output.WriteString("**To connect to ORC**: Run `orc attach` to attach to the orchestrator session.\n\n")
	}

	output.WriteString(getGitInstructions())
	output.WriteString(getCoreRules())

	output.WriteString("---\n")
	output.WriteString("⚠️ **Action Required**: Ask the user to configure the role or direct you to the correct context.\n")

	return output.String()
}
