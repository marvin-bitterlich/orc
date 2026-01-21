package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockQuestionRepository implements secondary.QuestionRepository for testing.
type mockQuestionRepository struct {
	questions                 map[string]*secondary.QuestionRecord
	createErr                 error
	getErr                    error
	updateErr                 error
	deleteErr                 error
	listErr                   error
	answerErr                 error
	missionExistsResult       bool
	missionExistsErr          error
	investigationExistsResult bool
	investigationExistsErr    error
}

func newMockQuestionRepository() *mockQuestionRepository {
	return &mockQuestionRepository{
		questions:                 make(map[string]*secondary.QuestionRecord),
		missionExistsResult:       true,
		investigationExistsResult: true,
	}
}

func (m *mockQuestionRepository) Create(ctx context.Context, question *secondary.QuestionRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.questions[question.ID] = question
	return nil
}

func (m *mockQuestionRepository) GetByID(ctx context.Context, id string) (*secondary.QuestionRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if question, ok := m.questions[id]; ok {
		return question, nil
	}
	return nil, errors.New("question not found")
}

func (m *mockQuestionRepository) List(ctx context.Context, filters secondary.QuestionFilters) ([]*secondary.QuestionRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.QuestionRecord
	for _, q := range m.questions {
		if filters.MissionID != "" && q.MissionID != filters.MissionID {
			continue
		}
		if filters.InvestigationID != "" && q.InvestigationID != filters.InvestigationID {
			continue
		}
		if filters.Status != "" && q.Status != filters.Status {
			continue
		}
		result = append(result, q)
	}
	return result, nil
}

func (m *mockQuestionRepository) Update(ctx context.Context, question *secondary.QuestionRecord) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if existing, ok := m.questions[question.ID]; ok {
		if question.Title != "" {
			existing.Title = question.Title
		}
		if question.Description != "" {
			existing.Description = question.Description
		}
	}
	return nil
}

func (m *mockQuestionRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.questions, id)
	return nil
}

func (m *mockQuestionRepository) Pin(ctx context.Context, id string) error {
	if question, ok := m.questions[id]; ok {
		question.Pinned = true
	}
	return nil
}

func (m *mockQuestionRepository) Unpin(ctx context.Context, id string) error {
	if question, ok := m.questions[id]; ok {
		question.Pinned = false
	}
	return nil
}

func (m *mockQuestionRepository) GetNextID(ctx context.Context) (string, error) {
	return "QUESTION-001", nil
}

func (m *mockQuestionRepository) Answer(ctx context.Context, id, answer string) error {
	if m.answerErr != nil {
		return m.answerErr
	}
	if question, ok := m.questions[id]; ok {
		question.Answer = answer
		question.Status = "answered"
		question.AnsweredAt = "2026-01-20T10:00:00Z"
	}
	return nil
}

