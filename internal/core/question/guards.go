// Package question contains the pure business logic for question operations.
// Guards are pure functions that evaluate preconditions without side effects.
package question

import "fmt"

// GuardResult represents the outcome of a guard evaluation.
type GuardResult struct {
	Allowed bool
	Reason  string
}

// Error converts the guard result to an error if not allowed.
func (r GuardResult) Error() error {
	if r.Allowed {
		return nil
	}
	return fmt.Errorf("%s", r.Reason)
}

// CreateQuestionContext provides context for question creation guards.
type CreateQuestionContext struct {
	MissionID           string
	MissionExists       bool
	InvestigationID     string // Optional - empty string means no investigation
	InvestigationExists bool   // Only checked if InvestigationID != ""
}

// AnswerQuestionContext provides context for question answering guards.
type AnswerQuestionContext struct {
	QuestionID string
	Status     string // "open", "answered"
	IsPinned   bool
}

// CanCreateQuestion evaluates whether a question can be created.
// Rules:
// - Mission must exist
// - Investigation must exist if provided
func CanCreateQuestion(ctx CreateQuestionContext) GuardResult {
	// Check mission exists
	if !ctx.MissionExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("mission %s not found", ctx.MissionID),
		}
	}

	// Check investigation exists if provided
	if ctx.InvestigationID != "" && !ctx.InvestigationExists {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("investigation %s not found", ctx.InvestigationID),
		}
	}

	return GuardResult{Allowed: true}
}

// CanAnswerQuestion evaluates whether a question can be answered.
// Rules:
// - Status must be "open" (cannot re-answer)
// - Question must not be pinned
func CanAnswerQuestion(ctx AnswerQuestionContext) GuardResult {
	// Check status is open (must check first - answered questions can't be answered regardless of pinned)
	if ctx.Status != "open" {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("can only answer open questions (current status: %s)", ctx.Status),
		}
	}

	// Check not pinned
	if ctx.IsPinned {
		return GuardResult{
			Allowed: false,
			Reason:  fmt.Sprintf("cannot answer pinned question %s. Unpin first with: orc question unpin %s", ctx.QuestionID, ctx.QuestionID),
		}
	}

	return GuardResult{Allowed: true}
}
