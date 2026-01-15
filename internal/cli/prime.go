package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/example/orc/internal/context"
	"github.com/example/orc/internal/models"
	"github.com/spf13/cobra"
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
- Active work order (if any)
- Latest handoff note (brief)

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

	// Detect grove context first (IMP mode)
	groveCtx, _ := context.DetectGroveContext()
	if groveCtx != nil {
		output := buildIMPPrimeOutput(groveCtx, cwd)
		fmt.Println(output)
		return nil
	}

	// Fallback to old mission/orchestrator context
	missionCtx, _ := context.DetectMissionContext()

	// Build prime output
	var output strings.Builder

	output.WriteString("# ORC Context (Session Prime)\n\n")

	// Current location
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context (if in git repo)
	output.WriteString(getGitContext())

	// Mission context
	if missionCtx != nil {
		output.WriteString(fmt.Sprintf("**Mission**: %s (Deputy context)\n", missionCtx.MissionID))
		output.WriteString(fmt.Sprintf("**Workspace**: %s\n\n", missionCtx.WorkspacePath))

		// Get mission details
		mission, err := models.GetMission(missionCtx.MissionID)
		if err == nil {
			output.WriteString(fmt.Sprintf("## %s\n\n", mission.Title))
			if mission.Description.Valid && mission.Description.String != "" {
				descLines := strings.Split(mission.Description.String, "\n")
				if len(descLines) > 3 {
					output.WriteString(strings.Join(descLines[:3], "\n"))
					output.WriteString("\n...\n\n")
				} else {
					output.WriteString(mission.Description.String)
					output.WriteString("\n\n")
				}
			}

			// Get active tasks
			tasks, err := models.ListTasks("", "", "implement")
			if err == nil && len(tasks) > 0 {
				// Filter to current mission
				for _, task := range tasks {
					if task.MissionID == missionCtx.MissionID {
						output.WriteString(fmt.Sprintf("### Active Work: %s\n\n", task.ID))
						output.WriteString(fmt.Sprintf("**%s** [%s]\n\n", task.Title, task.Status))
						break
					}
				}
			}

			// Get latest handoff
			ho, err := models.GetLatestHandoff()
			if err == nil && ho != nil {
				output.WriteString(fmt.Sprintf("### Latest Handoff: %s\n\n", ho.ID))
				output.WriteString(fmt.Sprintf("*Created: %s*\n\n", ho.CreatedAt.Format("Jan 2, 15:04")))

				// Include brief note excerpt
				noteLines := strings.Split(ho.HandoffNote, "\n")
				if len(noteLines) > 5 {
					output.WriteString(strings.Join(noteLines[:5], "\n"))
					output.WriteString("\n\n*(Showing first 5 lines)*\n\n")
				} else {
					output.WriteString(ho.HandoffNote)
					output.WriteString("\n\n")
				}
			}
		}
	} else {
		output.WriteString("**Context**: Master orchestrator (global)\n\n")
		output.WriteString("Run `orc mission list` to see available missions.\n\n")
	}

	// Footer note
	output.WriteString("\n---\n")
	output.WriteString("💡 **Note**: This is lightweight orientation context.\n")

	// Truncate to max lines if needed
	fullOutput := output.String()
	if format == "text" {
		lines := strings.Split(fullOutput, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, "...", "", "*(Output truncated to max lines)*")
		}
		fullOutput = strings.Join(lines, "\n")
	}

	fmt.Println(fullOutput)
	return nil
}

