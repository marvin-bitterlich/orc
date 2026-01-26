package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// OperationServiceImpl implements the OperationService interface.
type OperationServiceImpl struct {
	operationRepo secondary.OperationRepository
}

// NewOperationService creates a new OperationService with injected dependencies.
func NewOperationService(
	operationRepo secondary.OperationRepository,
) *OperationServiceImpl {
	return &OperationServiceImpl{
		operationRepo: operationRepo,
	}
}

// CreateOperation creates a new operation.
func (s *OperationServiceImpl) CreateOperation(ctx context.Context, req primary.CreateOperationRequest) (*primary.CreateOperationResponse, error) {
	// Validate commission exists
	exists, err := s.operationRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.operationRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate operation ID: %w", err)
	}

	// Create record
	record := &secondary.OperationRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       "ready",
	}

	if err := s.operationRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	// Fetch created operation
	created, err := s.operationRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created operation: %w", err)
	}

	return &primary.CreateOperationResponse{
		OperationID: created.ID,
		Operation:   s.recordToOperation(created),
	}, nil
}

// GetOperation retrieves an operation by ID.
func (s *OperationServiceImpl) GetOperation(ctx context.Context, operationID string) (*primary.Operation, error) {
	record, err := s.operationRepo.GetByID(ctx, operationID)
	if err != nil {
		return nil, err
	}
	return s.recordToOperation(record), nil
}

// ListOperations lists operations with optional filters.
func (s *OperationServiceImpl) ListOperations(ctx context.Context, filters primary.OperationFilters) ([]*primary.Operation, error) {
	records, err := s.operationRepo.List(ctx, secondary.OperationFilters{
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}

	operations := make([]*primary.Operation, len(records))
	for i, r := range records {
		operations[i] = s.recordToOperation(r)
	}
	return operations, nil
}

// UpdateOperationStatus updates the status of an operation.
func (s *OperationServiceImpl) UpdateOperationStatus(ctx context.Context, operationID, status string) error {
	// Check if completing
	setCompleted := status == "complete"
	return s.operationRepo.UpdateStatus(ctx, operationID, status, setCompleted)
}

// CompleteOperation marks an operation as complete.
func (s *OperationServiceImpl) CompleteOperation(ctx context.Context, operationID string) error {
	return s.operationRepo.UpdateStatus(ctx, operationID, "complete", true)
}

// Helper methods

func (s *OperationServiceImpl) recordToOperation(r *secondary.OperationRecord) *primary.Operation {
	return &primary.Operation{
		ID:           r.ID,
		CommissionID: r.CommissionID,
		Title:        r.Title,
		Description:  r.Description,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		CompletedAt:  r.CompletedAt,
	}
}

// Ensure OperationServiceImpl implements the interface
var _ primary.OperationService = (*OperationServiceImpl)(nil)
