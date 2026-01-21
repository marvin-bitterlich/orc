package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ConclaveServiceImpl implements the ConclaveService interface.
type ConclaveServiceImpl struct {
	conclaveRepo secondary.ConclaveRepository
}

// NewConclaveService creates a new ConclaveService with injected dependencies.
func NewConclaveService(
	conclaveRepo secondary.ConclaveRepository,
) *ConclaveServiceImpl {
	return &ConclaveServiceImpl{
		conclaveRepo: conclaveRepo,
	}
}

// CreateConclave creates a new conclave (ideation session).
func (s *ConclaveServiceImpl) CreateConclave(ctx context.Context, req primary.CreateConclaveRequest) (*primary.CreateConclaveResponse, error) {
	// Validate commission exists
	exists, err := s.conclaveRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.conclaveRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate conclave ID: %w", err)
	}

	// Create record
	record := &secondary.ConclaveRecord{
		ID:           nextID,
		CommissionID: req.CommissionID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       "open",
	}

	if err := s.conclaveRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create conclave: %w", err)
	}

	// Fetch created conclave
	created, err := s.conclaveRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created conclave: %w", err)
	}

	return &primary.CreateConclaveResponse{
		ConclaveID: created.ID,
		Conclave:   s.recordToConclave(created),
	}, nil
}

// GetConclave retrieves a conclave by ID.
func (s *ConclaveServiceImpl) GetConclave(ctx context.Context, conclaveID string) (*primary.Conclave, error) {
	record, err := s.conclaveRepo.GetByID(ctx, conclaveID)
	if err != nil {
		return nil, err
	}
	return s.recordToConclave(record), nil
}

