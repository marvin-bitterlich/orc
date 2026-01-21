package primary

import "context"

// ConclaveService defines the primary port for conclave operations.
type ConclaveService interface {
	// CreateConclave creates a new conclave (ideation session).
	CreateConclave(ctx context.Context, req CreateConclaveRequest) (*CreateConclaveResponse, error)

	// GetConclave retrieves a conclave by ID.
	GetConclave(ctx context.Context, conclaveID string) (*Conclave, error)

	// ListConclaves lists conclaves with optional filters.
	ListConclaves(ctx context.Context, filters ConclaveFilters) ([]*Conclave, error)

	// CompleteConclave marks a conclave as complete.
	CompleteConclave(ctx context.Context, conclaveID string) error

	// PauseConclave pauses an active conclave.
	PauseConclave(ctx context.Context, conclaveID string) error

	// ResumeConclave resumes a paused conclave.
	ResumeConclave(ctx context.Context, conclaveID string) error

	// UpdateConclave updates a conclave's title and/or description.
	UpdateConclave(ctx context.Context, req UpdateConclaveRequest) error

	// PinConclave pins a conclave to prevent completion.
	PinConclave(ctx context.Context, conclaveID string) error

	// UnpinConclave unpins a conclave.
	UnpinConclave(ctx context.Context, conclaveID string) error

	// DeleteConclave deletes a conclave.
	DeleteConclave(ctx context.Context, conclaveID string) error

	// GetConclavesByGrove retrieves conclaves assigned to a grove.
	GetConclavesByGrove(ctx context.Context, workbenchID string) ([]*Conclave, error)

	// GetConclaveTasks retrieves all tasks in a conclave.
	GetConclaveTasks(ctx context.Context, conclaveID string) ([]*ConclaveTask, error)

	// GetConclaveQuestions retrieves all questions in a conclave.
	GetConclaveQuestions(ctx context.Context, conclaveID string) ([]*ConclaveQuestion, error)

	// GetConclavePlans retrieves all plans in a conclave.
	GetConclavePlans(ctx context.Context, conclaveID string) ([]*ConclavePlan, error)
}

// CreateConclaveRequest contains parameters for creating a conclave.
type CreateConclaveRequest struct {
	CommissionID string
	Title        string
	Description  string
}

// CreateConclaveResponse contains the result of creating a conclave.
type CreateConclaveResponse struct {
	ConclaveID string
	Conclave   *Conclave
}

// UpdateConclaveRequest contains parameters for updating a conclave.
type UpdateConclaveRequest struct {
	ConclaveID  string
	Title       string
	Description string
}

// Conclave represents a conclave entity at the port boundary.
type Conclave struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         string
	Status              string
	AssignedWorkbenchID string
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string
}

// ConclaveFilters contains filter options for listing conclaves.
type ConclaveFilters struct {
	CommissionID string
	Status       string
}

// ConclaveTask represents a task associated with a conclave.
type ConclaveTask struct {
	ID                  string
	ShipmentID          string
	CommissionID        string
	Title               string
	Description         string
	Type                string
	Status              string
	Priority            string
	AssignedWorkbenchID string
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	ClaimedAt           string
	CompletedAt         string
	ConclaveID          string
	PromotedFromID      string
	PromotedFromType    string
}

// ConclaveQuestion represents a question associated with a conclave.
type ConclaveQuestion struct {
	ID               string
	InvestigationID  string
	CommissionID     string
	Title            string
	Description      string
	Status           string
	Answer           string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	AnsweredAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// ConclavePlan represents a plan associated with a conclave.
type ConclavePlan struct {
	ID               string
	ShipmentID       string
	CommissionID     string
	Title            string
	Description      string
	Status           string
	Content          string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ApprovedAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}
