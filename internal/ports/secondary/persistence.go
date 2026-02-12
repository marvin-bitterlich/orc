// Package secondary defines the secondary ports (driven adapters) for the application.
// These are the interfaces through which the application drives external systems.
package secondary

import "context"

// CommissionRepository defines the secondary port for commission persistence.
type CommissionRepository interface {
	// Create persists a new commission.
	Create(ctx context.Context, commission *CommissionRecord) error

	// GetByID retrieves a commission by its ID.
	GetByID(ctx context.Context, id string) (*CommissionRecord, error)

	// Update updates an existing commission.
	Update(ctx context.Context, commission *CommissionRecord) error

	// Delete removes a commission from persistence.
	Delete(ctx context.Context, id string) error

	// List retrieves commissions matching the given filters.
	List(ctx context.Context, filters CommissionFilters) ([]*CommissionRecord, error)

	// Pin pins a commission to keep it visible.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a commission.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available commission ID.
	GetNextID(ctx context.Context) (string, error)

	// CountShipments returns the number of shipments for a commission.
	CountShipments(ctx context.Context, commissionID string) (int, error)
}

// CommissionRecord represents a commission as stored in persistence.
type CommissionRecord struct {
	ID          string
	WorkshopID  string // FK to workshops - supports 1:many (workshop has many commissions)
	Title       string
	Description string
	Status      string
	Pinned      bool
	CreatedAt   string
	UpdatedAt   string
	StartedAt   string
	CompletedAt string
}

// CommissionFilters contains filter options for querying commissions.
type CommissionFilters struct {
	Status string
	Limit  int
}

// AgentIdentityProvider defines the secondary port for agent identity resolution.
// This abstracts the detection of current agent context (ORC vs IMP).
type AgentIdentityProvider interface {
	// GetCurrentIdentity returns the identity of the current agent.
	GetCurrentIdentity(ctx context.Context) (*AgentIdentity, error)
}

// AgentIdentity represents an agent's identity as provided by the secondary port.
type AgentIdentity struct {
	Type   AgentType
	ID     string // "ORC" for orchestrator, Workbench ID for IMP
	FullID string // Complete ID like "ORC" or "IMP-BENCH-001"
}

// AgentType represents the type of agent.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a workbench.
	AgentTypeIMP AgentType = "IMP"
)

// ShipmentRepository defines the secondary port for shipment persistence.
type ShipmentRepository interface {
	// Create persists a new shipment.
	Create(ctx context.Context, shipment *ShipmentRecord) error

	// GetByID retrieves a shipment by its ID.
	GetByID(ctx context.Context, id string) (*ShipmentRecord, error)

	// List retrieves shipments matching the given filters.
	List(ctx context.Context, filters ShipmentFilters) ([]*ShipmentRecord, error)

	// Update updates an existing shipment.
	Update(ctx context.Context, shipment *ShipmentRecord) error

	// Delete removes a shipment from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a shipment.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a shipment.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available shipment ID.
	GetNextID(ctx context.Context) (string, error)

	// GetByWorkbench retrieves shipments assigned to a workbench.
	GetByWorkbench(ctx context.Context, workbenchID string) ([]*ShipmentRecord, error)

	// AssignWorkbench assigns a shipment to a workbench.
	AssignWorkbench(ctx context.Context, shipmentID, workbenchID string) error

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// WorkbenchAssignedToOther checks if workbench is assigned to another shipment.
	WorkbenchAssignedToOther(ctx context.Context, workbenchID, excludeShipmentID string) (string, error)
}

// ShipmentRecord represents a shipment as stored in persistence.
// Shipments go directly under commissions (conclaves and shipyards are removed).
type ShipmentRecord struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         string // Empty string means null
	Status              string // draft, exploring, specced, tasked, ready_for_imp, implementing, auto_implementing, complete
	AssignedWorkbenchID string // Empty string means null
	RepoID              string // Empty string means null - FK to repos table
	Branch              string // Empty string means null - owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
	SpecNoteID          string // Empty string means null - spec note that generated this shipment (NOTE-xxx)
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string // Empty string means null
}

// ShipmentFilters contains filter options for querying shipments.
type ShipmentFilters struct {
	CommissionID string
	Status       string
}

