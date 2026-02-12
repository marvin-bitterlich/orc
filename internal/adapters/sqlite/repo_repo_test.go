package sqlite_test

import (
	"context"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

func TestRepoRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	t.Run("creates repository successfully", func(t *testing.T) {
		record := &secondary.RepoRecord{
			ID:            "REPO-001",
			Name:          "test-repo",
			URL:           "git@github.com:org/test-repo.git",
			LocalPath:     "/Users/test/src/test-repo",
			DefaultBranch: "main",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Verify it was created
		got, err := repo.GetByID(ctx, "REPO-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.Name != "test-repo" {
			t.Errorf("Name = %q, want %q", got.Name, "test-repo")
		}
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})

	t.Run("creates repository with default branch", func(t *testing.T) {
		record := &secondary.RepoRecord{
			ID:   "REPO-002",
			Name: "default-branch-repo",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REPO-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if got.DefaultBranch != "main" {
			t.Errorf("DefaultBranch = %q, want %q", got.DefaultBranch, "main")
		}
	})
}

func TestRepoRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create a test repo
	record := &secondary.RepoRecord{
		ID:   "REPO-001",
		Name: "named-repo",
	}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("finds repository by name", func(t *testing.T) {
		got, err := repo.GetByName(ctx, "named-repo")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if got == nil {
			t.Fatal("expected repository, got nil")
		}
		if got.ID != "REPO-001" {
			t.Errorf("ID = %q, want %q", got.ID, "REPO-001")
		}
	})

	t.Run("returns nil for non-existent name", func(t *testing.T) {
		got, err := repo.GetByName(ctx, "non-existent")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})
}

func TestRepoRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create test repos
	repos := []*secondary.RepoRecord{
		{ID: "REPO-001", Name: "alpha-repo", Status: "active"},
		{ID: "REPO-002", Name: "beta-repo", Status: "active"},
		{ID: "REPO-003", Name: "gamma-repo", Status: "archived"},
	}
	for _, r := range repos {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}
	// Archive REPO-003
	if err := repo.UpdateStatus(ctx, "REPO-003", "archived"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	t.Run("lists all repositories", func(t *testing.T) {
		got, err := repo.List(ctx, secondary.RepoFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(got) != 3 {
			t.Errorf("len = %d, want 3", len(got))
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		got, err := repo.List(ctx, secondary.RepoFilters{Status: "active"})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len = %d, want 2", len(got))
		}
	})
}

func TestRepoRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create a test repo
	record := &secondary.RepoRecord{
		ID:            "REPO-001",
		Name:          "update-test",
		DefaultBranch: "main",
	}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("updates repository URL", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.RepoRecord{
			ID:  "REPO-001",
			URL: "git@github.com:new/url.git",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REPO-001")
		if got.URL != "git@github.com:new/url.git" {
			t.Errorf("URL = %q, want %q", got.URL, "git@github.com:new/url.git")
		}
	})

	t.Run("fails for non-existent repository", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.RepoRecord{
			ID:  "REPO-999",
			URL: "test",
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRepoRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create a test repo
	record := &secondary.RepoRecord{
		ID:   "REPO-001",
		Name: "delete-test",
	}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("deletes repository", func(t *testing.T) {
		err := repo.Delete(ctx, "REPO-001")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = repo.GetByID(ctx, "REPO-001")
		if err == nil {
			t.Error("expected error after delete, got nil")
		}
	})

	t.Run("fails for non-existent repository", func(t *testing.T) {
		err := repo.Delete(ctx, "REPO-999")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRepoRepository_GetNextID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	t.Run("returns REPO-001 for empty table", func(t *testing.T) {
		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "REPO-001" {
			t.Errorf("ID = %q, want %q", id, "REPO-001")
		}
	})

	t.Run("increments after creation", func(t *testing.T) {
		record := &secondary.RepoRecord{ID: "REPO-001", Name: "first"}
		if err := repo.Create(ctx, record); err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		id, err := repo.GetNextID(ctx)
		if err != nil {
			t.Fatalf("GetNextID failed: %v", err)
		}
		if id != "REPO-002" {
			t.Errorf("ID = %q, want %q", id, "REPO-002")
		}
	})
}

func TestRepoRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	// Create a test repo
	record := &secondary.RepoRecord{
		ID:   "REPO-001",
		Name: "status-test",
	}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Run("updates status to archived", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "REPO-001", "archived")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REPO-001")
		if got.Status != "archived" {
			t.Errorf("Status = %q, want %q", got.Status, "archived")
		}
	})

	t.Run("updates status back to active", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "REPO-001", "active")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}

		got, _ := repo.GetByID(ctx, "REPO-001")
		if got.Status != "active" {
			t.Errorf("Status = %q, want %q", got.Status, "active")
		}
	})
}

