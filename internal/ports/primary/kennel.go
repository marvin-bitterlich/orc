package primary

import "context"

// KennelService defines the primary port for kennel operations.
// Kennels are 1:1 with workbenches (Watchdog seat).
type KennelService interface {
	// GetKennel retrieves a kennel by ID.
	GetKennel(ctx context.Context, kennelID string) (*Kennel, error)

	// GetKennelByWorkbench retrieves a kennel by workbench ID.
	GetKennelByWorkbench(ctx context.Context, workbenchID string) (*Kennel, error)

	// ListKennels lists kennels with optional filters.
	ListKennels(ctx context.Context, filters KennelFilters) ([]*Kennel, error)

	// CreateKennel creates a new kennel for a workbench.
	// Returns error if workbench already has a kennel.
	CreateKennel(ctx context.Context, workbenchID string) (*Kennel, error)

	// UpdateKennelStatus updates the status of a kennel.
	UpdateKennelStatus(ctx context.Context, kennelID, status string) error

	// EnsureAllWorkbenchesHaveKennels creates kennels for any workbenches missing them.
	// Used for data migration when introducing the kennel entity.
	EnsureAllWorkbenchesHaveKennels(ctx context.Context) ([]string, error)
}

// Kennel represents a kennel entity at the port boundary.
type Kennel struct {
	ID          string
	WorkbenchID string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}

// KennelFilters contains filter options for listing kennels.
type KennelFilters struct {
	WorkbenchID string
	Status      string
}

// Kennel status constants
const (
	KennelStatusVacant   = "vacant"
	KennelStatusOccupied = "occupied"
	KennelStatusAway     = "away"
)
