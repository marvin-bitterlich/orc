package primary

import "context"

// PatrolService defines the primary port for patrol operations.
// Patrols track monitoring sessions for kennels.
type PatrolService interface {
	// StartPatrol starts a new patrol for a workbench.
	// Resolves the kennel for the workbench and derives the TMux target.
	StartPatrol(ctx context.Context, workbenchID string) (*Patrol, error)

	// EndPatrol ends an active patrol.
	EndPatrol(ctx context.Context, patrolID string) error

	// GetPatrol retrieves a patrol by ID.
	GetPatrol(ctx context.Context, patrolID string) (*Patrol, error)

	// GetActivePatrolForKennel retrieves the active patrol for a kennel (if any).
	GetActivePatrolForKennel(ctx context.Context, kennelID string) (*Patrol, error)

	// ListPatrols lists patrols with optional filters.
	ListPatrols(ctx context.Context, filters PatrolFilters) ([]*Patrol, error)
}

// Patrol represents a patrol entity at the port boundary.
type Patrol struct {
	ID        string
	KennelID  string
	Target    string
	Status    string
	Config    string
	StartedAt string
	EndedAt   string
	CreatedAt string
	UpdatedAt string
}

// PatrolFilters contains filter options for listing patrols.
type PatrolFilters struct {
	KennelID string
	Status   string
}

// Patrol status constants.
const (
	PatrolStatusActive    = "active"
	PatrolStatusCompleted = "completed"
	PatrolStatusEscalated = "escalated"
)
