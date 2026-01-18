package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

type Tag struct {
	ID          string
	Name        string
	Description sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateTag creates a new tag
func CreateTag(name, description string) (*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Generate tag ID by finding max existing ID
	var maxID int
	err = database.QueryRow("SELECT COALESCE(MAX(CAST(SUBSTR(id, 5) AS INTEGER)), 0) FROM tags").Scan(&maxID)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("TAG-%03d", maxID+1)

	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	_, err = database.Exec(
		"INSERT INTO tags (id, name, description) VALUES (?, ?, ?)",
		id, name, desc,
	)
	if err != nil {
		return nil, err
	}

	return GetTag(id)
}

// GetTag retrieves a tag by ID
func GetTag(id string) (*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	tag := &Tag{}
	err = database.QueryRow(
		"SELECT id, name, description, created_at, updated_at FROM tags WHERE id = ?",
		id,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.CreatedAt, &tag.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetTagByName retrieves a tag by name
func GetTagByName(name string) (*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	tag := &Tag{}
	err = database.QueryRow(
		"SELECT id, name, description, created_at, updated_at FROM tags WHERE name = ?",
		name,
	).Scan(&tag.ID, &tag.Name, &tag.Description, &tag.CreatedAt, &tag.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return tag, nil
}

// ListTags retrieves all tags
func ListTags() ([]*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, name, description, created_at, updated_at FROM tags ORDER BY name ASC"

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Description, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// DeleteTag deletes a tag (cascades to task_tags)
func DeleteTag(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	_, err = database.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}

// GetEntityTag retrieves the tag for any entity type via entity_tags table
func GetEntityTag(entityID, entityType string) (*Tag, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var tagID string
	err = database.QueryRow(
		"SELECT tag_id FROM entity_tags WHERE entity_id = ? AND entity_type = ?",
		entityID, entityType,
	).Scan(&tagID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return GetTag(tagID)
}
