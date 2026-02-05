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
	noteService  primary.NoteService
}

// NewShipmentService creates a new ShipmentService with injected dependencies.
func NewShipmentService(
	shipmentRepo secondary.ShipmentRepository,
	taskRepo secondary.TaskRepository,
	noteService primary.NoteService,
) *ShipmentServiceImpl {
	return &ShipmentServiceImpl{
		shipmentRepo: shipmentRepo,
		taskRepo:     taskRepo,
		noteService:  noteService,
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

	// Create record - shipments go directly under commissions
	record := &secondary.ShipmentRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		Title:        req.Title,
		Description:  req.Description,
		RepoID:       req.RepoID,
		Branch:       branch,
		SpecNoteID:   req.SpecNoteID,
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
// If shipment has a SpecNoteID, the spec note is closed with reason "resolved".
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

	// Update shipment status to complete
	if err := s.shipmentRepo.UpdateStatus(ctx, shipmentID, "complete", true); err != nil {
		return err
	}

	// Close spec note if shipment was generated from one
	if record.SpecNoteID != "" && s.noteService != nil {
		closeReq := primary.CloseNoteRequest{
			NoteID: record.SpecNoteID,
			Reason: "resolved",
		}
		if err := s.noteService.CloseNote(ctx, closeReq); err != nil {
			// Log but don't fail - shipment is already complete
			fmt.Printf("Warning: failed to close spec note %s: %v\n", record.SpecNoteID, err)
		}
	}

	return nil
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

// DeployShipment marks a shipment as deployed (merged to master or deployed to prod).
func (s *ShipmentServiceImpl) DeployShipment(ctx context.Context, shipmentID string) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Get tasks to count open (non-complete) tasks
	taskRecords, err := s.taskRepo.List(ctx, secondary.TaskFilters{ShipmentID: shipmentID})
	if err != nil {
		return fmt.Errorf("failed to get tasks for shipment: %w", err)
	}

	openCount := 0
	for _, t := range taskRecords {
		if t.Status != "complete" {
			openCount++
		}
	}

	// Guard: check deploy preconditions
	guardCtx := coreshipment.StatusTransitionContext{
		ShipmentID:    shipmentID,
		Status:        record.Status,
		OpenTaskCount: openCount,
	}
	if result := coreshipment.CanDeployShipment(guardCtx); !result.Allowed {
		return result.Error()
	}

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, "deployed", false)
}

// VerifyShipment marks a shipment as verified (post-deploy verification passed).
func (s *ShipmentServiceImpl) VerifyShipment(ctx context.Context, shipmentID string) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: can only verify deployed shipments
	guardCtx := coreshipment.StatusTransitionContext{Status: record.Status}
	if result := coreshipment.CanVerifyShipment(guardCtx); !result.Allowed {
		return result.Error()
	}

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, "verified", false)
}

// UpdateShipment updates a shipment's title, description, and/or branch.
func (s *ShipmentServiceImpl) UpdateShipment(ctx context.Context, req primary.UpdateShipmentRequest) error {
	record := &secondary.ShipmentRecord{
		ID:          req.ShipmentID,
		Title:       req.Title,
		Description: req.Description,
		Branch:      req.Branch,
	}
	return s.shipmentRepo.Update(ctx, record)
}

// UpdateStatus sets a shipment's status directly (used for auto-transitions).
func (s *ShipmentServiceImpl) UpdateStatus(ctx context.Context, shipmentID, status string) error {
	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, status, false)
}

// SetStatus sets a shipment's status with escape hatch protection.
// If force is true, allows backwards transitions.
func (s *ShipmentServiceImpl) SetStatus(ctx context.Context, shipmentID, status string, force bool) error {
	record, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return err
	}

	// Guard: check for backwards transitions
	guardCtx := coreshipment.OverrideStatusContext{
		ShipmentID:    shipmentID,
		CurrentStatus: record.Status,
		NewStatus:     status,
		Force:         force,
	}
	if result := coreshipment.CanOverrideStatus(guardCtx); !result.Allowed {
		return result.Error()
	}

	// Set completed flag if transitioning to complete
	setCompleted := status == "complete"

	return s.shipmentRepo.UpdateStatus(ctx, shipmentID, status, setCompleted)
}

// TriggerAutoTransition evaluates and applies auto-transition for a shipment.
func (s *ShipmentServiceImpl) TriggerAutoTransition(ctx context.Context, shipmentID, triggerEvent string) (string, error) {
	ship, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return "", err
	}

	newStatus := coreshipment.GetAutoTransitionStatus(coreshipment.AutoTransitionContext{
		CurrentStatus: ship.Status,
		TriggerEvent:  triggerEvent,
	})
	if newStatus == "" {
		return "", nil // No transition needed
	}

	if err := s.shipmentRepo.UpdateStatus(ctx, shipmentID, newStatus, false); err != nil {
		return "", err
	}
	return newStatus, nil
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
		SpecNoteID:          r.SpecNoteID,
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

// Ensure ShipmentServiceImpl implements the interface
var _ primary.ShipmentService = (*ShipmentServiceImpl)(nil)

// Sentinel error for pinned shipment
var ErrShipmentPinned = errors.New("cannot complete pinned shipment")
