package filesystem_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/orc/internal/adapters/filesystem"
)

func TestWorkspaceAdapter_DirectoryOperations(t *testing.T) {
	tmpDir := t.TempDir()
	adapter, err := filesystem.NewWorkspaceAdapter(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	testDir := filepath.Join(tmpDir, "test-dir")

	// Directory should not exist initially
	exists, err := adapter.DirectoryExists(ctx, testDir)
	if err != nil {
		t.Fatalf("DirectoryExists failed: %v", err)
	}
	if exists {
		t.Error("expected directory to not exist")
	}

	// Create directory
	err = adapter.CreateDirectory(ctx, testDir)
	if err != nil {
		t.Fatalf("CreateDirectory failed: %v", err)
	}

	// Directory should exist now
	exists, err = adapter.DirectoryExists(ctx, testDir)
	if err != nil {
		t.Fatalf("DirectoryExists failed: %v", err)
	}
	if !exists {
		t.Error("expected directory to exist")
	}

	// Remove directory
	err = adapter.RemoveDirectory(ctx, testDir)
	if err != nil {
		t.Fatalf("RemoveDirectory failed: %v", err)
	}

	// Directory should not exist anymore
	exists, err = adapter.DirectoryExists(ctx, testDir)
	if err != nil {
		t.Fatalf("DirectoryExists failed: %v", err)
	}
	if exists {
		t.Error("expected directory to not exist after removal")
	}
}

func TestWorkspaceAdapter_PathResolution(t *testing.T) {
	worktreesBase := "/custom/worktrees"
	reposBase := "/custom/repos"

	adapter, err := filesystem.NewWorkspaceAdapter(worktreesBase, reposBase)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	// Test GetWorktreesBasePath
	if adapter.GetWorktreesBasePath() != worktreesBase {
		t.Errorf("expected %s, got %s", worktreesBase, adapter.GetWorktreesBasePath())
	}

	// Test GetRepoPath
	repoPath := adapter.GetRepoPath("main-app")
	expected := filepath.Join(reposBase, "main-app")
	if repoPath != expected {
		t.Errorf("expected %s, got %s", expected, repoPath)
	}

	// Test ResolveGrovePath
	grovePath := adapter.ResolveGrovePath("auth-backend")
	expected = filepath.Join(worktreesBase, "auth-backend")
	if grovePath != expected {
		t.Errorf("expected %s, got %s", expected, grovePath)
	}
}

func TestWorkspaceAdapter_DefaultPaths(t *testing.T) {
	// Create with empty paths to use defaults
	adapter, err := filesystem.NewWorkspaceAdapter("", "")
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	home, _ := os.UserHomeDir()

	// Check default worktrees path
	expectedWorktrees := filepath.Join(home, "src", "worktrees")
	if adapter.GetWorktreesBasePath() != expectedWorktrees {
		t.Errorf("expected default worktrees path %s, got %s", expectedWorktrees, adapter.GetWorktreesBasePath())
	}

	// Check default repo path
	expectedRepo := filepath.Join(home, "src", "test-repo")
	if adapter.GetRepoPath("test-repo") != expectedRepo {
		t.Errorf("expected repo path %s, got %s", expectedRepo, adapter.GetRepoPath("test-repo"))
	}
}

func TestWorkspaceAdapter_WorktreeExists(t *testing.T) {
	tmpDir := t.TempDir()
	adapter, err := filesystem.NewWorkspaceAdapter(tmpDir, tmpDir)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()

	// Non-existent path
	exists, err := adapter.WorktreeExists(ctx, filepath.Join(tmpDir, "nonexistent"))
	if err != nil {
		t.Fatalf("WorktreeExists failed: %v", err)
	}
	if exists {
		t.Error("expected worktree to not exist")
	}

	// Create a directory
	worktreePath := filepath.Join(tmpDir, "worktree")
	_ = os.MkdirAll(worktreePath, 0755)

	// Directory should exist
	exists, err = adapter.WorktreeExists(ctx, worktreePath)
	if err != nil {
		t.Fatalf("WorktreeExists failed: %v", err)
	}
	if !exists {
		t.Error("expected worktree to exist")
	}
}
