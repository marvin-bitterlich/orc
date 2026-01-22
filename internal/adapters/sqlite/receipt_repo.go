// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// ReceiptRepository implements secondary.ReceiptRepository with SQLite.
type ReceiptRepository struct {
	db *sql.DB
}

// NewReceiptRepository creates a new SQLite receipt repository.
func NewReceiptRepository(db *sql.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

// Create persists a new receipt.
func (r *ReceiptRepository) Create(ctx context.Context, rec *secondary.ReceiptRecord) error {
	var evidence, verificationNotes sql.NullString
	if rec.Evidence != "" {
		evidence = sql.NullString{String: rec.Evidence, Valid: true}
	}
	if rec.VerificationNotes != "" {
		verificationNotes = sql.NullString{String: rec.VerificationNotes, Valid: true}
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO receipts (id, shipment_id, delivered_outcome, evidence, verification_notes, status) VALUES (?, ?, ?, ?, ?, ?)`,
		rec.ID,
		rec.ShipmentID,
		rec.DeliveredOutcome,
		evidence,
		verificationNotes,
		rec.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create receipt: %w", err)
	}

	return nil
}

// GetByID retrieves a receipt by its ID.
func (r *ReceiptRepository) GetByID(ctx context.Context, id string) (*secondary.ReceiptRecord, error) {
	var (
		evidence          sql.NullString
		verificationNotes sql.NullString
		createdAt         time.Time
		updatedAt         time.Time
	)

	record := &secondary.ReceiptRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM receipts WHERE id = ?`,
		id,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.DeliveredOutcome,
		&evidence,
		&verificationNotes,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("receipt %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}
	record.Evidence = evidence.String
	record.VerificationNotes = verificationNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// GetByShipment retrieves a receipt by its shipment ID.
func (r *ReceiptRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.ReceiptRecord, error) {
	var (
		evidence          sql.NullString
		verificationNotes sql.NullString
		createdAt         time.Time
		updatedAt         time.Time
	)

	record := &secondary.ReceiptRecord{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM receipts WHERE shipment_id = ?`,
		shipmentID,
	).Scan(&record.ID,
		&record.ShipmentID,
		&record.DeliveredOutcome,
		&evidence,
		&verificationNotes,
		&record.Status,
		&createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("receipt for shipment %s not found", shipmentID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}
	record.Evidence = evidence.String
	record.VerificationNotes = verificationNotes.String
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)

	return record, nil
}

// List retrieves receipts matching the given filters.
func (r *ReceiptRepository) List(ctx context.Context, filters secondary.ReceiptFilters) ([]*secondary.ReceiptRecord, error) {
	query := `SELECT id, shipment_id, delivered_outcome, evidence, verification_notes, status, created_at, updated_at FROM receipts WHERE 1=1`
	args := []any{}

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
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}
	defer rows.Close()

	var recs []*secondary.ReceiptRecord
	for rows.Next() {
		var (
			evidence          sql.NullString
			verificationNotes sql.NullString
			createdAt         time.Time
			updatedAt         time.Time
		)

		record := &secondary.ReceiptRecord{}
		err := rows.Scan(&record.ID,
			&record.ShipmentID,
			&record.DeliveredOutcome,
			&evidence,
			&verificationNotes,
			&record.Status,
			&createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan receipt: %w", err)
		}
		record.Evidence = evidence.String
		record.VerificationNotes = verificationNotes.String
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)

		recs = append(recs, record)
	}

	return recs, nil
}

// Update updates an existing receipt.
func (r *ReceiptRepository) Update(ctx context.Context, rec *secondary.ReceiptRecord) error {
	query := "UPDATE receipts SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if rec.DeliveredOutcome != "" {
		query += ", delivered_outcome = ?"
		args = append(args, rec.DeliveredOutcome)
	}
	if rec.Evidence != "" {
		query += ", evidence = ?"
		args = append(args, sql.NullString{String: rec.Evidence, Valid: true})
	}
	if rec.VerificationNotes != "" {
		query += ", verification_notes = ?"
		args = append(args, sql.NullString{String: rec.VerificationNotes, Valid: true})
	}

	query += " WHERE id = ?"
	args = append(args, rec.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update receipt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("receipt %s not found", rec.ID)
	}

	return nil
}

// Delete removes a receipt from persistence.
func (r *ReceiptRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM receipts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete receipt: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("receipt %s not found", id)
	}

	return nil
}

// GetNextID returns the next available receipt ID.
func (r *ReceiptRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	prefixLen := len("REC-") + 1
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COALESCE(MAX(CAST(SUBSTR(id, %d) AS INTEGER)), 0) FROM receipts", prefixLen),
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next receipt ID: %w", err)
	}

	return fmt.Sprintf("REC-%03d", maxID+1), nil
}

// UpdateStatus updates the status of a receipt.
func (r *ReceiptRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE receipts SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update receipt status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("receipt %s not found", id)
	}

	return nil
}

// ShipmentExists checks if a shipment exists.
func (r *ReceiptRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentHasREC checks if a shipment already has a REC (for 1:1 constraint).
func (r *ReceiptRepository) ShipmentHasREC(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM receipts WHERE shipment_id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing receipt: %w", err)
	}
	return count > 0, nil
}

// GetWOStatus retrieves the status of a shipment's Work Order.
func (r *ReceiptRepository) GetWOStatus(ctx context.Context, shipmentID string) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, "SELECT status FROM work_orders WHERE shipment_id = ?", shipmentID).Scan(&status)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("Work Order for shipment %s not found", shipmentID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get Work Order status: %w", err)
	}
	return status, nil
}

// AllCRECsVerified checks if all CRECs for a shipment are verified.
// Returns true if there are no CRECs (vacuously true) or all CRECs are verified.
func (r *ReceiptRepository) AllCRECsVerified(ctx context.Context, shipmentID string) (bool, error) {
	// Count total CRECs and non-verified CRECs for the shipment
	// Use COALESCE to handle NULL when no rows match
	var total, nonVerified int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN status != 'verified' THEN 1 ELSE 0 END), 0) FROM cycle_receipts WHERE shipment_id = ?`,
		shipmentID,
	).Scan(&total, &nonVerified)
	if err != nil {
		return false, fmt.Errorf("failed to check CREC verification status: %w", err)
	}

	// If no CRECs exist, they are vacuously all verified
	if total == 0 {
		return true, nil
	}

	// All verified if none are non-verified
	return nonVerified == 0, nil
}

// Ensure ReceiptRepository implements the interface
var _ secondary.ReceiptRepository = (*ReceiptRepository)(nil)
