// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// QuestionRepository implements secondary.QuestionRepository with SQLite.
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new SQLite question repository.
func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// Create persists a new question.
func (r *QuestionRepository) Create(ctx context.Context, question *secondary.QuestionRecord) error {
	var desc sql.NullString
	if question.Description != "" {
		desc = sql.NullString{String: question.Description, Valid: true}
	}

	var investigationID sql.NullString
	if question.InvestigationID != "" {
		investigationID = sql.NullString{String: question.InvestigationID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO questions (id, investigation_id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?, ?)",
		question.ID, investigationID, question.MissionID, question.Title, desc, "open",
	)
	if err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}

	return nil
}

// GetByID retrieves a question by its ID.
func (r *QuestionRepository) GetByID(ctx context.Context, id string) (*secondary.QuestionRecord, error) {
	var (
		investigationID  sql.NullString
		desc             sql.NullString
		answer           sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		answeredAt       sql.NullTime
		conclaveID       sql.NullString
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.QuestionRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, investigation_id, mission_id, title, description, status, answer, pinned,
			created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type
		FROM questions WHERE id = ?`,
		id,
	).Scan(&record.ID, &investigationID, &record.MissionID, &record.Title, &desc, &record.Status, &answer, &pinned,
		&createdAt, &updatedAt, &answeredAt, &conclaveID, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("question %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	record.InvestigationID = investigationID.String
	record.Description = desc.String
	record.Answer = answer.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if answeredAt.Valid {
		record.AnsweredAt = answeredAt.Time.Format(time.RFC3339)
	}
	record.ConclaveID = conclaveID.String
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// List retrieves questions matching the given filters.
func (r *QuestionRepository) List(ctx context.Context, filters secondary.QuestionFilters) ([]*secondary.QuestionRecord, error) {
	query := `SELECT id, investigation_id, mission_id, title, description, status, answer, pinned,
		created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type
		FROM questions WHERE 1=1`
	args := []any{}

	if filters.InvestigationID != "" {
		query += " AND investigation_id = ?"
		args = append(args, filters.InvestigationID)
	}

	if filters.MissionID != "" {
		query += " AND mission_id = ?"
		args = append(args, filters.MissionID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}
	defer rows.Close()

	var questions []*secondary.QuestionRecord
	for rows.Next() {
		var (
			investigationID  sql.NullString
			desc             sql.NullString
			answer           sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			answeredAt       sql.NullTime
			conclaveID       sql.NullString
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.QuestionRecord{}
		err := rows.Scan(&record.ID, &investigationID, &record.MissionID, &record.Title, &desc, &record.Status, &answer, &pinned,
			&createdAt, &updatedAt, &answeredAt, &conclaveID, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		record.InvestigationID = investigationID.String
		record.Description = desc.String
		record.Answer = answer.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if answeredAt.Valid {
			record.AnsweredAt = answeredAt.Time.Format(time.RFC3339)
		}
		record.ConclaveID = conclaveID.String
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		questions = append(questions, record)
	}

	return questions, nil
}

// Update updates an existing question.
func (r *QuestionRepository) Update(ctx context.Context, question *secondary.QuestionRecord) error {
	query := "UPDATE questions SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if question.Title != "" {
		query += ", title = ?"
		args = append(args, question.Title)
	}

	if question.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: question.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, question.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("question %s not found", question.ID)
	}

	return nil
}

// Delete removes a question from persistence.
func (r *QuestionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM questions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	return nil
}

// Pin pins a question.
func (r *QuestionRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE questions SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin question: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	return nil
}

// Unpin unpins a question.
func (r *QuestionRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE questions SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin question: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	return nil
}

// GetNextID returns the next available question ID.
func (r *QuestionRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 3) AS INTEGER)), 0) FROM questions",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next question ID: %w", err)
	}

	return fmt.Sprintf("Q-%03d", maxID+1), nil
}

// Answer sets the answer for a question and marks it as answered.
func (r *QuestionRepository) Answer(ctx context.Context, id, answer string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE questions SET status = 'answered', answer = ?, answered_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		answer, id,
	)
	if err != nil {
		return fmt.Errorf("failed to answer question: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("question %s not found", id)
	}

	return nil
}

// MissionExists checks if a mission exists.
func (r *QuestionRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// InvestigationExists checks if an investigation exists.
func (r *QuestionRepository) InvestigationExists(ctx context.Context, investigationID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM investigations WHERE id = ?", investigationID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check investigation existence: %w", err)
	}
	return count > 0, nil
}

// Ensure QuestionRepository implements the interface
var _ secondary.QuestionRepository = (*QuestionRepository)(nil)
