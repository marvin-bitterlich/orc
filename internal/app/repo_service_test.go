package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockRepoRepository implements secondary.RepoRepository for testing.
type mockRepoRepository struct {
	repos        map[string]*secondary.RepoRecord
	reposByName  map[string]*secondary.RepoRecord
	nextID       int
	hasActivePRs bool
}

func newMockRepoRepository() *mockRepoRepository {
	return &mockRepoRepository{
		repos:       make(map[string]*secondary.RepoRecord),
		reposByName: make(map[string]*secondary.RepoRecord),
		nextID:      1,
	}
}

func (m *mockRepoRepository) Create(ctx context.Context, repo *secondary.RepoRecord) error {
	repo.Status = "active"
	m.repos[repo.ID] = repo
	m.reposByName[repo.Name] = repo
	return nil
}

func (m *mockRepoRepository) GetByID(ctx context.Context, id string) (*secondary.RepoRecord, error) {
	if r, ok := m.repos[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("repository %s not found", id)
}

func (m *mockRepoRepository) GetByName(ctx context.Context, name string) (*secondary.RepoRecord, error) {
	if r, ok := m.reposByName[name]; ok {
		return r, nil
	}
	return nil, nil // nil, nil means not found (not an error)
}

func (m *mockRepoRepository) List(ctx context.Context, filters secondary.RepoFilters) ([]*secondary.RepoRecord, error) {
	var result []*secondary.RepoRecord
	for _, r := range m.repos {
		if filters.Status == "" || r.Status == filters.Status {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRepoRepository) Update(ctx context.Context, repo *secondary.RepoRecord) error {
	if _, ok := m.repos[repo.ID]; !ok {
		return fmt.Errorf("repository %s not found", repo.ID)
	}
	existing := m.repos[repo.ID]
	if repo.URL != "" {
		existing.URL = repo.URL
	}
	if repo.LocalPath != "" {
		existing.LocalPath = repo.LocalPath
	}
	if repo.DefaultBranch != "" {
		existing.DefaultBranch = repo.DefaultBranch
	}
	if repo.UpstreamURL != "" {
		existing.UpstreamURL = repo.UpstreamURL
	}
	if repo.UpstreamBranch != "" {
		existing.UpstreamBranch = repo.UpstreamBranch
	}
	return nil
}

func (m *mockRepoRepository) Delete(ctx context.Context, id string) error {
	if r, ok := m.repos[id]; ok {
		delete(m.reposByName, r.Name)
		delete(m.repos, id)
		return nil
	}
	return fmt.Errorf("repository %s not found", id)
}

func (m *mockRepoRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("REPO-%03d", id), nil
}

func (m *mockRepoRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if r, ok := m.repos[id]; ok {
		r.Status = status
		return nil
	}
	return fmt.Errorf("repository %s not found", id)
}

func (m *mockRepoRepository) HasActivePRs(ctx context.Context, repoID string) (bool, error) {
	return m.hasActivePRs, nil
}

func TestRepoService_CreateRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("creates repository with valid name", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		resp, err := svc.CreateRepo(ctx, primary.CreateRepoRequest{
			Name:          "my-repo",
			URL:           "git@github.com:org/my-repo.git",
			DefaultBranch: "main",
		})

		if err != nil {
			t.Fatalf("CreateRepo failed: %v", err)
		}
		if resp.Repo.Name != "my-repo" {
			t.Errorf("Name = %q, want %q", resp.Repo.Name, "my-repo")
		}
		if resp.RepoID != "REPO-001" {
			t.Errorf("RepoID = %q, want %q", resp.RepoID, "REPO-001")
		}
	})

	t.Run("fails with empty name", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		_, err := svc.CreateRepo(ctx, primary.CreateRepoRequest{
			Name: "",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("fails with duplicate name", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create first repo
		_, err := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "duplicate"})
		if err != nil {
			t.Fatalf("first CreateRepo failed: %v", err)
		}

		// Try to create duplicate
		_, err = svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "duplicate"})
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})

	t.Run("uses default branch when not specified", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		resp, err := svc.CreateRepo(ctx, primary.CreateRepoRequest{
			Name: "no-branch",
		})

		if err != nil {
			t.Fatalf("CreateRepo failed: %v", err)
		}
		if resp.Repo.DefaultBranch != "main" {
			t.Errorf("DefaultBranch = %q, want %q", resp.Repo.DefaultBranch, "main")
		}
	})
}

func TestRepoService_ArchiveRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("archives active repository", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create a repo
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "to-archive"})

		err := svc.ArchiveRepo(ctx, resp.RepoID)
		if err != nil {
			t.Fatalf("ArchiveRepo failed: %v", err)
		}

		// Verify archived
		got, _ := svc.GetRepo(ctx, resp.RepoID)
		if got.Status != "archived" {
			t.Errorf("Status = %q, want %q", got.Status, "archived")
		}
	})

	t.Run("fails to archive already archived repository", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create and archive a repo
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "already-archived"})
		_ = svc.ArchiveRepo(ctx, resp.RepoID)

		// Try to archive again
		err := svc.ArchiveRepo(ctx, resp.RepoID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRepoService_RestoreRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("restores archived repository", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create and archive a repo
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "to-restore"})
		_ = svc.ArchiveRepo(ctx, resp.RepoID)

		err := svc.RestoreRepo(ctx, resp.RepoID)
		if err != nil {
			t.Fatalf("RestoreRepo failed: %v", err)
		}

		// Verify restored
		got, _ := svc.GetRepo(ctx, resp.RepoID)
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("fails to restore active repository", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create a repo (starts as active)
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "already-active"})

		err := svc.RestoreRepo(ctx, resp.RepoID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRepoService_DeleteRepo(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes repository with no active PRs", func(t *testing.T) {
		repo := newMockRepoRepository()
		repo.hasActivePRs = false
		svc := NewRepoService(repo)

		// Create a repo
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "to-delete"})

		err := svc.DeleteRepo(ctx, resp.RepoID)
		if err != nil {
			t.Fatalf("DeleteRepo failed: %v", err)
		}
	})

	t.Run("fails to delete repository with active PRs", func(t *testing.T) {
		repo := newMockRepoRepository()
		repo.hasActivePRs = true
		svc := NewRepoService(repo)

		// Create a repo
		resp, _ := svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "has-prs"})

		err := svc.DeleteRepo(ctx, resp.RepoID)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRepoService_GetRepoByName(t *testing.T) {
	ctx := context.Background()

	t.Run("finds repository by name", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		// Create a repo
		_, _ = svc.CreateRepo(ctx, primary.CreateRepoRequest{Name: "find-me"})

		got, err := svc.GetRepoByName(ctx, "find-me")
		if err != nil {
			t.Fatalf("GetRepoByName failed: %v", err)
		}
		if got.Name != "find-me" {
			t.Errorf("Name = %q, want %q", got.Name, "find-me")
		}
	})

	t.Run("returns error for non-existent name", func(t *testing.T) {
		repo := newMockRepoRepository()
		svc := NewRepoService(repo)

		_, err := svc.GetRepoByName(ctx, "non-existent")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
