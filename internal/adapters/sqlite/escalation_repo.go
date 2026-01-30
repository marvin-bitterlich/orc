// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// EscalationRepository implements secondary.EscalationRepository with SQLite.
type EscalationRepository struct {
	db *sql.DB
}

// NewEscalationRepository creates a new SQLite escalation repository.
func NewEscalationRepository(db *sql.DB) *EscalationRepository {
	return &EscalationRepository{db: db}
}

// Create persists a new escalation.
func (r *EscalationRepository) Create(ctx context.Context, escalation *secondary.EscalationRecord) error {
	var approvalID, targetActorID sql.NullString
	if escalation.ApprovalID != "" {
		approvalID = sql.NullString{String: escalation.ApprovalID, Valid: true}
	}
	if escalation.TargetActorID != "" {
		targetActorID = sql.NullString{String: escalation.TargetActorID, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO escalations (id, approval_id, plan_id, task_id, reason, status, routing_rule, origin_actor_id, target_actor_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		escalation.ID,
		approvalID,
		escalation.PlanID,
		escalation.TaskID,
		escalation.Reason,
		escalation.Status,
		escalation.RoutingRule,
		escalation.OriginActorID,
		targetActorID,
	)
	if err != nil {
		return fmt.Errorf("failed to create escalation: %w", err)
	}

	return nil
}

// GetByID retrieves an escalation by its ID.
func (r *EscalationRepository) GetByID(ctx context.Context, id string) (*secondary.EscalationRecord, error) {
	var (
		approvalID    sql.NullString
		targetActorID sql.NullString
		resolution    sql.NullString
		resolvedBy    sql.NullString
		createdAt     time.Time
		resolvedAt    sql.NullTime
	)

	record := &secondary.EscalationRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, approval_id, plan_id, task_id, reason, status, routing_rule, origin_actor_id, target_actor_id, resolution, resolved_by, created_at, resolved_at FROM escalations WHERE id = ?`,
		id,
	).Scan(&record.ID, &approvalID, &record.PlanID, &record.TaskID, &record.Reason, &record.Status, &record.RoutingRule, &record.OriginActorID, &targetActorID, &resolution, &resolvedBy, &createdAt, &resolvedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("escalation %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get escalation: %w", err)
	}
	record.ApprovalID = approvalID.String
	record.TargetActorID = targetActorID.String
	record.Resolution = resolution.String
	record.ResolvedBy = resolvedBy.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	if resolvedAt.Valid {
		record.ResolvedAt = resolvedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves escalations matching the given filters.
func (r *EscalationRepository) List(ctx context.Context, filters secondary.EscalationFilters) ([]*secondary.EscalationRecord, error) {
	query := `SELECT id, approval_id, plan_id, task_id, reason, status, routing_rule, origin_actor_id, target_actor_id, resolution, resolved_by, created_at, resolved_at FROM escalations WHERE 1=1`
	args := []any{}

	if filters.TaskID != "" {
		query += " AND task_id = ?"
		args = append(args, filters.TaskID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	if filters.TargetActorID != "" {
		query += " AND target_actor_id = ?"
		args = append(args, filters.TargetActorID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list escalations: %w", err)
	}
	defer rows.Close()

	var escalations []*secondary.EscalationRecord
	for rows.Next() {
		var (
			approvalID    sql.NullString
			targetActorID sql.NullString
			resolution    sql.NullString
			resolvedBy    sql.NullString
			createdAt     time.Time
			resolvedAt    sql.NullTime
		)

		record := &secondary.EscalationRecord{}
		err := rows.Scan(&record.ID, &approvalID, &record.PlanID, &record.TaskID, &record.Reason, &record.Status, &record.RoutingRule, &record.OriginActorID, &targetActorID, &resolution, &resolvedBy, &createdAt, &resolvedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan escalation: %w", err)
		}
		record.ApprovalID = approvalID.String
		record.TargetActorID = targetActorID.String
		record.Resolution = resolution.String
		record.ResolvedBy = resolvedBy.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		if resolvedAt.Valid {
			record.ResolvedAt = resolvedAt.Time.Format(time.RFC3339)
		}

		escalations = append(escalations, record)
	}

	return escalations, nil
}

// Update updates an existing escalation.
func (r *EscalationRepository) Update(ctx context.Context, escalation *secondary.EscalationRecord) error {
	query := "UPDATE escalations SET "
	args := []any{}
	setClauses := []string{}

	if escalation.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, escalation.Status)
	}
	if escalation.Resolution != "" {
		setClauses = append(setClauses, "resolution = ?")
		args = append(args, escalation.Resolution)
	}
	if escalation.ResolvedBy != "" {
		setClauses = append(setClauses, "resolved_by = ?")
		args = append(args, escalation.ResolvedBy)
	}
	if escalation.TargetActorID != "" {
		setClauses = append(setClauses, "target_actor_id = ?")
		args = append(args, escalation.TargetActorID)
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query += setClauses[0]
	for _, clause := range setClauses[1:] {
		query += ", " + clause
	}

	query += " WHERE id = ?"
	args = append(args, escalation.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update escalation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("escalation %s not found", escalation.ID)
	}

	return nil
}

// Delete removes an escalation from persistence.
func (r *EscalationRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM escalations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete escalation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("escalation %s not found", id)
	}

	return nil
}

// GetNextID returns the next available escalation ID.
func (r *EscalationRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("ESC-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM escalations", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next escalation ID: %w", err)
	}

	return fmt.Sprintf("ESC-%03d", maxID+1), nil
}

// UpdateStatus updates the status of an escalation.
func (r *EscalationRepository) UpdateStatus(ctx context.Context, id, status string, setResolved bool) error {
	var query string
	if setResolved {
		query = "UPDATE escalations SET status = ?, resolved_at = CURRENT_TIMESTAMP WHERE id = ?"
	} else {
		query = "UPDATE escalations SET status = ? WHERE id = ?"
	}

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update escalation status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("escalation %s not found", id)
	}

	return nil
}

// Resolve resolves an escalation with resolution text.
func (r *EscalationRepository) Resolve(ctx context.Context, id, resolution, resolvedBy string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE escalations SET status = 'resolved', resolution = ?, resolved_by = ?, resolved_at = CURRENT_TIMESTAMP WHERE id = ?",
		resolution, resolvedBy, id,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve escalation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("escalation %s not found", id)
	}

	return nil
}

// PlanExists checks if a plan exists.
func (r *EscalationRepository) PlanExists(ctx context.Context, planID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM plans WHERE id = ?", planID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check plan existence: %w", err)
	}
	return count > 0, nil
}

// TaskExists checks if a task exists.
func (r *EscalationRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence: %w", err)
	}
	return count > 0, nil
}

// ApprovalExists checks if an approval exists.
func (r *EscalationRepository) ApprovalExists(ctx context.Context, approvalID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM approvals WHERE id = ?", approvalID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check approval existence: %w", err)
	}
	return count > 0, nil
}

// Ensure EscalationRepository implements the interface
var _ secondary.EscalationRepository = (*EscalationRepository)(nil)
