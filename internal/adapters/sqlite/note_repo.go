// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// NoteRepository implements secondary.NoteRepository with SQLite.
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new SQLite note repository.
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Create persists a new note.
func (r *NoteRepository) Create(ctx context.Context, note *secondary.NoteRecord) error {
	var content, noteType sql.NullString
	var shipmentID, investigationID, conclaveID, tomeID sql.NullString

	if note.Content != "" {
		content = sql.NullString{String: note.Content, Valid: true}
	}
	if note.Type != "" {
		noteType = sql.NullString{String: note.Type, Valid: true}
	}
	if note.ShipmentID != "" {
		shipmentID = sql.NullString{String: note.ShipmentID, Valid: true}
	}
	if note.InvestigationID != "" {
		investigationID = sql.NullString{String: note.InvestigationID, Valid: true}
	}
	if note.ConclaveID != "" {
		conclaveID = sql.NullString{String: note.ConclaveID, Valid: true}
	}
	if note.TomeID != "" {
		tomeID = sql.NullString{String: note.TomeID, Valid: true}
	}

	status := "open"
	if note.Status != "" {
		status = note.Status
	}

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO notes (id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		note.ID, note.CommissionID, note.Title, content, noteType, status, shipmentID, investigationID, conclaveID, tomeID,
	)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// GetByID retrieves a note by its ID.
func (r *NoteRepository) GetByID(ctx context.Context, id string) (*secondary.NoteRecord, error) {
	var (
		content          sql.NullString
		noteType         sql.NullString
		status           string
		shipmentID       sql.NullString
		investigationID  sql.NullString
		conclaveID       sql.NullString
		tomeID           sql.NullString
		pinned           bool
		createdAt        time.Time
		updatedAt        time.Time
		closedAt         sql.NullTime
		promotedFromID   sql.NullString
		promotedFromType sql.NullString
	)

	record := &secondary.NoteRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE id = ?",
		id,
	).Scan(&record.ID, &record.CommissionID, &record.Title, &content, &noteType, &status, &shipmentID, &investigationID, &conclaveID, &tomeID, &pinned, &createdAt, &updatedAt, &closedAt, &promotedFromID, &promotedFromType)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	record.Content = content.String
	record.Type = noteType.String
	record.Status = status
	record.ShipmentID = shipmentID.String
	record.InvestigationID = investigationID.String
	record.ConclaveID = conclaveID.String
	record.TomeID = tomeID.String
	record.Pinned = pinned
	record.CreatedAt = createdAt.Format(time.RFC3339)
	record.UpdatedAt = updatedAt.Format(time.RFC3339)
	if closedAt.Valid {
		record.ClosedAt = closedAt.Time.Format(time.RFC3339)
	}
	record.PromotedFromID = promotedFromID.String
	record.PromotedFromType = promotedFromType.String

	return record, nil
}

// List retrieves notes matching the given filters.
func (r *NoteRepository) List(ctx context.Context, filters secondary.NoteFilters) ([]*secondary.NoteRecord, error) {
	query := "SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE 1=1"
	args := []any{}

	if filters.Type != "" {
		query += " AND type = ?"
		args = append(args, filters.Type)
	}

	if filters.CommissionID != "" {
		query += " AND commission_id = ?"
		args = append(args, filters.CommissionID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*secondary.NoteRecord
	for rows.Next() {
		var (
			content          sql.NullString
			noteType         sql.NullString
			status           string
			shipmentID       sql.NullString
			investigationID  sql.NullString
			conclaveID       sql.NullString
			tomeID           sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			closedAt         sql.NullTime
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.NoteRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &content, &noteType, &status, &shipmentID, &investigationID, &conclaveID, &tomeID, &pinned, &createdAt, &updatedAt, &closedAt, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		record.Content = content.String
		record.Type = noteType.String
		record.Status = status
		record.ShipmentID = shipmentID.String
		record.InvestigationID = investigationID.String
		record.ConclaveID = conclaveID.String
		record.TomeID = tomeID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		notes = append(notes, record)
	}

	return notes, nil
}

// Update updates an existing note.
func (r *NoteRepository) Update(ctx context.Context, note *secondary.NoteRecord) error {
	query := "UPDATE notes SET updated_at = CURRENT_TIMESTAMP"
	args := []any{}

	if note.Title != "" {
		query += ", title = ?"
		args = append(args, note.Title)
	}

	if note.Content != "" {
		query += ", content = ?"
		args = append(args, sql.NullString{String: note.Content, Valid: true})
	}

	// Container move: when moving to a new container, clear all other container IDs
	// to maintain mutual exclusivity (a note can only belong to one container)
	if note.ShipmentID != "" {
		query += ", shipment_id = ?, investigation_id = NULL, tome_id = NULL, conclave_id = NULL"
		args = append(args, note.ShipmentID)
	} else if note.TomeID != "" {
		query += ", tome_id = ?, shipment_id = NULL, investigation_id = NULL, conclave_id = NULL"
		args = append(args, note.TomeID)
	} else if note.ConclaveID != "" {
		query += ", conclave_id = ?, shipment_id = NULL, investigation_id = NULL, tome_id = NULL"
		args = append(args, note.ConclaveID)
	} else if note.InvestigationID != "" {
		query += ", investigation_id = ?, shipment_id = NULL, tome_id = NULL, conclave_id = NULL"
		args = append(args, note.InvestigationID)
	}

	query += " WHERE id = ?"
	args = append(args, note.ID)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note %s not found", note.ID)
	}

	return nil
}

// Delete removes a note from persistence.
func (r *NoteRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	return nil
}

// Pin pins a note.
func (r *NoteRepository) Pin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE notes SET pinned = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to pin note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	return nil
}

// Unpin unpins a note.
func (r *NoteRepository) Unpin(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE notes SET pinned = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to unpin note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	return nil
}

// GetNextID returns the next available note ID.
func (r *NoteRepository) GetNextID(ctx context.Context) (string, error) {
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(SUBSTR(id, 6) AS INTEGER)), 0) FROM notes",
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next note ID: %w", err)
	}

	return fmt.Sprintf("NOTE-%03d", maxID+1), nil
}

