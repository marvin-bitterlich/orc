package primary

import "context"

// TomeService defines the primary port for tome operations.
type TomeService interface {
	// CreateTome creates a new tome (knowledge container).
	CreateTome(ctx context.Context, req CreateTomeRequest) (*CreateTomeResponse, error)

	// GetTome retrieves a tome by ID.
	GetTome(ctx context.Context, tomeID string) (*Tome, error)

	// ListTomes lists tomes with optional filters.
	ListTomes(ctx context.Context, filters TomeFilters) ([]*Tome, error)

	// CloseTome marks a tome as closed.
	CloseTome(ctx context.Context, tomeID string) error

	// UpdateTome updates a tome's title and/or description.
	UpdateTome(ctx context.Context, req UpdateTomeRequest) error

	// PinTome pins a tome to prevent completion.
	PinTome(ctx context.Context, tomeID string) error

	// UnpinTome unpins a tome.
	UnpinTome(ctx context.Context, tomeID string) error

	// DeleteTome deletes a tome.
	DeleteTome(ctx context.Context, tomeID string) error

	// AssignTomeToWorkbench assigns a tome to a workbench.
	AssignTomeToWorkbench(ctx context.Context, tomeID, workbenchID string) error

	// GetTomesByWorkbench retrieves tomes assigned to a workbench.
	GetTomesByWorkbench(ctx context.Context, workbenchID string) ([]*Tome, error)

	// GetTomeNotes retrieves all notes in a tome.
	GetTomeNotes(ctx context.Context, tomeID string) ([]*Note, error)

	// UnparkTome moves a tome from commission root to a specific Conclave.
	UnparkTome(ctx context.Context, tomeID, conclaveID string) error
}

// CreateTomeRequest contains parameters for creating a tome.
type CreateTomeRequest struct {
	CommissionID  string
	ConclaveID    string // DEPRECATED: use ContainerID/ContainerType instead
	Title         string
	Description   string
	ContainerID   string // Optional: CON-xxx (empty for root tome)
	ContainerType string // Optional: "conclave" (empty for root tome)
}

// CreateTomeResponse contains the result of creating a tome.
type CreateTomeResponse struct {
	TomeID string
	Tome   *Tome
}

// UpdateTomeRequest contains parameters for updating a tome.
type UpdateTomeRequest struct {
	TomeID      string
	Title       string
	Description string
}

// Tome represents a tome entity at the port boundary.
type Tome struct {
	ID                  string
	CommissionID        string
	ConclaveID          string // DEPRECATED: use ContainerID/ContainerType instead
	Title               string
	Description         string
	Status              string
	AssignedWorkbenchID string
	Pinned              bool
	ContainerID         string // CON-xxx
	ContainerType       string // "conclave"
	CreatedAt           string
	UpdatedAt           string
	ClosedAt            string
}

// TomeFilters contains filter options for listing tomes.
type TomeFilters struct {
	CommissionID string
	ConclaveID   string
	Status       string
}
