package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockWorkshopLogRepository implements secondary.WorkshopLogRepository for testing.
type mockWorkshopLogRepository struct {
	logs           map[string]*secondary.WorkshopLogRecord
	workshopExists map[string]bool
	nextID         int
}

func newMockWorkshopLogRepository() *mockWorkshopLogRepository {
	return &mockWorkshopLogRepository{
		logs:           make(map[string]*secondary.WorkshopLogRecord),
		workshopExists: make(map[string]bool),
		nextID:         1,
	}
}

func (m *mockWorkshopLogRepository) Create(ctx context.Context, log *secondary.WorkshopLogRecord) error {
	m.logs[log.ID] = log
	return nil
}

func (m *mockWorkshopLogRepository) GetByID(ctx context.Context, id string) (*secondary.WorkshopLogRecord, error) {
	if l, ok := m.logs[id]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}

func (m *mockWorkshopLogRepository) List(ctx context.Context, filters secondary.WorkshopLogFilters) ([]*secondary.WorkshopLogRecord, error) {
	var result []*secondary.WorkshopLogRecord
	for _, l := range m.logs {
		if filters.WorkshopID != "" && l.WorkshopID != filters.WorkshopID {
			continue
		}
		if filters.EntityType != "" && l.EntityType != filters.EntityType {
			continue
		}
		if filters.EntityID != "" && l.EntityID != filters.EntityID {
			continue
		}
		if filters.ActorID != "" && l.ActorID != filters.ActorID {
			continue
		}
		if filters.Action != "" && l.Action != filters.Action {
			continue
		}
		result = append(result, l)
	}

	// Apply limit
	if filters.Limit > 0 && len(result) > filters.Limit {
		result = result[:filters.Limit]
	}

	return result, nil
}

func (m *mockWorkshopLogRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("WL-%04d", id), nil
}

func (m *mockWorkshopLogRepository) WorkshopExists(ctx context.Context, workshopID string) (bool, error) {
	return m.workshopExists[workshopID], nil
}

func (m *mockWorkshopLogRepository) PruneOlderThan(ctx context.Context, days int) (int, error) {
	// For testing, just pretend to prune
	count := 0
	cutoff := time.Now().AddDate(0, 0, -days)
	for id, log := range m.logs {
		ts, err := time.Parse(time.RFC3339, log.Timestamp)
		if err != nil {
			continue
		}
		if ts.Before(cutoff) {
			delete(m.logs, id)
			count++
		}
	}
	return count, nil
}

func newTestLogService() (*LogServiceImpl, *mockWorkshopLogRepository) {
	repo := newMockWorkshopLogRepository()
	service := NewLogService(repo)
	return service, repo
}

func TestLogService_GetLog(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.logs["WL-0001"] = &secondary.WorkshopLogRecord{
		ID:         "WL-0001",
		WorkshopID: "SHOP-001",
		Timestamp:  "2024-01-01T12:00:00Z",
		ActorID:    "IMP-BENCH-001",
		EntityType: "task",
		EntityID:   "TASK-001",
		Action:     "create",
	}

	log, err := service.GetLog(ctx, "WL-0001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log.EntityID != "TASK-001" {
		t.Errorf("expected entityID 'TASK-001', got %q", log.EntityID)
	}
}

func TestLogService_GetLog_NotFound(t *testing.T) {
	service, _ := newTestLogService()
	ctx := context.Background()

	_, err := service.GetLog(ctx, "WL-9999")
	if err == nil {
		t.Error("expected error for non-existent log")
	}
}

func TestLogService_ListLogs(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.logs["WL-0001"] = &secondary.WorkshopLogRecord{ID: "WL-0001", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-001", Action: "create", Timestamp: "2024-01-01T12:00:00Z"}
	repo.logs["WL-0002"] = &secondary.WorkshopLogRecord{ID: "WL-0002", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-002", Action: "create", Timestamp: "2024-01-01T12:01:00Z"}

	logs, err := service.ListLogs(ctx, primary.LogFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}

func TestLogService_ListLogs_WithFilters(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	repo.logs["WL-0001"] = &secondary.WorkshopLogRecord{ID: "WL-0001", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-001", Action: "create", ActorID: "IMP-BENCH-001", Timestamp: "2024-01-01T12:00:00Z"}
	repo.logs["WL-0002"] = &secondary.WorkshopLogRecord{ID: "WL-0002", WorkshopID: "SHOP-001", EntityType: "task", EntityID: "TASK-002", Action: "create", ActorID: "IMP-BENCH-002", Timestamp: "2024-01-01T12:01:00Z"}

	logs, err := service.ListLogs(ctx, primary.LogFilters{ActorID: "IMP-BENCH-001"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
	if logs[0].ActorID != "IMP-BENCH-001" {
		t.Errorf("expected actorID 'IMP-BENCH-001', got %q", logs[0].ActorID)
	}
}

func TestLogService_PruneLogs(t *testing.T) {
	service, repo := newTestLogService()
	ctx := context.Background()

	// Add old and new logs
	oldTime := time.Now().AddDate(0, 0, -60).Format(time.RFC3339) // 60 days old
	newTime := time.Now().Format(time.RFC3339)

	repo.logs["WL-0001"] = &secondary.WorkshopLogRecord{ID: "WL-0001", WorkshopID: "SHOP-001", EntityType: "task", Timestamp: oldTime}
	repo.logs["WL-0002"] = &secondary.WorkshopLogRecord{ID: "WL-0002", WorkshopID: "SHOP-001", EntityType: "task", Timestamp: newTime}

	count, err := service.PruneLogs(ctx, 30)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pruned, got %d", count)
	}
	if len(repo.logs) != 1 {
		t.Errorf("expected 1 log remaining, got %d", len(repo.logs))
	}
}