// TaskRepository defines the secondary port for task persistence.
type TaskRepository interface {
	// Create persists a new task.
	Create(ctx context.Context, task *TaskRecord) error

	// GetByID retrieves a task by its ID.
	GetByID(ctx context.Context, id string) (*TaskRecord, error)

	// List retrieves tasks matching the given filters.
	List(ctx context.Context, filters TaskFilters) ([]*TaskRecord, error)

	// Update updates an existing task.
	Update(ctx context.Context, task *TaskRecord) error

	// Delete removes a task from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a task.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a task.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available task ID.
	GetNextID(ctx context.Context) (string, error)

	// GetByWorkbench retrieves tasks assigned to a workbench.
	GetByWorkbench(ctx context.Context, workbenchID string) ([]*TaskRecord, error)

	// GetByShipment retrieves tasks for a shipment.
	GetByShipment(ctx context.Context, shipmentID string) ([]*TaskRecord, error)

	// UpdateStatus updates the status with optional timestamps.
	UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error

	// Claim claims a task for a workbench.
	Claim(ctx context.Context, id, workbenchID string) error

	// AssignWorkbenchByShipment assigns all tasks of a shipment to a workbench.
	AssignWorkbenchByShipment(ctx context.Context, shipmentID, workbenchID string) error

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// TomeExists checks if a tome exists (for validation).
	TomeExists(ctx context.Context, tomeID string) (bool, error)

	// GetTag retrieves the tag for a task (nil if none).
	GetTag(ctx context.Context, taskID string) (*TagRecord, error)

	// AddTag adds a tag to a task.
	AddTag(ctx context.Context, taskID, tagID string) error

	// RemoveTag removes the tag from a task.
	RemoveTag(ctx context.Context, taskID string) error

	// ListByTag retrieves tasks with a specific tag.
	ListByTag(ctx context.Context, tagID string) ([]*TaskRecord, error)

	// GetNextEntityTagID returns the next available entity tag ID.
	GetNextEntityTagID(ctx context.Context) (string, error)
}

// TaskRecord represents a task as stored in persistence.
type TaskRecord struct {
	ID                  string
	ShipmentID          string // Empty string means null
	CommissionID        string
	TomeID              string // Empty string means null
	ConclaveID          string // Empty string means null
	Title               string
	Description         string // Empty string means null
	Type                string // Empty string means null
	Status              string
	Priority            string // Empty string means null
	AssignedWorkbenchID string // Empty string means null
	Pinned              bool
	DependsOn           string // JSON array of task IDs, empty string means null
	CreatedAt           string
	UpdatedAt           string
	ClaimedAt           string // Empty string means null
	CompletedAt         string // Empty string means null
}

// TaskFilters contains filter options for querying tasks.
type TaskFilters struct {
	ShipmentID   string
	Status       string
	CommissionID string
}

// TagRecord represents a tag as stored in persistence.
type TagRecord struct {
	ID          string
	Name        string
	Description string // Empty string means null
	CreatedAt   string
	UpdatedAt   string
}

// TagRepository defines the secondary port for tag persistence.
type TagRepository interface {
	// Create persists a new tag.
	Create(ctx context.Context, tag *TagRecord) error

	// GetByID retrieves a tag by its ID.
	GetByID(ctx context.Context, id string) (*TagRecord, error)

	// GetByName retrieves a tag by its name.
	GetByName(ctx context.Context, name string) (*TagRecord, error)

	// List retrieves all tags ordered by name.
	List(ctx context.Context) ([]*TagRecord, error)

	// Delete removes a tag from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available tag ID.
	GetNextID(ctx context.Context) (string, error)

	// GetEntityTag retrieves the tag for an entity (nil if none).
	GetEntityTag(ctx context.Context, entityID, entityType string) (*TagRecord, error)
}

// NoteRepository defines the secondary port for note persistence.
type NoteRepository interface {
	// Create persists a new note.
	Create(ctx context.Context, note *NoteRecord) error

	// GetByID retrieves a note by its ID.
	GetByID(ctx context.Context, id string) (*NoteRecord, error)

	// List retrieves notes matching the given filters.
	List(ctx context.Context, filters NoteFilters) ([]*NoteRecord, error)

	// Update updates an existing note.
	Update(ctx context.Context, note *NoteRecord) error

	// Delete removes a note from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a note.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a note.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available note ID.
	GetNextID(ctx context.Context) (string, error)

	// GetByContainer retrieves notes for a specific container.
	GetByContainer(ctx context.Context, containerType, containerID string) ([]*NoteRecord, error)

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// TomeExists checks if a tome exists (for validation).
	TomeExists(ctx context.Context, tomeID string) (bool, error)

	// ConclaveExists checks if a conclave exists (for validation).
	ConclaveExists(ctx context.Context, conclaveID string) (bool, error)

	// UpdateStatus updates the status of a note (open/closed).
	UpdateStatus(ctx context.Context, id string, status string) error

	// CloseWithMerge closes a note and records it was merged into another note.
	CloseWithMerge(ctx context.Context, sourceID, targetID string) error

	// CloseWithReason closes a note with a reason and optional reference to another note.
	CloseWithReason(ctx context.Context, id, reason, byNoteID string) error
}

