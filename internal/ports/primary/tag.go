package primary

import "context"

// TagService defines the primary port for tag operations.
type TagService interface {
	// CreateTag creates a new tag.
	CreateTag(ctx context.Context, req CreateTagRequest) (*CreateTagResponse, error)

	// GetTag retrieves a tag by ID.
	GetTag(ctx context.Context, tagID string) (*Tag, error)

	// GetTagByName retrieves a tag by name.
	GetTagByName(ctx context.Context, name string) (*Tag, error)

	// ListTags retrieves all tags.
	ListTags(ctx context.Context) ([]*Tag, error)

	// DeleteTag deletes a tag.
	DeleteTag(ctx context.Context, tagID string) error

	// GetEntityTag retrieves the tag for an entity.
	GetEntityTag(ctx context.Context, entityID, entityType string) (*Tag, error)
}

// CreateTagRequest contains parameters for creating a tag.
type CreateTagRequest struct {
	Name        string
	Description string
}

// CreateTagResponse contains the result of creating a tag.
type CreateTagResponse struct {
	TagID string
	Tag   *Tag
}

// Tag represents a tag entity at the port boundary.
type Tag struct {
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}