// buildIMPPrimeOutput creates IMP-focused context prompt when in grove territory
func buildIMPPrimeOutput(groveCtx *context.GroveContext, cwd string) string {
	var output strings.Builder

	output.WriteString("# IMP Boot Context\n\n")

	// Section 1: IMP Identity
	output.WriteString("## Identity\n\n")
	output.WriteString(fmt.Sprintf("**Role**: Implementation Agent (IMP)\n"))
	output.WriteString(fmt.Sprintf("**Grove**: %s (`%s`)\n", groveCtx.Name, groveCtx.GroveID))
	output.WriteString(fmt.Sprintf("**Mission**: `%s`\n", groveCtx.MissionID))
	output.WriteString(fmt.Sprintf("**Location**: `%s`\n\n", cwd))

	// Git context
	output.WriteString(getGitContext())

	// Section 2: Assignment - Epics assigned to this grove
	output.WriteString("## Assignment\n\n")
	epics, err := models.GetEpicsByGrove(groveCtx.GroveID)
	if err == nil && len(epics) > 0 {
		for i, epic := range epics {
			output.WriteString(fmt.Sprintf("### Epic %d: %s\n\n", i+1, epic.ID))
			output.WriteString(fmt.Sprintf("**%s** [%s]\n\n", epic.Title, epic.Status))

			if epic.Description.Valid && epic.Description.String != "" {
				descLines := strings.Split(epic.Description.String, "\n")
				if len(descLines) > 5 {
					output.WriteString(strings.Join(descLines[:5], "\n"))
					output.WriteString("\n\n*(Description truncated)*\n\n")
				} else {
					output.WriteString(epic.Description.String)
					output.WriteString("\n\n")
				}
			}

			// Show ready tasks for this epic
			tasks, err := models.ListTasks(epic.MissionID, epic.ID, "ready")
			if err == nil && len(tasks) > 0 {
				output.WriteString("**Ready Tasks**:\n")
				for _, task := range tasks {
					output.WriteString(fmt.Sprintf("- %s - %s\n", task.ID, task.Title))
				}
				output.WriteString("\n")
			}
		}
	} else {
		output.WriteString("*No epics currently assigned to this grove.*\n\n")
		output.WriteString("Run `orc summary` to see the full mission tree.\n\n")
	}

	// Section 3: Handoff Context
	ho, err := models.GetLatestHandoffForGrove(groveCtx.GroveID)
	if err == nil && ho != nil {
		output.WriteString("## Handoff Context\n\n")
		output.WriteString(fmt.Sprintf("**From**: Previous session (%s)\n", ho.CreatedAt.Format("Jan 2, 15:04")))
		output.WriteString(fmt.Sprintf("**ID**: %s\n\n", ho.ID))

		noteLines := strings.Split(ho.HandoffNote, "\n")
		if len(noteLines) > 8 {
			output.WriteString(strings.Join(noteLines[:8], "\n"))
			output.WriteString("\n\n*(Showing first 8 lines)*\n\n")
		} else {
			output.WriteString(ho.HandoffNote)
			output.WriteString("\n\n")
		}
	}

	// Section 4: ORC CLI Primer
	output.WriteString("## ORC CLI Primer\n\n")
	output.WriteString("**Core Commands**:\n")
	output.WriteString("- `orc summary` - View mission tree and current assignments\n")
	output.WriteString("- `orc task list --epic EPIC-ID` - List tasks for your epic\n")
	output.WriteString("- `orc task show TASK-ID` - View task details\n")
	output.WriteString("- `orc task complete TASK-ID` - Mark task as completed\n")
	output.WriteString("- `orc epic check-assignment` - Verify your epic assignment\n")
	output.WriteString("- `/handoff` - Create session handoff note (via Claude Code skill)\n\n")

	// Section 5: Core Rules
	output.WriteString("## Core Rules\n\n")
	output.WriteString("- **Track all work in ORC ledger** - Use `orc task` commands to manage progress\n")
	output.WriteString("- **TodoWrite tool is NOT ALLOWED** - This tool is banned; use ORC ledger instead\n")
	output.WriteString("- **TODO markdown files are NOT ALLOWED** - No TODO.md or similar files\n")
	output.WriteString("- **Stay in grove territory** - Work within assigned epics only\n")
	output.WriteString("- **Handoff between sessions** - Use `/handoff` skill before ending work\n\n")

	// Footer
	output.WriteString("---\n")
	output.WriteString("💡 **You are an IMP** - Implementation agent working within a grove on assigned epics.\n")
	output.WriteString("🎯 **Focus**: Complete ready tasks, update ledger, maintain context through handoffs.\n")

	return output.String()
}

// getGitContext returns git context if in a git repository
func getGitContext() string {
	var output strings.Builder

	// Check if in git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return "" // Not in git repo
	}

	output.WriteString("**Git Context**:\n")

	// Get current branch
	cmd = exec.Command("git", "branch", "--show-current")
	if branchBytes, err := cmd.Output(); err == nil {
		branch := strings.TrimSpace(string(branchBytes))
		if branch != "" {
			output.WriteString(fmt.Sprintf("- Branch: `%s`\n", branch))
		}
	}

	// Get recent commits (last 3)
	cmd = exec.Command("git", "log", "--oneline", "-3")
	if commitBytes, err := cmd.Output(); err == nil {
		commits := strings.TrimSpace(string(commitBytes))
		if commits != "" {
			lines := strings.Split(commits, "\n")
			output.WriteString("- Recent commits:\n")
			for _, line := range lines {
				if line != "" {
					output.WriteString(fmt.Sprintf("  - %s\n", line))
				}
			}
		}
	}

	// Check for uncommitted changes
	cmd = exec.Command("git", "status", "--short")
	if statusBytes, err := cmd.Output(); err == nil {
		status := strings.TrimSpace(string(statusBytes))
		if status != "" {
			lines := strings.Split(status, "\n")
			output.WriteString(fmt.Sprintf("- Uncommitted changes: %d files\n", len(lines)))
		}
	}

	output.WriteString("\n")
	return output.String()
}
