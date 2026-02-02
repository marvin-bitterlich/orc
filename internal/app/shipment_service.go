package app

import (
	"context"
	"errors"
	"fmt"

	coreshipment "github.com/example/orc/internal/core/shipment"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ShipmentServiceImpl implements the ShipmentService interface.
type ShipmentServiceImpl struct {
	shipmentRepo secondary.ShipmentRepository
	taskRepo     secondary.TaskRepository
	shipyardRepo secondary.ShipyardRepository
}

// NewShipmentService creates a new ShipmentService with injected dependencies.
func NewShipmentService(
	shipmentRepo secondary.ShipmentRepository,
	taskRepo secondary.TaskRepository,
	shipyardRepo secondary.ShipyardRepository,
) *ShipmentServiceImpl {
	return &ShipmentServiceImpl{
		shipmentRepo: shipmentRepo,
		taskRepo:     taskRepo,
		shipyardRepo: shipyardRepo,
	}
}

// CreateShipment creates a new shipment for a commission.
func (s *ShipmentServiceImpl) CreateShipment(ctx context.Context, req primary.CreateShipmentRequest) (*primary.CreateShipmentResponse, error) {
	// Validate commission exists
	exists, err := s.shipmentRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.shipmentRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate shipment ID: %w", err)
	}

	// Generate branch name if repo is specified
	var branch string
	if req.RepoID != "" {
		if req.Branch != "" {
			branch = req.Branch // Use provided branch name
		} else {
			// Auto-generate branch name: {initials}/SHIP-{id}-{slug}
			branch = GenerateShipmentBranchName(UserInitials, nextID, req.Title)
		}
	}

	// Create record with explicit FKs
	// Status is determined by repo: draft (conclave) or queued (shipyard)
	record := &secondary.ShipmentRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		Title:        req.Title,
		Description:  req.Description,
		RepoID:       req.RepoID,
		Branch:       branch,
		ConclaveID:   req.ConclaveID,
		ShipyardID:   req.ShipyardID,
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
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
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
// If force is true, completes even if tasks are incomplete.
func (s *ShipmentServiceImpl) CompleteShipment(ctx context.Context, shipmentID string, force bool) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Get tasks for this shipment
	taskRecords, err := s.taskRepo.List(ctx, secondary.TaskFilters{ShipmentID: shipmentID})
	if err != nil {
		return fmt.Errorf("failed to get tasks for shipment: %w", err)
	}

	// Build task summaries for guard
	tasks := make([]coreshipment.TaskSummary, len(taskRecords))
	for i, t := range taskRecords {
		tasks[i] = coreshipment.TaskSummary{
			ID:     t.ID,
			Status: t.Status,
		}
	}

	// Guard: check all completion preconditions
	guardCtx := coreshipment.CompleteShipmentContext{
		ShipmentID:      shipmentID,
		IsPinned:        record.Pinned,
		Tasks:           tasks,
		ForceCompletion: force,
	}
	if result := coreshipment.CanCompleteShipment(guardCtx); !result.Allowed {
		return result.Error()
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

// AssignShipmentToWorkbench assigns a shipment to a workbench.
func (s *ShipmentServiceImpl) AssignShipmentToWorkbench(ctx context.Context, shipmentID, workbenchID string) error {
	// Verify shipment exists
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Check if workbench is already assigned to another shipment
	otherShipmentID, err := s.shipmentRepo.WorkbenchAssignedToOther(ctx, workbenchID, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to check workbench assignment: %w", err)
	}
	if otherShipmentID != "" {
		return fmt.Errorf("workbench already assigned to shipment %s", otherShipmentID)
	}

	// Assign workbench to shipment
	if err := s.shipmentRepo.AssignWorkbench(ctx, shipmentID, workbenchID); err != nil {
		return err
	}

	// Cascade to tasks
	return s.taskRepo.AssignWorkbenchByShipment(ctx, shipmentID, workbenchID)
}

// GetShipmentsByWorkbench retrieves shipments assigned to a workbench.
func (s *ShipmentServiceImpl) GetShipmentsByWorkbench(ctx context.Context, workbenchID string) ([]*primary.Shipment, error) {
	records, err := s.shipmentRepo.GetByWorkbench(ctx, workbenchID)
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

// ParkShipment moves a shipment to the commission's Shipyard.
func (s *ShipmentServiceImpl) ParkShipment(ctx context.Context, shipmentID string) error {
	// Get shipment to find commission and check current container
	shipment, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// If already queued (in shipyard), return early (no-op, CLI handles message)
	if shipment.Status == "queued" {
		return nil
	}

	// Look up factory for commission, then shipyard for factory
	factoryID, err := s.shipmentRepo.GetFactoryIDForCommission(ctx, shipment.CommissionID)
	if err != nil {
		return fmt.Errorf("failed to get factory for commission: %w", err)
	}

	yard, err := s.shipyardRepo.GetByFactoryID(ctx, factoryID)
	if err != nil {
		return fmt.Errorf("failed to get shipyard: %w", err)
	}

	// Set shipyard_id and status='queued'
	return s.shipmentRepo.SetShipyardID(ctx, shipmentID, yard.ID)
}

// UnparkShipment moves a shipment from Shipyard to a specific Conclave.
func (s *ShipmentServiceImpl) UnparkShipment(ctx context.Context, shipmentID, conclaveID string) error {
	// Get shipment to verify it exists
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Set conclave_id, clear shipyard_id, status='draft'
	return s.shipmentRepo.SetConclaveID(ctx, shipmentID, conclaveID)
}

// Helper methods

func (s *ShipmentServiceImpl) recordToShipment(r *secondary.ShipmentRecord) *primary.Shipment {
	return &primary.Shipment{
		ID:                  r.ID,
		CommissionID:        r.CommissionID,
		Title:               r.Title,
		Description:         r.Description,
		Status:              r.Status,
		AssignedWorkbenchID: r.AssignedWorkbenchID,
		RepoID:              r.RepoID,
		Branch:              r.Branch,
		Pinned:              r.Pinned,
		ConclaveID:          r.ConclaveID,
		ShipyardID:          r.ShipyardID,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		CompletedAt:         r.CompletedAt,
	}
}

// recordToTask converts a TaskRecord to a Task (shared helper).
func recordToTask(r *secondary.TaskRecord) *primary.Task {
	return &primary.Task{
		ID:                  r.ID,
		ShipmentID:          r.ShipmentID,
		TomeID:              r.TomeID,
		ConclaveID:          r.ConclaveID,
		CommissionID:        r.CommissionID,
		Title:               r.Title,
		Description:         r.Description,
		Type:                r.Type,
		Status:              r.Status,
		Priority:            r.Priority,
		AssignedWorkbenchID: r.AssignedWorkbenchID,
		Pinned:              r.Pinned,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		ClaimedAt:           r.ClaimedAt,
		CompletedAt:         r.CompletedAt,
	}
}

// ListShipyardQueue retrieves shipments in the shipyard queue, ordered by priority.
func (s *ShipmentServiceImpl) ListShipyardQueue(ctx context.Context, factoryID string) ([]*primary.ShipyardQueueEntry, error) {
	entries, err := s.shipmentRepo.ListShipyardQueue(ctx, factoryID)
	if err != nil {
		return nil, err
	}

	// Convert secondary records to primary types
	result := make([]*primary.ShipyardQueueEntry, len(entries))
	for i, e := range entries {
		result[i] = &primary.ShipyardQueueEntry{
			ID:           e.ID,
			CommissionID: e.CommissionID,
			Title:        e.Title,
			Priority:     e.Priority,
			TaskCount:    e.TaskCount,
			DoneCount:    e.DoneCount,
			CreatedAt:    e.CreatedAt,
		}
	}

	return result, nil
}

// SetShipmentPriority sets the priority for a shipment in the queue.
func (s *ShipmentServiceImpl) SetShipmentPriority(ctx context.Context, shipmentID string, priority *int) error {
	// Validate priority if set
	if priority != nil && *priority < 1 {
		return fmt.Errorf("priority must be at least 1, got %d", *priority)
	}

	return s.shipmentRepo.UpdatePriority(ctx, shipmentID, priority)
}

// Ensure ShipmentServiceImpl implements the interface
var _ primary.ShipmentService = (*ShipmentServiceImpl)(nil)

// Sentinel error for pinned shipment
var ErrShipmentPinned = errors.New("cannot complete pinned shipment")
