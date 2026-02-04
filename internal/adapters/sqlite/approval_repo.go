// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ApprovalRepository implements secondary.ApprovalRepository with SQLite.
type ApprovalRepository struct {
	db        *sql.DB
	logWriter secondary.LogWriter
}

// NewApprovalRepository creates a new SQLite approval repository.
// logWriter is optional - if nil, no audit logging is performed.
func NewApprovalRepository(db *sql.DB, logWriter secondary.LogWriter) *ApprovalRepository {
	return &ApprovalRepository{db: db, logWriter: logWriter}
}

// Create persists a new approval.
func (r *ApprovalRepository) Create(ctx context.Context, approval *secondary.ApprovalRecord) error {
	var reviewerInput, reviewerOutput sql.NullString
	if approval.ReviewerInput != "" {
		reviewerInput = sql.NullString{String: approval.ReviewerInput, Valid: true}
	}
	if approval.ReviewerOutput != "" {
		reviewerOutput = sql.NullString{String: approval.ReviewerOutput, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO approvals (id, plan_id, task_id, mechanism, reviewer_input, reviewer_output, outcome) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		approval.ID,
		approval.PlanID,
		approval.TaskID,
		approval.Mechanism,
		reviewerInput,
		reviewerOutput,
		approval.Outcome,
	)
	if err != nil {
		return fmt.Errorf("failed to create approval: %w", err)
	}

	// Log create operation
	if r.logWriter != nil {
		_ = r.logWriter.LogCreate(ctx, "approval", approval.ID)
	}

	return nil
}

// GetByID retrieves an approval by its ID.
func (r *ApprovalRepository) GetByID(ctx context.Context, id string) (*secondary.ApprovalRecord, error) {
	var (
		reviewerInput  sql.NullString
		reviewerOutput sql.NullString
		createdAt      time.Time
	)

	record := &secondary.ApprovalRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, plan_id, task_id, mechanism, reviewer_input, reviewer_output, outcome, created_at FROM approvals WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.PlanID, &record.TaskID, &record.Mechanism, &reviewerInput, &reviewerOutput, &record.Outcome, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("approval %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval: %w", err)
	}
	record.ReviewerInput = reviewerInput.String
	record.ReviewerOutput = reviewerOutput.String
	record.CreatedAt = createdAt.Format(time.RFC3339)

	return record, nil
}

// GetByPlan retrieves an approval by plan ID.
func (r *ApprovalRepository) GetByPlan(ctx context.Context, planID string) (*secondary.ApprovalRecord, error) {
	var (
		reviewerInput  sql.NullString
		reviewerOutput sql.NullString
		createdAt      time.Time
	)

	record := &secondary.ApprovalRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, plan_id, task_id, mechanism, reviewer_input, reviewer_output, outcome, created_at FROM approvals WHERE plan_id = ?`,
		planID,
	).Scan(&record.ID, &record.PlanID, &record.TaskID, &record.Mechanism, &reviewerInput, &reviewerOutput, &record.Outcome, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("approval for plan %s not found", planID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval: %w", err)
	}
	record.ReviewerInput = reviewerInput.String
	record.ReviewerOutput = reviewerOutput.String
	record.CreatedAt = createdAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves approvals matching the given filters.
func (r *ApprovalRepository) List(ctx context.Context, filters secondary.ApprovalFilters) ([]*secondary.ApprovalRecord, error) {
	query := `SELECT id, plan_id, task_id, mechanism, reviewer_input, reviewer_output, outcome, created_at FROM approvals WHERE 1=1`
	args := []any{}

	if filters.TaskID != "" {
		query += " AND task_id = ?"
		args = append(args, filters.TaskID)
	}

	if filters.Outcome != "" {
		query += " AND outcome = ?"
		args = append(args, filters.Outcome)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list approvals: %w", err)
	}
	defer rows.Close()

	var approvals []*secondary.ApprovalRecord
	for rows.Next() {
		var (
			reviewerInput  sql.NullString
			reviewerOutput sql.NullString
			createdAt      time.Time
		)

		record := &secondary.ApprovalRecord{}
		err := rows.Scan(&record.ID, &record.PlanID, &record.TaskID, &record.Mechanism, &reviewerInput, &reviewerOutput, &record.Outcome, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval: %w", err)
		}
		record.ReviewerInput = reviewerInput.String
		record.ReviewerOutput = reviewerOutput.String
		record.CreatedAt = createdAt.Format(time.RFC3339)

		approvals = append(approvals, record)
	}

	return approvals, nil
}

// Delete removes an approval from persistence.
func (r *ApprovalRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM approvals WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete approval: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("approval %s not found", id)
	}

	return nil
}

// GetNextID returns the next available approval ID.
func (r *ApprovalRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("APPR-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM approvals", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next approval ID: %w", err)
	}

	return fmt.Sprintf("APPR-%03d", maxID+1), nil
}

// PlanExists checks if a plan exists.
func (r *ApprovalRepository) PlanExists(ctx context.Context, planID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM plans WHERE id = ?", planID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check plan existence: %w", err)
	}
	return count > 0, nil
}

// TaskExists checks if a task exists.
func (r *ApprovalRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check task existence: %w", err)
	}
	return count > 0, nil
}

// PlanHasApproval checks if a plan already has an approval (for 1:1 constraint).
func (r *ApprovalRepository) PlanHasApproval(ctx context.Context, planID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM approvals WHERE plan_id = ?", planID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing approval: %w", err)
	}
	return count > 0, nil
}

// Ensure ApprovalRepository implements the interface
var _ secondary.ApprovalRepository = (*ApprovalRepository)(nil)
