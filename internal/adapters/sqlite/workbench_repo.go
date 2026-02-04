// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// WorkbenchRepository implements secondary.WorkbenchRepository with SQLite.
type WorkbenchRepository struct {
	db        *sql.DB
	logWriter secondary.LogWriter
}

// NewWorkbenchRepository creates a new SQLite workbench repository.
// logWriter is optional - if nil, no audit logging is performed.
func NewWorkbenchRepository(db *sql.DB, logWriter secondary.LogWriter) *WorkbenchRepository {
	return &WorkbenchRepository{db: db, logWriter: logWriter}
}

// Create persists a new workbench.
func (r *WorkbenchRepository) Create(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	// Verify workshop exists
	var exists int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM workshops WHERE id = ?", workbench.WorkshopID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify workshop: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("workshop %s not found", workbench.WorkshopID)
	}

	// Generate workbench ID by finding max existing ID
	var maxID int
	err = r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM workbenches",
	).Scan(&maxID)
	if err != nil {
		return fmt.Errorf("failed to generate workbench ID: %w", err)
	}

	id := fmt.Sprintf("BENCH-%03d", maxID+1)

	status := workbench.Status
	if status == "" {
		status = "active"
	}

	var repoID sql.NullString
	if workbench.RepoID != "" {
		repoID = sql.NullString{String: workbench.RepoID, Valid: true}
	}

	var homeBranch, currentBranch sql.NullString
	if workbench.HomeBranch != "" {
		homeBranch = sql.NullString{String: workbench.HomeBranch, Valid: true}
	}
	if workbench.CurrentBranch != "" {
		currentBranch = sql.NullString{String: workbench.CurrentBranch, Valid: true}
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO workbenches (id, workshop_id, name, repo_id, status, home_branch, current_branch) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, workbench.WorkshopID, workbench.Name, repoID, status, homeBranch, currentBranch,
	)
	if err != nil {
		return fmt.Errorf("failed to create workbench: %w", err)
	}

	// Update the record with the generated ID
	workbench.ID = id

	// Log create operation
	if r.logWriter != nil {
		_ = r.logWriter.LogCreate(ctx, "workbench", id)
	}

	return nil
}

// GetByID retrieves a workbench by its ID.
func (r *WorkbenchRepository) GetByID(ctx context.Context, id string) (*secondary.WorkbenchRecord, error) {
	var (
		createdAt     time.Time
		updatedAt     time.Time
		repoID        sql.NullString
		homeBranch    sql.NullString
		currentBranch sql.NullString
		focusedID     sql.NullString
	)

	record := &secondary.WorkbenchRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, workshop_id, name, repo_id, status, home_branch, current_branch, focused_id, created_at, updated_at FROM workbenches WHERE id = ?",
		id,
	).Scan(&record.ID, &record.WorkshopID, &record.Name, &repoID, &record.Status, &homeBranch, &currentBranch, &focusedID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workbench %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workbench: %w", err)
	}

	if repoID.Valid {
		record.RepoID = repoID.String
	}
	if homeBranch.Valid {
		record.HomeBranch = homeBranch.String
	}
	if currentBranch.Valid {
		record.CurrentBranch = currentBranch.String
	}
	record.FocusedID = focusedID.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByPath retrieves a workbench by its file path.
// Path is computed as ~/wb/{name}, so we extract name from path and query by name.
func (r *WorkbenchRepository) GetByPath(ctx context.Context, path string) (*secondary.WorkbenchRecord, error) {
	// Extract workbench name from path (last component)
	name := filepath.Base(path)
	if name == "" || name == "." || name == "/" {
		return nil, fmt.Errorf("invalid path: %s", path)
	}

	var (
		createdAt     time.Time
		updatedAt     time.Time
		repoID        sql.NullString
		homeBranch    sql.NullString
		currentBranch sql.NullString
		focusedID     sql.NullString
	)

	record := &secondary.WorkbenchRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, workshop_id, name, repo_id, status, home_branch, current_branch, focused_id, created_at, updated_at FROM workbenches WHERE name = ? AND status = 'active'",
		name,
	).Scan(&record.ID, &record.WorkshopID, &record.Name, &repoID, &record.Status, &homeBranch, &currentBranch, &focusedID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workbench with path %s not found", path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workbench by path: %w", err)
	}

	if repoID.Valid {
		record.RepoID = repoID.String
	}
	if homeBranch.Valid {
		record.HomeBranch = homeBranch.String
	}
	if currentBranch.Valid {
		record.CurrentBranch = currentBranch.String
	}
	record.FocusedID = focusedID.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByWorkshop retrieves all workbenches for a workshop.
func (r *WorkbenchRepository) GetByWorkshop(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, workshop_id, name, repo_id, status, home_branch, current_branch, focused_id, created_at, updated_at FROM workbenches WHERE workshop_id = ? AND status = 'active' ORDER BY created_at DESC",
		workshopID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list workbenches: %w", err)
	}
	defer rows.Close()

	var workbenches []*secondary.WorkbenchRecord
	for rows.Next() {
		var (
			createdAt     time.Time
			updatedAt     time.Time
			repoID        sql.NullString
			homeBranch    sql.NullString
			currentBranch sql.NullString
			focusedID     sql.NullString
		)

		record := &secondary.WorkbenchRecord{}
		err := rows.Scan(&record.ID, &record.WorkshopID, &record.Name, &repoID, &record.Status, &homeBranch, &currentBranch, &focusedID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workbench: %w", err)
		}

		if repoID.Valid {
			record.RepoID = repoID.String
		}
		if homeBranch.Valid {
			record.HomeBranch = homeBranch.String
		}
		if currentBranch.Valid {
			record.CurrentBranch = currentBranch.String
		}
		record.FocusedID = focusedID.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		workbenches = append(workbenches, record)
	}

	return workbenches, nil
}

