// Package sqlite contains SQLite implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/example/orc/internal/ports/secondary"
)

// MessageRepository implements secondary.MessageRepository with SQLite.
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new SQLite message repository.
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create persists a new message.
func (r *MessageRepository) Create(ctx context.Context, message *secondary.MessageRecord) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO messages (id, sender, recipient, subject, body, commission_id, read) VALUES (?, ?, ?, ?, ?, ?, ?)",
		message.ID, message.Sender, message.Recipient, message.Subject, message.Body, message.CommissionID, 0,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by its ID.
func (r *MessageRepository) GetByID(ctx context.Context, id string) (*secondary.MessageRecord, error) {
	var (
		timestamp time.Time
		readInt   int
	)

	record := &secondary.MessageRecord{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, sender, recipient, subject, body, timestamp, read, commission_id FROM messages WHERE id = ?",
		id,
	).Scan(&record.ID, &record.Sender, &record.Recipient, &record.Subject, &record.Body, &timestamp, &readInt, &record.CommissionID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	record.Timestamp = timestamp.Format(time.RFC3339)
	record.Read = readInt == 1

	return record, nil
}

// List retrieves messages for a recipient, optionally filtering to unread only.
func (r *MessageRepository) List(ctx context.Context, filters secondary.MessageFilters) ([]*secondary.MessageRecord, error) {
	query := "SELECT id, sender, recipient, subject, body, timestamp, read, commission_id FROM messages WHERE recipient = ?"
	args := []any{filters.Recipient}

	if filters.UnreadOnly {
		query += " AND read = 0"
	}

	query += " ORDER BY timestamp DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	var messages []*secondary.MessageRecord
	for rows.Next() {
		var (
			timestamp time.Time
			readInt   int
		)

		record := &secondary.MessageRecord{}
		err := rows.Scan(&record.ID, &record.Sender, &record.Recipient, &record.Subject, &record.Body, &timestamp, &readInt, &record.CommissionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		record.Timestamp = timestamp.Format(time.RFC3339)
		record.Read = readInt == 1

		messages = append(messages, record)
	}

	return messages, nil
}

// MarkRead marks a message as read.
func (r *MessageRepository) MarkRead(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE messages SET read = 1 WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message %s not found", id)
	}

	return nil
}

// GetConversation retrieves all messages between two agents.
func (r *MessageRepository) GetConversation(ctx context.Context, agent1, agent2 string) ([]*secondary.MessageRecord, error) {
	query := `
		SELECT id, sender, recipient, subject, body, timestamp, read, commission_id
		FROM messages
		WHERE (sender = ? AND recipient = ?) OR (sender = ? AND recipient = ?)
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, agent1, agent2, agent2, agent1)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	defer rows.Close()

	var messages []*secondary.MessageRecord
	for rows.Next() {
		var (
			timestamp time.Time
			readInt   int
		)

		record := &secondary.MessageRecord{}
		err := rows.Scan(&record.ID, &record.Sender, &record.Recipient, &record.Subject, &record.Body, &timestamp, &readInt, &record.CommissionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		record.Timestamp = timestamp.Format(time.RFC3339)
		record.Read = readInt == 1

		messages = append(messages, record)
	}

	return messages, nil
}

// GetUnreadCount returns the count of unread messages for a recipient.
func (r *MessageRepository) GetUnreadCount(ctx context.Context, recipient string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM messages WHERE recipient = ? AND read = 0",
		recipient,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetNextID returns the next available message ID for a commission.
func (r *MessageRepository) GetNextID(ctx context.Context, commissionID string) (string, error) {
	prefix := fmt.Sprintf("MSG-%s-", commissionID)
	var maxID int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(CAST(REPLACE(id, ?, '') AS INTEGER)), 0) FROM messages WHERE commission_id = ?",
		prefix, commissionID,
	).Scan(&maxID)
	if err != nil {
		return "", fmt.Errorf("failed to get next message ID: %w", err)
	}

	return fmt.Sprintf("MSG-%s-%03d", commissionID, maxID+1), nil
}

// CommissionExists checks if a commission exists.
func (r *MessageRepository) CommissionExists(ctx context.Context, commissionID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commissions WHERE id = ?", commissionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check commission existence: %w", err)
	}
	return count > 0, nil
}

// Ensure MessageRepository implements the interface.
var _ secondary.MessageRepository = (*MessageRepository)(nil)
