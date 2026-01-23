// Package secondary defines the secondary ports (driven adapters) for the application.
package secondary

import "context"

// WorkspaceAdapter defines the secondary port for filesystem and git worktree operations.
type WorkspaceAdapter interface {
	// Worktree operations
	CreateWorktree(ctx context.Context, repoName, branchName, targetPath string) error
	RemoveWorktree(ctx context.Context, path string) error
	WorktreeExists(ctx context.Context, path string) (bool, error)

	// Directory operations
	CreateDirectory(ctx context.Context, path string) error
	RemoveDirectory(ctx context.Context, path string) error
	DirectoryExists(ctx context.Context, path string) (bool, error)

	// Path resolution
	GetWorktreesBasePath() string
	GetRepoPath(repoName string) string
	ResolveWorkbenchPath(workbenchName string) string
}
