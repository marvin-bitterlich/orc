package app

import (
	"context"
	"fmt"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// MessageServiceImpl implements the MessageService interface.
type MessageServiceImpl struct {
	messageRepo secondary.MessageRepository
}

// NewMessageService creates a new MessageService with injected dependencies.
func NewMessageService(messageRepo secondary.MessageRepository) *MessageServiceImpl {
	return &MessageServiceImpl{
		messageRepo: messageRepo,
	}
}

// CreateMessage creates a new message.
func (s *MessageServiceImpl) CreateMessage(ctx context.Context, req primary.CreateMessageRequest) (*primary.CreateMessageResponse, error) {
	// Validate commission exists
	exists, err := s.messageRepo.CommissionExists(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate commission: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("commission %s not found", req.CommissionID)
	}

	// Get next ID
	nextID, err := s.messageRepo.GetNextID(ctx, req.CommissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message ID: %w", err)
	}

	// Create record
	record := &secondary.MessageRecord{
		ID:           nextID,
		Sender:       req.Sender,
		Recipient:    req.Recipient,
		Subject:      req.Subject,
		Body:         req.Body,
		CommissionID: req.CommissionID,
	}

	if err := s.messageRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Fetch created message
	created, err := s.messageRepo.GetByID(ctx, nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created message: %w", err)
	}

	return &primary.CreateMessageResponse{
		MessageID: created.ID,
		Message:   s.recordToMessage(created),
	}, nil
}

// GetMessage retrieves a message by ID.
func (s *MessageServiceImpl) GetMessage(ctx context.Context, messageID string) (*primary.Message, error) {
	record, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return s.recordToMessage(record), nil
}

// ListMessages lists messages for a recipient.
func (s *MessageServiceImpl) ListMessages(ctx context.Context, recipient string, unreadOnly bool) ([]*primary.Message, error) {
	records, err := s.messageRepo.List(ctx, secondary.MessageFilters{
		Recipient:  recipient,
		UnreadOnly: unreadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messages := make([]*primary.Message, len(records))
	for i, r := range records {
		messages[i] = s.recordToMessage(r)
	}
	return messages, nil
}

// MarkRead marks a message as read.
func (s *MessageServiceImpl) MarkRead(ctx context.Context, messageID string) error {
	return s.messageRepo.MarkRead(ctx, messageID)
}

// GetConversation retrieves all messages between two agents.
func (s *MessageServiceImpl) GetConversation(ctx context.Context, agent1, agent2 string) ([]*primary.Message, error) {
	records, err := s.messageRepo.GetConversation(ctx, agent1, agent2)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	messages := make([]*primary.Message, len(records))
	for i, r := range records {
		messages[i] = s.recordToMessage(r)
	}
	return messages, nil
}

// GetUnreadCount returns the count of unread messages for a recipient.
func (s *MessageServiceImpl) GetUnreadCount(ctx context.Context, recipient string) (int, error) {
	return s.messageRepo.GetUnreadCount(ctx, recipient)
}

// Helper methods

func (s *MessageServiceImpl) recordToMessage(r *secondary.MessageRecord) *primary.Message {
	return &primary.Message{
		ID:           r.ID,
		Sender:       r.Sender,
		Recipient:    r.Recipient,
		Subject:      r.Subject,
		Body:         r.Body,
		Timestamp:    r.Timestamp,
		Read:         r.Read,
		CommissionID: r.CommissionID,
	}
}

// Ensure MessageServiceImpl implements the interface.
var _ primary.MessageService = (*MessageServiceImpl)(nil)
