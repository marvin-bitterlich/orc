// Package persistence contains adapters that implement secondary port interfaces
// by wrapping the existing models package.
package persistence

import (
	"context"

	"github.com/example/orc/internal/models"
	"github.com/example/orc/internal/ports/secondary"
)

// MissionRepositoryAdapter wraps the models package to implement MissionRepository.
type MissionRepositoryAdapter struct{}

// NewMissionRepository creates a new MissionRepositoryAdapter.
func NewMissionRepository() *MissionRepositoryAdapter {
	return &MissionRepositoryAdapter{}
}

// Create persists a new mission using the models package.
func (r *MissionRepositoryAdapter) Create(ctx context.Context, mission *secondary.MissionRecord) error {
	created, err := models.CreateMission(mission.Title, mission.Description)
	if err != nil {
		return err
	}
	// Update the record with the generated ID
	mission.ID = created.ID
	return nil
}

// GetByID retrieves a mission by its ID.
func (r *MissionRepositoryAdapter) GetByID(ctx context.Context, id string) (*secondary.MissionRecord, error) {
	m, err := models.GetMission(id)
	if err != nil {
		return nil, err
	}
	return r.toRecord(m), nil
}

// Update updates an existing mission.
func (r *MissionRepositoryAdapter) Update(ctx context.Context, mission *secondary.MissionRecord) error {
	// Handle status updates
	if mission.Status != "" {
		if err := models.UpdateMissionStatus(mission.ID, mission.Status); err != nil {
			return err
		}
	}
	// Handle title/description updates
	if mission.Title != "" || mission.Description != "" {
		if err := models.UpdateMission(mission.ID, mission.Title, mission.Description); err != nil {
			return err
		}
	}
	// Handle pinned status
	if mission.Pinned {
		return models.PinMission(mission.ID)
	}
	return nil
}

// Delete removes a mission from persistence.
func (r *MissionRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return models.DeleteMission(id)
}

// List retrieves missions matching the given filters.
func (r *MissionRepositoryAdapter) List(ctx context.Context, filters secondary.MissionFilters) ([]*secondary.MissionRecord, error) {
	missions, err := models.ListMissions(filters.Status)
	if err != nil {
		return nil, err
	}

	records := make([]*secondary.MissionRecord, len(missions))
	for i, m := range missions {
		records[i] = r.toRecord(m)
	}
	return records, nil
}

// GetNextID returns the next available mission ID.
// Note: The models package generates IDs internally in Create, so this is a placeholder.
func (r *MissionRepositoryAdapter) GetNextID(ctx context.Context) (string, error) {
	// ID generation happens in Create - this returns a hint
	return "", nil
}

// CountShipments returns the number of shipments for a mission.
func (r *MissionRepositoryAdapter) CountShipments(ctx context.Context, missionID string) (int, error) {
	shipments, err := models.ListShipments(missionID, "")
	if err != nil {
		return 0, err
	}
	return len(shipments), nil
}

// toRecord converts a models.Mission to a secondary.MissionRecord.
func (r *MissionRepositoryAdapter) toRecord(m *models.Mission) *secondary.MissionRecord {
	desc := ""
	if m.Description.Valid {
		desc = m.Description.String
	}
	completedAt := ""
	if m.CompletedAt.Valid {
		completedAt = m.CompletedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	return &secondary.MissionRecord{
		ID:          m.ID,
		Title:       m.Title,
		Description: desc,
		Status:      m.Status,
		Pinned:      m.Pinned,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		StartedAt:   "", // Mission model doesn't track start time
		CompletedAt: completedAt,
	}
}

// Pin pins a mission.
func (r *MissionRepositoryAdapter) Pin(ctx context.Context, id string) error {
	return models.PinMission(id)
}

// Unpin unpins a mission.
func (r *MissionRepositoryAdapter) Unpin(ctx context.Context, id string) error {
	return models.UnpinMission(id)
}

// Ensure MissionRepositoryAdapter implements the interface
var _ secondary.MissionRepository = (*MissionRepositoryAdapter)(nil)
