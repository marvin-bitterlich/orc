package primary

import "context"

// StuckService defines the primary port for stuck tracking operations.
// Tracks consecutive failures and coordinates escalation.
type StuckService interface {
	// ProcessOutcome processes a check outcome and manages stuck state.
	// Returns the result indicating what action to take.
	ProcessOutcome(ctx context.Context, patrolID, checkID, outcome string) (*StuckResult, error)

	// GetStuck retrieves a stuck by ID.
	GetStuck(ctx context.Context, stuckID string) (*Stuck, error)

	// GetOpenStuck retrieves the open stuck for a patrol (if any).
	GetOpenStuck(ctx context.Context, patrolID string) (*Stuck, error)

	// ResolveStuck marks a stuck as resolved.
	ResolveStuck(ctx context.Context, stuckID string) error

	// EscalateStuck marks a stuck as escalated.
	EscalateStuck(ctx context.Context, stuckID string) error
}

// StuckResult contains the result of processing an outcome.
type StuckResult struct {
	// Action is the recommended action to take (from detection.ActionXxx constants).
	Action string

	// StuckID is the ID of the current stuck episode (if any).
	StuckID string

	// CheckCount is the current consecutive failure count.
	CheckCount int

	// NeedsEscalation indicates if escalation threshold was reached.
	NeedsEscalation bool

	// Message is the action-specific message (for nudge/escalate).
	Message string
}

// Stuck represents a stuck episode at the port boundary.
type Stuck struct {
	ID           string
	PatrolID     string
	FirstCheckID string
	CheckCount   int
	Status       string
	ResolvedAt   string
	CreatedAt    string
	UpdatedAt    string
}

// Stuck status constants.
const (
	StuckStatusOpen      = "open"
	StuckStatusResolved  = "resolved"
	StuckStatusEscalated = "escalated"
)
