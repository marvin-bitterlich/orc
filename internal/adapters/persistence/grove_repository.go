package persistence

import (
	"context"

	"github.com/example/orc/internal/models"
	"github.com/example/orc/internal/ports/secondary"
)

// GroveRepositoryAdapter wraps the models package to implement GroveRepository.
type GroveRepositoryAdapter struct{}

// NewGroveRepository creates a new GroveRepositoryAdapter.
func NewGroveRepository() *GroveRepositoryAdapter {
	return &GroveRepositoryAdapter{}
}

// Create persists a new grove.
func (r *GroveRepositoryAdapter) Create(ctx context.Context, grove *secondary.GroveRecord) error {
	// Extract repos if stored (not in current GroveRecord)
	created, err := models.CreateGrove(grove.MissionID, grove.Name, grove.WorktreePath, []string{})
	if err != nil {
		return err
	}
	grove.ID = created.ID
	return nil
}

// GetByID retrieves a grove by its ID.
func (r *GroveRepositoryAdapter) GetByID(ctx context.Context, id string) (*secondary.GroveRecord, error) {
	g, err := models.GetGrove(id)
	if err != nil {
		return nil, err
	}
	return r.toRecord(g), nil
}

// GetByMission retrieves all groves for a mission.
func (r *GroveRepositoryAdapter) GetByMission(ctx context.Context, missionID string) ([]*secondary.GroveRecord, error) {
	groves, err := models.GetGrovesByMission(missionID)
	if err != nil {
		return nil, err
	}

	records := make([]*secondary.GroveRecord, len(groves))
	for i, g := range groves {
		records[i] = r.toRecord(g)
	}
	return records, nil
}

// Update updates an existing grove.
func (r *GroveRepositoryAdapter) Update(ctx context.Context, grove *secondary.GroveRecord) error {
	return models.UpdateGrovePath(grove.ID, grove.WorktreePath)
}

// Delete removes a grove from persistence.
func (r *GroveRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return models.DeleteGrove(id)
}

// GetNextID returns the next available grove ID.
func (r *GroveRepositoryAdapter) GetNextID(ctx context.Context) (string, error) {
	return "", nil
}

// toRecord converts a models.Grove to a secondary.GroveRecord.
func (r *GroveRepositoryAdapter) toRecord(g *models.Grove) *secondary.GroveRecord {
	return &secondary.GroveRecord{
		ID:           g.ID,
		Name:         g.Name,
		MissionID:    g.MissionID,
		WorktreePath: g.Path,
		Status:       g.Status,
		CreatedAt:    g.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Ensure GroveRepositoryAdapter implements the interface
var _ secondary.GroveRepository = (*GroveRepositoryAdapter)(nil)
