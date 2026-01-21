// Package filesystem contains filesystem-based adapter implementations.
package filesystem

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/orc/internal/ports/secondary"
)

// WorkspaceAdapter implements secondary.WorkspaceAdapter for filesystem operations.
type WorkspaceAdapter struct {
	worktreesBasePath string
	reposBasePath     string
}

// NewWorkspaceAdapter creates a new filesystem workspace adapter.
// If basePaths are empty, defaults to ~/src/worktrees and ~/src respectively.
func NewWorkspaceAdapter(worktreesBasePath, reposBasePath string) (*WorkspaceAdapter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	if worktreesBasePath == "" {
		worktreesBasePath = filepath.Join(home, "src", "worktrees")
	}
	if reposBasePath == "" {
		reposBasePath = filepath.Join(home, "src")
	}

	return &WorkspaceAdapter{
		worktreesBasePath: worktreesBasePath,
		reposBasePath:     reposBasePath,
	}, nil
}

// CreateWorktree creates a git worktree for a repository.
func (a *WorkspaceAdapter) CreateWorktree(ctx context.Context, repoName, branchName, targetPath string) error {
	repoPath := a.GetRepoPath(repoName)

	// Check if repo exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("repo not found at %s", repoPath)
	}

	// Create worktree with new branch
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", targetPath, "-b", branchName)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %w: %s", err, string(output))
	}

	return nil
}

// RemoveWorktree removes a git worktree.
func (a *WorkspaceAdapter) RemoveWorktree(ctx context.Context, path string) error {
	// Try git worktree remove first
	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", path, "--force")
	if err := cmd.Run(); err != nil {
		// Fall back to direct directory removal
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove worktree directory: %w", err)
		}
	}

	return nil
}

// WorktreeExists checks if a worktree exists at the given path.
func (a *WorkspaceAdapter) WorktreeExists(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check worktree: %w", err)
	}
	return info.IsDir(), nil
}

// CreateDirectory creates a directory with all parent directories.
func (a *WorkspaceAdapter) CreateDirectory(ctx context.Context, path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// RemoveDirectory removes a directory and all contents.
func (a *WorkspaceAdapter) RemoveDirectory(ctx context.Context, path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

// DirectoryExists checks if a directory exists.
func (a *WorkspaceAdapter) DirectoryExists(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check directory: %w", err)
	}
	return info.IsDir(), nil
}

// GetWorktreesBasePath returns the base path for worktrees (e.g., ~/src/worktrees).
func (a *WorkspaceAdapter) GetWorktreesBasePath() string {
	return a.worktreesBasePath
}

// GetRepoPath returns the path to a repository (e.g., ~/src/main-app).
func (a *WorkspaceAdapter) GetRepoPath(repoName string) string {
	return filepath.Join(a.reposBasePath, repoName)
}

// ResolveGrovePath returns the full path for a grove (e.g., ~/src/worktrees/auth-backend).
func (a *WorkspaceAdapter) ResolveGrovePath(groveName string) string {
	return filepath.Join(a.worktreesBasePath, groveName)
}

// Ensure WorkspaceAdapter implements the interface
var _ secondary.WorkspaceAdapter = (*WorkspaceAdapter)(nil)