// NoteRecord represents a note as stored in persistence.
type NoteRecord struct {
	ID                  string
	CommissionID        string
	Title               string
	Content             string // Empty string means null
	Type                string // Empty string means null
	Status              string // "open" or "closed"
	ShipmentID          string // Empty string means null
	ConclaveID          string // Empty string means null
	TomeID              string // Empty string means null
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	ClosedAt            string // Empty string means null
	PromotedFromID      string // Empty string means null
	PromotedFromType    string // Empty string means null
	CloseReason         string // Empty string means null
	ClosedByNoteID      string // Empty string means null
	PromoteToCommission bool   // When true, clear all container associations to make commission-level
}

// NoteFilters contains filter options for querying notes.
type NoteFilters struct {
	Type         string
	CommissionID string
}

// HandoffRepository defines the secondary port for handoff persistence.
// Handoffs are immutable - no Update or Delete operations.
type HandoffRepository interface {
	// Create persists a new handoff.
	Create(ctx context.Context, handoff *HandoffRecord) error

	// GetByID retrieves a handoff by its ID.
	GetByID(ctx context.Context, id string) (*HandoffRecord, error)

	// GetLatest retrieves the most recent handoff.
	GetLatest(ctx context.Context) (*HandoffRecord, error)

	// GetLatestForWorkbench retrieves the most recent handoff for a workbench.
	GetLatestForWorkbench(ctx context.Context, workbenchID string) (*HandoffRecord, error)

	// List retrieves handoffs with optional limit.
	List(ctx context.Context, limit int) ([]*HandoffRecord, error)

	// GetNextID returns the next available handoff ID.
	GetNextID(ctx context.Context) (string, error)
}

// HandoffRecord represents a handoff as stored in persistence.
type HandoffRecord struct {
	ID                 string
	CreatedAt          string
	HandoffNote        string
	ActiveCommissionID string // Empty string means null
	ActiveWorkbenchID  string // Empty string means null
	TodosSnapshot      string // Empty string means null
}

// TomeRepository defines the secondary port for tome persistence.
type TomeRepository interface {
	// Create persists a new tome.
	Create(ctx context.Context, tome *TomeRecord) error

	// GetByID retrieves a tome by its ID.
	GetByID(ctx context.Context, id string) (*TomeRecord, error)

	// List retrieves tomes matching the given filters.
	List(ctx context.Context, filters TomeFilters) ([]*TomeRecord, error)

	// Update updates an existing tome.
	Update(ctx context.Context, tome *TomeRecord) error

	// Delete removes a tome from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a tome.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a tome.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available tome ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// GetByWorkbench retrieves tomes assigned to a workbench.
	GetByWorkbench(ctx context.Context, workbenchID string) ([]*TomeRecord, error)

	// GetByConclave retrieves tomes belonging to a conclave.
	GetByConclave(ctx context.Context, conclaveID string) ([]*TomeRecord, error)

	// AssignWorkbench assigns a tome to a workbench.
	AssignWorkbench(ctx context.Context, tomeID, workbenchID string) error

	// UpdateContainer updates the container assignment for a tome.
	// Used for unpark (→ conclave) operations.
	UpdateContainer(ctx context.Context, id, containerID, containerType string) error

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)
}

// TomeRecord represents a tome as stored in persistence.
type TomeRecord struct {
	ID                  string
	CommissionID        string
	ConclaveID          string // Empty string means null - optional parent conclave (legacy, use ContainerID)
	Title               string
	Description         string // Empty string means null
	Status              string
	AssignedWorkbenchID string // Empty string means null
	Pinned              bool
	ContainerID         string // Empty string means null - CON-xxx
	ContainerType       string // Empty string means null - "conclave"
	CreatedAt           string
	UpdatedAt           string
	ClosedAt            string // Empty string means null
}

