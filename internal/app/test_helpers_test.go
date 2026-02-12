package app

import (
	"context"

	"github.com/example/orc/internal/ports/secondary"
)

// Ensure mockWorkspaceAdapter implements the interface
var _ secondary.WorkspaceAdapter = (*mockWorkspaceAdapter)(nil)

// mockWorkspaceAdapter implements secondary.WorkspaceAdapter for testing.
type mockWorkspaceAdapter struct {
	worktrees            map[string]bool
	worktreeExistsResult bool
	createWorktreeErr    error
	removeWorktreeErr    error
}

func newMockWorkspaceAdapter() *mockWorkspaceAdapter {
	return &mockWorkspaceAdapter{
		worktrees:            make(map[string]bool),
		worktreeExistsResult: false,
	}
}

func (m *mockWorkspaceAdapter) CreateWorktree(ctx context.Context, repoPath, branchName, targetPath string) error {
	if m.createWorktreeErr != nil {
		return m.createWorktreeErr
	}
	m.worktrees[targetPath] = true
	return nil
}

func (m *mockWorkspaceAdapter) RemoveWorktree(ctx context.Context, path string) error {
	if m.removeWorktreeErr != nil {
		return m.removeWorktreeErr
	}
	delete(m.worktrees, path)
	return nil
}

func (m *mockWorkspaceAdapter) WorktreeExists(ctx context.Context, path string) (bool, error) {
	if m.worktreeExistsResult {
		return m.worktreeExistsResult, nil
	}
	return m.worktrees[path], nil
}

func (m *mockWorkspaceAdapter) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (m *mockWorkspaceAdapter) RemoveDirectory(ctx context.Context, path string) error {
	return nil
}

func (m *mockWorkspaceAdapter) DirectoryExists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

func (m *mockWorkspaceAdapter) GetWorktreesBasePath() string {
	return "/tmp/worktrees"
}

func (m *mockWorkspaceAdapter) GetRepoPath(repoName string) string {
	return "/tmp/repos/" + repoName
}

func (m *mockWorkspaceAdapter) ResolveWorkbenchPath(workbenchName string) string {
	return "/tmp/worktrees/" + workbenchName
}
