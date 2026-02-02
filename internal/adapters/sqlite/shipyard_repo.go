// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	coreshipyard "github.com/example/orc/internal/core/shipyard"
	"github.com/example/orc/internal/ports/secondary"
)

// ShipyardRepository implements secondary.ShipyardRepository with SQLite.
type ShipyardRepository struct {
	db *sql.DB
}

// NewShipyardRepository creates a new SQLite shipyard repository.
func NewShipyardRepository(db *sql.DB) *ShipyardRepository {
	return &ShipyardRepository{db: db}
}

// Create persists a new shipyard.
func (r *ShipyardRepository) Create(ctx context.Context, shipyard *secondary.ShipyardRecord) error {
	// Verify factory exists
	exists, err := r.FactoryExists(ctx, shipyard.FactoryID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("factory %s not found", shipyard.FactoryID)
	}

	// Check if shipyard already exists for this factory
	var count int
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM shipyards WHERE factory_id = ?", shipyard.FactoryID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing shipyard: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("shipyard already exists for factory %s", shipyard.FactoryID)
	}

	// Generate shipyard ID if not provided
	id := shipyard.ID
	if id == "" {
		var maxID int
		err = r.db.QueryRowContext(ctx,
			"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipyards",
		).Scan(&maxID)
		if err != nil {
			return fmt.Errorf("failed to generate shipyard ID: %w", err)
		}
		id = coreshipyard.GenerateShipyardID(maxID)
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO shipyards (id, factory_id) VALUES (?, ?)",
		id, shipyard.FactoryID,
	)
	if err != nil {
		return fmt.Errorf("failed to create shipyard: %w", err)
	}

	shipyard.ID = id
	return nil
}

// GetByID retrieves a shipyard by its ID.
func (r *ShipyardRepository) GetByID(ctx context.Context, id string) (*secondary.ShipyardRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.ShipyardRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, factory_id, created_at, updated_at FROM shipyards WHERE id = ?",
		id,
	).Scan(&record.ID, &record.FactoryID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipyard %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shipyard: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetByFactoryID retrieves the shipyard for a factory.
func (r *ShipyardRepository) GetByFactoryID(ctx context.Context, factoryID string) (*secondary.ShipyardRecord, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)

	record := &secondary.ShipyardRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, factory_id, created_at, updated_at FROM shipyards WHERE factory_id = ?",
		factoryID,
	).Scan(&record.ID, &record.FactoryID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipyard for factory %s not found", factoryID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shipyard by factory: %w", err)
	}

	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// GetNextID returns the next available shipyard ID.
func (r *ShipyardRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM shipyards",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next shipyard ID: %w", err)
	}

	return coreshipyard.GenerateShipyardID(maxID), nil
}

// FactoryExists checks if a factory exists.
func (r *ShipyardRepository) FactoryExists(ctx context.Context, factoryID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM factories WHERE id = ?",
		factoryID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check factory existence: %w", err)
	}

	return count > 0, nil
}

// Ensure ShipyardRepository implements the interface
var _ secondary.ShipyardRepository = (*ShipyardRepository)(nil)