// TomeFilters contains filter options for querying tomes.
type TomeFilters struct {
	CommissionID string
	ConclaveID   string
	Status       string
}

// ShipyardRecord represents a shipyard as stored in persistence.
// One shipyard per factory, auto-created.
type ShipyardRecord struct {
	ID        string
	FactoryID string
	CreatedAt string
	UpdatedAt string
}

// ShipyardRepository defines the secondary port for shipyard persistence.
type ShipyardRepository interface {
	// Create persists a new shipyard.
	Create(ctx context.Context, shipyard *ShipyardRecord) error

	// GetByID retrieves a shipyard by its ID.
	GetByID(ctx context.Context, id string) (*ShipyardRecord, error)

	// GetByFactoryID retrieves the shipyard for a factory.
	GetByFactoryID(ctx context.Context, factoryID string) (*ShipyardRecord, error)

	// GetNextID returns the next available shipyard ID.
	GetNextID(ctx context.Context) (string, error)

	// FactoryExists checks if a factory exists (for validation).
	FactoryExists(ctx context.Context, factoryID string) (bool, error)
}

// ConclaveRepository defines the secondary port for conclave persistence.
type ConclaveRepository interface {
	// Create persists a new conclave.
	Create(ctx context.Context, conclave *ConclaveRecord) error

	// GetByID retrieves a conclave by its ID.
	GetByID(ctx context.Context, id string) (*ConclaveRecord, error)

	// List retrieves conclaves matching the given filters.
	List(ctx context.Context, filters ConclaveFilters) ([]*ConclaveRecord, error)

	// Update updates an existing conclave.
	Update(ctx context.Context, conclave *ConclaveRecord) error

	// Delete removes a conclave from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a conclave.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a conclave.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available conclave ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status and optionally decided_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setDecided bool) error

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// GetTasksByConclave retrieves tasks belonging to a conclave.
	GetTasksByConclave(ctx context.Context, conclaveID string) ([]*ConclaveTaskRecord, error)

	// GetPlansByConclave retrieves plans belonging to a conclave.
	GetPlansByConclave(ctx context.Context, conclaveID string) ([]*ConclavePlanRecord, error)
}

// ConclaveRecord represents a conclave as stored in persistence.
type ConclaveRecord struct {
	ID           string
	CommissionID string
	ShipmentID   string // Empty string means null
	Title        string
	Description  string // Empty string means null
	Status       string
	Decision     string // Empty string means null
	Pinned       bool
	CreatedAt    string
	UpdatedAt    string
	DecidedAt    string // Empty string means null
}

// ConclaveFilters contains filter options for querying conclaves.
type ConclaveFilters struct {
	CommissionID string
	Status       string
}

// ConclaveTaskRecord represents a task as returned from conclave cross-entity query.
type ConclaveTaskRecord struct {
	ID                  string
	ShipmentID          string
	CommissionID        string
	Title               string
	Description         string
	Type                string
	Status              string
	Priority            string
	AssignedWorkbenchID string
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	ClaimedAt           string
	CompletedAt         string
	ConclaveID          string
	PromotedFromID      string
	PromotedFromType    string
}

// ConclavePlanRecord represents a plan as returned from conclave cross-entity query.
type ConclavePlanRecord struct {
	ID               string
	TaskID           string
	CommissionID     string
	Title            string
	Description      string
	Status           string
	Content          string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ApprovedAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// OperationRepository defines the secondary port for operation persistence.
// Operations are minimal entities with no Delete operation.
type OperationRepository interface {
	// Create persists a new operation.
	Create(ctx context.Context, operation *OperationRecord) error

	// GetByID retrieves an operation by its ID.
	GetByID(ctx context.Context, id string) (*OperationRecord, error)

	// List retrieves operations matching the given filters.
	List(ctx context.Context, filters OperationFilters) ([]*OperationRecord, error)

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// GetNextID returns the next available operation ID.
	GetNextID(ctx context.Context) (string, error)

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)
}

// OperationRecord represents an operation as stored in persistence.
type OperationRecord struct {
	ID           string
	CommissionID string
	Title        string
	Description  string // Empty string means null
	Status       string
	CreatedAt    string
	UpdatedAt    string
	CompletedAt  string // Empty string means null
}

// OperationFilters contains filter options for querying operations.
type OperationFilters struct {
	CommissionID string
	Status       string
}

