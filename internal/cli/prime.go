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
		Long: `Generate concise context snapshot for injecting into new Claude sessions.

Designed for SessionStart hook injection after /clear - provides just enough
context to orient the new session without overwhelming it.

Output includes:
- Current location (cwd)
- Mission context (if any)
- Active work order (if any)
- Latest handoff note (brief)

This is NOT a replacement for /g-bootstrap (which provides full context).
This is for automatic injection after /clear to maintain basic orientation.

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

	// Detect mission context
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

			// Get active work order
			workOrders, err := models.ListWorkOrders(missionCtx.MissionID, "in_progress")
			if err == nil && len(workOrders) > 0 {
				wo := workOrders[0]
				output.WriteString(fmt.Sprintf("### Active Work: %s\n\n", wo.ID))
				output.WriteString(fmt.Sprintf("**%s** [%s]\n\n", wo.Title, wo.Status))
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
					output.WriteString("\n\n*(Use `/g-bootstrap` for full context)*\n\n")
				} else {
					output.WriteString(ho.HandoffNote)
					output.WriteString("\n\n")
				}
			}
		}
	} else {
		output.WriteString("**Context**: Master orchestrator (global)\n\n")
		output.WriteString("Run `orc mission list` to see available missions.\n\n")
		output.WriteString("Use `/g-bootstrap` to load full session context.\n")
	}

	// Footer note
	output.WriteString("\n---\n")
	output.WriteString("💡 **Note**: This is lightweight orientation context.\n")
	output.WriteString("   For full context (Graphiti memory + deep analysis), use `/g-bootstrap`\n")

	// Truncate to max lines if needed
	fullOutput := output.String()
	if format == "text" {
		lines := strings.Split(fullOutput, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, "...", "", "*(Output truncated - use /g-bootstrap for full context)*")
		}
		fullOutput = strings.Join(lines, "\n")
	}

	fmt.Println(fullOutput)
	return nil
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
