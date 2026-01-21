package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/core/repo"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// RepoServiceImpl implements the RepoService interface.
type RepoServiceImpl struct {
	repoRepo secondary.RepoRepository
}

// NewRepoService creates a new RepoService with injected dependencies.
func NewRepoService(repoRepo secondary.RepoRepository) *RepoServiceImpl {
	return &RepoServiceImpl{
		repoRepo: repoRepo,
	}
}

// CreateRepo creates a new repository.
func (s *RepoServiceImpl) CreateRepo(ctx context.Context, req primary.CreateRepoRequest) (*primary.CreateRepoResponse, error) {
	// Check if name already exists
	existing, err := s.repoRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}

	// Evaluate guard
	result := repo.CanCreateRepo(repo.CreateRepoContext{
		Name:       req.Name,
		NameExists: existing != nil,
	})
	if err := result.Error(); err != nil {
		return nil, err
	}

	// Get next ID
	nextID, err := s.repoRepo.GetNextID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate repository ID: %w", err)
	}

	// Set default branch
	defaultBranch := req.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Build record
	record := &secondary.RepoRecord{
		ID:            nextID,
		Name:          req.Name,
		URL:           req.URL,
		LocalPath:     req.LocalPath,
		DefaultBranch: defaultBranch,
	}

	if err := s.repoRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Fetch created repository
	created, err := s.repoRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created repository: %w", err)
	}

	return &primary.CreateRepoResponse{
		RepoID: created.ID,
		Repo:   s.recordToRepo(created),
	}, nil
}

// GetRepo retrieves a repository by ID.
func (s *RepoServiceImpl) GetRepo(ctx context.Context, repoID string) (*primary.Repo, error) {
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return s.recordToRepo(record), nil
}

// GetRepoByName retrieves a repository by its unique name.
func (s *RepoServiceImpl) GetRepoByName(ctx context.Context, name string) (*primary.Repo, error) {
	record, err := s.repoRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("repository with name %q not found", name)
	}
	return s.recordToRepo(record), nil
}

// ListRepos lists repositories with optional filters.
func (s *RepoServiceImpl) ListRepos(ctx context.Context, filters primary.RepoFilters) ([]*primary.Repo, error) {
	records, err := s.repoRepo.List(ctx, secondary.RepoFilters{
		Status: filters.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	repos := make([]*primary.Repo, len(records))
	for i, r := range records {
		repos[i] = s.recordToRepo(r)
	}
	return repos, nil
}

// UpdateRepo updates a repository's configuration.
func (s *RepoServiceImpl) UpdateRepo(ctx context.Context, req primary.UpdateRepoRequest) error {
	// Verify repository exists
	_, err := s.repoRepo.GetByID(ctx, req.RepoID)
	if err != nil {
		return err
	}

	record := &secondary.RepoRecord{
		ID:            req.RepoID,
		URL:           req.URL,
		LocalPath:     req.LocalPath,
		DefaultBranch: req.DefaultBranch,
	}
	return s.repoRepo.Update(ctx, record)
}

// ArchiveRepo archives a repository.
func (s *RepoServiceImpl) ArchiveRepo(ctx context.Context, repoID string) error {
	// Get current repository
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := repo.CanArchiveRepo(repo.ArchiveRepoContext{
		RepoID: repoID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.UpdateStatus(ctx, repoID, "archived")
}

// RestoreRepo restores an archived repository.
func (s *RepoServiceImpl) RestoreRepo(ctx context.Context, repoID string) error {
	// Get current repository
	record, err := s.repoRepo.GetByID(ctx, repoID)
	if err != nil {
		return err
	}

	// Evaluate guard
	result := repo.CanRestoreRepo(repo.RestoreRepoContext{
		RepoID: repoID,
		Status: record.Status,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.UpdateStatus(ctx, repoID, "active")
}

// DeleteRepo hard-deletes a repository.
func (s *RepoServiceImpl) DeleteRepo(ctx context.Context, repoID string) error {
	// Check for active PRs
	hasActivePRs, err := s.repoRepo.HasActivePRs(ctx, repoID)
	if err != nil {
		return fmt.Errorf("failed to check active PRs: %w", err)
	}

	// Evaluate guard
	result := repo.CanDeleteRepo(repo.DeleteRepoContext{
		RepoID:       repoID,
		HasActivePRs: hasActivePRs,
	})
	if err := result.Error(); err != nil {
		return err
	}

	return s.repoRepo.Delete(ctx, repoID)
}

// Helper methods

func (s *RepoServiceImpl) recordToRepo(r *secondary.RepoRecord) *primary.Repo {
	return &primary.Repo{
		ID:            r.ID,
		Name:          r.Name,
		URL:           r.URL,
		LocalPath:     r.LocalPath,
		DefaultBranch: r.DefaultBranch,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// Ensure RepoServiceImpl implements the interface
var _ primary.RepoService = (*RepoServiceImpl)(nil)