// PlanRepository defines the secondary port for plan persistence.
type PlanRepository interface {
	// Create persists a new plan.
	Create(ctx context.Context, plan *PlanRecord) error

	// GetByID retrieves a plan by its ID.
	GetByID(ctx context.Context, id string) (*PlanRecord, error)

	// List retrieves plans matching the given filters.
	List(ctx context.Context, filters PlanFilters) ([]*PlanRecord, error)

	// Update updates an existing plan.
	Update(ctx context.Context, plan *PlanRecord) error

	// Delete removes a plan from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a plan.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a plan.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available plan ID.
	GetNextID(ctx context.Context) (string, error)

	// Approve approves a plan and sets the approved_at timestamp.
	Approve(ctx context.Context, id string) error

	// GetActivePlanForTask retrieves the active (draft) plan for a task.
	GetActivePlanForTask(ctx context.Context, taskID string) (*PlanRecord, error)

	// HasActivePlanForTask checks if a task has an active (draft) plan.
	HasActivePlanForTask(ctx context.Context, taskID string) (bool, error)

	// UpdateStatus updates the plan status.
	UpdateStatus(ctx context.Context, id, status string) error

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// TaskExists checks if a task exists (for validation).
	TaskExists(ctx context.Context, taskID string) (bool, error)
}

// PlanRecord represents a plan as stored in persistence.
type PlanRecord struct {
	ID               string
	TaskID           string // FK to tasks
	CommissionID     string
	Title            string
	Description      string // Empty string means null
	Status           string
	Content          string // Empty string means null
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ApprovedAt       string // Empty string means null
	ConclaveID       string // Empty string means null
	PromotedFromID   string // Empty string means null
	PromotedFromType string // Empty string means null
	SupersedesPlanID string // Empty string means null - FK to plans
}

// PlanFilters contains filter options for querying plans.
type PlanFilters struct {
	TaskID       string
	CommissionID string
	Status       string
}

// RepoRepository defines the secondary port for repository persistence.
type RepoRepository interface {
	// Create persists a new repository.
	Create(ctx context.Context, repo *RepoRecord) error

	// GetByID retrieves a repository by its ID.
	GetByID(ctx context.Context, id string) (*RepoRecord, error)

	// GetByName retrieves a repository by its unique name.
	GetByName(ctx context.Context, name string) (*RepoRecord, error)

	// List retrieves repositories matching the given filters.
	List(ctx context.Context, filters RepoFilters) ([]*RepoRecord, error)

	// Update updates an existing repository.
	Update(ctx context.Context, repo *RepoRecord) error

	// Delete removes a repository from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available repository ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of a repository.
	UpdateStatus(ctx context.Context, id, status string) error

	// HasActivePRs checks if a repository has active (non-terminal) PRs.
	HasActivePRs(ctx context.Context, repoID string) (bool, error)
}

// RepoRecord represents a repository as stored in persistence.
type RepoRecord struct {
	ID            string
	Name          string
	URL           string // Empty string means null
	LocalPath     string // Empty string means null
	DefaultBranch string
	Status        string
	CreatedAt     string
	UpdatedAt     string
}

// RepoFilters contains filter options for querying repositories.
type RepoFilters struct {
	Status string
}

// PRRepository defines the secondary port for pull request persistence.
type PRRepository interface {
	// Create persists a new pull request.
	Create(ctx context.Context, pr *PRRecord) error

	// GetByID retrieves a pull request by its ID.
	GetByID(ctx context.Context, id string) (*PRRecord, error)

	// GetByShipment retrieves a pull request by shipment ID.
	GetByShipment(ctx context.Context, shipmentID string) (*PRRecord, error)

	// List retrieves pull requests matching the given filters.
	List(ctx context.Context, filters PRFilters) ([]*PRRecord, error)

	// Update updates an existing pull request.
	Update(ctx context.Context, pr *PRRecord) error

	// Delete removes a pull request from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available pull request ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of a PR with optional timestamps.
	UpdateStatus(ctx context.Context, id, status string, setMerged, setClosed bool) error

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// RepoExists checks if a repository exists (for validation).
	RepoExists(ctx context.Context, repoID string) (bool, error)

	// ShipmentHasPR checks if a shipment already has a PR.
	ShipmentHasPR(ctx context.Context, shipmentID string) (bool, error)

	// GetShipmentStatus retrieves the status of a shipment.
	GetShipmentStatus(ctx context.Context, shipmentID string) (string, error)
}

