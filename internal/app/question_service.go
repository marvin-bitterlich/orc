package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// QuestionServiceImpl implements the QuestionService interface.
type QuestionServiceImpl struct {
	questionRepo secondary.QuestionRepository
}

// NewQuestionService creates a new QuestionService with injected dependencies.
func NewQuestionService(
	questionRepo secondary.QuestionRepository,
) *QuestionServiceImpl {
	return &QuestionServiceImpl{
		questionRepo: questionRepo,
	}
}

// CreateQuestion creates a new question.
func (s *QuestionServiceImpl) CreateQuestion(ctx context.Context, req primary.CreateQuestionRequest) (*primary.CreateQuestionResponse, error) {
	// Validate mission exists
	exists, err := s.questionRepo.MissionExists(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate mission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("mission %s not found", req.MissionID)
	}

	// Validate investigation exists if provided
	if req.InvestigationID != "" {
		invExists, err := s.questionRepo.InvestigationExists(ctx, req.InvestigationID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate investigation: %w", err)
		}
		if !invExists {
			return nil, fmt.Errorf("investigation %s not found", req.InvestigationID)
		}
	}

	// Get next ID
	nextID, err := s.questionRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question ID: %w", err)
	}

	// Create record
	record := &secondary.QuestionRecord{
		ID:              nextID,
		MissionID:       req.MissionID,
		InvestigationID: req.InvestigationID,
		Title:           req.Title,
		Description:     req.Description,
		Status:          "open",
	}

	if err := s.questionRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	// Fetch created question
	created, err := s.questionRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created question: %w", err)
	}

	return &primary.CreateQuestionResponse{
		QuestionID: created.ID,
		Question:   s.recordToQuestion(created),
	}, nil
}

// GetQuestion retrieves a question by ID.
func (s *QuestionServiceImpl) GetQuestion(ctx context.Context, questionID string) (*primary.Question, error) {
	record, err := s.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	return s.recordToQuestion(record), nil
}

// ListQuestions lists questions with optional filters.
func (s *QuestionServiceImpl) ListQuestions(ctx context.Context, filters primary.QuestionFilters) ([]*primary.Question, error) {
	records, err := s.questionRepo.List(ctx, secondary.QuestionFilters{
		InvestigationID: filters.InvestigationID,
		MissionID:       filters.MissionID,
		Status:          filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}

	questions := make([]*primary.Question, len(records))
	for i, r := range records {
		questions[i] = s.recordToQuestion(r)
	}
	return questions, nil
}

// AnswerQuestion answers a question (marks it as answered).
func (s *QuestionServiceImpl) AnswerQuestion(ctx context.Context, questionID, answer string) error {
	return s.questionRepo.Answer(ctx, questionID, answer)
}

// UpdateQuestion updates a question's title and/or description.
func (s *QuestionServiceImpl) UpdateQuestion(ctx context.Context, req primary.UpdateQuestionRequest) error {
	record := &secondary.QuestionRecord{
		ID:          req.QuestionID,
		Title:       req.Title,
		Description: req.Description,
	}
	return s.questionRepo.Update(ctx, record)
}

// PinQuestion pins a question.
func (s *QuestionServiceImpl) PinQuestion(ctx context.Context, questionID string) error {
	return s.questionRepo.Pin(ctx, questionID)
}

// UnpinQuestion unpins a question.
func (s *QuestionServiceImpl) UnpinQuestion(ctx context.Context, questionID string) error {
	return s.questionRepo.Unpin(ctx, questionID)
}

// DeleteQuestion deletes a question.
func (s *QuestionServiceImpl) DeleteQuestion(ctx context.Context, questionID string) error {
	return s.questionRepo.Delete(ctx, questionID)
}

// Helper methods

func (s *QuestionServiceImpl) recordToQuestion(r *secondary.QuestionRecord) *primary.Question {
	return &primary.Question{
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

// Ensure QuestionServiceImpl implements the interface
var _ primary.QuestionService = (*QuestionServiceImpl)(nil)
