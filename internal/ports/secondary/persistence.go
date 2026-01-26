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
	Type         AgentType
	ID           string // "ORC" for orchestrator, Workbench ID for IMP
	FullID       string // Complete ID like "ORC" or "IMP-BENCH-001"
	CommissionID string // Commission ID (empty for ORC outside commission)
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
type ShipmentRecord struct {
	ID                  string
	CommissionID        string
	Title               string
	Description         string // Empty string means null
	Status              string
	AssignedWorkbenchID string // Empty string means null
	RepoID              string // Empty string means null - FK to repos table
	Branch              string // Empty string means null - owned branch (e.g., ml/SHIP-001-feature-name)
	Pinned              bool
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

	// GetByInvestigation retrieves tasks for an investigation.
	GetByInvestigation(ctx context.Context, investigationID string) ([]*TaskRecord, error)

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

	// ConclaveExists checks if a conclave exists (for validation).
	ConclaveExists(ctx context.Context, conclaveID string) (bool, error)

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
	InvestigationID     string // Empty string means null
	TomeID              string // Empty string means null
	ConclaveID          string // Empty string means null
	Title               string
	Description         string // Empty string means null
	Type                string // Empty string means null
	Status              string
	Priority            string // Empty string means null
	AssignedWorkbenchID string // Empty string means null
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	ClaimedAt           string // Empty string means null
	CompletedAt         string // Empty string means null
}

// TaskFilters contains filter options for querying tasks.
type TaskFilters struct {
	ShipmentID      string
	InvestigationID string
	Status          string
	CommissionID    string
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
}

// NoteRecord represents a note as stored in persistence.
type NoteRecord struct {
	ID               string
	CommissionID     string
	Title            string
	Content          string // Empty string means null
	Type             string // Empty string means null
	Status           string // "open" or "closed"
	ShipmentID       string // Empty string means null
	InvestigationID  string // Empty string means null
	ConclaveID       string // Empty string means null
	TomeID           string // Empty string means null
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ClosedAt         string // Empty string means null
	PromotedFromID   string // Empty string means null
	PromotedFromType string // Empty string means null
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

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)
}

// TomeRecord represents a tome as stored in persistence.
type TomeRecord struct {
	ID                  string
	CommissionID        string
	ConclaveID          string // Empty string means null - optional parent conclave
	Title               string
	Description         string // Empty string means null
	Status              string
	AssignedWorkbenchID string // Empty string means null
	Pinned              bool
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
	ShipmentID       string
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

// InvestigationRepository defines the secondary port for investigation persistence.
type InvestigationRepository interface {
	// Create persists a new investigation.
	Create(ctx context.Context, investigation *InvestigationRecord) error

	// GetByID retrieves an investigation by its ID.
	GetByID(ctx context.Context, id string) (*InvestigationRecord, error)

	// List retrieves investigations matching the given filters.
	List(ctx context.Context, filters InvestigationFilters) ([]*InvestigationRecord, error)

	// Update updates an existing investigation.
	Update(ctx context.Context, investigation *InvestigationRecord) error

	// Delete removes an investigation from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins an investigation.
	Pin(ctx context.Context, id string) error

	// Unpin unpins an investigation.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available investigation ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// GetByWorkbench retrieves investigations assigned to a workbench.
	GetByWorkbench(ctx context.Context, workbenchID string) ([]*InvestigationRecord, error)

	// AssignWorkbench assigns an investigation to a workbench.
	AssignWorkbench(ctx context.Context, investigationID, workbenchID string) error

	// GetByConclave retrieves investigations for a conclave.
	GetByConclave(ctx context.Context, conclaveID string) ([]*InvestigationRecord, error)

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)
}

// InvestigationRecord represents an investigation as stored in persistence.
type InvestigationRecord struct {
	ID                  string
	CommissionID        string
	ConclaveID          string // Empty string means null
	Title               string
	Description         string // Empty string means null
	Status              string
	AssignedWorkbenchID string // Empty string means null
	Pinned              bool
	CreatedAt           string
	UpdatedAt           string
	CompletedAt         string // Empty string means null
}

