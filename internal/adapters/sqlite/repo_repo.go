// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// RepoRepository implements secondary.RepoRepository with SQLite.
type RepoRepository struct {
	db *sql.DB
}

// NewRepoRepository creates a new SQLite repository repository.
func NewRepoRepository(db *sql.DB) *RepoRepository {
	return &RepoRepository{db: db}
}

// Create persists a new repository.
func (r *RepoRepository) Create(ctx context.Context, repo *secondary.RepoRecord) error {
	var url, localPath, upstreamURL, upstreamBranch sql.NullString

	if repo.URL != "" {
		url = sql.NullString{String: repo.URL, Valid: true}
	}
	if repo.LocalPath != "" {
		localPath = sql.NullString{String: repo.LocalPath, Valid: true}
	}
	if repo.UpstreamURL != "" {
		upstreamURL = sql.NullString{String: repo.UpstreamURL, Valid: true}
	}
	if repo.UpstreamBranch != "" {
		upstreamBranch = sql.NullString{String: repo.UpstreamBranch, Valid: true}
	}

	defaultBranch := repo.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO repos (id, name, url, local_path, default_branch, upstream_url, upstream_branch, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		repo.ID, repo.Name, url, localPath, defaultBranch, upstreamURL, upstreamBranch, "active",
	)
	if err != nil {
		return fmt.Errorf("failed to create repo: %w", err)
	}

	return nil
}

// GetByID retrieves a repository by its ID.
func (r *RepoRepository) GetByID(ctx context.Context, id string) (*secondary.RepoRecord, error) {
	var (
		url            sql.NullString
		localPath      sql.NullString
		upstreamURL    sql.NullString
		upstreamBranch sql.NullString
		defaultBranch  string
		status         string
		createdAt      time.Time
		updatedAt      time.Time
	)

	record := &secondary.RepoRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, url, local_path, default_branch, upstream_url, upstream_branch, status, created_at, updated_at FROM repos WHERE id = ?",
		id,
	).Scan(&record.ID, &record.Name, &url, &localPath, &defaultBranch, &upstreamURL, &upstreamBranch, &status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	record.URL = url.String
	record.LocalPath = localPath.String
	record.DefaultBranch = defaultBranch
	record.UpstreamURL = upstreamURL.String
	record.UpstreamBranch = upstreamBranch.String
	record.Status = status
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByName retrieves a repository by its unique name.
func (r *RepoRepository) GetByName(ctx context.Context, name string) (*secondary.RepoRecord, error) {
	var (
		url            sql.NullString
		localPath      sql.NullString
		upstreamURL    sql.NullString
		upstreamBranch sql.NullString
		defaultBranch  string
		status         string
		createdAt      time.Time
		updatedAt      time.Time
	)

	record := &secondary.RepoRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, url, local_path, default_branch, upstream_url, upstream_branch, status, created_at, updated_at FROM repos WHERE name = ?",
		name,
	).Scan(&record.ID, &record.Name, &url, &localPath, &defaultBranch, &upstreamURL, &upstreamBranch, &status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil for "not found" to distinguish from errors
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository by name: %w", err)
	}

	record.URL = url.String
	record.LocalPath = localPath.String
	record.DefaultBranch = defaultBranch
	record.UpstreamURL = upstreamURL.String
	record.UpstreamBranch = upstreamBranch.String
	record.Status = status
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves repositories matching the given filters.
func (r *RepoRepository) List(ctx context.Context, filters secondary.RepoFilters) ([]*secondary.RepoRecord, error) {
	query := "SELECT id, name, url, local_path, default_branch, upstream_url, upstream_branch, status, created_at, updated_at FROM repos WHERE 1=1"
	args := []any{}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer rows.Close()

	var repos []*secondary.RepoRecord
	for rows.Next() {
		var (
			url            sql.NullString
			localPath      sql.NullString
			upstreamURL    sql.NullString
			upstreamBranch sql.NullString
			defaultBranch  string
			status         string
			createdAt      time.Time
			updatedAt      time.Time
		)

		record := &secondary.RepoRecord{}
		err := rows.Scan(&record.ID, &record.Name, &url, &localPath, &defaultBranch, &upstreamURL, &upstreamBranch, &status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}

		record.URL = url.String
		record.LocalPath = localPath.String
		record.DefaultBranch = defaultBranch
		record.UpstreamURL = upstreamURL.String
		record.UpstreamBranch = upstreamBranch.String
		record.Status = status
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		repos = append(repos, record)
	}

	return repos, nil
}

// Update updates an existing repository.
func (r *RepoRepository) Update(ctx context.Context, repo *secondary.RepoRecord) error {
	query := "UPDATE repos SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if repo.URL != "" {
		query += ", url = ?"
		args = append(args, sql.NullString{String: repo.URL, Valid: true})
	}

	if repo.LocalPath != "" {
		query += ", local_path = ?"
		args = append(args, sql.NullString{String: repo.LocalPath, Valid: true})
	}

	if repo.DefaultBranch != "" {
		query += ", default_branch = ?"
		args = append(args, repo.DefaultBranch)
	}

	if repo.UpstreamURL != "" {
		query += ", upstream_url = ?"
		args = append(args, sql.NullString{String: repo.UpstreamURL, Valid: true})
	}

	if repo.UpstreamBranch != "" {
		query += ", upstream_branch = ?"
		args = append(args, sql.NullString{String: repo.UpstreamBranch, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, repo.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("repository %s not found", repo.ID)
	}

	return nil
}

// Delete removes a repository from persistence.
func (r *RepoRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM repos WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("repository %s not found", id)
	}

	return nil
}

// GetNextID returns the next available repository ID.
func (r *RepoRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM repos",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next repository ID: %w", err)
	}

	return fmt.Sprintf("REPO-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a repository.
func (r *RepoRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE repos SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("repository %s not found", id)
	}

	return nil
}

// HasActivePRs checks if a repository has active (non-terminal) PRs.
func (r *RepoRepository) HasActivePRs(ctx context.Context, repoID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM prs WHERE repo_id = ? AND status NOT IN ('merged', 'closed')",
		repoID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check active PRs: %w", err)
	}
	return count > 0, nil
}

// Ensure RepoRepository implements the interface
var _ secondary.RepoRepository = (*RepoRepository)(nil)
