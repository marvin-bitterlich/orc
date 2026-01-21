package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// InvestigationServiceImpl implements the InvestigationService interface.
type InvestigationServiceImpl struct {
	investigationRepo secondary.InvestigationRepository
}

// NewInvestigationService creates a new InvestigationService with injected dependencies.
func NewInvestigationService(
	investigationRepo secondary.InvestigationRepository,
) *InvestigationServiceImpl {
	return &InvestigationServiceImpl{
		investigationRepo: investigationRepo,
	}
}

// CreateInvestigation creates a new investigation (research container).
func (s *InvestigationServiceImpl) CreateInvestigation(ctx context.Context, req primary.CreateInvestigationRequest) (*primary.CreateInvestigationResponse, error) {
	// Validate mission exists
	exists, err := s.investigationRepo.MissionExists(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate mission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("mission %s not found", req.MissionID)
	}

	// Get next ID
	nextID, err := s.investigationRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate investigation ID: %w", err)
	}

	// Create record
	record := &secondary.InvestigationRecord{
		ID:          nextID,
		MissionID:   req.MissionID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "active",
	}

	if err := s.investigationRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create investigation: %w", err)
	}

	// Fetch created investigation
	created, err := s.investigationRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created investigation: %w", err)
	}

	return &primary.CreateInvestigationResponse{
		InvestigationID: created.ID,
		Investigation:   s.recordToInvestigation(created),
	}, nil
}

// GetInvestigation retrieves an investigation by ID.
func (s *InvestigationServiceImpl) GetInvestigation(ctx context.Context, investigationID string) (*primary.Investigation, error) {
	record, err := s.investigationRepo.GetByID(ctx, investigationID)
	if err != nil {
		return nil, err
	}
	return s.recordToInvestigation(record), nil
}

// ListInvestigations lists investigations with optional filters.
func (s *InvestigationServiceImpl) ListInvestigations(ctx context.Context, filters primary.InvestigationFilters) ([]*primary.Investigation, error) {
	records, err := s.investigationRepo.List(ctx, secondary.InvestigationFilters{
		MissionID: filters.MissionID,
		Status:    filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list investigations: %w", err)
	}

	investigations := make([]*primary.Investigation, len(records))
	for i, r := range records {
		investigations[i] = s.recordToInvestigation(r)
	}
	return investigations, nil
}

// CompleteInvestigation marks an investigation as complete.
func (s *InvestigationServiceImpl) CompleteInvestigation(ctx context.Context, investigationID string) error {
	record, err := s.investigationRepo.GetByID(ctx, investigationID)
	if err != nil {
		return err
	}

	// Guard: cannot complete pinned investigation
	if record.Pinned {
		return fmt.Errorf("cannot complete pinned investigation %s. Unpin first with: orc investigation unpin %s", investigationID, investigationID)
	}

	return s.investigationRepo.UpdateStatus(ctx, investigationID, "complete", true)
}

// PauseInvestigation pauses an active investigation.
func (s *InvestigationServiceImpl) PauseInvestigation(ctx context.Context, investigationID string) error {
	record, err := s.investigationRepo.GetByID(ctx, investigationID)
	if err != nil {
		return err
	}

	// Guard: can only pause active investigations
	if record.Status != "active" {
		return fmt.Errorf("can only pause active investigations (current status: %s)", record.Status)
	}

	return s.investigationRepo.UpdateStatus(ctx, investigationID, "paused", false)
}

// ResumeInvestigation resumes a paused investigation.
func (s *InvestigationServiceImpl) ResumeInvestigation(ctx context.Context, investigationID string) error {
	record, err := s.investigationRepo.GetByID(ctx, investigationID)
	if err != nil {
		return err
	}

	// Guard: can only resume paused investigations
	if record.Status != "paused" {
		return fmt.Errorf("can only resume paused investigations (current status: %s)", record.Status)
	}

	return s.investigationRepo.UpdateStatus(ctx, investigationID, "active", false)
}

// UpdateInvestigation updates an investigation's title and/or description.
func (s *InvestigationServiceImpl) UpdateInvestigation(ctx context.Context, req primary.UpdateInvestigationRequest) error {
	record := &secondary.InvestigationRecord{
		ID:          req.InvestigationID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.investigationRepo.Update(ctx, record)
}

// PinInvestigation pins an investigation.
func (s *InvestigationServiceImpl) PinInvestigation(ctx context.Context, investigationID string) error {
	return s.investigationRepo.Pin(ctx, investigationID)
}

// UnpinInvestigation unpins an investigation.
func (s *InvestigationServiceImpl) UnpinInvestigation(ctx context.Context, investigationID string) error {
	return s.investigationRepo.Unpin(ctx, investigationID)
}

// DeleteInvestigation deletes an investigation.
func (s *InvestigationServiceImpl) DeleteInvestigation(ctx context.Context, investigationID string) error {
	return s.investigationRepo.Delete(ctx, investigationID)
}

// AssignInvestigationToGrove assigns an investigation to a grove.
func (s *InvestigationServiceImpl) AssignInvestigationToGrove(ctx context.Context, investigationID, groveID string) error {
	return s.investigationRepo.AssignGrove(ctx, investigationID, groveID)
}

// GetInvestigationsByGrove retrieves investigations assigned to a grove.
func (s *InvestigationServiceImpl) GetInvestigationsByGrove(ctx context.Context, groveID string) ([]*primary.Investigation, error) {
	records, err := s.investigationRepo.GetByGrove(ctx, groveID)
	if err != nil {
		return nil, err
	}

	investigations := make([]*primary.Investigation, len(records))
	for i, r := range records {
		investigations[i] = s.recordToInvestigation(r)
	}
	return investigations, nil
}

// GetInvestigationQuestions retrieves all questions in an investigation.
func (s *InvestigationServiceImpl) GetInvestigationQuestions(ctx context.Context, investigationID string) ([]*primary.InvestigationQuestion, error) {
	records, err := s.investigationRepo.GetQuestionsByInvestigation(ctx, investigationID)
	if err != nil {
		return nil, err
	}

	questions := make([]*primary.InvestigationQuestion, len(records))
	for i, r := range records {
		questions[i] = s.questionRecordToInvestigationQuestion(r)
	}
	return questions, nil
}

// Helper methods

func (s *InvestigationServiceImpl) recordToInvestigation(r *secondary.InvestigationRecord) *primary.Investigation {
	return &primary.Investigation{
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

func (s *InvestigationServiceImpl) questionRecordToInvestigationQuestion(r *secondary.InvestigationQuestionRecord) *primary.InvestigationQuestion {
	return &primary.InvestigationQuestion{
		ID:               r.ID,
		InvestigationID:  r.InvestigationID,
		MissionID:        r.MissionID,
		Title:            r.Title,
		Description:      r.Description,
		Status:           r.Status,
		Answer:           r.Answer,
		Pinned:           r.Pinned,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		AnsweredAt:       r.AnsweredAt,
		ConclaveID:       r.ConclaveID,
		PromotedFromID:   r.PromotedFromID,
		PromotedFromType: r.PromotedFromType,
	}
}

// Ensure InvestigationServiceImpl implements the interface
var _ primary.InvestigationService = (*InvestigationServiceImpl)(nil)
