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

	// GetByMission retrieves all groves for a mission.
	GetByMission(ctx context.Context, missionID string) ([]*GroveRecord, error)

	// Update updates an existing grove.
	Update(ctx context.Context, grove *GroveRecord) error

	// Delete removes a grove from persistence.
	Delete(ctx context.Context, id string) error

	// GetNextID returns the next available grove ID.
	GetNextID(ctx context.Context) (string, error)
}

// GroveRecord represents a grove as stored in persistence.
type GroveRecord struct {
	ID          string
	Name        string
	MissionID   string
	WorktreePath string
	Status      string
	CreatedAt   string
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
	ID        string   // "ORC" for orchestrator, Grove ID for IMP
	FullID    string   // Complete ID like "ORC" or "IMP-GROVE-001"
	MissionID string   // Mission ID (empty for ORC outside mission)
}

// AgentType represents the type of agent.
type AgentType string

const (
	// AgentTypeORC represents the orchestrator agent.
	AgentTypeORC AgentType = "ORC"
	// AgentTypeIMP represents an implementation agent in a grove.
	AgentTypeIMP AgentType = "IMP"
)