// GetByContainer retrieves notes for a specific container.
func (r *NoteRepository) GetByContainer(ctx context.Context, containerType, containerID string) ([]*secondary.NoteRecord, error) {
	var query string
	switch containerType {
	case "shipment":
		query = "SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE shipment_id = ? ORDER BY created_at DESC"
	case "investigation":
		query = "SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE investigation_id = ? ORDER BY created_at DESC"
	case "conclave":
		// Exclude notes that have been filed to a tome - they appear under the tome instead
		query = "SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE conclave_id = ? AND tome_id IS NULL ORDER BY created_at DESC"
	case "tome":
		query = "SELECT id, commission_id, title, content, type, status, shipment_id, investigation_id, conclave_id, tome_id, pinned, created_at, updated_at, closed_at, promoted_from_id, promoted_from_type FROM notes WHERE tome_id = ? ORDER BY created_at DESC"
	default:
		return nil, fmt.Errorf("unknown container type: %s", containerType)
	}

	rows, err := r.db.QueryContext(ctx, query, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by container: %w", err)
	}
	defer rows.Close()

	var notes []*secondary.NoteRecord
	for rows.Next() {
		var (
			content          sql.NullString
			noteType         sql.NullString
			status           string
			shipmentID       sql.NullString
			investigationID  sql.NullString
			conclaveID       sql.NullString
			tomeID           sql.NullString
			pinned           bool
			createdAt        time.Time
			updatedAt        time.Time
			closedAt         sql.NullTime
			promotedFromID   sql.NullString
			promotedFromType sql.NullString
		)

		record := &secondary.NoteRecord{}
		err := rows.Scan(&record.ID, &record.CommissionID, &record.Title, &content, &noteType, &status, &shipmentID, &investigationID, &conclaveID, &tomeID, &pinned, &createdAt, &updatedAt, &closedAt, &promotedFromID, &promotedFromType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		record.Content = content.String
		record.Type = noteType.String
		record.Status = status
		record.ShipmentID = shipmentID.String
		record.InvestigationID = investigationID.String
		record.ConclaveID = conclaveID.String
		record.TomeID = tomeID.String
		record.Pinned = pinned
		record.CreatedAt = createdAt.Format(time.RFC3339)
		record.UpdatedAt = updatedAt.Format(time.RFC3339)
		if closedAt.Valid {
			record.ClosedAt = closedAt.Time.Format(time.RFC3339)
		}
		record.PromotedFromID = promotedFromID.String
		record.PromotedFromType = promotedFromType.String

		notes = append(notes, record)
	}

	return notes, nil
}

// CommissionExists checks if a commission exists.
func (r *NoteRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// ShipmentExists checks if a shipment exists.
func (r *NoteRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shipments WHERE id = ?", shipmentID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check shipment existence: %w", err)
	}
	return count > 0, nil
}

// TomeExists checks if a tome exists.
func (r *NoteRepository) TomeExists(ctx context.Context, tomeID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tomes WHERE id = ?", tomeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check tome existence: %w", err)
	}
	return count > 0, nil
}

// ConclaveExists checks if a conclave exists.
func (r *NoteRepository) ConclaveExists(ctx context.Context, conclaveID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM conclaves WHERE id = ?", conclaveID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check conclave existence: %w", err)
	}
	return count > 0, nil
}

// UpdateStatus updates the status of a note (open/closed).
func (r *NoteRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	var query string
	if status == "closed" {
		query = "UPDATE notes SET status = ?, closed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	} else {
		query = "UPDATE notes SET status = ?, closed_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	}

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update note status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note %s not found", id)
	}

	return nil
}

// Ensure NoteRepository implements the interface
var _ secondary.NoteRepository = (*NoteRepository)(nil)
