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
}

// InfraPlanRequest contains parameters for planning infrastructure.
type InfraPlanRequest struct {
	WorkshopID string
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

// InfraApplyResponse contains the result of applying infrastructure.
type InfraApplyResponse struct {
	WorkshopID         string
	WorkshopName       string
	GatehouseCreated   bool
	WorkbenchesCreated int
	ConfigsCreated     int
	NothingToDo        bool
}
