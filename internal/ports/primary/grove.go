package primary

import "context"

// GroveService defines the primary port for grove operations.
type GroveService interface {
	// CreateGrove creates a new grove for a mission.
	CreateGrove(ctx context.Context, req CreateGroveRequest) (*CreateGroveResponse, error)

	// OpenGrove opens a grove in a TMux window with IMP layout.
	OpenGrove(ctx context.Context, req OpenGroveRequest) (*OpenGroveResponse, error)

	// GetGrove retrieves a grove by ID.
	GetGrove(ctx context.Context, groveID string) (*Grove, error)

	// GetGroveByPath retrieves a grove by its filesystem path.
	GetGroveByPath(ctx context.Context, path string) (*Grove, error)

	// ListGroves lists groves with optional filters.
	ListGroves(ctx context.Context, filters GroveFilters) ([]*Grove, error)

	// RenameGrove renames a grove.
	RenameGrove(ctx context.Context, req RenameGroveRequest) error

	// UpdateGrovePath updates the filesystem path of a grove.
	UpdateGrovePath(ctx context.Context, groveID, newPath string) error

	// DeleteGrove deletes a grove.
	DeleteGrove(ctx context.Context, req DeleteGroveRequest) error
}

// CreateGroveRequest contains parameters for creating a grove.
type CreateGroveRequest struct {
	Name      string
	MissionID string   // Required or derived from context
	Repos     []string // Optional repository names
	BasePath  string   // Optional, defaults to ~/src/worktrees
}

// CreateGroveResponse contains the result of grove creation.
type CreateGroveResponse struct {
	GroveID string
	Grove   *Grove
	Path    string // Materialized grove path
}

// OpenGroveRequest contains parameters for opening a grove.
type OpenGroveRequest struct {
	GroveID string
}

// OpenGroveResponse contains the result of opening a grove.
type OpenGroveResponse struct {
	Grove       *Grove
	SessionName string
	WindowName  string
}

// RenameGroveRequest contains parameters for renaming a grove.
type RenameGroveRequest struct {
	GroveID      string
	NewName      string
	UpdateConfig bool // Also update .orc/config.json
}

// DeleteGroveRequest contains parameters for deleting a grove.
type DeleteGroveRequest struct {
	GroveID        string
	Force          bool
	RemoveWorktree bool // Also remove filesystem worktree
}

// Grove represents a grove entity at the port boundary.
type Grove struct {
	ID        string
	Name      string
	MissionID string
	Path      string
	Repos     []string
	Status    string
	CreatedAt string
}

// GroveFilters contains filter options for listing groves.
type GroveFilters struct {
	MissionID string
	Status    string
	Limit     int
}
