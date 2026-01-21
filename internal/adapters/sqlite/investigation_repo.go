// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// InvestigationRepository implements secondary.InvestigationRepository with SQLite.
type InvestigationRepository struct {
	db *sql.DB
}

// NewInvestigationRepository creates a new SQLite investigation repository.
func NewInvestigationRepository(db *sql.DB) *InvestigationRepository {
	return &InvestigationRepository{db: db}
}

// Create persists a new investigation.
func (r *InvestigationRepository) Create(ctx context.Context, investigation *secondary.InvestigationRecord) error {
	var desc sql.NullString
	if investigation.Description != "" {
		desc = sql.NullString{String: investigation.Description, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO investigations (id, mission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		investigation.ID, investigation.MissionID, investigation.Title, desc, "active",
	)
	if err != nil {
		return fmt.Errorf("failed to create investigation: %w", err)
	}

	return nil
}

// GetByID retrieves an investigation by its ID.
func (r *InvestigationRepository) GetByID(ctx context.Context, id string) (*secondary.InvestigationRecord, error) {
	var (
		desc            sql.NullString
		assignedGroveID sql.NullString
		pinned          bool
		createdAt       time.Time
		updatedAt       time.Time
		completedAt     sql.NullTime
	)

	record := &secondary.InvestigationRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE id = ?",
		id,
	).Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("investigation %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get investigation: %w", err)
	}

	record.Description = desc.String
	record.AssignedGroveID = assignedGroveID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves investigations matching the given filters.
func (r *InvestigationRepository) List(ctx context.Context, filters secondary.InvestigationFilters) ([]*secondary.InvestigationRecord, error) {
	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE 1=1"
	args := []any{}

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
		return nil, fmt.Errorf("failed to list investigations: %w", err)
	}
	defer rows.Close()

	var investigations []*secondary.InvestigationRecord
	for rows.Next() {
		var (
			desc            sql.NullString
			assignedGroveID sql.NullString
			pinned          bool
			createdAt       time.Time
			updatedAt       time.Time
			completedAt     sql.NullTime
		)

		record := &secondary.InvestigationRecord{}
		err := rows.Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan investigation: %w", err)
		}

		record.Description = desc.String
		record.AssignedGroveID = assignedGroveID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		investigations = append(investigations, record)
	}

	return investigations, nil
}

// Update updates an existing investigation.
func (r *InvestigationRepository) Update(ctx context.Context, investigation *secondary.InvestigationRecord) error {
	query := "UPDATE investigations SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if investigation.Title != "" {
		query += ", title = ?"
		args = append(args, investigation.Title)
	}

	if investigation.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: investigation.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, investigation.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update investigation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", investigation.ID)
	}

	return nil
}

// Delete removes an investigation from persistence.
func (r *InvestigationRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM investigations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete investigation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	return nil
}

// Pin pins an investigation.
func (r *InvestigationRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE investigations SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin investigation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	return nil
}

// Unpin unpins an investigation.
func (r *InvestigationRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE investigations SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin investigation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	return nil
}

// GetNextID returns the next available investigation ID.
func (r *InvestigationRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM investigations",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next investigation ID: %w", err)
	}

	return fmt.Sprintf("INV-%03d", maxID+1), nil
}

// UpdateStatus updates the status and optionally completed_at timestamp.
func (r *InvestigationRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE investigations SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE investigations SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update investigation status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", id)
	}

	return nil
}

// GetByGrove retrieves investigations assigned to a grove.
func (r *InvestigationRepository) GetByGrove(ctx context.Context, groveID string) ([]*secondary.InvestigationRecord, error) {
	query := "SELECT id, mission_id, title, description, status, assigned_grove_id, pinned, created_at, updated_at, completed_at FROM investigations WHERE assigned_grove_id = ?"
	rows, err := r.db.QueryContext(ctx, query, groveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get investigations by grove: %w", err)
	}
	defer rows.Close()

	var investigations []*secondary.InvestigationRecord
	for rows.Next() {
		var (
			desc            sql.NullString
			assignedGroveID sql.NullString
			pinned          bool
			createdAt       time.Time
			updatedAt       time.Time
			completedAt     sql.NullTime
		)

		record := &secondary.InvestigationRecord{}
		err := rows.Scan(&record.ID, &record.MissionID, &record.Title, &desc, &record.Status, &assignedGroveID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan investigation: %w", err)
		}

		record.Description = desc.String
		record.AssignedGroveID = assignedGroveID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		investigations = append(investigations, record)
	}

	return investigations, nil
}

// AssignGrove assigns an investigation to a grove.
func (r *InvestigationRepository) AssignGrove(ctx context.Context, investigationID, groveID string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE investigations SET assigned_grove_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		groveID, investigationID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign investigation to grove: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("investigation %s not found", investigationID)
	}

	return nil
}

// MissionExists checks if a mission exists.
func (r *InvestigationRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// GetQuestionsByInvestigation retrieves questions for an investigation.
func (r *InvestigationRepository) GetQuestionsByInvestigation(ctx context.Context, investigationID string) ([]*secondary.InvestigationQuestionRecord, error) {
	query := `SELECT id, investigation_id, mission_id, title, description, status, answer, pinned,
		created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type
		FROM questions WHERE investigation_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, investigationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions by investigation: %w", err)
	}
	defer rows.Close()

	var questions []*secondary.InvestigationQuestionRecord
	for rows.Next() {
		var (
			investigationIDCol sql.NullString
			desc               sql.NullString
			answer             sql.NullString
			pinned             bool
			createdAt          time.Time
			updatedAt          time.Time
			answeredAt         sql.NullTime
			conclaveID         sql.NullString
			promotedFromID     sql.NullString
			promotedFromType   sql.NullString
		)

		record := &secondary.InvestigationQuestionRecord{}
		err := rows.Scan(&record.ID, &investigationIDCol, &record.MissionID, &record.Title, &desc, &record.Status, &answer, &pinned,
			&createdAt, &updatedAt, &answeredAt, &conclaveID, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		record.InvestigationID = investigationIDCol.String
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

// Ensure InvestigationRepository implements the interface
var _ secondary.InvestigationRepository = (*InvestigationRepository)(nil)
