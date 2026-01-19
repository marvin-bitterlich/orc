// Package primary defines the primary ports (driving adapters) for the application.
// These are the interfaces through which the outside world drives the application.
package primary

import "context"

// MissionService defines the primary port for mission operations.
// This interface documents the intended contract for mission management.
// Implementations will live in the application layer, adapters in CLI/API layers.
type MissionService interface {
	// CreateMission creates a new mission with the given parameters.
	CreateMission(ctx context.Context, req CreateMissionRequest) (*CreateMissionResponse, error)

	// StartMission begins execution of a mission (activates it).
	StartMission(ctx context.Context, req StartMissionRequest) (*StartMissionResponse, error)

	// LaunchMission creates and immediately starts a mission.
	LaunchMission(ctx context.Context, req LaunchMissionRequest) (*LaunchMissionResponse, error)

	// GetMission retrieves a mission by ID.
	GetMission(ctx context.Context, missionID string) (*Mission, error)

	// ListMissions lists missions with optional filters.
	ListMissions(ctx context.Context, filters MissionFilters) ([]*Mission, error)

	// CompleteMission marks a mission as complete.
	CompleteMission(ctx context.Context, missionID string) error

	// ArchiveMission archives a completed mission.
	ArchiveMission(ctx context.Context, missionID string) error

	// UpdateMission updates mission title and/or description.
	UpdateMission(ctx context.Context, req UpdateMissionRequest) error

	// DeleteMission deletes a mission.
	DeleteMission(ctx context.Context, req DeleteMissionRequest) error

	// PinMission pins a mission to prevent completion/archival.
	PinMission(ctx context.Context, missionID string) error

	// UnpinMission unpins a mission.
	UnpinMission(ctx context.Context, missionID string) error
}

// CreateMissionRequest contains parameters for creating a mission.
type CreateMissionRequest struct {
	Title       string
	Description string
}

// CreateMissionResponse contains the result of creating a mission.
type CreateMissionResponse struct {
	MissionID string
	Mission   *Mission
}

// StartMissionRequest contains parameters for starting a mission.
type StartMissionRequest struct {
	MissionID string
}

// StartMissionResponse contains the result of starting a mission.
type StartMissionResponse struct {
	Mission *Mission
}

// LaunchMissionRequest contains parameters for launching a mission.
type LaunchMissionRequest struct {
	Title       string
	Description string
}

// LaunchMissionResponse contains the result of launching a mission.
type LaunchMissionResponse struct {
	MissionID string
	Mission   *Mission
}

// Mission represents a mission entity at the port boundary.
type Mission struct {
	ID          string
	Title       string
	Description string
	Status      string
	CreatedAt   string
	StartedAt   string
	CompletedAt string
}

// MissionFilters contains filter options for listing missions.
type MissionFilters struct {
	Status string
	Limit  int
}

// UpdateMissionRequest contains parameters for updating a mission.
type UpdateMissionRequest struct {
	MissionID   string
	Title       string
	Description string
}

// DeleteMissionRequest contains parameters for deleting a mission.
type DeleteMissionRequest struct {
	MissionID string
	Force     bool
}
