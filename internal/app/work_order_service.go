package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkOrderServiceImpl implements the WorkOrderService interface.
type WorkOrderServiceImpl struct {
	workOrderRepo secondary.WorkOrderRepository
}

// NewWorkOrderService creates a new WorkOrderService with injected dependencies.
func NewWorkOrderService(workOrderRepo secondary.WorkOrderRepository) *WorkOrderServiceImpl {
	return &WorkOrderServiceImpl{
		workOrderRepo: workOrderRepo,
	}
}

// CreateWorkOrder creates a new workOrder.
func (s *WorkOrderServiceImpl) CreateWorkOrder(ctx context.Context, req primary.CreateWorkOrderRequest) (*primary.CreateWorkOrderResponse, error) {
	// Validate shipment exists
	exists, err := s.workOrderRepo.ShipmentExists(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate shipment: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("shipment %s not found", req.ShipmentID)
	}

	// Check if shipment already has a workOrder (1:1 relationship)
	hasWorkOrder, err := s.workOrderRepo.ShipmentHasWorkOrder(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing workOrder: %w", err)
	}
	if hasWorkOrder {
		return nil, fmt.Errorf("shipment %s already has a work order", req.ShipmentID)
	}

	// Get next ID
	nextID, err := s.workOrderRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workOrder ID: %w", err)
	}

	// Create record
	record := &secondary.WorkOrderRecord{
		ID:                 nextID,
		ShipmentID:         req.ShipmentID,
		Outcome:            req.Outcome,
		AcceptanceCriteria: req.AcceptanceCriteria,
		Status:             "draft",
	}

	if err := s.workOrderRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create workOrder: %w", err)
	}

	// Fetch created workOrder
	created, err := s.workOrderRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created workOrder: %w", err)
	}

	return &primary.CreateWorkOrderResponse{
		WorkOrderID: created.ID,
		WorkOrder:   s.recordToWorkOrder(created),
	}, nil
}

// GetWorkOrder retrieves a workOrder by ID.
func (s *WorkOrderServiceImpl) GetWorkOrder(ctx context.Context, workOrderID string) (*primary.WorkOrder, error) {
	record, err := s.workOrderRepo.GetByID(ctx, workOrderID)
	if err != nil {
		return nil, err
	}
	return s.recordToWorkOrder(record), nil
}

// GetWorkOrderByShipment retrieves a workOrder by shipment ID.
func (s *WorkOrderServiceImpl) GetWorkOrderByShipment(ctx context.Context, shipmentID string) (*primary.WorkOrder, error) {
	record, err := s.workOrderRepo.GetByShipment(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return s.recordToWorkOrder(record), nil
}

// ListWorkOrders lists work_orders with optional filters.
func (s *WorkOrderServiceImpl) ListWorkOrders(ctx context.Context, filters primary.WorkOrderFilters) ([]*primary.WorkOrder, error) {
	records, err := s.workOrderRepo.List(ctx, secondary.WorkOrderFilters{
		ShipmentID: filters.ShipmentID,
		Status:     filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list work_orders: %w", err)
	}

	workOrders := make([]*primary.WorkOrder, len(records))
	for i, r := range records {
		workOrders[i] = s.recordToWorkOrder(r)
	}
	return workOrders, nil
}

// UpdateWorkOrder updates a workOrder.
// Draft and active work orders can be updated.
func (s *WorkOrderServiceImpl) UpdateWorkOrder(ctx context.Context, req primary.UpdateWorkOrderRequest) error {
	// Verify work order exists and is in active status
	existing, err := s.workOrderRepo.GetByID(ctx, req.WorkOrderID)
	if err != nil {
		return err
	}

	if existing.Status != "draft" && existing.Status != "active" {
		return fmt.Errorf("cannot update work order %s: current status is %s (must be draft or active)", req.WorkOrderID, existing.Status)
	}

	record := &secondary.WorkOrderRecord{
		ID:                 req.WorkOrderID,
		Outcome:            req.Outcome,
		AcceptanceCriteria: req.AcceptanceCriteria,
	}
	return s.workOrderRepo.Update(ctx, record)
}

// DeleteWorkOrder deletes a workOrder.
func (s *WorkOrderServiceImpl) DeleteWorkOrder(ctx context.Context, workOrderID string) error {
	return s.workOrderRepo.Delete(ctx, workOrderID)
}

// ActivateWorkOrder transitions a work order from draft to active.
func (s *WorkOrderServiceImpl) ActivateWorkOrder(ctx context.Context, workOrderID string) error {
	// Verify work order exists and is in draft status
	record, err := s.workOrderRepo.GetByID(ctx, workOrderID)
	if err != nil {
		return err
	}

	if record.Status != "draft" {
		return fmt.Errorf("cannot activate work order %s: current status is %s (must be draft)", workOrderID, record.Status)
	}

	return s.workOrderRepo.UpdateStatus(ctx, workOrderID, "active")
}

// CompleteWorkOrder transitions a work order from active to complete.
func (s *WorkOrderServiceImpl) CompleteWorkOrder(ctx context.Context, workOrderID string) error {
	// Verify work order exists and is in active status
	record, err := s.workOrderRepo.GetByID(ctx, workOrderID)
	if err != nil {
		return err
	}

	if record.Status != "active" {
		return fmt.Errorf("cannot complete work order %s: current status is %s (must be active)", workOrderID, record.Status)
	}

	return s.workOrderRepo.UpdateStatus(ctx, workOrderID, "complete")
}

// Helper methods

func (s *WorkOrderServiceImpl) recordToWorkOrder(r *secondary.WorkOrderRecord) *primary.WorkOrder {
	return &primary.WorkOrder{
		ID:                 r.ID,
		ShipmentID:         r.ShipmentID,
		Outcome:            r.Outcome,
		AcceptanceCriteria: r.AcceptanceCriteria,
		Status:             r.Status,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}

// Ensure WorkOrderServiceImpl implements the interface
var _ primary.WorkOrderService = (*WorkOrderServiceImpl)(nil)
