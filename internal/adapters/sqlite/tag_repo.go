// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// TagRepository implements secondary.TagRepository with SQLite.
type TagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new SQLite tag repository.
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create persists a new tag.
func (r *TagRepository) Create(ctx context.Context, tag *secondary.TagRecord) error {
	var desc sql.NullString
	if tag.Description != "" {
		desc = sql.NullString{String: tag.Description, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO tags (id, name, description) VALUES (?, ?, ?)",
		tag.ID, tag.Name, desc,
	)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// GetByID retrieves a tag by its ID.
func (r *TagRepository) GetByID(ctx context.Context, id string) (*secondary.TagRecord, error) {
	var (
		desc      sql.NullString
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.TagRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM tags WHERE id = ?",
		id,
	).Scan(&record.ID, &record.Name, &desc, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	record.Description = desc.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByName retrieves a tag by its name.
func (r *TagRepository) GetByName(ctx context.Context, name string) (*secondary.TagRecord, error) {
	var (
		desc      sql.NullString
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.TagRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM tags WHERE name = ?",
		name,
	).Scan(&record.ID, &record.Name, &desc, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag '%s' not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	record.Description = desc.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves all tags ordered by name.
func (r *TagRepository) List(ctx context.Context) ([]*secondary.TagRecord, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM tags ORDER BY name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*secondary.TagRecord
	for rows.Next() {
		var (
			desc      sql.NullString
			createdAt time.Time
			updatedAt time.Time
		)

		record := &secondary.TagRecord{}
		err := rows.Scan(&record.ID, &record.Name, &desc, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		record.Description = desc.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		tags = append(tags, record)
	}

	return tags, nil
}

// Delete removes a tag from persistence.
func (r *TagRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("tag %s not found", id)
	}

	return nil
}

// GetNextID returns the next available tag ID.
func (r *TagRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM tags",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next tag ID: %w", err)
	}

	return fmt.Sprintf("TAG-%03d", maxID+1), nil
}

// GetEntityTag retrieves the tag for an entity (nil if none).
func (r *TagRepository) GetEntityTag(ctx context.Context, entityID, entityType string) (*secondary.TagRecord, error) {
	var tagID string
	err := r.db.QueryRowContext(ctx,
		"SELECT tag_id FROM entity_tags WHERE entity_id = ? AND entity_type = ?",
		entityID, entityType,
	).Scan(&tagID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entity tag: %w", err)
	}

	return r.GetByID(ctx, tagID)
}

// Ensure TagRepository implements the interface.
var _ secondary.TagRepository = (*TagRepository)(nil)
