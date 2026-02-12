// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	coreworkshop "github.com/example/orc/internal/core/workshop"
	"github.com/example/orc/internal/ports/secondary"
)

// WorkshopRepository implements secondary.WorkshopRepository with SQLite.
type WorkshopRepository struct {
	db *sql.DB
}

// NewWorkshopRepository creates a new SQLite workshop repository.
func NewWorkshopRepository(db *sql.DB) *WorkshopRepository {
	return &WorkshopRepository{db: db}
}

// Create persists a new workshop.
func (r *WorkshopRepository) Create(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	// Verify factory exists
	var exists int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM factories WHERE id = ?", workshop.FactoryID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify factory: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("factory %s not found", workshop.FactoryID)
	}

	// Generate workshop ID by finding max existing ID
	var maxID int
	err = r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM workshops",
	).Scan(&maxID)
	if err != nil {
		return fmt.Errorf("failed to generate workshop ID: %w", err)
	}

	id := coreworkshop.GenerateWorkshopID(maxID)

	// If no name provided, use name pool
	name := workshop.Name
	if name == "" {
		// Count existing workshops to get next name
		var count int
		err = r.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM workshops",
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count workshops: %w", err)
		}
		name = coreworkshop.GetNextWorkshopName(count)
	}

	status := workshop.Status
	if status == "" {
		status = "active"
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO workshops (id, factory_id, name, status) VALUES (?, ?, ?, ?)",
		id, workshop.FactoryID, name, status,
	)
	if err != nil {
		return fmt.Errorf("failed to create workshop: %w", err)
	}

	// Update the record with the generated ID and name
	workshop.ID = id
	workshop.Name = name
	return nil
}

// GetByID retrieves a workshop by its ID.
func (r *WorkshopRepository) GetByID(ctx context.Context, id string) (*secondary.WorkshopRecord, error) {
	var (
		activeCommissionID sql.NullString
		createdAt          time.Time
		updatedAt          time.Time
	)

	record := &secondary.WorkshopRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, factory_id, name, status, active_commission_id, created_at, updated_at FROM workshops WHERE id = ?",
		id,
	).Scan(&record.ID, &record.FactoryID, &record.Name, &record.Status, &activeCommissionID, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("workshop %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workshop: %w", err)
	}

	record.ActiveCommissionID = activeCommissionID.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	return record, nil
}

// List retrieves workshops matching the given filters.
func (r *WorkshopRepository) List(ctx context.Context, filters secondary.WorkshopFilters) ([]*secondary.WorkshopRecord, error) {
	query := "SELECT id, factory_id, name, status, active_commission_id, created_at, updated_at FROM workshops WHERE 1=1"
	args := []any{}

	if filters.FactoryID != "" {
		query += " AND factory_id = ?"
		args = append(args, filters.FactoryID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workshops: %w", err)
	}
	defer rows.Close()

	var workshops []*secondary.WorkshopRecord
	for rows.Next() {
		var (
			activeCommissionID sql.NullString
			createdAt          time.Time
			updatedAt          time.Time
		)

		record := &secondary.WorkshopRecord{}
		err := rows.Scan(&record.ID, &record.FactoryID, &record.Name, &record.Status, &activeCommissionID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workshop: %w", err)
		}

		record.ActiveCommissionID = activeCommissionID.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		workshops = append(workshops, record)
	}

	return workshops, nil
}

// Update updates an existing workshop.
func (r *WorkshopRepository) Update(ctx context.Context, workshop *secondary.WorkshopRecord) error {
	query := "UPDATE workshops SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if workshop.Name != "" {
		query += ", name = ?"
		args = append(args, workshop.Name)
	}

	if workshop.Status != "" {
		query += ", status = ?"
		args = append(args, workshop.Status)
	}

	query += " WHERE id = ?"
	args = append(args, workshop.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workshop: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workshop %s not found", workshop.ID)
	}

	return nil
}

// Delete removes a workshop from persistence.
func (r *WorkshopRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM workshops WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete workshop: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workshop %s not found", id)
	}

	return nil
}

