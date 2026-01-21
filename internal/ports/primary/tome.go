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

	// CompleteTome marks a tome as complete.
	CompleteTome(ctx context.Context, tomeID string) error

	// PauseTome pauses an active tome.
	PauseTome(ctx context.Context, tomeID string) error

	// ResumeTome resumes a paused tome.
	ResumeTome(ctx context.Context, tomeID string) error

	// UpdateTome updates a tome's title and/or description.
	UpdateTome(ctx context.Context, req UpdateTomeRequest) error

	// PinTome pins a tome to prevent completion.
	PinTome(ctx context.Context, tomeID string) error

	// UnpinTome unpins a tome.
	UnpinTome(ctx context.Context, tomeID string) error

	// DeleteTome deletes a tome.
	DeleteTome(ctx context.Context, tomeID string) error

	// AssignTomeToGrove assigns a tome to a grove.
	AssignTomeToGrove(ctx context.Context, tomeID, groveID string) error

	// GetTomesByGrove retrieves tomes assigned to a grove.
	GetTomesByGrove(ctx context.Context, groveID string) ([]*Tome, error)

	// GetTomeNotes retrieves all notes in a tome.
	GetTomeNotes(ctx context.Context, tomeID string) ([]*Note, error)
}

// CreateTomeRequest contains parameters for creating a tome.
type CreateTomeRequest struct {
	MissionID   string
	Title       string
	Description string
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
	ID              string
	MissionID       string
	Title           string
	Description     string
	Status          string
	AssignedGroveID string
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string
}

// TomeFilters contains filter options for listing tomes.
type TomeFilters struct {
	MissionID string
	Status    string
}
