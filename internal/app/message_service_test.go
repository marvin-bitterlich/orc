package app

import (
	"context"
	"errors"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// mockMessageRepository implements secondary.MessageRepository for testing.
type mockMessageRepository struct {
	messages            map[string]*secondary.MessageRecord
	createErr           error
	getErr              error
	listErr             error
	markReadErr         error
	missionExistsResult bool
	missionExistsErr    error
}

func newMockMessageRepository() *mockMessageRepository {
	return &mockMessageRepository{
		messages:            make(map[string]*secondary.MessageRecord),
		missionExistsResult: true,
	}
}

func (m *mockMessageRepository) Create(ctx context.Context, message *secondary.MessageRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.messages[message.ID] = message
	return nil
}

func (m *mockMessageRepository) GetByID(ctx context.Context, id string) (*secondary.MessageRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if message, ok := m.messages[id]; ok {
		return message, nil
	}
	return nil, errors.New("message not found")
}

func (m *mockMessageRepository) List(ctx context.Context, filters secondary.MessageFilters) ([]*secondary.MessageRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.MessageRecord
	for _, msg := range m.messages {
		if filters.Recipient != "" && msg.Recipient != filters.Recipient {
			continue
		}
		if filters.UnreadOnly && msg.Read {
			continue
		}
		result = append(result, msg)
	}
	return result, nil
}

func (m *mockMessageRepository) MarkRead(ctx context.Context, id string) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	if message, ok := m.messages[id]; ok {
		message.Read = true
	}
	return nil
}

func (m *mockMessageRepository) GetConversation(ctx context.Context, agent1, agent2 string) ([]*secondary.MessageRecord, error) {
	var result []*secondary.MessageRecord
	for _, msg := range m.messages {
		// Match messages between the two agents (in either direction)
		if (msg.Sender == agent1 && msg.Recipient == agent2) ||
			(msg.Sender == agent2 && msg.Recipient == agent1) {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *mockMessageRepository) GetUnreadCount(ctx context.Context, recipient string) (int, error) {
	count := 0
	for _, msg := range m.messages {
		if msg.Recipient == recipient && !msg.Read {
			count++
		}
	}
	return count, nil
}

func (m *mockMessageRepository) GetNextID(ctx context.Context, missionID string) (string, error) {
	return "MSG-MISSION-001-001", nil
}

func (m *mockMessageRepository) MissionExists(ctx context.Context, missionID string) (bool, error) {
	if m.missionExistsErr != nil {
		return false, m.missionExistsErr
	}
	return m.missionExistsResult, nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestMessageService() (*MessageServiceImpl, *mockMessageRepository) {
	messageRepo := newMockMessageRepository()
	service := NewMessageService(messageRepo)
	return service, messageRepo
}

// ============================================================================
// CreateMessage Tests
// ============================================================================

func TestCreateMessage_Success(t *testing.T) {
	service, _ := newTestMessageService()
	ctx := context.Background()

	resp, err := service.CreateMessage(ctx, primary.CreateMessageRequest{
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Task Assignment",
		Body:      "Please work on TASK-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.MessageID == "" {
		t.Error("expected message ID to be set")
	}
	if resp.Message.Subject != "Task Assignment" {
		t.Errorf("expected subject 'Task Assignment', got '%s'", resp.Message.Subject)
	}
	if resp.Message.Read {
		t.Error("expected message to be unread initially")
	}
}

func TestCreateMessage_MissionNotFound(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.missionExistsResult = false

	_, err := service.CreateMessage(ctx, primary.CreateMessageRequest{
		MissionID: "MISSION-NONEXISTENT",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Test",
		Body:      "Test message",
	})

	if err == nil {
		t.Fatal("expected error for non-existent mission, got nil")
	}
}

// ============================================================================
// GetMessage Tests
// ============================================================================

func TestGetMessage_Found(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Test Message",
		Body:      "Hello",
	}

	message, err := service.GetMessage(ctx, "MSG-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if message.Subject != "Test Message" {
		t.Errorf("expected subject 'Test Message', got '%s'", message.Subject)
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	service, _ := newTestMessageService()
	ctx := context.Background()

	_, err := service.GetMessage(ctx, "MSG-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent message, got nil")
	}
}

// ============================================================================
// ListMessages Tests
// ============================================================================

func TestListMessages_FilterByRecipient(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Message 1",
	}
	messageRepo.messages["MSG-002"] = &secondary.MessageRecord{
		ID:        "MSG-002",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-002",
		Subject:   "Message 2",
	}

	messages, err := service.ListMessages(ctx, "IMP-GROVE-001", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}

func TestListMessages_UnreadOnly(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Unread Message",
		Read:      false,
	}
	messageRepo.messages["MSG-002"] = &secondary.MessageRecord{
		ID:        "MSG-002",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Read Message",
		Read:      true,
	}

	messages, err := service.ListMessages(ctx, "IMP-GROVE-001", true)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 unread message, got %d", len(messages))
	}
}

// ============================================================================
// MarkRead Tests
// ============================================================================

func TestMarkRead_Success(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Test Message",
		Read:      false,
	}

	err := service.MarkRead(ctx, "MSG-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !messageRepo.messages["MSG-001"].Read {
		t.Error("expected message to be marked as read")
	}
}

// ============================================================================
// GetConversation Tests
// ============================================================================

func TestGetConversation_Success(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "From ORC",
	}
	messageRepo.messages["MSG-002"] = &secondary.MessageRecord{
		ID:        "MSG-002",
		MissionID: "MISSION-001",
		Sender:    "IMP-GROVE-001",
		Recipient: "ORC",
		Subject:   "Reply from IMP",
	}
	messageRepo.messages["MSG-003"] = &secondary.MessageRecord{
		ID:        "MSG-003",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-002",
		Subject:   "Different conversation",
	}

	messages, err := service.GetConversation(ctx, "ORC", "IMP-GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages in conversation, got %d", len(messages))
	}
}

func TestGetConversation_ReverseOrder(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Subject:   "Message",
	}

	// Should work regardless of argument order
	messages, err := service.GetConversation(ctx, "IMP-GROVE-001", "ORC")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}
}

// ============================================================================
// GetUnreadCount Tests
// ============================================================================

func TestGetUnreadCount_Success(t *testing.T) {
	service, messageRepo := newTestMessageService()
	ctx := context.Background()

	messageRepo.messages["MSG-001"] = &secondary.MessageRecord{
		ID:        "MSG-001",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Read:      false,
	}
	messageRepo.messages["MSG-002"] = &secondary.MessageRecord{
		ID:        "MSG-002",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Read:      false,
	}
	messageRepo.messages["MSG-003"] = &secondary.MessageRecord{
		ID:        "MSG-003",
		MissionID: "MISSION-001",
		Sender:    "ORC",
		Recipient: "IMP-GROVE-001",
		Read:      true,
	}

	count, err := service.GetUnreadCount(ctx, "IMP-GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 unread messages, got %d", count)
	}
}

func TestGetUnreadCount_NoMessages(t *testing.T) {
	service, _ := newTestMessageService()
	ctx := context.Background()

	count, err := service.GetUnreadCount(ctx, "IMP-GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 unread messages, got %d", count)
	}
}
