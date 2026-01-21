package primary

import "context"

// RepoService defines the primary port for repository operations.
type RepoService interface {
	// CreateRepo creates a new repository configuration.
	CreateRepo(ctx context.Context, req CreateRepoRequest) (*CreateRepoResponse, error)

	// GetRepo retrieves a repository by ID.
	GetRepo(ctx context.Context, repoID string) (*Repo, error)

	// GetRepoByName retrieves a repository by its unique name.
	GetRepoByName(ctx context.Context, name string) (*Repo, error)

	// ListRepos lists repositories with optional filters.
	ListRepos(ctx context.Context, filters RepoFilters) ([]*Repo, error)

	// UpdateRepo updates a repository's configuration.
	UpdateRepo(ctx context.Context, req UpdateRepoRequest) error

	// ArchiveRepo archives a repository (soft delete).
	ArchiveRepo(ctx context.Context, repoID string) error

	// RestoreRepo restores an archived repository.
	RestoreRepo(ctx context.Context, repoID string) error

	// DeleteRepo hard-deletes a repository.
	DeleteRepo(ctx context.Context, repoID string) error
}

// CreateRepoRequest contains parameters for creating a repository.
type CreateRepoRequest struct {
	Name          string
	URL           string
	LocalPath     string
	DefaultBranch string
}

// CreateRepoResponse contains the result of creating a repository.
type CreateRepoResponse struct {
	RepoID string
	Repo   *Repo
}

// UpdateRepoRequest contains parameters for updating a repository.
type UpdateRepoRequest struct {
	RepoID        string
	URL           string
	LocalPath     string
	DefaultBranch string
}

// Repo represents a repository entity at the port boundary.
type Repo struct {
	ID            string
	Name          string
	URL           string
	LocalPath     string
	DefaultBranch string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}

// RepoFilters contains filter options for listing repositories.
type RepoFilters struct {
	Status string
}

// Repository status constants
const (
	RepoStatusActive   = "active"
	RepoStatusArchived = "archived"
)
