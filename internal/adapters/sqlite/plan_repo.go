// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// PlanRepository implements secondary.PlanRepository with SQLite.
type PlanRepository struct {
	db        *sql.DB
	logWriter secondary.LogWriter
}

// NewPlanRepository creates a new SQLite plan repository.
// logWriter is optional - if nil, no audit logging is performed.
func NewPlanRepository(db *sql.DB, logWriter secondary.LogWriter) *PlanRepository {
	return &PlanRepository{db: db, logWriter: logWriter}
}

// Create persists a new plan.
func (r *PlanRepository) Create(ctx context.Context, plan *secondary.PlanRecord) error {
	var desc sql.NullString
	if plan.Description != "" {
		desc = sql.NullString{String: plan.Description, Valid: true}
	}

	var content sql.NullString
	if plan.Content != "" {
		content = sql.NullString{String: plan.Content, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO plans (id, task_id, commission_id, title, description, content, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		plan.ID, plan.TaskID, plan.CommissionID, plan.Title, desc, content, "draft",
	)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Log create operation
	if r.logWriter != nil {
		_ = r.logWriter.LogCreate(ctx, "plan", plan.ID)
	}

	return nil
}

// GetByID retrieves a plan by its ID.
func (r *PlanRepository) GetByID(ctx context.Context, id string) (*secondary.PlanRecord, error) {
	var (
		desc             sql.NullString
		content          sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		approvedAt       sql.NullTime
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.PlanRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, task_id, commission_id, title, description, status, content, pinned,
			created_at, updated_at, approved_at, promoted_from_id, promoted_from_type
		FROM plans WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.TaskID, &record.CommissionID, &record.Title, &desc, &record.Status, &content, &pinned,
		&createdAt, &updatedAt, &approvedAt, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	record.Description = desc.String
	record.Content = content.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if approvedAt.Valid {
		record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
	}
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// List retrieves plans matching the given filters.
func (r *PlanRepository) List(ctx context.Context, filters secondary.PlanFilters) ([]*secondary.PlanRecord, error) {
	query := `SELECT id, task_id, commission_id, title, description, status, content, pinned,
		created_at, updated_at, approved_at, promoted_from_id, promoted_from_type
		FROM plans WHERE 1=1`
	args := []any{}

	if filters.TaskID != "" {
		query += " AND task_id = ?"
		args = append(args, filters.TaskID)
	}

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
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []*secondary.PlanRecord
	for rows.Next() {
		var (
			desc             sql.NullString
			content          sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			approvedAt       sql.NullTime
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.PlanRecord{}
		err := rows.Scan(&record.ID, &record.TaskID, &record.CommissionID, &record.Title, &desc, &record.Status, &content, &pinned,
			&createdAt, &updatedAt, &approvedAt, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		record.Description = desc.String
		record.Content = content.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if approvedAt.Valid {
			record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
		}
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		plans = append(plans, record)
	}

	return plans, nil
}

// Update updates an existing plan.
func (r *PlanRepository) Update(ctx context.Context, plan *secondary.PlanRecord) error {
	query := "UPDATE plans SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if plan.Title != "" {
		query += ", title = ?"
		args = append(args, plan.Title)
	}

	if plan.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: plan.Description, Valid: true})
	}

	if plan.Content != "" {
		query += ", content = ?"
		args = append(args, sql.NullString{String: plan.Content, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, plan.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", plan.ID)
	}

	return nil
}

// Delete removes a plan from persistence.
func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM plans WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// Pin pins a plan.
func (r *PlanRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// Unpin unpins a plan.
func (r *PlanRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// GetNextID returns the next available plan ID.
func (r *PlanRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM plans",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next plan ID: %w", err)
	}

	return fmt.Sprintf("PLAN-%03d", maxID+1), nil
}

// Approve approves a plan and sets the approved_at timestamp.
func (r *PlanRepository) Approve(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET status = 'approved', approved_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// UpdateStatus updates the plan status.
func (r *PlanRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE plans SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("plan %s not found", id)
	}

	return nil
}

// GetActivePlanForTask retrieves the active (draft) plan for a task.
func (r *PlanRepository) GetActivePlanForTask(ctx context.Context, taskID string) (*secondary.PlanRecord, error) {
	var (
		desc             sql.NullString
		content          sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		approvedAt       sql.NullTime
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.PlanRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, task_id, commission_id, title, description, status, content, pinned,
			created_at, updated_at, approved_at, promoted_from_id, promoted_from_type
		FROM plans WHERE task_id = ? AND status = 'draft' LIMIT 1`,
		taskID,
	).Scan(&record.ID, &record.TaskID, &record.CommissionID, &record.Title, &desc, &record.Status, &content, &pinned,
		&createdAt, &updatedAt, &approvedAt, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, nil // No active plan is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active plan for task: %w", err)
	}

	record.Description = desc.String
	record.Content = content.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if approvedAt.Valid {
		record.ApprovedAt = approvedAt.Time.Format(time.RFC3339)
	}
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// HasActivePlanForTask checks if a task has an active (draft) plan.
func (r *PlanRepository) HasActivePlanForTask(ctx context.Context, taskID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM plans WHERE task_id = ? AND status = 'draft'",
		taskID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check active plan for task: %w", err)
	}
	return count > 0, nil
}

// CommissionExists checks if a commission exists.
func (r *PlanRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// TaskExists checks if a task exists.
func (r *PlanRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence: %w", err)
	}
	return count > 0, nil
}

// Ensure PlanRepository implements the interface
var _ secondary.PlanRepository = (*PlanRepository)(nil)
