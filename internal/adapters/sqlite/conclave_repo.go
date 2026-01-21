// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ConclaveRepository implements secondary.ConclaveRepository with SQLite.
type ConclaveRepository struct {
	db *sql.DB
}

// NewConclaveRepository creates a new SQLite conclave repository.
func NewConclaveRepository(db *sql.DB) *ConclaveRepository {
	return &ConclaveRepository{db: db}
}

// Create persists a new conclave.
func (r *ConclaveRepository) Create(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	var desc sql.NullString
	if conclave.Description != "" {
		desc = sql.NullString{String: conclave.Description, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO conclaves (id, commission_id, title, description, status) VALUES (?, ?, ?, ?, ?)",
		conclave.ID, conclave.CommissionID, conclave.Title, desc, "active",
	)
	if err != nil {
		return fmt.Errorf("failed to create conclave: %w", err)
	}

	return nil
}

// GetByID retrieves a conclave by its ID.
func (r *ConclaveRepository) GetByID(ctx context.Context, id string) (*secondary.ConclaveRecord, error) {
	var (
		desc                sql.NullString
		assignedWorkbenchID sql.NullString
		pinned              bool
		createdAt           time.Time
		updatedAt           time.Time
		completedAt         sql.NullTime
	)

	record := &secondary.ConclaveRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conclave %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conclave: %w", err)
	}

	record.Description = desc.String
	record.AssignedWorkbenchID = assignedWorkbenchID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if completedAt.Valid {
		record.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves conclaves matching the given filters.
func (r *ConclaveRepository) List(ctx context.Context, filters secondary.ConclaveFilters) ([]*secondary.ConclaveRecord, error) {
	query := "SELECT id, commission_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE 1=1"
	args := []any{}

	if filters.CommissionID != "" {
		query += " AND commission_id = ?"
		args = append(args, filters.CommissionID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list conclaves: %w", err)
	}
	defer rows.Close()

	var conclaves []*secondary.ConclaveRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			completedAt         sql.NullTime
		)

		record := &secondary.ConclaveRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conclave: %w", err)
		}

		record.Description = desc.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		conclaves = append(conclaves, record)
	}

	return conclaves, nil
}

// Update updates an existing conclave.
func (r *ConclaveRepository) Update(ctx context.Context, conclave *secondary.ConclaveRecord) error {
	query := "UPDATE conclaves SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if conclave.Title != "" {
		query += ", title = ?"
		args = append(args, conclave.Title)
	}

	if conclave.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: conclave.Description, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, conclave.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", conclave.ID)
	}

	return nil
}

// Delete removes a conclave from persistence.
func (r *ConclaveRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM conclaves WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// Pin pins a conclave.
func (r *ConclaveRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE conclaves SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// Unpin unpins a conclave.
func (r *ConclaveRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE conclaves SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin conclave: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// GetNextID returns the next available conclave ID.
func (r *ConclaveRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM conclaves",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next conclave ID: %w", err)
	}

	return fmt.Sprintf("CON-%03d", maxID+1), nil
}

// UpdateStatus updates the status and optionally completed_at timestamp.
func (r *ConclaveRepository) UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error {
	var query string
	var args []any

	if setCompleted {
		query = "UPDATE conclaves SET status = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	} else {
		query = "UPDATE conclaves SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
		args = []any{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update conclave status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("conclave %s not found", id)
	}

	return nil
}

// GetByWorkbench retrieves conclaves assigned to a grove.
func (r *ConclaveRepository) GetByWorkbench(ctx context.Context, workbenchID string) ([]*secondary.ConclaveRecord, error) {
	query := "SELECT id, commission_id, title, description, status, assigned_workbench_id, pinned, created_at, updated_at, completed_at FROM conclaves WHERE assigned_workbench_id = ?"
	rows, err := r.db.QueryContext(ctx, query, workbenchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conclaves by grove: %w", err)
	}
	defer rows.Close()

	var conclaves []*secondary.ConclaveRecord
	for rows.Next() {
		var (
			desc                sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			completedAt         sql.NullTime
		)

		record := &secondary.ConclaveRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &desc, &record.Status, &assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conclave: %w", err)
		}

		record.Description = desc.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}

		conclaves = append(conclaves, record)
	}

	return conclaves, nil
}

// CommissionExists checks if a mission exists.
func (r *ConclaveRepository) CommissionExists(ctx context.Context, missionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", missionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check mission existence: %w", err)
	}
	return count > 0, nil
}