// ListConclaves lists conclaves with optional filters.
func (s *ConclaveServiceImpl) ListConclaves(ctx context.Context, filters primary.ConclaveFilters) ([]*primary.Conclave, error) {
	records, err := s.conclaveRepo.List(ctx, secondary.ConclaveFilters{
		CommissionID: filters.CommissionID,
		Status:       filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list conclaves: %w", err)
	}

	conclaves := make([]*primary.Conclave, len(records))
	for i, r := range records {
		conclaves[i] = s.recordToConclave(r)
	}
	return conclaves, nil
}

// CompleteConclave marks a conclave as complete.
func (s *ConclaveServiceImpl) CompleteConclave(ctx context.Context, conclaveID string) error {
	record, err := s.conclaveRepo.GetByID(ctx, conclaveID)
	if err != nil {
		return err
	}

	// Guard: cannot complete pinned conclave
	if record.Pinned {
		return fmt.Errorf("cannot complete pinned conclave %s. Unpin first with: orc conclave unpin %s", conclaveID, conclaveID)
	}

	return s.conclaveRepo.UpdateStatus(ctx, conclaveID, "closed", true)
}

// PauseConclave pauses an active conclave.
func (s *ConclaveServiceImpl) PauseConclave(ctx context.Context, conclaveID string) error {
	record, err := s.conclaveRepo.GetByID(ctx, conclaveID)
	if err != nil {
		return err
	}

	// Guard: can only pause open conclaves
	if record.Status != "open" {
		return fmt.Errorf("can only pause open conclaves (current status: %s)", record.Status)
	}

	return s.conclaveRepo.UpdateStatus(ctx, conclaveID, "paused", false)
}

// ResumeConclave resumes a paused conclave.
func (s *ConclaveServiceImpl) ResumeConclave(ctx context.Context, conclaveID string) error {
	record, err := s.conclaveRepo.GetByID(ctx, conclaveID)
	if err != nil {
		return err
	}

	// Guard: can only resume paused conclaves
	if record.Status != "paused" {
		return fmt.Errorf("can only resume paused conclaves (current status: %s)", record.Status)
	}

	return s.conclaveRepo.UpdateStatus(ctx, conclaveID, "open", false)
}

// UpdateConclave updates a conclave's title and/or description.
func (s *ConclaveServiceImpl) UpdateConclave(ctx context.Context, req primary.UpdateConclaveRequest) error {
	record := &secondary.ConclaveRecord{
		ID:          req.ConclaveID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.conclaveRepo.Update(ctx, record)
}

// PinConclave pins a conclave.
func (s *ConclaveServiceImpl) PinConclave(ctx context.Context, conclaveID string) error {
	return s.conclaveRepo.Pin(ctx, conclaveID)
}

// UnpinConclave unpins a conclave.
func (s *ConclaveServiceImpl) UnpinConclave(ctx context.Context, conclaveID string) error {
	return s.conclaveRepo.Unpin(ctx, conclaveID)
}

// DeleteConclave deletes a conclave.
func (s *ConclaveServiceImpl) DeleteConclave(ctx context.Context, conclaveID string) error {
	return s.conclaveRepo.Delete(ctx, conclaveID)
}

// GetConclavesByShipment retrieves conclaves belonging to a shipment.
func (s *ConclaveServiceImpl) GetConclavesByShipment(ctx context.Context, shipmentID string) ([]*primary.Conclave, error) {
	// Get all conclaves and filter by shipment
	records, err := s.conclaveRepo.List(ctx, secondary.ConclaveFilters{})
	if err != nil {
		return nil, err
	}

	var conclaves []*primary.Conclave
	for _, r := range records {
		if r.ShipmentID == shipmentID {
			conclaves = append(conclaves, s.recordToConclave(r))
		}
	}
	return conclaves, nil
}

// GetConclaveTasks retrieves all tasks in a conclave.
func (s *ConclaveServiceImpl) GetConclaveTasks(ctx context.Context, conclaveID string) ([]*primary.ConclaveTask, error) {
	records, err := s.conclaveRepo.GetTasksByConclave(ctx, conclaveID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*primary.ConclaveTask, len(records))
	for i, r := range records {
		tasks[i] = s.taskRecordToConclaveTask(r)
	}
	return tasks, nil
}

// GetConclaveQuestions retrieves all questions in a conclave.
func (s *ConclaveServiceImpl) GetConclaveQuestions(ctx context.Context, conclaveID string) ([]*primary.ConclaveQuestion, error) {
	records, err := s.conclaveRepo.GetQuestionsByConclave(ctx, conclaveID)
	if err != nil {
		return nil, err
	}

	questions := make([]*primary.ConclaveQuestion, len(records))
	for i, r := range records {
		questions[i] = s.questionRecordToConclaveQuestion(r)
	}
	return questions, nil
}

// GetConclavePlans retrieves all plans in a conclave.
func (s *ConclaveServiceImpl) GetConclavePlans(ctx context.Context, conclaveID string) ([]*primary.ConclavePlan, error) {
	records, err := s.conclaveRepo.GetPlansByConclave(ctx, conclaveID)
	if err != nil {
		return nil, err
	}

	plans := make([]*primary.ConclavePlan, len(records))
	for i, r := range records {
		plans[i] = s.planRecordToConclavePlan(r)
	}
	return plans, nil
}

// Helper methods

func (s *ConclaveServiceImpl) recordToConclave(r *secondary.ConclaveRecord) *primary.Conclave {
	return &primary.Conclave{
		ID:           r.ID,
		CommissionID: r.CommissionID,
		ShipmentID:   r.ShipmentID,
		Title:        r.Title,
		Description:  r.Description,
		Status:       r.Status,
		Decision:     r.Decision,
		Pinned:       r.Pinned,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		DecidedAt:    r.DecidedAt,
	}
}

func (s *ConclaveServiceImpl) taskRecordToConclaveTask(r *secondary.ConclaveTaskRecord) *primary.ConclaveTask {
	return &primary.ConclaveTask{
		ID:                  r.ID,
		ShipmentID:          r.ShipmentID,
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

func (s *ConclaveServiceImpl) questionRecordToConclaveQuestion(r *secondary.ConclaveQuestionRecord) *primary.ConclaveQuestion {
	return &primary.ConclaveQuestion{
		ID:              r.ID,
		CommissionID:    r.CommissionID,
		ShipmentID:      r.ShipmentID,
		InvestigationID: r.InvestigationID,
		ConclaveID:      r.ConclaveID,
		Title:           r.Title,
		Content:         r.Content,
		Answer:          r.Answer,
		Status:          r.Status,
		Pinned:          r.Pinned,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		AnsweredAt:      r.AnsweredAt,
	}
}

func (s *ConclaveServiceImpl) planRecordToConclavePlan(r *secondary.ConclavePlanRecord) *primary.ConclavePlan {
	return &primary.ConclavePlan{
		ID:           r.ID,
		CommissionID: r.CommissionID,
		ShipmentID:   r.ShipmentID,
		Title:        r.Title,
		Description:  r.Description,
		Content:      r.Content,
		Status:       r.Status,
		Pinned:       r.Pinned,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		ApprovedAt:   r.ApprovedAt,
	}
}

// Ensure ConclaveServiceImpl implements the interface
var _ primary.ConclaveService = (*ConclaveServiceImpl)(nil)
