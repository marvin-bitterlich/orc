package primary

import "context"

// QuestionService defines the primary port for question operations.
type QuestionService interface {
	// CreateQuestion creates a new question.
	CreateQuestion(ctx context.Context, req CreateQuestionRequest) (*CreateQuestionResponse, error)

	// GetQuestion retrieves a question by ID.
	GetQuestion(ctx context.Context, questionID string) (*Question, error)

	// ListQuestions lists questions with optional filters.
	ListQuestions(ctx context.Context, filters QuestionFilters) ([]*Question, error)

	// AnswerQuestion answers a question (marks it as answered).
	AnswerQuestion(ctx context.Context, questionID, answer string) error

	// UpdateQuestion updates a question's title and/or description.
	UpdateQuestion(ctx context.Context, req UpdateQuestionRequest) error

	// PinQuestion pins a question.
	PinQuestion(ctx context.Context, questionID string) error

	// UnpinQuestion unpins a question.
	UnpinQuestion(ctx context.Context, questionID string) error

	// DeleteQuestion deletes a question.
	DeleteQuestion(ctx context.Context, questionID string) error
}

// CreateQuestionRequest contains parameters for creating a question.
type CreateQuestionRequest struct {
	MissionID       string
	InvestigationID string // Optional
	Title           string
	Description     string
}

// CreateQuestionResponse contains the result of creating a question.
type CreateQuestionResponse struct {
	QuestionID string
	Question   *Question
}

// UpdateQuestionRequest contains parameters for updating a question.
type UpdateQuestionRequest struct {
	QuestionID  string
	Title       string
	Description string
}

// Question represents a question entity at the port boundary.
type Question struct {
	ID               string
	InvestigationID  string
	MissionID        string
	Title            string
	Description      string
	Status           string
	Answer           string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	AnsweredAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// QuestionFilters contains filter options for listing questions.
type QuestionFilters struct {
	InvestigationID string
	MissionID       string
	Status          string
}