// PRRecord represents a pull request as stored in persistence.
type PRRecord struct {
	ID           string
	ShipmentID   string
	RepoID       string
	CommissionID string
	Number       int // 0 means null (for draft PRs without GitHub PR number)
	Title        string
	Description  string // Empty string means null
	Branch       string
	TargetBranch string // Empty string means null (defaults to repo default)
	URL          string // Empty string means null
	Status       string
	CreatedAt    string
	UpdatedAt    string
	MergedAt     string // Empty string means null
	ClosedAt     string // Empty string means null
}

// PRFilters contains filter options for querying pull requests.
type PRFilters struct {
	ShipmentID   string
	RepoID       string
	CommissionID string
	Status       string
}

// FactoryRepository defines the secondary port for factory persistence.
// A Factory is a TMux session - the persistent runtime environment.
type FactoryRepository interface {
	// Create persists a new factory.
	Create(ctx context.Context, factory *FactoryRecord) error

	// GetByID retrieves a factory by its ID.
	GetByID(ctx context.Context, id string) (*FactoryRecord, error)

	// GetByName retrieves a factory by its unique name.
	GetByName(ctx context.Context, name string) (*FactoryRecord, error)

	// List retrieves factories matching the given filters.
	List(ctx context.Context, filters FactoryFilters) ([]*FactoryRecord, error)

	// Update updates an existing factory.
	Update(ctx context.Context, factory *FactoryRecord) error

	// Delete removes a factory from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available factory ID.
	GetNextID(ctx context.Context) (string, error)

	// CountWorkshops returns the number of workshops for a factory.
	CountWorkshops(ctx context.Context, factoryID string) (int, error)

	// CountCommissions returns the number of commissions for a factory.
	CountCommissions(ctx context.Context, factoryID string) (int, error)
}

// FactoryRecord represents a factory as stored in persistence.
type FactoryRecord struct {
	ID        string
	Name      string
	Status    string
	CreatedAt string
	UpdatedAt string
}

// FactoryFilters contains filter options for querying factories.
type FactoryFilters struct {
	Status string
	Limit  int
}

// WorkshopRepository defines the secondary port for workshop persistence.
// A Workshop is a persistent place within a Factory, hosting Workbenches.
type WorkshopRepository interface {
	// Create persists a new workshop.
	Create(ctx context.Context, workshop *WorkshopRecord) error

	// GetByID retrieves a workshop by its ID.
	GetByID(ctx context.Context, id string) (*WorkshopRecord, error)

	// List retrieves workshops matching the given filters.
	List(ctx context.Context, filters WorkshopFilters) ([]*WorkshopRecord, error)

	// Update updates an existing workshop.
	Update(ctx context.Context, workshop *WorkshopRecord) error

	// Delete removes a workshop from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available workshop ID.
	GetNextID(ctx context.Context) (string, error)

	// CountWorkbenches returns the number of workbenches for a workshop.
	CountWorkbenches(ctx context.Context, workshopID string) (int, error)

	// CountByFactory returns the number of workshops for a factory.
	CountByFactory(ctx context.Context, factoryID string) (int, error)

	// FactoryExists checks if a factory exists (for validation).
	FactoryExists(ctx context.Context, factoryID string) (bool, error)

	// SetActiveCommissionID updates the active commission for a workshop (Goblin context).
	// Pass empty string to clear.
	SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error

	// GetActiveCommissions returns commission IDs derived from focus:
	// - Gatehouse focused_id (resolved to commission)
	// - All workbench focused_ids in workshop (resolved to commission)
	// Returns deduplicated commission IDs.
	GetActiveCommissions(ctx context.Context, workshopID string) ([]string, error)
}

// WorkshopRecord represents a workshop as stored in persistence.
type WorkshopRecord struct {
	ID                 string
	FactoryID          string
	Name               string
	Status             string
	ActiveCommissionID string // Empty string means null - Goblin commission context
	CreatedAt          string
	UpdatedAt          string
}

// WorkshopFilters contains filter options for querying workshops.
type WorkshopFilters struct {
	FactoryID string
	Status    string
	Limit     int
}

