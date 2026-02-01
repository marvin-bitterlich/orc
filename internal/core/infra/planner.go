// Package infra contains pure business logic for infrastructure planning.
package infra

// PlanInput contains pre-fetched data for infrastructure plan generation.
// All values must be gathered by the caller - no I/O in the planner.
type PlanInput struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	// Gatehouse state
	GatehouseID           string
	GatehousePath         string
	GatehousePathExists   bool
	GatehouseConfigExists bool

	// Workbench state
	Workbenches []WorkbenchPlanInput
}

// WorkbenchPlanInput contains pre-fetched data for a single workbench.
type WorkbenchPlanInput struct {
	ID             string
	Name           string
	WorktreePath   string
	RepoName       string
	HomeBranch     string
	WorktreeExists bool
	ConfigExists   bool
}

// Plan describes infrastructure state for a workshop.
type Plan struct {
	WorkshopID   string
	WorkshopName string
	FactoryID    string
	FactoryName  string

	Gatehouse   *GatehouseOp
	Workbenches []WorkbenchOp
}

// GatehouseOp describes gatehouse infrastructure state.
type GatehouseOp struct {
	ID           string
	Path         string
	Exists       bool
	ConfigExists bool
}

// WorkbenchOp describes workbench infrastructure state.
type WorkbenchOp struct {
	ID           string
	Name         string
	Path         string
	Exists       bool
	ConfigExists bool
	RepoName     string
	Branch       string
}

// GeneratePlan creates an infrastructure plan.
// This is a pure function - all input data must be pre-fetched.
func GeneratePlan(input PlanInput) Plan {
	plan := Plan{
		WorkshopID:   input.WorkshopID,
		WorkshopName: input.WorkshopName,
		FactoryID:    input.FactoryID,
		FactoryName:  input.FactoryName,
	}

	// Gatehouse
	plan.Gatehouse = &GatehouseOp{
		ID:           input.GatehouseID,
		Path:         input.GatehousePath,
		Exists:       input.GatehousePathExists,
		ConfigExists: input.GatehouseConfigExists,
	}

	// Workbenches
	for _, wb := range input.Workbenches {
		plan.Workbenches = append(plan.Workbenches, WorkbenchOp{
			ID:           wb.ID,
			Name:         wb.Name,
			Path:         wb.WorktreePath,
			Exists:       wb.WorktreeExists,
			ConfigExists: wb.ConfigExists,
			RepoName:     wb.RepoName,
			Branch:       wb.HomeBranch,
		})
	}

	return plan
}