// InvestigationFilters contains filter options for querying investigations.
type InvestigationFilters struct {
	CommissionID string
	ConclaveID   string
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

	// GetActivePlanForShipment retrieves the active (draft) plan for a shipment.
	GetActivePlanForShipment(ctx context.Context, shipmentID string) (*PlanRecord, error)

	// HasActivePlanForShipment checks if a shipment has an active (draft) plan.
	HasActivePlanForShipment(ctx context.Context, shipmentID string) (bool, error)

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// CycleExists checks if a cycle exists (for validation).
	CycleExists(ctx context.Context, cycleID string) (bool, error)

	// GetByCycle retrieves plans for a cycle.
	GetByCycle(ctx context.Context, cycleID string) ([]*PlanRecord, error)
}

// PlanRecord represents a plan as stored in persistence.
type PlanRecord struct {
	ID               string
	ShipmentID       string // Empty string means null
	CycleID          string // Empty string means null
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
}

// PlanFilters contains filter options for querying plans.
type PlanFilters struct {
	ShipmentID   string
	CycleID      string
	CommissionID string
	Status       string
}

// MessageRepository defines the secondary port for message persistence.
type MessageRepository interface {
	// Create persists a new message.
	Create(ctx context.Context, message *MessageRecord) error

	// GetByID retrieves a message by its ID.
	GetByID(ctx context.Context, id string) (*MessageRecord, error)

	// List retrieves messages for a recipient, optionally filtering to unread only.
	List(ctx context.Context, filters MessageFilters) ([]*MessageRecord, error)

	// MarkRead marks a message as read.
	MarkRead(ctx context.Context, id string) error

	// GetConversation retrieves all messages between two agents.
	GetConversation(ctx context.Context, agent1, agent2 string) ([]*MessageRecord, error)

	// GetUnreadCount returns the count of unread messages for a recipient.
	GetUnreadCount(ctx context.Context, recipient string) (int, error)

	// GetNextID returns the next available message ID for a commission.
	GetNextID(ctx context.Context, commissionID string) (string, error)

	// CommissionExists checks if a commission exists (for validation).
	CommissionExists(ctx context.Context, commissionID string) (bool, error)
}

// MessageRecord represents a message as stored in persistence.
type MessageRecord struct {
	ID           string
	Sender       string
	Recipient    string
	Subject      string
	Body         string
	Timestamp    string
	Read         bool
	CommissionID string
}

// MessageFilters contains filter options for querying messages.
type MessageFilters struct {
	Recipient  string
	UnreadOnly bool
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
}

// WorkshopRecord represents a workshop as stored in persistence.
type WorkshopRecord struct {
	ID        string
	FactoryID string
	Name      string
	Status    string
	CreatedAt string
	UpdatedAt string
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
	CreatedAt     string
	UpdatedAt     string
}

// WorkOrderRepository defines the secondary port for workOrder persistence.
type WorkOrderRepository interface {
	// Create persists a new workOrder.
	Create(ctx context.Context, workOrder *WorkOrderRecord) error

	// GetByID retrieves a workOrder by its ID.
	GetByID(ctx context.Context, id string) (*WorkOrderRecord, error)

	// List retrieves work_orders matching the given filters.
	List(ctx context.Context, filters WorkOrderFilters) ([]*WorkOrderRecord, error)

	// Update updates an existing workOrder.
	Update(ctx context.Context, workOrder *WorkOrderRecord) error

	// Delete removes a workOrder from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available workOrder ID.
	GetNextID(ctx context.Context) (string, error)

	// ShipmentExists checks if a shipment exists.
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// ShipmentHasWorkOrder checks if a shipment already has a workOrder (for 1:1 relationships).
	ShipmentHasWorkOrder(ctx context.Context, shipmentID string) (bool, error)

	// GetByShipment retrieves a work order by its shipment ID.
	GetByShipment(ctx context.Context, shipmentID string) (*WorkOrderRecord, error)

	// UpdateStatus updates the status of a work order.
	UpdateStatus(ctx context.Context, id, status string) error
}

// WorkOrderRecord represents a workOrder as stored in persistence.
type WorkOrderRecord struct {
	ID                 string
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria string // JSON array, empty string means null
	Status             string
	CreatedAt          string
	UpdatedAt          string
}