// WorkbenchRepository defines the secondary port for workbench persistence.
// A Workbench is a git worktree - replaces the Grove concept.
type WorkbenchRepository interface {
	// Create persists a new workbench.
	Create(ctx context.Context, workbench *WorkbenchRecord) error

	// GetByID retrieves a workbench by its ID.
	GetByID(ctx context.Context, id string) (*WorkbenchRecord, error)

	// GetByPath retrieves a workbench by its file path.
	GetByPath(ctx context.Context, path string) (*WorkbenchRecord, error)

	// GetByWorkshop retrieves all workbenches for a workshop.
	GetByWorkshop(ctx context.Context, workshopID string) ([]*WorkbenchRecord, error)

	// List retrieves all workbenches, optionally filtered by workshop.
	List(ctx context.Context, workshopID string) ([]*WorkbenchRecord, error)

	// Update updates an existing workbench.
	Update(ctx context.Context, workbench *WorkbenchRecord) error

	// Delete removes a workbench from persistence.
	Delete(ctx context.Context, id string) error

	// Rename updates the name of a workbench.
	Rename(ctx context.Context, id, newName string) error

	// UpdatePath updates the path of a workbench.
	UpdatePath(ctx context.Context, id, newPath string) error

	// UpdateFocusedID updates the focused container ID for a workbench.
	// Pass empty string to clear focus.
	UpdateFocusedID(ctx context.Context, id, focusedID string) error

	// GetByFocusedID retrieves all active workbenches focusing a specific container.
	// Used to check for focus exclusivity conflicts.
	GetByFocusedID(ctx context.Context, focusedID string) ([]*WorkbenchRecord, error)

	// GetNextID returns the next available workbench ID.
	GetNextID(ctx context.Context) (string, error)

	// WorkshopExists checks if a workshop exists (for validation).
	WorkshopExists(ctx context.Context, workshopID string) (bool, error)
}

// WorkbenchRecord represents a workbench as stored in persistence.
type WorkbenchRecord struct {
	ID            string
	Name          string
	WorkshopID    string
	RepoID        string // Optional - linked repo
	WorktreePath  string
	Status        string
	HomeBranch    string // Git home branch for this workbench (e.g., ml/BENCH-name)
	CurrentBranch string // Currently checked out branch
	FocusedID     string // Empty string means null - IMP focus (CON-xxx or SHIP-xxx)
	CreatedAt     string
	UpdatedAt     string
}

// ApprovalRepository defines the secondary port for approval persistence.
// Approvals are 1:1 with plans.
type ApprovalRepository interface {
	// Create persists a new approval.
	Create(ctx context.Context, approval *ApprovalRecord) error

	// GetByID retrieves an approval by its ID.
	GetByID(ctx context.Context, id string) (*ApprovalRecord, error)

	// GetByPlan retrieves an approval by plan ID.
	GetByPlan(ctx context.Context, planID string) (*ApprovalRecord, error)

	// List retrieves approvals matching the given filters.
	List(ctx context.Context, filters ApprovalFilters) ([]*ApprovalRecord, error)

	// Delete removes an approval from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available approval ID.
	GetNextID(ctx context.Context) (string, error)

	// PlanExists checks if a plan exists (for validation).
	PlanExists(ctx context.Context, planID string) (bool, error)

	// TaskExists checks if a task exists (for validation).
	TaskExists(ctx context.Context, taskID string) (bool, error)

	// PlanHasApproval checks if a plan already has an approval (for 1:1 constraint).
	PlanHasApproval(ctx context.Context, planID string) (bool, error)
}

// ApprovalRecord represents an approval as stored in persistence.
type ApprovalRecord struct {
	ID             string
	PlanID         string
	TaskID         string
	Mechanism      string // 'subagent' or 'manual'
	ReviewerInput  string // Empty string means null
	ReviewerOutput string // Empty string means null
	Outcome        string // 'approved' or 'escalated'
	CreatedAt      string
}

// ApprovalFilters contains filter options for querying approvals.
type ApprovalFilters struct {
	TaskID  string
	Outcome string
}