// GetTasksByConclave retrieves tasks belonging to a conclave.
func (r *ConclaveRepository) GetTasksByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveTaskRecord, error) {
	query := `SELECT id, shipment_id, commission_id, title, description, type, status, priority,
		assigned_workbench_id, pinned, created_at, updated_at, claimed_at, completed_at,
		conclave_id, promoted_from_id, promoted_from_type
		FROM tasks WHERE conclave_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by conclave: %w", err)
	}
	defer rows.Close()

	var tasks []*secondary.ConclaveTaskRecord
	for rows.Next() {
		var (
			shipmentID          sql.NullString
			desc                sql.NullString
			taskType            sql.NullString
			priority            sql.NullString
			assignedWorkbenchID sql.NullString
			pinned              bool
			createdAt           time.Time
			updatedAt           time.Time
			claimedAt           sql.NullTime
			completedAt         sql.NullTime
			conclaveIDCol       sql.NullString
			promotedFromID      sql.NullString
			promotedFromType    sql.NullString
		)

		record := &secondary.ConclaveTaskRecord{}
		err := rows.Scan(&record.ID, &shipmentID, &record.CommissionID, &record.Title, &desc, &taskType, &record.Status, &priority,
			&assignedWorkbenchID, &pinned, &createdAt, &updatedAt, &claimedAt, &completedAt,
			&conclaveIDCol, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		record.ShipmentID = shipmentID.String
		record.Description = desc.String
		record.Type = taskType.String
		record.Priority = priority.String
		record.AssignedWorkbenchID = assignedWorkbenchID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if claimedAt.Valid {
			record.ClaimedAt = claimedAt.Time.Format(time.RFC3339)
		}
		if completedAt.Valid {
			record.CompletedAt = completedAt.Time.Format(time.RFC3339)
		}
		record.ConclaveID = conclaveIDCol.String
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		tasks = append(tasks, record)
	}

	return tasks, nil
}

// GetQuestionsByConclave retrieves questions belonging to a conclave.
func (r *ConclaveRepository) GetQuestionsByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclaveQuestionRecord, error) {
	query := `SELECT id, investigation_id, commission_id, title, description, status, answer, pinned,
		created_at, updated_at, answered_at, conclave_id, promoted_from_id, promoted_from_type
		FROM questions WHERE conclave_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions by conclave: %w", err)
	}
	defer rows.Close()

	var questions []*secondary.ConclaveQuestionRecord
	for rows.Next() {
		var (
			investigationID  sql.NullString
			desc             sql.NullString
			answer           sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			answeredAt       sql.NullTime
			conclaveIDCol    sql.NullString
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.ConclaveQuestionRecord{}
		err := rows.Scan(&record.ID, &investigationID, &record.CommissionID, &record.Title, &desc, &record.Status, &answer, &pinned,
			&createdAt, &updatedAt, &answeredAt, &conclaveIDCol, &promotedFromID, &promotedFromType)
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
		record.ConclaveID = conclaveIDCol.String
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		questions = append(questions, record)
	}

	return questions, nil
}

// GetPlansByConclave retrieves plans belonging to a conclave.
func (r *ConclaveRepository) GetPlansByConclave(ctx context.Context, conclaveID string) ([]*secondary.ConclavePlanRecord, error) {
	query := `SELECT id, shipment_id, commission_id, title, description, status, content, pinned,
		created_at, updated_at, approved_at, conclave_id, promoted_from_id, promoted_from_type
		FROM plans WHERE conclave_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, conclaveID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans by conclave: %w", err)
	}
	defer rows.Close()

	var plans []*secondary.ConclavePlanRecord
	for rows.Next() {
		var (
			shipmentID       sql.NullString
			desc             sql.NullString
			content          sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			approvedAt       sql.NullTime
			conclaveIDCol    sql.NullString
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.ConclavePlanRecord{}
		err := rows.Scan(&record.ID, &shipmentID, &record.CommissionID, &record.Title, &desc, &record.Status, &content, &pinned,
			&createdAt, &updatedAt, &approvedAt, &conclaveIDCol, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		record.ShipmentID = shipmentID.String
		record.Description = desc.String
		record.Content = content.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
		}
		record.ConclaveID = conclaveIDCol.String
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		plans = append(plans, record)
	}

	return plans, nil
}

// Ensure ConclaveRepository implements the interface
var _ secondary.ConclaveRepository = (*ConclaveRepository)(nil)