func (m *mockQuestionRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

func (m *mockQuestionRepository) InvestigationExists(ctx context.Context, investigationID string) (bool, error) {
	if m.investigationExistsErr != nil {
		return false, m.investigationExistsErr
	}
	return m.investigationExistsResult, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestQuestionService() (*QuestionServiceImpl, *mockQuestionRepository) {
	questionRepo := newMockQuestionRepository()
	service := NewQuestionService(questionRepo)
	return service, questionRepo
}

// ============================================================================
// CreateQuestion Tests
// ============================================================================

func TestCreateQuestion_Success(t *testing.T) {
	service, _ := newTestQuestionService()
	ctx := context.Background()

	resp, err := service.CreateQuestion(ctx, primary.CreateQuestionRequest{
		MissionID:   "MISSION-001",
		Title:       "Test Question",
		Description: "A test question",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.QuestionID == "" {
		t.Error("expected question ID to be set")
	}
	if resp.Question.Title != "Test Question" {
		t.Errorf("expected title 'Test Question', got '%s'", resp.Question.Title)
	}
	if resp.Question.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", resp.Question.Status)
	}
}

func TestCreateQuestion_WithInvestigation(t *testing.T) {
	service, _ := newTestQuestionService()
	ctx := context.Background()

	resp, err := service.CreateQuestion(ctx, primary.CreateQuestionRequest{
		MissionID:       "MISSION-001",
		InvestigationID: "INV-001",
		Title:           "Investigation Question",
		Description:     "A question for an investigation",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Question.InvestigationID != "INV-001" {
		t.Errorf("expected investigation ID 'INV-001', got '%s'", resp.Question.InvestigationID)
	}
}

func TestCreateQuestion_MissionNotFound(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.missionExistsResult = false

	_, err := service.CreateQuestion(ctx, primary.CreateQuestionRequest{
		MissionID:   "MISSION-NONEXISTENT",
		Title:       "Test Question",
		Description: "A test question",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

func TestCreateQuestion_InvestigationNotFound(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.investigationExistsResult = false

	_, err := service.CreateQuestion(ctx, primary.CreateQuestionRequest{
		MissionID:       "MISSION-001",
		InvestigationID: "INV-NONEXISTENT",
		Title:           "Test Question",
		Description:     "A test question",
	})

	if err == nil {
		t.Fatal("expected error for non-existent investigation, got nil")
	}
}

// ============================================================================
// GetQuestion Tests
// ============================================================================

func TestGetQuestion_Found(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Test Question",
		Status:    "open",
	}

	question, err := service.GetQuestion(ctx, "QUESTION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if question.Title != "Test Question" {
		t.Errorf("expected title 'Test Question', got '%s'", question.Title)
	}
}

func TestGetQuestion_NotFound(t *testing.T) {
	service, _ := newTestQuestionService()
	ctx := context.Background()

	_, err := service.GetQuestion(ctx, "QUESTION-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent question, got nil")
	}
}

// ============================================================================
// ListQuestions Tests
// ============================================================================

func TestListQuestions_FilterByMission(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Question 1",
		Status:    "open",
	}
	questionRepo.questions["QUESTION-002"] = &secondary.QuestionRecord{
		ID:        "QUESTION-002",
		MissionID: "MISSION-002",
		Title:     "Question 2",
		Status:    "open",
	}

	questions, err := service.ListQuestions(ctx, primary.QuestionFilters{MissionID: "MISSION-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("expected 1 question, got %d", len(questions))
	}
}

func TestListQuestions_FilterByInvestigation(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:              "QUESTION-001",
		MissionID:       "MISSION-001",
		InvestigationID: "INV-001",
		Title:           "Investigation Question",
		Status:          "open",
	}
	questionRepo.questions["QUESTION-002"] = &secondary.QuestionRecord{
		ID:        "QUESTION-002",
		MissionID: "MISSION-001",
		Title:     "Standalone Question",
		Status:    "open",
	}

	questions, err := service.ListQuestions(ctx, primary.QuestionFilters{InvestigationID: "INV-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("expected 1 investigation question, got %d", len(questions))
	}
}

func TestListQuestions_FilterByStatus(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Open Question",
		Status:    "open",
	}
	questionRepo.questions["QUESTION-002"] = &secondary.QuestionRecord{
		ID:        "QUESTION-002",
		MissionID: "MISSION-001",
		Title:     "Answered Question",
		Status:    "answered",
	}

	questions, err := service.ListQuestions(ctx, primary.QuestionFilters{Status: "open"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("expected 1 open question, got %d", len(questions))
	}
}

// ============================================================================
// AnswerQuestion Tests
// ============================================================================

func TestAnswerQuestion_Success(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Test Question",
		Status:    "open",
	}

	err := service.AnswerQuestion(ctx, "QUESTION-001", "This is the answer")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if questionRepo.questions["QUESTION-001"].Answer != "This is the answer" {
		t.Errorf("expected answer 'This is the answer', got '%s'", questionRepo.questions["QUESTION-001"].Answer)
	}
	if questionRepo.questions["QUESTION-001"].Status != "answered" {
		t.Errorf("expected status 'answered', got '%s'", questionRepo.questions["QUESTION-001"].Status)
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestPinQuestion(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Test Question",
		Status:    "open",
		Pinned:    false,
	}

	err := service.PinQuestion(ctx, "QUESTION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !questionRepo.questions["QUESTION-001"].Pinned {
		t.Error("expected question to be pinned")
	}
}

func TestUnpinQuestion(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Pinned Question",
		Status:    "open",
		Pinned:    true,
	}

	err := service.UnpinQuestion(ctx, "QUESTION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if questionRepo.questions["QUESTION-001"].Pinned {
		t.Error("expected question to be unpinned")
	}
}

// ============================================================================
// UpdateQuestion Tests
// ============================================================================

func TestUpdateQuestion_Title(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:          "QUESTION-001",
		MissionID:   "MISSION-001",
		Title:       "Old Title",
		Description: "Original description",
		Status:      "open",
	}

	err := service.UpdateQuestion(ctx, primary.UpdateQuestionRequest{
		QuestionID: "QUESTION-001",
		Title:      "New Title",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if questionRepo.questions["QUESTION-001"].Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", questionRepo.questions["QUESTION-001"].Title)
	}
}

// ============================================================================
// DeleteQuestion Tests
// ============================================================================

func TestDeleteQuestion_Success(t *testing.T) {
	service, questionRepo := newTestQuestionService()
	ctx := context.Background()

	questionRepo.questions["QUESTION-001"] = &secondary.QuestionRecord{
		ID:        "QUESTION-001",
		MissionID: "MISSION-001",
		Title:     "Test Question",
		Status:    "open",
	}

	err := service.DeleteQuestion(ctx, "QUESTION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := questionRepo.questions["QUESTION-001"]; exists {
		t.Error("expected question to be deleted")
	}
}
