package primary

import "context"

// OperationService defines the primary port for operation operations.
// Operations are minimal entities with no Delete or pinning operations.
type OperationService interface {
	// CreateOperation creates a new operation.
	CreateOperation(ctx context.Context, req CreateOperationRequest) (*CreateOperationResponse, error)

	// GetOperation retrieves an operation by ID.
	GetOperation(ctx context.Context, operationID string) (*Operation, error)

	// ListOperations lists operations with optional filters.
	ListOperations(ctx context.Context, filters OperationFilters) ([]*Operation, error)

	// UpdateOperationStatus updates the status of an operation.
	UpdateOperationStatus(ctx context.Context, operationID, status string) error

	// CompleteOperation marks an operation as complete.
	CompleteOperation(ctx context.Context, operationID string) error
}

// CreateOperationRequest contains parameters for creating an operation.
type CreateOperationRequest struct {
	MissionID   string
	Title       string
	Description string
}

// CreateOperationResponse contains the result of creating an operation.
type CreateOperationResponse struct {
	OperationID string
	Operation   *Operation
}

// Operation represents an operation entity at the port boundary.
type Operation struct {
	ID          string
	MissionID   string
	Title       string
	Description string
	Status      string
	CreatedAt   string
	UpdatedAt   string
	CompletedAt string
}

// OperationFilters contains filter options for listing operations.
type OperationFilters struct {
	MissionID string
	Status    string
}
