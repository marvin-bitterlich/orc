package primary

import "context"

// InfraService defines the primary port for infrastructure planning operations.
// It compares DB state to filesystem state and generates plans with OpStatus.
type InfraService interface {
	// PlanInfra generates a plan showing infrastructure state for a workshop.
	// Shows what would need to be created/already exists.
	PlanInfra(ctx context.Context, req InfraPlanRequest) (*InfraPlan, error)

	// ApplyInfra executes the infrastructure plan, creating directories,
	// worktrees, and configs as needed.
	ApplyInfra(ctx context.Context, plan *InfraPlan) (*InfraApplyResponse, error)

	// CleanupWorkbench removes the worktree directory for a workbench.
	// Called before DB deletion to ensure filesystem cleanup happens first.
	CleanupWorkbench(ctx context.Context, req CleanupWorkbenchRequest) error

	// CleanupWorkshop removes all infrastructure for a workshop (workbenches, gatehouse, tmux).
	// Called before DB deletion to ensure filesystem cleanup happens first.
	CleanupWorkshop(ctx context.Context, req CleanupWorkshopRequest) error

	// CleanupOrphans scans for and removes orphaned infrastructure.
	// Useful for manual recovery when system is in inconsistent state.
	CleanupOrphans(ctx context.Context, req CleanupOrphansRequest) (*CleanupOrphansResponse, error)
}

// InfraPlanRequest contains parameters for planning infrastructure.
type InfraPlanRequest struct {
	WorkshopID string
	Force      bool // Force deletion of dirty worktrees
	NoDelete   bool // Skip DELETE operations (only perform CREATE)
}

// InfraPlan describes the infrastructure state for a workshop.
type InfraPlan struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	// Gatehouse infrastructure
	Gatehouse *InfraGatehouseOp

	// Workbench infrastructure
	Workbenches []InfraWorkbenchOp

	// Orphan infrastructure (exists on disk but not in DB)
	OrphanWorkbenches []InfraWorkbenchOp
	OrphanGatehouses  []InfraGatehouseOp

	// TMux infrastructure
	TMuxSession *InfraTMuxSessionOp

	// Force deletion of dirty worktrees
	Force bool

	// Skip DELETE operations
	NoDelete bool
}

// InfraGatehouseOp describes gatehouse infrastructure state.
type InfraGatehouseOp struct {
	ID           string   // GATE-XXX
	Path         string   // ~/.orc/ws/WORK-xxx-slug
	Status       OpStatus // EXISTS or CREATE
	ConfigStatus OpStatus // EXISTS or CREATE
}

// InfraWorkbenchOp describes workbench infrastructure state.
type InfraWorkbenchOp struct {
	ID           string // BENCH-XXX
	Name         string
	Path         string   // Worktree path
	Status       OpStatus // EXISTS, CREATE, or MISSING
	ConfigStatus OpStatus // EXISTS or CREATE
	RepoName     string   // Source repo name (if linked)
	Branch       string   // Home branch
}

// InfraTMuxSessionOp describes tmux session infrastructure state.
type InfraTMuxSessionOp struct {
	SessionName   string
	Status        OpStatus // EXISTS or CREATE
	Windows       []InfraTMuxWindowOp
	OrphanWindows []InfraTMuxWindowOp // Windows that exist but shouldn't (DELETE)
}

// InfraTMuxWindowOp describes tmux window infrastructure state.
type InfraTMuxWindowOp struct {
	Name          string
	Path          string
	Status        OpStatus // EXISTS, CREATE, or DELETE
	Panes         []InfraTMuxPaneOp
	AgentOK       bool   // @orc_agent matches expected
	ActualAgent   string // Current @orc_agent value
	ExpectedAgent string // Expected agent (e.g., "IMP-name@BENCH-xxx")
}

// InfraTMuxPaneOp describes tmux pane verification state.
type InfraTMuxPaneOp struct {
	Index           int    // Pane index (1-based)
	PathOK          bool   // StartPath matches expected
	CommandOK       bool   // StartCommand matches expected (true if no expected command)
	ActualPath      string // Actual pane_start_path
	ActualCommand   string // Actual pane_start_command
	ExpectedPath    string // Expected path
	ExpectedCommand string // Expected command (empty if shell)
}

// InfraApplyResponse contains the result of applying infrastructure.
type InfraApplyResponse struct {
	WorkshopID         string
	WorkshopName       string
	GatehouseCreated   bool
	WorkbenchesCreated int
	ConfigsCreated     int
	OrphansDeleted     int
	NothingToDo        bool
}

// CleanupWorkbenchRequest contains parameters for cleaning up workbench infrastructure.
type CleanupWorkbenchRequest struct {
	WorkbenchID string
	Force       bool // Force deletion even if worktree has uncommitted changes
}

// CleanupWorkshopRequest contains parameters for cleaning up workshop infrastructure.
type CleanupWorkshopRequest struct {
	WorkshopID string
	Force      bool // Force deletion even if worktrees have uncommitted changes
}

// CleanupOrphansRequest contains parameters for cleaning up orphaned infrastructure.
type CleanupOrphansRequest struct {
	Force bool // Force deletion even if worktrees have uncommitted changes
}

// CleanupOrphansResponse contains the result of cleaning up orphans.
type CleanupOrphansResponse struct {
	WorkbenchesDeleted int
	GatehousesDeleted  int
}
