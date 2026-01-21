// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// PRRepository implements secondary.PRRepository with SQLite.
type PRRepository struct {
	db *sql.DB
}

// NewPRRepository creates a new SQLite PR repository.
func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

// Create persists a new pull request.
func (r *PRRepository) Create(ctx context.Context, pr *secondary.PRRecord) error {
	var description, targetBranch, url sql.NullString
	var number sql.NullInt64

	if pr.Description != "" {
		description = sql.NullString{String: pr.Description, Valid: true}
	}
	if pr.TargetBranch != "" {
		targetBranch = sql.NullString{String: pr.TargetBranch, Valid: true}
	}
	if pr.URL != "" {
		url = sql.NullString{String: pr.URL, Valid: true}
	}
	if pr.Number > 0 {
		number = sql.NullInt64{Int64: int64(pr.Number), Valid: true}
	}

	status := pr.Status
	if status == "" {
		status = "open"
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO prs (id, shipment_id, repo_id, commission_id, number, title, description, branch, target_branch, url, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pr.ID, pr.ShipmentID, pr.RepoID, pr.CommissionID, number, pr.Title, description, pr.Branch, targetBranch, url, status,
	)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	return nil
}

// GetByID retrieves a pull request by its ID.
func (r *PRRepository) GetByID(ctx context.Context, id string) (*secondary.PRRecord, error) {
	var (
		number       sql.NullInt64
		description  sql.NullString
		targetBranch sql.NullString
		url          sql.NullString
		status       string
		createdAt    time.Time
		updatedAt    time.Time
		mergedAt     sql.NullTime
		closedAt     sql.NullTime
	)

	record := &secondary.PRRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, repo_id, commission_id, number, title, description, branch, target_branch, url, status, created_at, updated_at, merged_at, closed_at
		 FROM prs WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.ShipmentID, &record.RepoID, &record.CommissionID, &number, &record.Title, &description, &record.Branch, &targetBranch, &url, &status, &createdAt, &updatedAt, &mergedAt, &closedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("PR %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if number.Valid {
		record.Number = int(number.Int64)
	}
	record.Description = description.String
	record.TargetBranch = targetBranch.String
	record.URL = url.String
	record.Status = status
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if mergedAt.Valid {
		record.MergedAt = mergedAt.Time.Format(time.RFC3339)
	}
	if closedAt.Valid {
		record.ClosedAt = closedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// GetByShipment retrieves a pull request by shipment ID.
func (r *PRRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.PRRecord, error) {
	var (
		number       sql.NullInt64
		description  sql.NullString
		targetBranch sql.NullString
		url          sql.NullString
		status       string
		createdAt    time.Time
		updatedAt    time.Time
		mergedAt     sql.NullTime
		closedAt     sql.NullTime
	)

	record := &secondary.PRRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, repo_id, commission_id, number, title, description, branch, target_branch, url, status, created_at, updated_at, merged_at, closed_at
		 FROM prs WHERE shipment_id = ?`,
		shipmentID,
	).Scan(&record.ID, &record.ShipmentID, &record.RepoID, &record.CommissionID, &number, &record.Title, &description, &record.Branch, &targetBranch, &url, &status, &createdAt, &updatedAt, &mergedAt, &closedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil for "not found"
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PR by shipment: %w", err)
	}

	if number.Valid {
		record.Number = int(number.Int64)
	}
	record.Description = description.String
	record.TargetBranch = targetBranch.String
	record.URL = url.String
	record.Status = status
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if mergedAt.Valid {
		record.MergedAt = mergedAt.Time.Format(time.RFC3339)
	}
	if closedAt.Valid {
		record.ClosedAt = closedAt.Time.Format(time.RFC3339)
	}

	return record, nil
}

// List retrieves pull requests matching the given filters.
func (r *PRRepository) List(ctx context.Context, filters secondary.PRFilters) ([]*secondary.PRRecord, error) {
	query := `SELECT id, shipment_id, repo_id, commission_id, number, title, description, branch, target_branch, url, status, created_at, updated_at, merged_at, closed_at
			  FROM prs WHERE 1=1`
	args := []any{}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.RepoID != "" {
		query += " AND repo_id = ?"
		args = append(args, filters.RepoID)
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
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}
	defer rows.Close()

	var prs []*secondary.PRRecord
	for rows.Next() {
		var (
			number       sql.NullInt64
			description  sql.NullString
			targetBranch sql.NullString
			url          sql.NullString
			status       string
			createdAt    time.Time
			updatedAt    time.Time
			mergedAt     sql.NullTime
			closedAt     sql.NullTime
		)

		record := &secondary.PRRecord{}
		err := rows.Scan(&record.ID, &record.ShipmentID, &record.RepoID, &record.CommissionID, &number, &record.Title, &description, &record.Branch, &targetBranch, &url, &status, &createdAt, &updatedAt, &mergedAt, &closedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}

		if number.Valid {
			record.Number = int(number.Int64)
		}
		record.Description = description.String
		record.TargetBranch = targetBranch.String
		record.URL = url.String
		record.Status = status
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if mergedAt.Valid {
			record.MergedAt = mergedAt.Time.Format(time.RFC3339)
		}
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}

		prs = append(prs, record)
	}

	return prs, nil
}

// Update updates an existing pull request.
func (r *PRRepository) Update(ctx context.Context, pr *secondary.PRRecord) error {
	query := "UPDATE prs SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if pr.Title != "" {
		query += ", title = ?"
		args = append(args, pr.Title)
	}

	if pr.Description != "" {
		query += ", description = ?"
		args = append(args, sql.NullString{String: pr.Description, Valid: true})
	}

	if pr.URL != "" {
		query += ", url = ?"
		args = append(args, sql.NullString{String: pr.URL, Valid: true})
	}

	if pr.Number > 0 {
		query += ", number = ?"
		args = append(args, sql.NullInt64{Int64: int64(pr.Number), Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, pr.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("PR %s not found", pr.ID)
	}

	return nil
}

// Delete removes a pull request from persistence.
func (r *PRRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM prs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("PR %s not found", id)
	}

	return nil
}

// GetNextID returns the next available PR ID.
func (r *PRRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 4) AS INTEGER)), 0) FROM prs",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next PR ID: %w", err)
	}

	return fmt.Sprintf("PR-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a PR with optional timestamps.
func (r *PRRepository) UpdateStatus(ctx context.Context, id, status string, setMerged, setClosed bool) error {
	query := "UPDATE prs SET status = ?, updated_at = CURRENT_TIMESTAMP"
	args := []any{status}

	if setMerged {
		query += ", merged_at = CURRENT_TIMESTAMP"
	}

	if setClosed {
		query += ", closed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = ?"
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("PR %s not found", id)
	}

	return nil
}

// ShipmentExists checks if a shipment exists.
func (r *PRRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// RepoExists checks if a repository exists.
func (r *PRRepository) RepoExists(ctx context.Context, repoID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM repos WHERE id = ?", repoID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check repo existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentHasPR checks if a shipment already has a PR.
func (r *PRRepository) ShipmentHasPR(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM prs WHERE shipment_id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment PR: %w", err)
	}
	return count > 0, nil
}

// GetShipmentStatus retrieves the status of a shipment.
func (r *PRRepository) GetShipmentStatus(ctx context.Context, shipmentID string) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, "SELECT status FROM shipments WHERE id = ?", shipmentID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("shipment %s not found", shipmentID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get shipment status: %w", err)
	}
	return status, nil
}

// Ensure PRRepository implements the interface
var _ secondary.PRRepository = (*PRRepository)(nil)
