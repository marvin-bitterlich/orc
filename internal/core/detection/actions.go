package detection

import (
	"context"
	"fmt"
)

// Action types for watchdog interventions.
const (
	ActionNone     = "none"     // No action needed
	ActionBTab     = "btab"     // Send BTab (shift+tab) for menu
	ActionEnter    = "enter"    // Send Enter for typed command
	ActionNudge    = "nudge"    // Send nudge message for idle
	ActionEscalate = "escalate" // Escalate to gatehouse
)

// Action represents a watchdog intervention.
type Action struct {
	Type    string // ActionNone, ActionBTab, ActionEnter, ActionNudge, ActionEscalate
	Target  string // TMux target (session:window.pane)
	Message string // For nudge/escalate actions
}

// ActionExecutor interface for executing watchdog actions.
// This allows mocking in tests.
type ActionExecutor interface {
	SendKeys(ctx context.Context, target, keys string) error
	NudgeSession(ctx context.Context, target, message string) error
}

// SelectAction determines the appropriate action for a given outcome.
// stuckCount is used for escalation threshold checking.
func SelectAction(outcome string, target string, stuckCount int, stuckThreshold int) Action {
	switch outcome {
	case OutcomeMenu:
		return Action{
			Type:   ActionBTab,
			Target: target,
		}

	case OutcomeTyped:
		return Action{
			Type:   ActionEnter,
			Target: target,
		}

	case OutcomeIdle:
		return Action{
			Type:    ActionNudge,
			Target:  target,
			Message: "You appear to be idle. Continue with the current task or run /imp-nudge for guidance.",
		}

	case OutcomeError:
		if stuckCount >= stuckThreshold {
			return Action{
				Type:    ActionEscalate,
				Target:  target,
				Message: fmt.Sprintf("IMP has been stuck on error for %d checks. Needs human intervention.", stuckCount),
			}
		}
		// Below threshold, try a nudge first
		return Action{
			Type:    ActionNudge,
			Target:  target,
			Message: "An error was detected. Please review and fix the issue, then continue.",
		}

	case OutcomeWorking:
		return Action{
			Type:   ActionNone,
			Target: target,
		}

	default:
		return Action{
			Type:   ActionNone,
			Target: target,
		}
	}
}

// ExecuteAction performs the selected action using the provided executor.
func ExecuteAction(ctx context.Context, executor ActionExecutor, action Action) error {
	switch action.Type {
	case ActionNone:
		return nil

	case ActionBTab:
		// BTab is Shift+Tab in tmux
		return executor.SendKeys(ctx, action.Target, "BTab")

	case ActionEnter:
		return executor.SendKeys(ctx, action.Target, "Enter")

	case ActionNudge:
		return executor.NudgeSession(ctx, action.Target, action.Message)

	case ActionEscalate:
		// Escalation is handled by the caller (patrol service)
		// This just signals that escalation is needed
		return nil

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// DefaultStuckThreshold is the number of consecutive failures before escalation.
const DefaultStuckThreshold = 5

// NudgeMessages provides context-aware nudge messages based on outcome.
var NudgeMessages = map[string]string{
	OutcomeIdle:  "You appear to be idle. Continue with the current task or run /imp-nudge for guidance.",
	OutcomeError: "An error was detected. Please review and fix the issue, then continue.",
}
