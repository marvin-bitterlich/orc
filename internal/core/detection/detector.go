// Package detection provides pane content classification for watchdog monitoring.
package detection

import (
	"regexp"
	"strings"
)

// Outcome constants for pane content classification.
const (
	OutcomeWorking = "working" // Claude is actively processing
	OutcomeIdle    = "idle"    // Waiting for input, no activity
	OutcomeMenu    = "menu"    // Interactive menu displayed
	OutcomeTyped   = "typed"   // Text typed but not submitted
	OutcomeError   = "error"   // Error condition detected
)

// ANSI escape code pattern for stripping colors.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// Menu patterns - Claude is showing an interactive menu.
var menuPatterns = []string{
	"Would you like to proceed?",
	"Yes, clear context",
	"Yes, auto-accept edits",
	"shift+tab",
	"❯ 1.",
	"❯ 2.",
}

// Working patterns - Claude is actively processing.
var workingPatterns = []string{
	"Thundering",
	"Kneading",
	"Contemplating",
	"Cogitating",
	"Ruminating",
	"Pondering",
	"(esc to interrupt)",
	"✶",
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏", // Spinner chars
}

// Idle patterns - Claude has finished and is waiting.
var idlePatterns = []string{
	"Cogitated for",
	"✻ Cogitated",
	"───────────────────",
}

// Error patterns - Something went wrong.
var errorPatterns = []string{
	"Error:",
	"ERROR:",
	"error:",
	"panic:",
	"PANIC:",
	"FAILED",
	"Failed",
	"fatal:",
	"FATAL:",
	"Exception:",
	"Traceback",
	"stack trace",
}

// DetectOutcome classifies pane content into an outcome category.
// Priority order: error > menu > working > typed > idle
func DetectOutcome(content string) string {
	// Strip ANSI codes for cleaner matching
	clean := stripANSI(content)

	// Check for errors first (highest priority)
	if detectError(clean) {
		return OutcomeError
	}

	// Check for menu prompt
	if detectMenu(clean) {
		return OutcomeMenu
	}

	// Check for active work
	if detectWorking(clean) {
		return OutcomeWorking
	}

	// Check for typed but not submitted
	if detectTyped(clean) {
		return OutcomeTyped
	}

	// Default to idle if completed message present or empty prompt
	if detectIdle(clean) {
		return OutcomeIdle
	}

	// If nothing matches, assume working (safest default)
	return OutcomeWorking
}

// stripANSI removes ANSI escape codes from content.
func stripANSI(content string) string {
	return ansiPattern.ReplaceAllString(content, "")
}

// detectError checks for error patterns in content.
func detectError(content string) bool {
	for _, pattern := range errorPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// detectMenu checks for interactive menu patterns.
func detectMenu(content string) bool {
	for _, pattern := range menuPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// detectWorking checks for active processing patterns.
func detectWorking(content string) bool {
	for _, pattern := range workingPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// detectIdle checks for idle/waiting patterns.
func detectIdle(content string) bool {
	// Check for completion message
	for _, pattern := range idlePatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	// Check for empty prompt at end
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) > 0 {
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		// Empty prompt line
		if lastLine == "❯" || lastLine == ">" || lastLine == "$" || lastLine == "#" {
			return true
		}
	}

	return false
}

// detectTyped checks if there's text typed but not submitted.
// This is detected when the last line has a prompt with text after it.
func detectTyped(content string) bool {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 {
		return false
	}

	lastLine := lines[len(lines)-1]

	// Check for prompt followed by text
	prompts := []string{"❯ ", "> ", "$ ", "# "}
	for _, prompt := range prompts {
		if strings.Contains(lastLine, prompt) {
			// Find text after prompt
			idx := strings.Index(lastLine, prompt)
			afterPrompt := strings.TrimSpace(lastLine[idx+len(prompt):])
			if afterPrompt != "" {
				return true
			}
		}
	}

	return false
}

// DetectOutcomeWithContext uses previous outcome for better classification.
// For example, typed detection requires comparing to previous check.
func DetectOutcomeWithContext(content, previousContent string, previousOutcome string) string {
	outcome := DetectOutcome(content)

	// If current looks typed but previous was also typed with same content,
	// it's been typed for multiple checks - keep as typed
	if outcome == OutcomeTyped && previousOutcome == OutcomeTyped {
		// Same typed content means user hasn't submitted
		return OutcomeTyped
	}

	return outcome
}
