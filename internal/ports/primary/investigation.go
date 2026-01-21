package primary

import "context"

// InvestigationService defines the primary port for investigation operations.
type InvestigationService interface {
	// CreateInvestigation creates a new investigation (research container).
	CreateInvestigation(ctx context.Context, req CreateInvestigationRequest) (*CreateInvestigationResponse, error)

	// GetInvestigation retrieves an investigation by ID.
	GetInvestigation(ctx context.Context, investigationID string) (*Investigation, error)

	// ListInvestigations lists investigations with optional filters.
	ListInvestigations(ctx context.Context, filters InvestigationFilters) ([]*Investigation, error)

	// CompleteInvestigation marks an investigation as complete.
	CompleteInvestigation(ctx context.Context, investigationID string) error

	// PauseInvestigation pauses an active investigation.
	PauseInvestigation(ctx context.Context, investigationID string) error

	// ResumeInvestigation resumes a paused investigation.
	ResumeInvestigation(ctx context.Context, investigationID string) error

	// UpdateInvestigation updates an investigation's title and/or description.
	UpdateInvestigation(ctx context.Context, req UpdateInvestigationRequest) error

	// PinInvestigation pins an investigation to prevent completion.
	PinInvestigation(ctx context.Context, investigationID string) error

	// UnpinInvestigation unpins an investigation.
	UnpinInvestigation(ctx context.Context, investigationID string) error

	// DeleteInvestigation deletes an investigation.
	DeleteInvestigation(ctx context.Context, investigationID string) error

	// AssignInvestigationToGrove assigns an investigation to a grove.
	AssignInvestigationToGrove(ctx context.Context, investigationID, groveID string) error

	// GetInvestigationsByGrove retrieves investigations assigned to a grove.
	GetInvestigationsByGrove(ctx context.Context, groveID string) ([]*Investigation, error)

	// GetInvestigationQuestions retrieves all questions in an investigation.
	GetInvestigationQuestions(ctx context.Context, investigationID string) ([]*InvestigationQuestion, error)
}

// CreateInvestigationRequest contains parameters for creating an investigation.
type CreateInvestigationRequest struct {
	MissionID   string
	Title       string
	Description string
}

// CreateInvestigationResponse contains the result of creating an investigation.
type CreateInvestigationResponse struct {
	InvestigationID string
	Investigation   *Investigation
}

// UpdateInvestigationRequest contains parameters for updating an investigation.
type UpdateInvestigationRequest struct {
	InvestigationID string
	Title           string
	Description     string
}

// Investigation represents an investigation entity at the port boundary.
type Investigation struct {
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

// InvestigationFilters contains filter options for listing investigations.
type InvestigationFilters struct {
	MissionID string
	Status    string
}

// InvestigationQuestion represents a question associated with an investigation.
type InvestigationQuestion struct {
	ID               string
	InvestigationID  string
	MissionID        string
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
