package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/orc/internal/ports/primary"
)

// mockGroveService implements primary.GroveService for testing
type mockGroveService struct {
	createGroveFn func(ctx context.Context, req primary.CreateGroveRequest) (*primary.CreateGroveResponse, error)
	openGroveFn   func(ctx context.Context, req primary.OpenGroveRequest) (*primary.OpenGroveResponse, error)
	getGroveFn    func(ctx context.Context, groveID string) (*primary.Grove, error)
	listGrovesFn  func(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error)
	renameGroveFn func(ctx context.Context, req primary.RenameGroveRequest) error
	deleteGroveFn func(ctx context.Context, req primary.DeleteGroveRequest) error

	// Track calls for verification
	lastRenameReq primary.RenameGroveRequest
	lastDeleteReq primary.DeleteGroveRequest
}

func (m *mockGroveService) CreateGrove(ctx context.Context, req primary.CreateGroveRequest) (*primary.CreateGroveResponse, error) {
	if m.createGroveFn != nil {
		return m.createGroveFn(ctx, req)
	}
	return &primary.CreateGroveResponse{
		GroveID: "GROVE-001",
		Grove:   &primary.Grove{ID: "GROVE-001", Name: req.Name},
		Path:    "/tmp/groves/" + req.Name,
	}, nil
}

func (m *mockGroveService) OpenGrove(ctx context.Context, req primary.OpenGroveRequest) (*primary.OpenGroveResponse, error) {
	if m.openGroveFn != nil {
		return m.openGroveFn(ctx, req)
	}
	return nil, errors.New("not implemented in adapter")
}

func (m *mockGroveService) GetGrove(ctx context.Context, groveID string) (*primary.Grove, error) {
	if m.getGroveFn != nil {
		return m.getGroveFn(ctx, groveID)
	}
	return &primary.Grove{
		ID:        groveID,
		Name:      "test-grove",
		MissionID: "MISSION-001",
		Path:      "/tmp/groves/test-grove",
		Status:    "active",
		CreatedAt: "2026-01-19",
	}, nil
}

func (m *mockGroveService) ListGroves(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
	if m.listGrovesFn != nil {
		return m.listGrovesFn(ctx, filters)
	}
	return []*primary.Grove{}, nil
}

func (m *mockGroveService) RenameGrove(ctx context.Context, req primary.RenameGroveRequest) error {
	m.lastRenameReq = req
	if m.renameGroveFn != nil {
		return m.renameGroveFn(ctx, req)
	}
	return nil
}

func (m *mockGroveService) DeleteGrove(ctx context.Context, req primary.DeleteGroveRequest) error {
	m.lastDeleteReq = req
	if m.deleteGroveFn != nil {
		return m.deleteGroveFn(ctx, req)
	}
	return nil
}

func (m *mockGroveService) GetGroveByPath(ctx context.Context, path string) (*primary.Grove, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *mockGroveService) UpdateGrovePath(ctx context.Context, groveID, newPath string) error {
	return nil
}

// ============================================================================
// List Tests
// ============================================================================

