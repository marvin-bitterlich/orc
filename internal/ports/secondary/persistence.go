// Package secondary defines the secondary ports (driven adapters) for the application.
// These are the interfaces through which the application drives external systems.
package secondary

import "context"

// MissionRepository defines the secondary port for mission persistence.
type MissionRepository interface {
	// Create persists a new mission.
	Create(ctx context.Context, mission *MissionRecord) error

	// GetByID retrieves a mission by its ID.
	GetByID(ctx context.Context, id string) (*MissionRecord, error)

	// Update updates an existing mission.
	Update(ctx context.Context, mission *MissionRecord) error

	// Delete removes a mission from persistence.
	Delete(ctx context.Context, id string) error

	// List retrieves missions matching the given filters.
	List(ctx context.Context, filters MissionFilters) ([]*MissionRecord, error)

	// Pin pins a mission to keep it visible.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a mission.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available mission ID.
	GetNextID(ctx context.Context) (string, error)

	// CountShipments returns the number of shipments for a mission.
	CountShipments(ctx context.Context, missionID string) (int, error)
}

// MissionRecord represents a mission as stored in persistence.
type MissionRecord struct {
	ID          string
	Title       string
	Description string
	Status      string
	Pinned      bool
	CreatedAt   string
	StartedAt   string
	CompletedAt string
}

// MissionFilters contains filter options for querying missions.
type MissionFilters struct {
	Status string
	Limit  int
}

// GroveRepository defines the secondary port for grove persistence.
type GroveRepository interface {
	// Create persists a new grove.
	Create(ctx context.Context, grove *GroveRecord) error

	// GetByID retrieves a grove by its ID.
	GetByID(ctx context.Context, id string) (*GroveRecord, error)

	// GetByPath retrieves a grove by its file path.
	GetByPath(ctx context.Context, path string) (*GroveRecord, error)

	// GetByMission retrieves all groves for a mission.
	GetByMission(ctx context.Context, missionID string) ([]*GroveRecord, error)

	// List retrieves all groves, optionally filtered by mission.
	List(ctx context.Context, missionID string) ([]*GroveRecord, error)

	// Update updates an existing grove.
	Update(ctx context.Context, grove *GroveRecord) error

	// Delete removes a grove from persistence.
	Delete(ctx context.Context, id string) error

	// Rename updates the name of a grove.
	Rename(ctx context.Context, id, newName string) error

	// UpdatePath updates the path of a grove.
	UpdatePath(ctx context.Context, id, newPath string) error

	// GetNextID returns the next available grove ID.
	GetNextID(ctx context.Context) (string, error)
}

// GroveRecord represents a grove as stored in persistence.
type GroveRecord struct {
	ID           string
	Name         string
	MissionID    string
	WorktreePath string
	Status       string
	CreatedAt    string
}

// AgentIdentityProvider defines the secondary port for agent identity resolution.
// This abstracts the detection of current agent context (ORC vs IMP).
type AgentIdentityProvider interface {
	// GetCurrentIdentity returns the identity of the current agent.
	GetCurrentIdentity(ctx context.Context) (*AgentIdentity, error)
}

// AgentIdentity represents an agent's identity as provided by the secondary port.
type AgentIdentity struct {
	Type      AgentType
	ID        string // "ORC" for orchestrator, Grove ID for IMP
	FullID    string // Complete ID like "ORC" or "IMP-GROVE-001"
	MissionID string // Mission ID (empty for ORC outside mission)
}

// AgentType represents the type of agent.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a grove.
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

	// GetByGrove retrieves shipments assigned to a grove.
	GetByGrove(ctx context.Context, groveID string) ([]*ShipmentRecord, error)

	// AssignGrove assigns a shipment to a grove.
	AssignGrove(ctx context.Context, shipmentID, groveID string) error

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// GroveAssignedToOther checks if grove is assigned to another shipment.
	GroveAssignedToOther(ctx context.Context, groveID, excludeShipmentID string) (string, error)
}

// ShipmentRecord represents a shipment as stored in persistence.
type ShipmentRecord struct {
	ID              string
	MissionID       string
	Title           string
	Description     string // Empty string means null
	Status          string
	AssignedGroveID string // Empty string means null
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string // Empty string means null
}