// WorkOrderFilters contains filter options for querying work_orders.
type WorkOrderFilters struct {
	ShipmentID string
	Status     string
}

// CycleRepository defines the secondary port for cycle persistence.
type CycleRepository interface {
	// Create persists a new cycle.
	Create(ctx context.Context, cycle *CycleRecord) error

	// GetByID retrieves a cycle by its ID.
	GetByID(ctx context.Context, id string) (*CycleRecord, error)

	// List retrieves cycles matching the given filters.
	List(ctx context.Context, filters CycleFilters) ([]*CycleRecord, error)

	// Update updates an existing cycle.
	Update(ctx context.Context, cycle *CycleRecord) error

	// Delete removes a cycle from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available cycle ID.
	GetNextID(ctx context.Context) (string, error)

	// ShipmentExists checks if a shipment exists.
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// GetNextSequenceNumber returns the next sequence number for a shipment.
	GetNextSequenceNumber(ctx context.Context, shipmentID string) (int64, error)

	// GetActiveCycle returns the active cycle for a shipment (if any).
	GetActiveCycle(ctx context.Context, shipmentID string) (*CycleRecord, error)

	// GetByShipmentAndSequence returns a specific cycle by shipment and sequence number.
	GetByShipmentAndSequence(ctx context.Context, shipmentID string, seq int64) (*CycleRecord, error)

	// UpdateStatus updates cycle status and optional timestamps.
	UpdateStatus(ctx context.Context, id, status string, setStarted, setCompleted bool) error
}

// CycleRecord represents a cycle as stored in persistence.
type CycleRecord struct {
	ID             string
	ShipmentID     string
	SequenceNumber int64
	Status         string
	CreatedAt      string
	UpdatedAt      string
	StartedAt      string // Empty string means null
	CompletedAt    string // Empty string means null
}

// CycleFilters contains filter options for querying cycles.
type CycleFilters struct {
	ShipmentID string
	Status     string
}

// CycleWorkOrderRepository defines the secondary port for cycle work order persistence.
type CycleWorkOrderRepository interface {
	// Create persists a new cycle work order.
	Create(ctx context.Context, cwo *CycleWorkOrderRecord) error

	// GetByID retrieves a cycle work order by its ID.
	GetByID(ctx context.Context, id string) (*CycleWorkOrderRecord, error)

	// GetByCycle retrieves a cycle work order by its cycle ID.
	GetByCycle(ctx context.Context, cycleID string) (*CycleWorkOrderRecord, error)

	// List retrieves cycle work orders matching the given filters.
	List(ctx context.Context, filters CycleWorkOrderFilters) ([]*CycleWorkOrderRecord, error)

	// Update updates an existing cycle work order.
	Update(ctx context.Context, cwo *CycleWorkOrderRecord) error

	// Delete removes a cycle work order from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available cycle work order ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of a cycle work order.
	UpdateStatus(ctx context.Context, id, status string) error

	// Validation helpers (for guards to query)

	// CycleExists checks if a cycle exists.
	CycleExists(ctx context.Context, cycleID string) (bool, error)

	// ShipmentExists checks if a shipment exists.
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// CycleHasCWO checks if a cycle already has a CWO (for 1:1 constraint).
	CycleHasCWO(ctx context.Context, cycleID string) (bool, error)

	// GetCycleStatus retrieves the status of a cycle.
	GetCycleStatus(ctx context.Context, cycleID string) (string, error)

	// GetCycleShipmentID retrieves the shipment ID for a cycle.
	GetCycleShipmentID(ctx context.Context, cycleID string) (string, error)
}

// CycleWorkOrderRecord represents a cycle work order as stored in persistence.
type CycleWorkOrderRecord struct {
	ID                 string
	CycleID            string
	ShipmentID         string
	Outcome            string
	AcceptanceCriteria string // JSON array, empty string means null
	Status             string
	CreatedAt          string
	UpdatedAt          string
}

// CycleWorkOrderFilters contains filter options for querying cycle work orders.
type CycleWorkOrderFilters struct {
	CycleID    string
	ShipmentID string
	Status     string
}

