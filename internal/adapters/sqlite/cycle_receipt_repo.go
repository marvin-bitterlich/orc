// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// CycleReceiptRepository implements secondary.CycleReceiptRepository with SQLite.
type CycleReceiptRepository struct {
	db *sql.DB
}

// NewCycleReceiptRepository creates a new SQLite cycle receipt repository.
func NewCycleReceiptRepository(db *sql.DB) *CycleReceiptRepository {
	return &CycleReceiptRepository{db: db}
}

// Create persists a new cycle receipt.
func (r *CycleReceiptRepository) Create(ctx context.Context, crec *secondary.CycleReceiptRecord) error {
	var evidence, verificationNotes sql.NullString
	if crec.Evidence != "" {
		evidence = sql.NullString{String: crec.Evidence, Valid: true}
	}
	if crec.VerificationNotes != "" {
		verificationNotes = sql.NullString{String: crec.VerificationNotes, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cycle_receipts (id, cwo_id, shipment_id, delivered_outcome, evidence, verification_notes, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		crec.ID,
		crec.CWOID,
		crec.ShipmentID,
		crec.DeliveredOutcome,
		evidence,
		verificationNotes,
		crec.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create cycle receipt: %w", err)
	}

	return nil
}

// GetByID retrieves a cycle receipt by its ID.
func (r *CycleReceiptRepository) GetByID(ctx context.Context, id string) (*secondary.CycleReceiptRecord, error) {
	var (
		evidence          sql.NullString
		verificationNotes sql.NullString
		createdAt         time.Time
		updatedAt         time.Time
	)

	record := &secondary.CycleReceiptRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, cwo_id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM cycle_receipts WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.CWOID,
		&record.ShipmentID,
		&record.DeliveredOutcome,
		&evidence,
		&verificationNotes,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle receipt %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle receipt: %w", err)
	}
	record.Evidence = evidence.String
	record.VerificationNotes = verificationNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByCWO retrieves a cycle receipt by its CWO ID.
func (r *CycleReceiptRepository) GetByCWO(ctx context.Context, cwoID string) (*secondary.CycleReceiptRecord, error) {
	var (
		evidence          sql.NullString
		verificationNotes sql.NullString
		createdAt         time.Time
		updatedAt         time.Time
	)

	record := &secondary.CycleReceiptRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, cwo_id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM cycle_receipts WHERE cwo_id = ?`,
		cwoID,
	).Scan(&record.ID,
		&record.CWOID,
		&record.ShipmentID,
		&record.DeliveredOutcome,
		&evidence,
		&verificationNotes,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cycle receipt for CWO %s not found", cwoID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cycle receipt: %w", err)
	}
	record.Evidence = evidence.String
	record.VerificationNotes = verificationNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves cycle receipts matching the given filters.
func (r *CycleReceiptRepository) List(ctx context.Context, filters secondary.CycleReceiptFilters) ([]*secondary.CycleReceiptRecord, error) {
	query := `SELECT id, cwo_id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM cycle_receipts WHERE 1=1`
	args := []any{}

	if filters.CWOID != "" {
		query += " AND cwo_id = ?"
		args = append(args, filters.CWOID)
	}

	if filters.ShipmentID != "" {
		query += " AND shipment_id = ?"
		args = append(args, filters.ShipmentID)
	}

	if filters.Status != "" {
		query += " AND status = ?"
		args = append(args, filters.Status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cycle receipts: %w", err)
	}
	defer rows.Close()

	var crecs []*secondary.CycleReceiptRecord
	for rows.Next() {
		var (
			evidence          sql.NullString
			verificationNotes sql.NullString
			createdAt         time.Time
			updatedAt         time.Time
		)

		record := &secondary.CycleReceiptRecord{}
		err := rows.Scan(&record.ID,
			&record.CWOID,
			&record.ShipmentID,
			&record.DeliveredOutcome,
			&evidence,
			&verificationNotes,
			&record.Status,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cycle receipt: %w", err)
		}
		record.Evidence = evidence.String
		record.VerificationNotes = verificationNotes.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		crecs = append(crecs, record)
	}

	return crecs, nil
}

// Update updates an existing cycle receipt.
func (r *CycleReceiptRepository) Update(ctx context.Context, crec *secondary.CycleReceiptRecord) error {
	query := "UPDATE cycle_receipts SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if crec.DeliveredOutcome != "" {
		query += ", delivered_outcome = ?"
		args = append(args, crec.DeliveredOutcome)
	}
	if crec.Evidence != "" {
		query += ", evidence = ?"
		args = append(args, sql.NullString{String: crec.Evidence, Valid: true})
	}
	if crec.VerificationNotes != "" {
		query += ", verification_notes = ?"
		args = append(args, sql.NullString{String: crec.VerificationNotes, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, crec.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update cycle receipt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle receipt %s not found", crec.ID)
	}

	return nil
}

// Delete removes a cycle receipt from persistence.
func (r *CycleReceiptRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM cycle_receipts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete cycle receipt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle receipt %s not found", id)
	}

	return nil
}

// GetNextID returns the next available cycle receipt ID.
func (r *CycleReceiptRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("CREC-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM cycle_receipts", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next cycle receipt ID: %w", err)
	}

	return fmt.Sprintf("CREC-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a cycle receipt.
func (r *CycleReceiptRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE cycle_receipts SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update cycle receipt status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("cycle receipt %s not found", id)
	}

	return nil
}

// CWOExists checks if a CWO exists.
func (r *CycleReceiptRepository) CWOExists(ctx context.Context, cwoID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cycle_work_orders WHERE id = ?", cwoID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check CWO existence: %w", err)
	}
	return count > 0, nil
}

// CWOHasCREC checks if a CWO already has a CREC (for 1:1 constraint).
func (r *CycleReceiptRepository) CWOHasCREC(ctx context.Context, cwoID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cycle_receipts WHERE cwo_id = ?", cwoID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing cycle receipt: %w", err)
	}
	return count > 0, nil
}

// GetCWOStatus retrieves the status of a CWO.
func (r *CycleReceiptRepository) GetCWOStatus(ctx context.Context, cwoID string) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, "SELECT status FROM cycle_work_orders WHERE id = ?", cwoID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("CWO %s not found", cwoID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get CWO status: %w", err)
	}
	return status, nil
}

// GetCWOShipmentID retrieves the shipment ID for a CWO.
func (r *CycleReceiptRepository) GetCWOShipmentID(ctx context.Context, cwoID string) (string, error) {
	var shipmentID string
	err := r.db.QueryRowContext(ctx, "SELECT shipment_id FROM cycle_work_orders WHERE id = ?", cwoID).Scan(&shipmentID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("CWO %s not found", cwoID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get CWO shipment ID: %w", err)
	}
	return shipmentID, nil
}

// Ensure CycleReceiptRepository implements the interface
var _ secondary.CycleReceiptRepository = (*CycleReceiptRepository)(nil)