// ShipmentFilters contains filter options for querying shipments.
type ShipmentFilters struct {
	MissionID string
	Status    string
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

	// GetByGrove retrieves tasks assigned to a grove.
	GetByGrove(ctx context.Context, groveID string) ([]*TaskRecord, error)

	// GetByShipment retrieves tasks for a shipment.
	GetByShipment(ctx context.Context, shipmentID string) ([]*TaskRecord, error)

	// UpdateStatus updates the status with optional timestamps.
	UpdateStatus(ctx context.Context, id, status string, setClaimed, setCompleted bool) error

	// Claim claims a task for a grove.
	Claim(ctx context.Context, id, groveID string) error

	// AssignGroveByShipment assigns all tasks of a shipment to a grove.
	AssignGroveByShipment(ctx context.Context, shipmentID, groveID string) error

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)

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
	ID               string
	ShipmentID       string // Empty string means null
	MissionID        string
	Title            string
	Description      string // Empty string means null
	Type             string // Empty string means null
	Status           string
	Priority         string // Empty string means null
	AssignedGroveID  string // Empty string means null
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ClaimedAt        string // Empty string means null
	CompletedAt      string // Empty string means null
	ConclaveID       string // Empty string means null
	PromotedFromID   string // Empty string means null
	PromotedFromType string // Empty string means null
}

// TaskFilters contains filter options for querying tasks.
type TaskFilters struct {
	ShipmentID string
	Status     string
	MissionID  string
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

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// UpdateStatus updates the status of a note (open/closed).
	UpdateStatus(ctx context.Context, id string, status string) error
}

// NoteRecord represents a note as stored in persistence.
type NoteRecord struct {
	ID               string
	MissionID        string
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
	Type      string
	MissionID string
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

	// GetLatestForGrove retrieves the most recent handoff for a grove.
	GetLatestForGrove(ctx context.Context, groveID string) (*HandoffRecord, error)

	// List retrieves handoffs with optional limit.
	List(ctx context.Context, limit int) ([]*HandoffRecord, error)

	// GetNextID returns the next available handoff ID.
	GetNextID(ctx context.Context) (string, error)
}

// HandoffRecord represents a handoff as stored in persistence.
type HandoffRecord struct {
	ID              string
	CreatedAt       string
	HandoffNote     string
	ActiveMissionID string // Empty string means null
	ActiveGroveID   string // Empty string means null
	TodosSnapshot   string // Empty string means null
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

	// GetByGrove retrieves tomes assigned to a grove.
	GetByGrove(ctx context.Context, groveID string) ([]*TomeRecord, error)

	// AssignGrove assigns a tome to a grove.
	AssignGrove(ctx context.Context, tomeID, groveID string) error

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)
}

// TomeRecord represents a tome as stored in persistence.
type TomeRecord struct {
	ID              string
	MissionID       string
	Title           string
	Description     string // Empty string means null
	Status          string
	AssignedGroveID string // Empty string means null
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string // Empty string means null
}

// TomeFilters contains filter options for querying tomes.
type TomeFilters struct {
	MissionID string
	Status    string
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

	// UpdateStatus updates the status and optionally completed_at timestamp.
	UpdateStatus(ctx context.Context, id, status string, setCompleted bool) error

	// GetByGrove retrieves conclaves assigned to a grove.
	GetByGrove(ctx context.Context, groveID string) ([]*ConclaveRecord, error)

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// GetTasksByConclave retrieves tasks belonging to a conclave.
	GetTasksByConclave(ctx context.Context, conclaveID string) ([]*ConclaveTaskRecord, error)

	// GetQuestionsByConclave retrieves questions belonging to a conclave.
	GetQuestionsByConclave(ctx context.Context, conclaveID string) ([]*ConclaveQuestionRecord, error)

	// GetPlansByConclave retrieves plans belonging to a conclave.
	GetPlansByConclave(ctx context.Context, conclaveID string) ([]*ConclavePlanRecord, error)
}

// ConclaveRecord represents a conclave as stored in persistence.
type ConclaveRecord struct {
	ID              string
	MissionID       string
	Title           string
	Description     string // Empty string means null
	Status          string
	AssignedGroveID string // Empty string means null
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string // Empty string means null
}

// ConclaveFilters contains filter options for querying conclaves.
type ConclaveFilters struct {
	MissionID string
	Status    string
}

