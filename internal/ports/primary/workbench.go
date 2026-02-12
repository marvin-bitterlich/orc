package primary

import "context"

// WorkbenchService defines the primary port for workbench operations.
// A Workbench is a git worktree.
type WorkbenchService interface {
	// CreateWorkbench creates a new workbench in a workshop.
	CreateWorkbench(ctx context.Context, req CreateWorkbenchRequest) (*CreateWorkbenchResponse, error)

	// GetWorkbench retrieves a workbench by ID.
	GetWorkbench(ctx context.Context, workbenchID string) (*Workbench, error)

	// GetWorkbenchByPath retrieves a workbench by its filesystem path.
	GetWorkbenchByPath(ctx context.Context, path string) (*Workbench, error)

	// ListWorkbenches lists workbenches with optional filters.
	ListWorkbenches(ctx context.Context, filters WorkbenchFilters) ([]*Workbench, error)

	// RenameWorkbench renames a workbench.
	RenameWorkbench(ctx context.Context, req RenameWorkbenchRequest) error

	// UpdateWorkbenchPath updates the filesystem path of a workbench.
	UpdateWorkbenchPath(ctx context.Context, workbenchID, newPath string) error

	// DeleteWorkbench deletes a workbench.
	DeleteWorkbench(ctx context.Context, req DeleteWorkbenchRequest) error

	// CheckoutBranch switches to a target branch using stash dance (stash, checkout, pop).
	CheckoutBranch(ctx context.Context, req CheckoutBranchRequest) (*CheckoutBranchResponse, error)

	// GetWorkbenchStatus returns the current git status of a workbench.
	GetWorkbenchStatus(ctx context.Context, workbenchID string) (*WorkbenchGitStatus, error)

	// UpdateFocusedID sets or clears the focused container ID for a workbench.
	// Pass empty string to clear focus.
	UpdateFocusedID(ctx context.Context, workbenchID, focusedID string) error

	// GetFocusedID returns the currently focused container ID for a workbench.
	GetFocusedID(ctx context.Context, workbenchID string) (string, error)

	// GetWorkbenchesByFocusedID returns all active workbenches focused on a given container.
	// Used for focus exclusivity checks (IMP cannot focus on container already focused by another IMP).
	GetWorkbenchesByFocusedID(ctx context.Context, focusedID string) ([]*Workbench, error)

	// ArchiveWorkbench soft-deletes a workbench by setting status to 'archived'.
	// The record remains in DB so infra plan can detect it as a DELETE target.
	ArchiveWorkbench(ctx context.Context, workbenchID string) error
}

// CreateWorkbenchRequest contains parameters for creating a workbench.
type CreateWorkbenchRequest struct {
	Name       string   // Optional - auto-generated as {repo}-{number} if empty and RepoID is set
	WorkshopID string   // Required
	RepoID     string   // Optional - link to repo (required for auto-generated name)
	Repos      []string // Optional repository names for worktree creation
}

// CreateWorkbenchResponse contains the result of workbench creation.
type CreateWorkbenchResponse struct {
	WorkbenchID string
	Workbench   *Workbench
	Path        string // Materialized workbench path
}

// RenameWorkbenchRequest contains parameters for renaming a workbench.
type RenameWorkbenchRequest struct {
	WorkbenchID  string
	NewName      string
	UpdateConfig bool // Also update .orc/config.json
}

// DeleteWorkbenchRequest contains parameters for deleting a workbench.
type DeleteWorkbenchRequest struct {
	WorkbenchID string
	Force       bool
}

// Workbench represents a workbench entity at the port boundary.
// A Workbench is a git worktree.
type Workbench struct {
	ID            string
	Name          string
	WorkshopID    string
	RepoID        string
	Path          string
	Status        string
	HomeBranch    string // Git home branch (e.g., ml/BENCH-name)
	CurrentBranch string // Currently checked out branch
	CreatedAt     string
	UpdatedAt     string
}

// WorkbenchFilters contains filter options for listing workbenches.
type WorkbenchFilters struct {
	WorkshopID string
	RepoID     string
	Status     string
	Limit      int
}

// CheckoutBranchRequest contains parameters for switching branches.
type CheckoutBranchRequest struct {
	WorkbenchID  string
	TargetBranch string
}

// CheckoutBranchResponse contains the result of a branch checkout.
type CheckoutBranchResponse struct {
	PreviousBranch string
	CurrentBranch  string
	StashApplied   bool // True if changes were stashed and reapplied
}

// WorkbenchGitStatus represents the git status of a workbench.
type WorkbenchGitStatus struct {
	WorkbenchID   string
	CurrentBranch string
	HomeBranch    string
	IsDirty       bool // Has uncommitted changes
	DirtyFiles    int  // Number of modified/untracked files
	AheadBy       int  // Commits ahead of remote
	BehindBy      int  // Commits behind remote
}
