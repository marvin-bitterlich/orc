package primary

import "context"

// EscalationService defines the primary port for escalation operations.
type EscalationService interface {
	// GetEscalation retrieves an escalation by ID.
	GetEscalation(ctx context.Context, escalationID string) (*Escalation, error)

	// ListEscalations lists escalations with optional filters.
	ListEscalations(ctx context.Context, filters EscalationFilters) ([]*Escalation, error)
}

// Escalation represents an escalation entity at the port boundary.
type Escalation struct {
	ID            string
	ApprovalID    string // May be empty
	PlanID        string
	TaskID        string
	Reason        string
	Status        string // 'pending', 'resolved', 'dismissed'
	RoutingRule   string
	OriginActorID string
	TargetActorID string // May be empty
	Resolution    string // May be empty
	ResolvedBy    string // May be empty
	CreatedAt     string
	ResolvedAt    string // May be empty
}

// EscalationFilters contains filter options for listing escalations.
type EscalationFilters struct {
	TaskID        string
	Status        string
	TargetActorID string
}

// Escalation status constants
const (
	EscalationStatusPending   = "pending"
	EscalationStatusResolved  = "resolved"
	EscalationStatusDismissed = "dismissed"
)