// ConclaveTaskRecord represents a task as returned from conclave cross-entity query.
type ConclaveTaskRecord struct {
	ID               string
	ShipmentID       string
	MissionID        string
	Title            string
	Description      string
	Type             string
	Status           string
	Priority         string
	AssignedGroveID  string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	ClaimedAt        string
	CompletedAt      string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// ConclaveQuestionRecord represents a question as returned from conclave cross-entity query.
type ConclaveQuestionRecord struct {
	ID               string
	InvestigationID  string
	MissionID        string
	Title            string
	Description      string
	Status           string
	Answer           string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	AnsweredAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// ConclavePlanRecord represents a plan as returned from conclave cross-entity query.
type ConclavePlanRecord struct {
	ID               string
	ShipmentID       string
	MissionID        string
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

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)
}

// OperationRecord represents an operation as stored in persistence.
type OperationRecord struct {
	ID          string
	MissionID   string
	Title       string
	Description string // Empty string means null
	Status      string
	CreatedAt   string
	UpdatedAt   string
	CompletedAt string // Empty string means null
}

// OperationFilters contains filter options for querying operations.
type OperationFilters struct {
	MissionID string
	Status    string
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

	// GetByGrove retrieves investigations assigned to a grove.
	GetByGrove(ctx context.Context, groveID string) ([]*InvestigationRecord, error)

	// AssignGrove assigns an investigation to a grove.
	AssignGrove(ctx context.Context, investigationID, groveID string) error

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// GetQuestionsByInvestigation retrieves questions for an investigation.
	GetQuestionsByInvestigation(ctx context.Context, investigationID string) ([]*InvestigationQuestionRecord, error)
}

// InvestigationRecord represents an investigation as stored in persistence.
type InvestigationRecord struct {
	ID              string
	MissionID       string
	Title           string
	Description     string // Empty string means null
	Status          string
	AssignedGroveID string // Empty string means null
	Pinned          bool
	CreatedAt       string
	UpdatedAt       string
	CompletedAt     string // Empty string means null
}

// InvestigationFilters contains filter options for querying investigations.
type InvestigationFilters struct {
	MissionID string
	Status    string
}

// InvestigationQuestionRecord represents a question as returned from investigation cross-entity query.
type InvestigationQuestionRecord struct {
	ID               string
	InvestigationID  string
	MissionID        string
	Title            string
	Description      string
	Status           string
	Answer           string
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	AnsweredAt       string
	ConclaveID       string
	PromotedFromID   string
	PromotedFromType string
}

// QuestionRepository defines the secondary port for question persistence.
type QuestionRepository interface {
	// Create persists a new question.
	Create(ctx context.Context, question *QuestionRecord) error

	// GetByID retrieves a question by its ID.
	GetByID(ctx context.Context, id string) (*QuestionRecord, error)

	// List retrieves questions matching the given filters.
	List(ctx context.Context, filters QuestionFilters) ([]*QuestionRecord, error)

	// Update updates an existing question.
	Update(ctx context.Context, question *QuestionRecord) error

	// Delete removes a question from persistence.
	Delete(ctx context.Context, id string) error

	// Pin pins a question.
	Pin(ctx context.Context, id string) error

	// Unpin unpins a question.
	Unpin(ctx context.Context, id string) error

	// GetNextID returns the next available question ID.
	GetNextID(ctx context.Context) (string, error)

	// Answer sets the answer for a question and marks it as answered.
	Answer(ctx context.Context, id, answer string) error

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// InvestigationExists checks if an investigation exists (for validation).
	InvestigationExists(ctx context.Context, investigationID string) (bool, error)
}

// QuestionRecord represents a question as stored in persistence.
type QuestionRecord struct {
	ID               string
	InvestigationID  string // Empty string means null
	MissionID        string
	Title            string
	Description      string // Empty string means null
	Status           string
	Answer           string // Empty string means null
	Pinned           bool
	CreatedAt        string
	UpdatedAt        string
	AnsweredAt       string // Empty string means null
	ConclaveID       string // Empty string means null
	PromotedFromID   string // Empty string means null
	PromotedFromType string // Empty string means null
}

// QuestionFilters contains filter options for querying questions.
type QuestionFilters struct {
	InvestigationID string
	MissionID       string
	Status          string
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

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)

	// ShipmentExists checks if a shipment exists (for validation).
	ShipmentExists(ctx context.Context, shipmentID string) (bool, error)
}

// PlanRecord represents a plan as stored in persistence.
type PlanRecord struct {
	ID               string
	ShipmentID       string // Empty string means null
	MissionID        string
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
	ShipmentID string
	MissionID  string
	Status     string
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

	// GetNextID returns the next available message ID for a mission.
	GetNextID(ctx context.Context, missionID string) (string, error)

	// MissionExists checks if a mission exists (for validation).
	MissionExists(ctx context.Context, missionID string) (bool, error)
}

// MessageRecord represents a message as stored in persistence.
type MessageRecord struct {
	ID        string
	Sender    string
	Recipient string
	Subject   string
	Body      string
	Timestamp string
	Read      bool
	MissionID string
}

// MessageFilters contains filter options for querying messages.
type MessageFilters struct {
	Recipient  string
	UnreadOnly bool
}