// List retrieves all workbenches, optionally filtered by workshop.
func (r *WorkbenchRepository) List(ctx context.Context, workshopID string) ([]*secondary.WorkbenchRecord, error) {
	query := "SELECT id, workshop_id, name, repo_id, status, home_branch, current_branch, focused_id, created_at, updated_at FROM workbenches WHERE 1=1"
	args := []any{}

	if workshopID != "" {
		query += " AND workshop_id = ?"
		args = append(args, workshopID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workbenches: %w", err)
	}
	defer rows.Close()

	var workbenches []*secondary.WorkbenchRecord
	for rows.Next() {
		var (
			createdAt     time.Time
			updatedAt     time.Time
			repoID        sql.NullString
			homeBranch    sql.NullString
			currentBranch sql.NullString
			focusedID     sql.NullString
		)

		record := &secondary.WorkbenchRecord{}
		err := rows.Scan(&record.ID, &record.WorkshopID, &record.Name, &repoID, &record.Status, &homeBranch, &currentBranch, &focusedID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workbench: %w", err)
		}

		if repoID.Valid {
			record.RepoID = repoID.String
		}
		if homeBranch.Valid {
			record.HomeBranch = homeBranch.String
		}
		if currentBranch.Valid {
			record.CurrentBranch = currentBranch.String
		}
		record.FocusedID = focusedID.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		workbenches = append(workbenches, record)
	}

	return workbenches, nil
}

// Update updates an existing workbench.
func (r *WorkbenchRepository) Update(ctx context.Context, workbench *secondary.WorkbenchRecord) error {
	query := "UPDATE workbenches SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if workbench.Name != "" {
		query += ", name = ?"
		args = append(args, workbench.Name)
	}

	if workbench.Status != "" {
		query += ", status = ?"
		args = append(args, workbench.Status)
	}

	if workbench.HomeBranch != "" {
		query += ", home_branch = ?"
		args = append(args, workbench.HomeBranch)
	}

	if workbench.CurrentBranch != "" {
		query += ", current_branch = ?"
		args = append(args, workbench.CurrentBranch)
	}

	query += " WHERE id = ?"
	args = append(args, workbench.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workbench: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workbench %s not found", workbench.ID)
	}

	return nil
}

// Delete removes a workbench from persistence.
func (r *WorkbenchRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM workbenches WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workbench: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workbench %s not found", id)
	}

	return nil
}

// Rename updates the name of a workbench.
func (r *WorkbenchRepository) Rename(ctx context.Context, id, newName string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE workbenches SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		newName, id,
	)
	if err != nil {
		return fmt.Errorf("failed to rename workbench: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workbench %s not found", id)
	}

	return nil
}

// UpdatePath is deprecated - path is now computed dynamically as ~/wb/{name}.
// This method exists for interface compatibility but does nothing.
func (r *WorkbenchRepository) UpdatePath(ctx context.Context, id, newPath string) error {
	// Path is computed dynamically, not stored in DB
	return nil
}

// UpdateFocusedID updates the focused container ID for a workbench.
// Pass empty string to clear focus.
func (r *WorkbenchRepository) UpdateFocusedID(ctx context.Context, id, focusedID string) error {
	var focusedValue any
	if focusedID == "" {
		focusedValue = nil
	} else {
		focusedValue = focusedID
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE workbenches SET focused_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		focusedValue, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update workbench focus: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workbench %s not found", id)
	}

	return nil
}

// GetByFocusedID retrieves all active workbenches focusing a specific container.
func (r *WorkbenchRepository) GetByFocusedID(ctx context.Context, focusedID string) ([]*secondary.WorkbenchRecord, error) {
	if focusedID == "" {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workshop_id, name, repo_id, status, home_branch, current_branch, focused_id, created_at, updated_at
		FROM workbenches WHERE focused_id = ? AND status = 'active'`,
		focusedID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workbenches by focused_id: %w", err)
	}
	defer rows.Close()

	var workbenches []*secondary.WorkbenchRecord
	for rows.Next() {
		var (
			repoID        sql.NullString
			homeBranch    sql.NullString
			currentBranch sql.NullString
			focusedIDVal  sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
		)

		record := &secondary.WorkbenchRecord{}
		err := rows.Scan(
			&record.ID, &record.WorkshopID, &record.Name,
			&repoID, &record.Status, &homeBranch, &currentBranch, &focusedIDVal,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workbench: %w", err)
		}

		record.RepoID = repoID.String
		record.HomeBranch = homeBranch.String
		record.CurrentBranch = currentBranch.String
		record.FocusedID = focusedIDVal.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		workbenches = append(workbenches, record)
	}

	return workbenches, nil
}

// GetNextID returns the next available workbench ID.
func (r *WorkbenchRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 7) AS INTEGER)), 0) FROM workbenches",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next workbench ID: %w", err)
	}

	return fmt.Sprintf("BENCH-%03d", maxID+1), nil
}

// WorkshopExists checks if a workshop exists.
func (r *WorkbenchRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM workshops WHERE id = ?",
		workshopID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check workshop existence: %w", err)
	}

	return count > 0, nil
}

// Ensure WorkbenchRepository implements the interface
var _ secondary.WorkbenchRepository = (*WorkbenchRepository)(nil)
