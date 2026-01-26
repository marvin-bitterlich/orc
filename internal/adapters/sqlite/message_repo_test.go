package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/ports/secondary"
)

// setupMessageTestDB creates the test database with required seed data.
func setupMessageTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB := setupTestDB(t)
	seedCommission(t, testDB, "COMM-001", "Test Commission")
	return testDB
}

// createTestMessage is a helper that creates a message with a generated ID.
func createTestMessage(t *testing.T, repo *sqlite.MessageRepository, ctx context.Context, commissionID, sender, recipient, subject, body string) *secondary.MessageRecord {
	t.Helper()

	nextID, err := repo.GetNextID(ctx, commissionID)
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}

	msg := &secondary.MessageRecord{
		ID:           nextID,
		CommissionID: commissionID,
		Sender:       sender,
		Recipient:    recipient,
		Subject:      subject,
		Body:         body,
	}

	err = repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	return msg
}

func TestMessageRepository_Create(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	msg := &secondary.MessageRecord{
		ID:           "MSG-COMM-001-001",
		CommissionID: "COMM-001",
		Sender:       "ORC",
		Recipient:    "IMP-001",
		Subject:      "Task Assignment",
		Body:         "Please complete the following task...",
	}

	err := repo.Create(ctx, msg)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify message was created
	retrieved, err := repo.GetByID(ctx, "MSG-COMM-001-001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Subject != "Task Assignment" {
		t.Errorf("expected subject 'Task Assignment', got '%s'", retrieved.Subject)
	}
	if retrieved.Read {
		t.Error("expected message to be unread")
	}
}

func TestMessageRepository_GetByID(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	msg := createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Test Subject", "Test Body")

	retrieved, err := repo.GetByID(ctx, msg.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Sender != "ORC" {
		t.Errorf("expected sender 'ORC', got '%s'", retrieved.Sender)
	}
	if retrieved.Recipient != "IMP-001" {
		t.Errorf("expected recipient 'IMP-001', got '%s'", retrieved.Recipient)
	}
	if retrieved.Timestamp == "" {
		t.Error("expected Timestamp to be set")
	}
}

func TestMessageRepository_GetByID_NotFound(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "MSG-999")
	if err == nil {
		t.Error("expected error for non-existent message")
	}
}

func TestMessageRepository_List(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 1", "Body 1")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 2", "Body 2")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-002", "Message 3", "Body 3")

	// List for IMP-001
	messages, err := repo.List(ctx, secondary.MessageFilters{Recipient: "IMP-001"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages for IMP-001, got %d", len(messages))
	}
}

func TestMessageRepository_List_UnreadOnly(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	msg1 := createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 1", "Body 1")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 2", "Body 2")

	// Mark msg1 as read
	_ = repo.MarkRead(ctx, msg1.ID)

	// List unread only
	messages, err := repo.List(ctx, secondary.MessageFilters{Recipient: "IMP-001", UnreadOnly: true})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 unread message, got %d", len(messages))
	}
}

func TestMessageRepository_MarkRead(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	msg := createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Test", "Body")

	// Initially unread
	retrieved, _ := repo.GetByID(ctx, msg.ID)
	if retrieved.Read {
		t.Error("expected message to be unread initially")
	}

	// Mark as read
	err := repo.MarkRead(ctx, msg.ID)
	if err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}

	// Verify marked as read
	retrieved, _ = repo.GetByID(ctx, msg.ID)
	if !retrieved.Read {
		t.Error("expected message to be read")
	}
}

func TestMessageRepository_MarkRead_NotFound(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	err := repo.MarkRead(ctx, "MSG-999")
	if err == nil {
		t.Error("expected error for non-existent message")
	}
}

func TestMessageRepository_GetConversation(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	// Create a conversation between ORC and IMP-001
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Hello", "Hi IMP")
	createTestMessage(t, repo, ctx, "COMM-001", "IMP-001", "ORC", "Re: Hello", "Hi ORC")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Re: Hello", "How's it going?")

	// Create a message with different participants
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-002", "Different", "To another IMP")

	// Get conversation between ORC and IMP-001
	messages, err := repo.GetConversation(ctx, "ORC", "IMP-001")
	if err != nil {
		t.Fatalf("GetConversation failed: %v", err)
	}

	if len(messages) != 3 {
		t.Errorf("expected 3 messages in conversation, got %d", len(messages))
	}

	// Verify order (oldest first for conversation)
	if messages[0].Subject != "Hello" {
		t.Errorf("expected first message 'Hello', got '%s'", messages[0].Subject)
	}
}

func TestMessageRepository_GetConversation_Symmetric(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Hello", "Body")

	// Get conversation from either direction
	messages1, _ := repo.GetConversation(ctx, "ORC", "IMP-001")
	messages2, _ := repo.GetConversation(ctx, "IMP-001", "ORC")

	if len(messages1) != len(messages2) {
		t.Errorf("expected same conversation from both directions")
	}
}

func TestMessageRepository_GetUnreadCount(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	// Initially 0
	count, err := repo.GetUnreadCount(ctx, "IMP-001")
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 unread, got %d", count)
	}

	// Create some messages
	msg1 := createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 1", "Body")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Message 2", "Body")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-002", "Message 3", "Body") // different recipient

	// Count unread for IMP-001
	count, err = repo.GetUnreadCount(ctx, "IMP-001")
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 unread, got %d", count)
	}

	// Mark one as read
	_ = repo.MarkRead(ctx, msg1.ID)

	count, err = repo.GetUnreadCount(ctx, "IMP-001")
	if err != nil {
		t.Fatalf("GetUnreadCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 unread after marking one read, got %d", count)
	}
}

func TestMessageRepository_GetNextID(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	id, err := repo.GetNextID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "MSG-COMM-001-001" {
		t.Errorf("expected MSG-COMM-001-001, got %s", id)
	}

	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Test", "Body")

	id, err = repo.GetNextID(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "MSG-COMM-001-002" {
		t.Errorf("expected MSG-COMM-001-002, got %s", id)
	}
}

func TestMessageRepository_GetNextID_DifferentCommissions(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	// Add another commission
	_, _ = db.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-002', 'Commission 2', 'active')")

	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Test", "Body")
	createTestMessage(t, repo, ctx, "COMM-001", "ORC", "IMP-001", "Test 2", "Body")

	// ID for COMM-002 should still be 001
	id, err := repo.GetNextID(ctx, "COMM-002")
	if err != nil {
		t.Fatalf("GetNextID failed: %v", err)
	}
	if id != "MSG-COMM-002-001" {
		t.Errorf("expected MSG-COMM-002-001, got %s", id)
	}
}

func TestMessageRepository_CommissionExists(t *testing.T) {
	db := setupMessageTestDB(t)
	repo := sqlite.NewMessageRepository(db)
	ctx := context.Background()

	exists, err := repo.CommissionExists(ctx, "COMM-001")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if !exists {
		t.Error("expected commission to exist")
	}

	exists, err = repo.CommissionExists(ctx, "COMM-999")
	if err != nil {
		t.Fatalf("CommissionExists failed: %v", err)
	}
	if exists {
		t.Error("expected commission to not exist")
	}
}
