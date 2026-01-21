package primary

import "context"

// HandoffService defines the primary port for handoff operations.
// Handoffs are immutable - no update or delete operations.
type HandoffService interface {
	// CreateHandoff creates a new handoff note for session continuity.
	CreateHandoff(ctx context.Context, req CreateHandoffRequest) (*CreateHandoffResponse, error)

	// GetHandoff retrieves a handoff by ID.
	GetHandoff(ctx context.Context, handoffID string) (*Handoff, error)

	// GetLatestHandoff retrieves the most recent handoff.
	GetLatestHandoff(ctx context.Context) (*Handoff, error)

	// GetLatestHandoffForGrove retrieves the most recent handoff for a grove.
	GetLatestHandoffForGrove(ctx context.Context, groveID string) (*Handoff, error)

	// ListHandoffs lists handoffs with optional limit.
	ListHandoffs(ctx context.Context, limit int) ([]*Handoff, error)
}

// CreateHandoffRequest contains parameters for creating a handoff.
type CreateHandoffRequest struct {
	HandoffNote     string
	ActiveMissionID string
	ActiveGroveID   string
	TodosSnapshot   string // JSON snapshot of todos state
}

// CreateHandoffResponse contains the result of creating a handoff.
type CreateHandoffResponse struct {
	HandoffID string
	Handoff   *Handoff
}

// Handoff represents a handoff entity at the port boundary.
type Handoff struct {
	ID              string
	CreatedAt       string
	HandoffNote     string
	ActiveMissionID string
	ActiveGroveID   string
	TodosSnapshot   string
}
