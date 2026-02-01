package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockKennelRepository implements secondary.KennelRepository for testing.
type mockKennelRepository struct {
	kennels            map[string]*secondary.KennelRecord
	kennelsByWorkbench map[string]*secondary.KennelRecord
	workbenchExists    map[string]bool
	workbenchHasKennel map[string]bool
	nextID             int
}

func newMockKennelRepository() *mockKennelRepository {
	return &mockKennelRepository{
		kennels:            make(map[string]*secondary.KennelRecord),
		kennelsByWorkbench: make(map[string]*secondary.KennelRecord),
		workbenchExists:    make(map[string]bool),
		workbenchHasKennel: make(map[string]bool),
		nextID:             1,
	}
}

func (m *mockKennelRepository) Create(ctx context.Context, kennel *secondary.KennelRecord) error {
	m.kennels[kennel.ID] = kennel
	m.kennelsByWorkbench[kennel.WorkbenchID] = kennel
	m.workbenchHasKennel[kennel.WorkbenchID] = true
	return nil
}

func (m *mockKennelRepository) GetByID(ctx context.Context, id string) (*secondary.KennelRecord, error) {
	if k, ok := m.kennels[id]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}

func (m *mockKennelRepository) GetByWorkbench(ctx context.Context, workbenchID string) (*secondary.KennelRecord, error) {
	if k, ok := m.kennelsByWorkbench[workbenchID]; ok {
		return k, nil
	}
	return nil, errors.New("not found")
}

func (m *mockKennelRepository) List(ctx context.Context, filters secondary.KennelFilters) ([]*secondary.KennelRecord, error) {
	var result []*secondary.KennelRecord
	for _, k := range m.kennels {
		if filters.WorkbenchID != "" && k.WorkbenchID != filters.WorkbenchID {
			continue
		}
		if filters.Status != "" && k.Status != filters.Status {
			continue
		}
		result = append(result, k)
	}
	return result, nil
}

func (m *mockKennelRepository) Update(ctx context.Context, kennel *secondary.KennelRecord) error {
	if _, ok := m.kennels[kennel.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockKennelRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.kennels[id]; !ok {
		return errors.New("not found")
	}
	k := m.kennels[id]
	delete(m.kennelsByWorkbench, k.WorkbenchID)
	delete(m.kennels, id)
	m.workbenchHasKennel[k.WorkbenchID] = false
	return nil
}

func (m *mockKennelRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("KENNEL-%03d", id), nil
}

func (m *mockKennelRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if k, ok := m.kennels[id]; ok {
		k.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockKennelRepository) WorkbenchExists(ctx context.Context, workbenchID string) (bool, error) {
	return m.workbenchExists[workbenchID], nil
}

func (m *mockKennelRepository) WorkbenchHasKennel(ctx context.Context, workbenchID string) (bool, error) {
	return m.workbenchHasKennel[workbenchID], nil
}

func newTestKennelService() (*KennelServiceImpl, *mockKennelRepository) {
	repo := newMockKennelRepository()
	service := NewKennelService(repo)
	return service, repo
}

func TestKennelService_GetKennel(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	}

	kennel, err := service.GetKennel(ctx, "KENNEL-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if kennel.WorkbenchID != "BENCH-001" {
		t.Errorf("expected workbenchID 'BENCH-001', got %q", kennel.WorkbenchID)
	}
}

func TestKennelService_GetKennel_NotFound(t *testing.T) {
	service, _ := newTestKennelService()
	ctx := context.Background()

	_, err := service.GetKennel(ctx, "KENNEL-999")
	if err == nil {
		t.Error("expected error for non-existent kennel")
	}
}

func TestKennelService_GetKennelByWorkbench(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	}
	repo.kennelsByWorkbench["BENCH-001"] = repo.kennels["KENNEL-001"]

	kennel, err := service.GetKennelByWorkbench(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if kennel.ID != "KENNEL-001" {
		t.Errorf("expected ID 'KENNEL-001', got %q", kennel.ID)
	}
}

func TestKennelService_ListKennels(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.kennels["KENNEL-001"] = &secondary.KennelRecord{ID: "KENNEL-001", WorkbenchID: "BENCH-001", Status: "vacant"}
	repo.kennels["KENNEL-002"] = &secondary.KennelRecord{ID: "KENNEL-002", WorkbenchID: "BENCH-002", Status: "occupied"}

	kennels, err := service.ListKennels(ctx, primary.KennelFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(kennels) != 2 {
		t.Errorf("expected 2 kennels, got %d", len(kennels))
	}
}

func TestKennelService_CreateKennel(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	// Setup: workbench exists
	repo.workbenchExists["BENCH-001"] = true

	kennel, err := service.CreateKennel(ctx, "BENCH-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if kennel.ID != "KENNEL-001" {
		t.Errorf("expected ID 'KENNEL-001', got %q", kennel.ID)
	}
	if kennel.Status != "vacant" {
		t.Errorf("expected status 'vacant', got %q", kennel.Status)
	}
}

func TestKennelService_CreateKennel_WorkbenchNotFound(t *testing.T) {
	service, _ := newTestKennelService()
	ctx := context.Background()

	_, err := service.CreateKennel(ctx, "BENCH-999")
	if err == nil {
		t.Error("expected error for non-existent workbench")
	}
}

func TestKennelService_CreateKennel_AlreadyExists(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.workbenchExists["BENCH-001"] = true
	repo.workbenchHasKennel["BENCH-001"] = true

	_, err := service.CreateKennel(ctx, "BENCH-001")
	if err == nil {
		t.Error("expected error for existing kennel")
	}
}

func TestKennelService_UpdateKennelStatus(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	}

	err := service.UpdateKennelStatus(ctx, "KENNEL-001", "occupied")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.kennels["KENNEL-001"].Status != "occupied" {
		t.Errorf("expected status 'occupied', got %q", repo.kennels["KENNEL-001"].Status)
	}
}

func TestKennelService_UpdateKennelStatus_InvalidStatus(t *testing.T) {
	service, repo := newTestKennelService()
	ctx := context.Background()

	repo.kennels["KENNEL-001"] = &secondary.KennelRecord{
		ID:          "KENNEL-001",
		WorkbenchID: "BENCH-001",
		Status:      "vacant",
	}

	err := service.UpdateKennelStatus(ctx, "KENNEL-001", "invalid")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}
