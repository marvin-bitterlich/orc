package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ShipmentServiceImpl implements the ShipmentService interface.
type ShipmentServiceImpl struct {
	shipmentRepo secondary.ShipmentRepository
	taskRepo     secondary.TaskRepository
}

// NewShipmentService creates a new ShipmentService with injected dependencies.
func NewShipmentService(
	shipmentRepo secondary.ShipmentRepository,
	taskRepo secondary.TaskRepository,
) *ShipmentServiceImpl {
	return &ShipmentServiceImpl{
		shipmentRepo: shipmentRepo,
		taskRepo:     taskRepo,
	}
}

// CreateShipment creates a new shipment for a mission.
func (s *ShipmentServiceImpl) CreateShipment(ctx context.Context, req primary.CreateShipmentRequest) (*primary.CreateShipmentResponse, error) {
	// Validate mission exists
	exists, err := s.shipmentRepo.MissionExists(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate mission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("mission %s not found", req.MissionID)
	}

	// Get next ID
	nextID, err := s.shipmentRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate shipment ID: %w", err)
	}

	// Create record
	record := &secondary.ShipmentRecord{
		ID:          nextID,
		MissionID:   req.MissionID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "active",
	}

	if err := s.shipmentRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Fetch created shipment
	created, err := s.shipmentRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created shipment: %w", err)
	}

	return &primary.CreateShipmentResponse{
		ShipmentID: created.ID,
		Shipment:   s.recordToShipment(created),
	}, nil
}

// GetShipment retrieves a shipment by ID.
func (s *ShipmentServiceImpl) GetShipment(ctx context.Context, shipmentID string) (*primary.Shipment, error) {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	return s.recordToShipment(record), nil
}

// ListShipments lists shipments with optional filters.
func (s *ShipmentServiceImpl) ListShipments(ctx context.Context, filters primary.ShipmentFilters) ([]*primary.Shipment, error) {
	records, err := s.shipmentRepo.List(ctx, secondary.ShipmentFilters{
		MissionID: filters.MissionID,
		Status:    filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	shipments := make([]*primary.Shipment, len(records))
	for i, r := range records {
		shipments[i] = s.recordToShipment(r)
	}
	return shipments, nil
}

// CompleteShipment marks a shipment as complete.
func (s *ShipmentServiceImpl) CompleteShipment(ctx context.Context, shipmentID string) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: cannot complete pinned shipment
	if record.Pinned {
		return fmt.Errorf("cannot complete pinned shipment %s. Unpin first with: orc shipment unpin %s", shipmentID, shipmentID)
	}

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, "complete", true)
}

// PauseShipment pauses an active shipment.
func (s *ShipmentServiceImpl) PauseShipment(ctx context.Context, shipmentID string) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: can only pause active shipments
	if record.Status != "active" {
		return fmt.Errorf("can only pause active shipments (current status: %s)", record.Status)
	}

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, "paused", false)
}

// ResumeShipment resumes a paused shipment.
func (s *ShipmentServiceImpl) ResumeShipment(ctx context.Context, shipmentID string) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: can only resume paused shipments
	if record.Status != "paused" {
		return fmt.Errorf("can only resume paused shipments (current status: %s)", record.Status)
	}

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, "active", false)
}

// UpdateShipment updates a shipment's title and/or description.
func (s *ShipmentServiceImpl) UpdateShipment(ctx context.Context, req primary.UpdateShipmentRequest) error {
	record := &secondary.ShipmentRecord{
		ID:          req.ShipmentID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.shipmentRepo.Update(ctx, record)
}

// PinShipment pins a shipment.
func (s *ShipmentServiceImpl) PinShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Pin(ctx, shipmentID)
}

// UnpinShipment unpins a shipment.
func (s *ShipmentServiceImpl) UnpinShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Unpin(ctx, shipmentID)
}

// AssignShipmentToGrove assigns a shipment to a grove.
func (s *ShipmentServiceImpl) AssignShipmentToGrove(ctx context.Context, shipmentID, groveID string) error {
	// Verify shipment exists
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Check if grove is already assigned to another shipment
	otherShipmentID, err := s.shipmentRepo.GroveAssignedToOther(ctx, groveID, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to check grove assignment: %w", err)
	}
	if otherShipmentID != "" {
		return fmt.Errorf("grove already assigned to shipment %s", otherShipmentID)
	}

	// Assign grove to shipment
	if err := s.shipmentRepo.AssignGrove(ctx, shipmentID, groveID); err != nil {
		return err
	}

	// Cascade to tasks
	return s.taskRepo.AssignGroveByShipment(ctx, shipmentID, groveID)
}

// GetShipmentsByGrove retrieves shipments assigned to a grove.
func (s *ShipmentServiceImpl) GetShipmentsByGrove(ctx context.Context, groveID string) ([]*primary.Shipment, error) {
	records, err := s.shipmentRepo.GetByGrove(ctx, groveID)
	if err != nil {
		return nil, err
	}

	shipments := make([]*primary.Shipment, len(records))
	for i, r := range records {
		shipments[i] = s.recordToShipment(r)
	}
	return shipments, nil
}

// GetShipmentTasks retrieves all tasks for a shipment.
func (s *ShipmentServiceImpl) GetShipmentTasks(ctx context.Context, shipmentID string) ([]*primary.Task, error) {
	records, err := s.taskRepo.GetByShipment(ctx, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipment tasks: %w", err)
	}

	tasks := make([]*primary.Task, len(records))
	for i, r := range records {
		tasks[i] = recordToTask(r)
	}
	return tasks, nil
}

// DeleteShipment deletes a shipment.
func (s *ShipmentServiceImpl) DeleteShipment(ctx context.Context, shipmentID string) error {
	return s.shipmentRepo.Delete(ctx, shipmentID)
}

// Helper methods

func (s *ShipmentServiceImpl) recordToShipment(r *secondary.ShipmentRecord) *primary.Shipment {
	return &primary.Shipment{
		ID:              r.ID,
		MissionID:       r.MissionID,
		Title:           r.Title,
		Description:     r.Description,
		Status:          r.Status,
		AssignedGroveID: r.AssignedGroveID,
		Pinned:          r.Pinned,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		CompletedAt:     r.CompletedAt,
	}
}

// recordToTask converts a TaskRecord to a Task (shared helper).
func recordToTask(r *secondary.TaskRecord) *primary.Task {
	return &primary.Task{
		ID:               r.ID,
		ShipmentID:       r.ShipmentID,
		MissionID:        r.MissionID,
		Title:            r.Title,
		Description:      r.Description,
		Type:             r.Type,
		Status:           r.Status,
		Priority:         r.Priority,
		AssignedGroveID:  r.AssignedGroveID,
		Pinned:           r.Pinned,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		ClaimedAt:        r.ClaimedAt,
		CompletedAt:      r.CompletedAt,
		ConclaveID:       r.ConclaveID,
		PromotedFromID:   r.PromotedFromID,
		PromotedFromType: r.PromotedFromType,
	}
}

// Ensure ShipmentServiceImpl implements the interface
var _ primary.ShipmentService = (*ShipmentServiceImpl)(nil)

// Sentinel error for pinned shipment
var ErrShipmentPinned = errors.New("cannot complete pinned shipment")
