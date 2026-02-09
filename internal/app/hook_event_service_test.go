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

// mockHookEventRepository implements secondary.HookEventRepository for testing.
type mockHookEventRepository struct {
	events    map[string]*secondary.HookEventRecord
	createErr error
	getErr    error
	listErr   error
	nextID    int
}

func newMockHookEventRepository() *mockHookEventRepository {
	return &mockHookEventRepository{
		events: make(map[string]*secondary.HookEventRecord),
		nextID: 1,
	}
}

func (m *mockHookEventRepository) Create(ctx context.Context, event *secondary.HookEventRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockHookEventRepository) GetByID(ctx context.Context, id string) (*secondary.HookEventRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if event, ok := m.events[id]; ok {
		return event, nil
	}
	return nil, errors.New("hook event not found")
}

func (m *mockHookEventRepository) List(ctx context.Context, filters secondary.HookEventFilters) ([]*secondary.HookEventRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*secondary.HookEventRecord
	for _, event := range m.events {
		if filters.WorkbenchID != "" && event.WorkbenchID != filters.WorkbenchID {
			continue
		}
		if filters.HookType != "" && event.HookType != filters.HookType {
			continue
		}
		result = append(result, event)
	}
	if filters.Limit > 0 && len(result) > filters.Limit {
		result = result[:filters.Limit]
	}
	return result, nil
}

func (m *mockHookEventRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return "HEV-" + string('0'+byte(id/1000)) + string('0'+byte((id/100)%10)) + string('0'+byte((id/10)%10)) + string('0'+byte(id%10)), nil
}

// ============================================================================
// Test Helper
// ============================================================================

func newTestHookEventService() (*HookEventServiceImpl, *mockHookEventRepository) {
	repo := newMockHookEventRepository()
	service := NewHookEventService(repo)
	return service, repo
}

// ============================================================================
// LogHookEvent Tests
// ============================================================================

func TestLogHookEvent_Success(t *testing.T) {
	service, _ := newTestHookEventService()
	ctx := context.Background()

	resp, err := service.LogHookEvent(ctx, primary.LogHookEventRequest{
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		PayloadJSON:         `{"hook_type":"Stop"}`,
		Cwd:                 "/Users/test",
		SessionID:           "sess-123",
		ShipmentID:          "SHIP-001",
		ShipmentStatus:      "implementing",
		TaskCountIncomplete: 3,
		Decision:            "block",
		Reason:              "Incomplete tasks",
		DurationMs:          42,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.EventID == "" {
		t.Error("expected event ID to be set")
	}
	if resp.Event.WorkbenchID != "BENCH-001" {
		t.Errorf("expected workbench ID 'BENCH-001', got '%s'", resp.Event.WorkbenchID)
	}
	if resp.Event.HookType != "Stop" {
		t.Errorf("expected hook type 'Stop', got '%s'", resp.Event.HookType)
	}
	if resp.Event.Decision != "block" {
		t.Errorf("expected decision 'block', got '%s'", resp.Event.Decision)
	}
	if resp.Event.TaskCountIncomplete != 3 {
		t.Errorf("expected task count 3, got %d", resp.Event.TaskCountIncomplete)
	}
}

func TestLogHookEvent_MinimalFields(t *testing.T) {
	service, _ := newTestHookEventService()
	ctx := context.Background()

	resp, err := service.LogHookEvent(ctx, primary.LogHookEventRequest{
		WorkbenchID:         "BENCH-001",
		HookType:            "UserPromptSubmit",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Event.HookType != "UserPromptSubmit" {
		t.Errorf("expected hook type 'UserPromptSubmit', got '%s'", resp.Event.HookType)
	}
	if resp.Event.Decision != "allow" {
		t.Errorf("expected decision 'allow', got '%s'", resp.Event.Decision)
	}
}

// ============================================================================
// GetHookEvent Tests
// ============================================================================

func TestGetHookEvent_Found(t *testing.T) {
	service, repo := newTestHookEventService()
	ctx := context.Background()

	repo.events["HEV-0001"] = &secondary.HookEventRecord{
		ID:                  "HEV-0001",
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		Decision:            "block",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	event, err := service.GetHookEvent(ctx, "HEV-0001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.ID != "HEV-0001" {
		t.Errorf("expected ID 'HEV-0001', got '%s'", event.ID)
	}
}

func TestGetHookEvent_NotFound(t *testing.T) {
	service, _ := newTestHookEventService()
	ctx := context.Background()

	_, err := service.GetHookEvent(ctx, "HEV-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error for non-existent event, got nil")
	}
}

// ============================================================================
// ListHookEvents Tests
// ============================================================================

func TestListHookEvents_Success(t *testing.T) {
	service, repo := newTestHookEventService()
	ctx := context.Background()

	repo.events["HEV-0001"] = &secondary.HookEventRecord{
		ID:                  "HEV-0001",
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		Decision:            "block",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}
	repo.events["HEV-0002"] = &secondary.HookEventRecord{
		ID:                  "HEV-0002",
		WorkbenchID:         "BENCH-001",
		HookType:            "UserPromptSubmit",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}
	repo.events["HEV-0003"] = &secondary.HookEventRecord{
		ID:                  "HEV-0003",
		WorkbenchID:         "BENCH-002",
		HookType:            "Stop",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	events, err := service.ListHookEvents(ctx, primary.HookEventFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestListHookEvents_FilterByWorkbench(t *testing.T) {
	service, repo := newTestHookEventService()
	ctx := context.Background()

	repo.events["HEV-0001"] = &secondary.HookEventRecord{
		ID:                  "HEV-0001",
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		Decision:            "block",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}
	repo.events["HEV-0002"] = &secondary.HookEventRecord{
		ID:                  "HEV-0002",
		WorkbenchID:         "BENCH-002",
		HookType:            "Stop",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	events, err := service.ListHookEvents(ctx, primary.HookEventFilters{WorkbenchID: "BENCH-001"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestListHookEvents_FilterByHookType(t *testing.T) {
	service, repo := newTestHookEventService()
	ctx := context.Background()

	repo.events["HEV-0001"] = &secondary.HookEventRecord{
		ID:                  "HEV-0001",
		WorkbenchID:         "BENCH-001",
		HookType:            "Stop",
		Decision:            "block",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}
	repo.events["HEV-0002"] = &secondary.HookEventRecord{
		ID:                  "HEV-0002",
		WorkbenchID:         "BENCH-001",
		HookType:            "UserPromptSubmit",
		Decision:            "allow",
		TaskCountIncomplete: -1,
		DurationMs:          -1,
	}

	events, err := service.ListHookEvents(ctx, primary.HookEventFilters{HookType: "Stop"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestListHookEvents_Empty(t *testing.T) {
	service, _ := newTestHookEventService()
	ctx := context.Background()

	events, err := service.ListHookEvents(ctx, primary.HookEventFilters{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if events == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}