func TestGroveAdapter_List_WithResults(t *testing.T) {
	mock := &mockGroveService{
		listGrovesFn: func(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
			return []*primary.Grove{
				{ID: "GROVE-001", Name: "backend", MissionID: "MISSION-001", Status: "active", Path: "/tmp/backend"},
				{ID: "GROVE-002", Name: "frontend", MissionID: "MISSION-001", Status: "active", Path: "/tmp/frontend"},
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	groves, err := adapter.List(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(groves) != 2 {
		t.Errorf("expected 2 groves, got %d", len(groves))
	}
	output := buf.String()
	if !strings.Contains(output, "GROVE-001") {
		t.Errorf("expected output to contain 'GROVE-001', got '%s'", output)
	}
	if !strings.Contains(output, "backend") {
		t.Errorf("expected output to contain 'backend', got '%s'", output)
	}
}

func TestGroveAdapter_List_Empty(t *testing.T) {
	mock := &mockGroveService{
		listGrovesFn: func(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
			return []*primary.Grove{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	groves, err := adapter.List(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(groves) != 0 {
		t.Errorf("expected 0 groves, got %d", len(groves))
	}
	if !strings.Contains(buf.String(), "No groves found") {
		t.Errorf("expected 'No groves found', got '%s'", buf.String())
	}
}

func TestGroveAdapter_List_FilterByMission(t *testing.T) {
	var capturedMissionID string
	mock := &mockGroveService{
		listGrovesFn: func(ctx context.Context, filters primary.GroveFilters) ([]*primary.Grove, error) {
			capturedMissionID = filters.MissionID
			return []*primary.Grove{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, _ = adapter.List(context.Background(), "MISSION-002")

	if capturedMissionID != "MISSION-002" {
		t.Errorf("expected mission filter 'MISSION-002', got '%s'", capturedMissionID)
	}
}

// ============================================================================
// Show Tests
// ============================================================================

func TestGroveAdapter_Show_Success(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return &primary.Grove{
				ID:        groveID,
				Name:      "auth-backend",
				MissionID: "MISSION-001",
				Path:      "/home/user/src/worktrees/auth-backend",
				Status:    "active",
				CreatedAt: "2026-01-19",
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	grove, err := adapter.Show(context.Background(), "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grove.Name != "auth-backend" {
		t.Errorf("expected name 'auth-backend', got '%s'", grove.Name)
	}
	output := buf.String()
	if !strings.Contains(output, "auth-backend") {
		t.Errorf("expected output to contain name, got '%s'", output)
	}
	if !strings.Contains(output, "MISSION-001") {
		t.Errorf("expected output to contain mission, got '%s'", output)
	}
}

func TestGroveAdapter_Show_NotFound(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return nil, errors.New("grove not found")
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, err := adapter.Show(context.Background(), "GROVE-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// Rename Tests
// ============================================================================

func TestGroveAdapter_Rename_Success(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return &primary.Grove{
				ID:   groveID,
				Name: "old-name",
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	oldName, err := adapter.Rename(context.Background(), "GROVE-001", "new-name")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oldName != "old-name" {
		t.Errorf("expected old name 'old-name', got '%s'", oldName)
	}
	if mock.lastRenameReq.NewName != "new-name" {
		t.Errorf("expected new name 'new-name', got '%s'", mock.lastRenameReq.NewName)
	}
	output := buf.String()
	if !strings.Contains(output, "renamed") {
		t.Errorf("expected rename message, got '%s'", output)
	}
	if !strings.Contains(output, "old-name → new-name") {
		t.Errorf("expected name transition in output, got '%s'", output)
	}
}

func TestGroveAdapter_Rename_NotFound(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return nil, errors.New("grove not found")
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, err := adapter.Rename(context.Background(), "GROVE-NONEXISTENT", "new-name")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroveAdapter_Rename_ServiceError(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return &primary.Grove{ID: groveID, Name: "old-name"}, nil
		},
		renameGroveFn: func(ctx context.Context, req primary.RenameGroveRequest) error {
			return errors.New("rename failed")
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, err := adapter.Rename(context.Background(), "GROVE-001", "new-name")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestGroveAdapter_Delete_Success(t *testing.T) {
	mock := &mockGroveService{}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	grove, err := adapter.Delete(context.Background(), "GROVE-001", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grove == nil {
		t.Fatal("expected grove to be returned")
	}
	if !strings.Contains(buf.String(), "Deleted grove") {
		t.Errorf("expected delete message, got '%s'", buf.String())
	}
}

func TestGroveAdapter_Delete_WithForce(t *testing.T) {
	mock := &mockGroveService{}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, _ = adapter.Delete(context.Background(), "GROVE-001", true)

	if !mock.lastDeleteReq.Force {
		t.Error("expected force flag to be true")
	}
}

func TestGroveAdapter_Delete_NotFound(t *testing.T) {
	mock := &mockGroveService{
		getGroveFn: func(ctx context.Context, groveID string) (*primary.Grove, error) {
			return nil, errors.New("grove not found")
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, err := adapter.Delete(context.Background(), "GROVE-NONEXISTENT", false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGroveAdapter_Delete_GuardError(t *testing.T) {
	mock := &mockGroveService{
		deleteGroveFn: func(ctx context.Context, req primary.DeleteGroveRequest) error {
			return errors.New("grove has active tasks")
		},
	}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	_, err := adapter.Delete(context.Background(), "GROVE-001", false)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "active tasks") {
		t.Errorf("expected guard error, got '%s'", err.Error())
	}
}

// ============================================================================
// GetGrove Tests
// ============================================================================

func TestGroveAdapter_GetGrove_Success(t *testing.T) {
	mock := &mockGroveService{}
	var buf bytes.Buffer
	adapter := NewGroveAdapter(mock, &buf)

	grove, err := adapter.GetGrove(context.Background(), "GROVE-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if grove.ID != "GROVE-001" {
		t.Errorf("expected grove ID 'GROVE-001', got '%s'", grove.ID)
	}
	// GetGrove should not produce output
	if buf.String() != "" {
		t.Errorf("expected no output, got '%s'", buf.String())
	}
}