// CycleReceiptRepository defines the secondary port for cycle receipt persistence.
type CycleReceiptRepository interface {
	// Create persists a new cycle receipt.
	Create(ctx context.Context, crec *CycleReceiptRecord) error

	// GetByID retrieves a cycle receipt by its ID.
	GetByID(ctx context.Context, id string) (*CycleReceiptRecord, error)

	// GetByCWO retrieves a cycle receipt by its CWO ID.
	GetByCWO(ctx context.Context, cwoID string) (*CycleReceiptRecord, error)

	// List retrieves cycle receipts matching the given filters.
	List(ctx context.Context, filters CycleReceiptFilters) ([]*CycleReceiptRecord, error)

	// Update updates an existing cycle receipt.
	Update(ctx context.Context, crec *CycleReceiptRecord) error

	// Delete removes a cycle receipt from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available cycle receipt ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of a cycle receipt.
	UpdateStatus(ctx context.Context, id, status string) error

	// Validation helpers (for guards to query)

	// CWOExists checks if a CWO exists.
	CWOExists(ctx context.Context, cwoID string) (bool, error)

	// CWOHasCREC checks if a CWO already has a CREC (for 1:1 constraint).
	CWOHasCREC(ctx context.Context, cwoID string) (bool, error)

	// GetCWOStatus retrieves the status of a CWO.
	GetCWOStatus(ctx context.Context, cwoID string) (string, error)

	// GetCWOShipmentID retrieves the shipment ID for a CWO.
	GetCWOShipmentID(ctx context.Context, cwoID string) (string, error)

	// GetCWOCycleID retrieves the cycle ID for a CWO.
	GetCWOCycleID(ctx context.Context, cwoID string) (string, error)
}

// CycleReceiptRecord represents a cycle receipt as stored in persistence.
type CycleReceiptRecord struct {
	ID                string
	CWOID             string
	ShipmentID        string
	DeliveredOutcome  string
	Evidence          string // Empty string means null
	VerificationNotes string // Empty string means null
	Status            string
	CreatedAt         string
	UpdatedAt         string
}

// CycleReceiptFilters contains filter options for querying cycle receipts.
type CycleReceiptFilters struct {
	CWOID      string
	ShipmentID string
	Status     string
}

// ReceiptRepository defines the secondary port for receipt persistence.
type ReceiptRepository interface {
	// Create persists a new receipt.
	Create(ctx context.Context, rec *ReceiptRecord) error

	// GetByID retrieves a receipt by its ID.
	GetByID(ctx context.Context, id string) (*ReceiptRecord, error)

	// GetByShipment retrieves a receipt by its shipment ID.
	GetByShipment(ctx context.Context, shipmentID string) (*ReceiptRecord, error)

	// List retrieves receipts matching the given filters.
	List(ctx context.Context, filters ReceiptFilters) ([]*ReceiptRecord, error)

	// Update updates an existing receipt.
	Update(ctx context.Context, rec *ReceiptRecord) error

	// Delete removes a receipt from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available receipt ID.
	GetNextID(ctx context.Context) (string, error)

	// UpdateStatus updates the status of a receipt.
	UpdateStatus(ctx context.Context, id, status string) error

	// Validation helpers (for guards to query)

	// ShipmentExists checks if a shipment exists.
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

	// ShipmentHasREC checks if a shipment already has a REC (for 1:1 constraint).
	ShipmentHasREC(ctx context.Context, shipmentID string) (bool, error)

	// GetWOStatus retrieves the status of a shipment's Work Order.
	GetWOStatus(ctx context.Context, shipmentID string) (string, error)

	// AllCRECsVerified checks if all CRECs for a shipment are verified.
	AllCRECsVerified(ctx context.Context, shipmentID string) (bool, error)
}

// ReceiptRecord represents a receipt as stored in persistence.
type ReceiptRecord struct {
	ID                string
	ShipmentID        string
	DeliveredOutcome  string
	Evidence          string // Empty string means null
	VerificationNotes string // Empty string means null
	Status            string
	CreatedAt         string
	UpdatedAt         string
}

// ReceiptFilters contains filter options for querying receipts.
type ReceiptFilters struct {
	ShipmentID string
	Status     string
}