// EscalationRepository defines the secondary port for escalation persistence.
type EscalationRepository interface {
	// Create persists a new escalation.
	Create(ctx context.Context, escalation *EscalationRecord) error

	// GetByID retrieves an escalation by its ID.
	GetByID(ctx context.Context, id string) (*EscalationRecord, error)

	// List retrieves escalations matching the given filters.
	List(ctx context.Context, filters EscalationFilters) ([]*EscalationRecord, error)

	// Update updates an existing escalation.
	Update(ctx context.Context, escalation *EscalationRecord) error

	// Delete removes an escalation from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available escalation ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of an escalation.
	UpdateStatus(ctx context.Context, id, status string, setResolved bool) error

	// Resolve resolves an escalation with resolution text.
	Resolve(ctx context.Context, id, resolution, resolvedBy string) error

	// PlanExists checks if a plan exists (for validation).
	PlanExists(ctx context.Context, planID string) (bool, error)

	// TaskExists checks if a task exists (for validation).
	TaskExists(ctx context.Context, taskID string) (bool, error)

	// ApprovalExists checks if an approval exists (for validation).
	ApprovalExists(ctx context.Context, approvalID string) (bool, error)
}

// EscalationRecord represents an escalation as stored in persistence.
type EscalationRecord struct {
	ID            string
	ApprovalID    string // Empty string means null
	PlanID        string
	TaskID        string
	Reason        string
	Status        string // 'pending', 'resolved', 'dismissed'
	RoutingRule   string // e.g. 'workshop_gatehouse'
	OriginActorID string
	TargetActorID string // Empty string means null
	Resolution    string // Empty string means null
	ResolvedBy    string // Empty string means null
	CreatedAt     string
	ResolvedAt    string // Empty string means null
}

// EscalationFilters contains filter options for querying escalations.
type EscalationFilters struct {
	TaskID        string
	Status        string
	TargetActorID string
}

// WorkshopLogRepository defines the secondary port for workshop log (audit trail) persistence.
// Logs are immutable - no Update operations, but old entries can be pruned.
type WorkshopLogRepository interface {
	// Create persists a new workshop log entry.
	Create(ctx context.Context, log *WorkshopLogRecord) error

	// GetByID retrieves a log entry by its ID.
	GetByID(ctx context.Context, id string) (*WorkshopLogRecord, error)

	// List retrieves log entries matching the given filters.
	List(ctx context.Context, filters WorkshopLogFilters) ([]*WorkshopLogRecord, error)

	// GetNextID returns the next available log ID.
	GetNextID(ctx context.Context) (string, error)

	// WorkshopExists checks if a workshop exists (for validation).
	WorkshopExists(ctx context.Context, workshopID string) (bool, error)

	// PruneOlderThan deletes log entries older than the given number of days.
	// Returns the number of deleted entries.
	PruneOlderThan(ctx context.Context, days int) (int, error)
}

// WorkshopLogRecord represents a workshop log entry as stored in persistence.
type WorkshopLogRecord struct {
	ID         string
	WorkshopID string
	Timestamp  string
	ActorID    string // Empty string means null
	EntityType string
	EntityID   string
	Action     string // 'create', 'update', 'delete'
	FieldName  string // Empty string means null - for updates only
	OldValue   string // Empty string means null
	NewValue   string // Empty string means null
	CreatedAt  string
}

// WorkshopLogFilters contains filter options for querying logs.
type WorkshopLogFilters struct {
	WorkshopID string
	EntityType string
	EntityID   string
	ActorID    string
	Action     string
	Limit      int
}

// HookEventRepository defines the secondary port for hook event persistence.
// Hook events are immutable audit records of Claude Code hook invocations.
type HookEventRepository interface {
	// Create persists a new hook event.
	Create(ctx context.Context, event *HookEventRecord) error

	// GetByID retrieves a hook event by its ID.
	GetByID(ctx context.Context, id string) (*HookEventRecord, error)

	// List retrieves hook events matching the given filters.
	List(ctx context.Context, filters HookEventFilters) ([]*HookEventRecord, error)

	// GetNextID returns the next available hook event ID.
	GetNextID(ctx context.Context) (string, error)
}

// HookEventRecord represents a hook event as stored in persistence.
type HookEventRecord struct {
	ID                  string
	WorkbenchID         string
	HookType            string // 'Stop', 'UserPromptSubmit'
	Timestamp           string
	PayloadJSON         string // Empty string means null
	Cwd                 string // Empty string means null
	SessionID           string // Empty string means null
	ShipmentID          string // Empty string means null
	ShipmentStatus      string // Empty string means null
	TaskCountIncomplete int    // -1 means null
	Decision            string // 'allow', 'block'
	Reason              string // Empty string means null
	DurationMs          int    // -1 means null
	Error               string // Empty string means null
	CreatedAt           string
}

// HookEventFilters contains filter options for querying hook events.
type HookEventFilters struct {
	WorkbenchID string
	HookType    string
	Limit       int
}