func TestRepoRepository_UpstreamFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRepoRepository(db)
	ctx := context.Background()

	t.Run("creates repository with upstream fields", func(t *testing.T) {
		record := &secondary.RepoRecord{
			ID:             "REPO-001",
			Name:           "forked-repo",
			URL:            "git@github.com:me/forked.git",
			DefaultBranch:  "main",
			UpstreamURL:    "git@github.com:upstream/original.git",
			UpstreamBranch: "develop",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REPO-001")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.UpstreamURL != "git@github.com:upstream/original.git" {
			t.Errorf("UpstreamURL = %q, want %q", got.UpstreamURL, "git@github.com:upstream/original.git")
		}
		if got.UpstreamBranch != "develop" {
			t.Errorf("UpstreamBranch = %q, want %q", got.UpstreamBranch, "develop")
		}
	})

	t.Run("creates repository without upstream fields", func(t *testing.T) {
		record := &secondary.RepoRecord{
			ID:   "REPO-002",
			Name: "no-upstream",
		}

		err := repo.Create(ctx, record)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REPO-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.UpstreamURL != "" {
			t.Errorf("UpstreamURL = %q, want empty", got.UpstreamURL)
		}
		if got.UpstreamBranch != "" {
			t.Errorf("UpstreamBranch = %q, want empty", got.UpstreamBranch)
		}
	})

	t.Run("updates upstream fields on existing repo", func(t *testing.T) {
		err := repo.Update(ctx, &secondary.RepoRecord{
			ID:             "REPO-002",
			UpstreamURL:    "git@github.com:upstream/repo.git",
			UpstreamBranch: "main",
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		got, err := repo.GetByID(ctx, "REPO-002")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if got.UpstreamURL != "git@github.com:upstream/repo.git" {
			t.Errorf("UpstreamURL = %q, want %q", got.UpstreamURL, "git@github.com:upstream/repo.git")
		}
		if got.UpstreamBranch != "main" {
			t.Errorf("UpstreamBranch = %q, want %q", got.UpstreamBranch, "main")
		}
	})

	t.Run("upstream fields round-trip through GetByName", func(t *testing.T) {
		got, err := repo.GetByName(ctx, "forked-repo")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if got.UpstreamURL != "git@github.com:upstream/original.git" {
			t.Errorf("UpstreamURL = %q, want %q", got.UpstreamURL, "git@github.com:upstream/original.git")
		}
		if got.UpstreamBranch != "develop" {
			t.Errorf("UpstreamBranch = %q, want %q", got.UpstreamBranch, "develop")
		}
	})

	t.Run("upstream fields round-trip through List", func(t *testing.T) {
		repos, err := repo.List(ctx, secondary.RepoFilters{})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		// Find forked-repo in the list
		var found *secondary.RepoRecord
		for _, r := range repos {
			if r.Name == "forked-repo" {
				found = r
				break
			}
		}
		if found == nil {
			t.Fatal("forked-repo not found in list")
		}
		if found.UpstreamURL != "git@github.com:upstream/original.git" {
			t.Errorf("UpstreamURL = %q, want %q", found.UpstreamURL, "git@github.com:upstream/original.git")
		}
	})
}