// GetNextID returns the next available workshop ID.
func (r *WorkshopRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM workshops",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next workshop ID: %w", err)
	}

	return coreworkshop.GenerateWorkshopID(maxID), nil
}

// CountWorkbenches returns the number of workbenches for a workshop.
func (r *WorkshopRepository) CountWorkbenches(ctx context.Context, workshopID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM workbenches WHERE workshop_id = ?",
		workshopID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workbenches: %w", err)
	}

	return count, nil
}

// CountByFactory returns the number of workshops for a factory.
func (r *WorkshopRepository) CountByFactory(ctx context.Context, factoryID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM workshops WHERE factory_id = ?",
		factoryID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workshops by factory: %w", err)
	}

	return count, nil
}

// FactoryExists checks if a factory exists.
func (r *WorkshopRepository) FactoryExists(ctx context.Context, factoryID string) (bool, error) {
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

// SetActiveCommissionID updates the active commission for a workshop (Goblin context).
// Pass empty string to clear.
func (r *WorkshopRepository) SetActiveCommissionID(ctx context.Context, workshopID, commissionID string) error {
	var value any
	if commissionID == "" {
		value = nil
	} else {
		value = commissionID
	}

	result, err := r.db.ExecContext(ctx,
		"UPDATE workshops SET active_commission_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		value, workshopID,
	)
	if err != nil {
		return fmt.Errorf("failed to update workshop active commission: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("workshop %s not found", workshopID)
	}

	return nil
}

// GetActiveCommissions returns commission IDs derived from focus:
// - All workbench focused_ids in workshop (resolved to commission)
// Returns deduplicated commission IDs.
func (r *WorkshopRepository) GetActiveCommissions(ctx context.Context, workshopID string) ([]string, error) {
	// Collect all focused_ids from workbenches
	focusedIDs := make(map[string]bool)

	// Get all workbench focused_ids for this workshop
	rows, err := r.db.QueryContext(ctx,
		"SELECT focused_id FROM workbenches WHERE workshop_id = ? AND status = 'active' AND focused_id IS NOT NULL AND focused_id != ''",
		workshopID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workbench focuses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var focusedID string
		if err := rows.Scan(&focusedID); err != nil {
			return nil, fmt.Errorf("failed to scan workbench focused_id: %w", err)
		}
		if focusedID != "" {
			focusedIDs[focusedID] = true
		}
	}

	if len(focusedIDs) == 0 {
		return nil, nil
	}

	// Resolve focused_ids to commission IDs
	commissionIDs := make(map[string]bool)

	for focusedID := range focusedIDs {
		commissionID, err := r.resolveToCommission(ctx, focusedID)
		if err != nil {
			// Skip unresolvable focuses (entity might have been deleted)
			continue
		}
		if commissionID != "" {
			commissionIDs[commissionID] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(commissionIDs))
	for id := range commissionIDs {
		result = append(result, id)
	}

	return result, nil
}

// resolveToCommission resolves a focused_id to its commission ID.
// - COMM-xxx: returns directly
// - SHIP-xxx: returns shipment.commission_id
// - TOME-xxx: returns tome.commission_id
func (r *WorkshopRepository) resolveToCommission(ctx context.Context, focusedID string) (string, error) {
	if len(focusedID) < 5 {
		return "", nil
	}

	prefix := focusedID[:5]
	switch prefix {
	case "COMM-":
		// Direct commission reference
		return focusedID, nil

	case "SHIP-":
		// Resolve via shipment
		var commissionID sql.NullString
		err := r.db.QueryRowContext(ctx,
			"SELECT commission_id FROM shipments WHERE id = ?",
			focusedID,
		).Scan(&commissionID)
		if err != nil {
			return "", err
		}
		return commissionID.String, nil

	case "TOME-":
		// Resolve via tome
		var commissionID sql.NullString
		err := r.db.QueryRowContext(ctx,
			"SELECT commission_id FROM tomes WHERE id = ?",
			focusedID,
		).Scan(&commissionID)
		if err != nil {
			return "", err
		}
		return commissionID.String, nil

	default:
		return "", nil
	}
}

// Ensure WorkshopRepository implements the interface
var _ secondary.WorkshopRepository = (*WorkshopRepository)(nil)
