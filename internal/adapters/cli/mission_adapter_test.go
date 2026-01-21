package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/orc/internal/ports/primary"
)

// mockMissionService implements primary.MissionService for testing
type mockMissionService struct {
	createMissionFn   func(ctx context.Context, req primary.CreateMissionRequest) (*primary.CreateMissionResponse, error)
	listMissionsFn    func(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error)
	getMissionFn      func(ctx context.Context, missionID string) (*primary.Mission, error)
	updateMissionFn   func(ctx context.Context, req primary.UpdateMissionRequest) error
	completeMissionFn func(ctx context.Context, missionID string) error
	archiveMissionFn  func(ctx context.Context, missionID string) error
	deleteMissionFn   func(ctx context.Context, req primary.DeleteMissionRequest) error
	pinMissionFn      func(ctx context.Context, missionID string) error
	unpinMissionFn    func(ctx context.Context, missionID string) error

	// Track calls for verification
	lastCreateReq primary.CreateMissionRequest
	lastUpdateReq primary.UpdateMissionRequest
	lastDeleteReq primary.DeleteMissionRequest
}

func (m *mockMissionService) CreateMission(ctx context.Context, req primary.CreateMissionRequest) (*primary.CreateMissionResponse, error) {
	m.lastCreateReq = req
	if m.createMissionFn != nil {
		return m.createMissionFn(ctx, req)
	}
	return &primary.CreateMissionResponse{
		MissionID: "MISSION-001",
		Mission:   &primary.Mission{ID: "MISSION-001", Title: req.Title},
	}, nil
}

func (m *mockMissionService) ListMissions(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
	if m.listMissionsFn != nil {
		return m.listMissionsFn(ctx, filters)
	}
	return []*primary.Mission{}, nil
}

func (m *mockMissionService) GetMission(ctx context.Context, missionID string) (*primary.Mission, error) {
	if m.getMissionFn != nil {
		return m.getMissionFn(ctx, missionID)
	}
	return &primary.Mission{ID: missionID, Title: "Test Mission", Status: "active"}, nil
}

func (m *mockMissionService) StartMission(ctx context.Context, req primary.StartMissionRequest) (*primary.StartMissionResponse, error) {
	return nil, errors.New("not implemented in adapter")
}

func (m *mockMissionService) LaunchMission(ctx context.Context, req primary.LaunchMissionRequest) (*primary.LaunchMissionResponse, error) {
	return nil, errors.New("not implemented in adapter")
}

func (m *mockMissionService) UpdateMission(ctx context.Context, req primary.UpdateMissionRequest) error {
	m.lastUpdateReq = req
	if m.updateMissionFn != nil {
		return m.updateMissionFn(ctx, req)
	}
	return nil
}

func (m *mockMissionService) CompleteMission(ctx context.Context, missionID string) error {
	if m.completeMissionFn != nil {
		return m.completeMissionFn(ctx, missionID)
	}
	return nil
}

func (m *mockMissionService) ArchiveMission(ctx context.Context, missionID string) error {
	if m.archiveMissionFn != nil {
		return m.archiveMissionFn(ctx, missionID)
	}
	return nil
}

func (m *mockMissionService) DeleteMission(ctx context.Context, req primary.DeleteMissionRequest) error {
	m.lastDeleteReq = req
	if m.deleteMissionFn != nil {
		return m.deleteMissionFn(ctx, req)
	}
	return nil
}

func (m *mockMissionService) PinMission(ctx context.Context, missionID string) error {
	if m.pinMissionFn != nil {
		return m.pinMissionFn(ctx, missionID)
	}
	return nil
}

func (m *mockMissionService) UnpinMission(ctx context.Context, missionID string) error {
	if m.unpinMissionFn != nil {
		return m.unpinMissionFn(ctx, missionID)
	}
	return nil
}

// ============================================================================
// Create Tests
// ============================================================================

func TestMissionAdapter_Create_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Create(context.Background(), "Test Mission", "A description")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastCreateReq.Title != "Test Mission" {
		t.Errorf("expected title 'Test Mission', got '%s'", mock.lastCreateReq.Title)
	}
	if mock.lastCreateReq.Description != "A description" {
		t.Errorf("expected description 'A description', got '%s'", mock.lastCreateReq.Description)
	}
	if !strings.Contains(buf.String(), "Created mission MISSION-001") {
		t.Errorf("expected output to contain 'Created mission MISSION-001', got '%s'", buf.String())
	}
}

func TestMissionAdapter_Create_ServiceError(t *testing.T) {
	mock := &mockMissionService{
		createMissionFn: func(ctx context.Context, req primary.CreateMissionRequest) (*primary.CreateMissionResponse, error) {
			return nil, errors.New("IMPs cannot create missions")
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Create(context.Background(), "Test", "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "IMPs cannot create missions") {
		t.Errorf("expected error to contain guard message, got '%s'", err.Error())
	}
}

// ============================================================================
// List Tests
// ============================================================================

func TestMissionAdapter_List_WithResults(t *testing.T) {
	mock := &mockMissionService{
		listMissionsFn: func(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
			return []*primary.Mission{
				{ID: "MISSION-001", Title: "First", Status: "active"},
				{ID: "MISSION-002", Title: "Second", Status: "complete"},
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.List(context.Background(), "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "MISSION-001") {
		t.Errorf("expected output to contain 'MISSION-001', got '%s'", output)
	}
	if !strings.Contains(output, "MISSION-002") {
		t.Errorf("expected output to contain 'MISSION-002', got '%s'", output)
	}
}

func TestMissionAdapter_List_Empty(t *testing.T) {
	mock := &mockMissionService{
		listMissionsFn: func(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
			return []*primary.Mission{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.List(context.Background(), "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "No missions found") {
		t.Errorf("expected 'No missions found', got '%s'", buf.String())
	}
}

func TestMissionAdapter_List_FilterByStatus(t *testing.T) {
	var capturedStatus string
	mock := &mockMissionService{
		listMissionsFn: func(ctx context.Context, filters primary.MissionFilters) ([]*primary.Mission, error) {
			capturedStatus = filters.Status
			return []*primary.Mission{}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	_ = adapter.List(context.Background(), "active")

	if capturedStatus != "active" {
		t.Errorf("expected status filter 'active', got '%s'", capturedStatus)
	}
}

// ============================================================================
// Show Tests
// ============================================================================

func TestMissionAdapter_Show_Success(t *testing.T) {
	mock := &mockMissionService{
		getMissionFn: func(ctx context.Context, missionID string) (*primary.Mission, error) {
			return &primary.Mission{
				ID:          missionID,
				Title:       "Test Mission",
				Description: "Detailed description",
				Status:      "active",
				CreatedAt:   "2026-01-19",
			}, nil
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	mission, err := adapter.Show(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mission.ID != "MISSION-001" {
		t.Errorf("expected mission ID 'MISSION-001', got '%s'", mission.ID)
	}
	output := buf.String()
	if !strings.Contains(output, "Test Mission") {
		t.Errorf("expected output to contain title, got '%s'", output)
	}
	if !strings.Contains(output, "Detailed description") {
		t.Errorf("expected output to contain description, got '%s'", output)
	}
}

func TestMissionAdapter_Show_NotFound(t *testing.T) {
	mock := &mockMissionService{
		getMissionFn: func(ctx context.Context, missionID string) (*primary.Mission, error) {
			return nil, errors.New("mission not found")
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	_, err := adapter.Show(context.Background(), "MISSION-NONEXISTENT")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// Update Tests
// ============================================================================

func TestMissionAdapter_Update_TitleOnly(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Update(context.Background(), "MISSION-001", "New Title", "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastUpdateReq.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", mock.lastUpdateReq.Title)
	}
}

func TestMissionAdapter_Update_NoFieldsError(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Update(context.Background(), "MISSION-001", "", "")

	if err == nil {
		t.Fatal("expected error when no fields specified, got nil")
	}
	if !strings.Contains(err.Error(), "must specify") {
		t.Errorf("expected validation error, got '%s'", err.Error())
	}
}

// ============================================================================
// Complete Tests
// ============================================================================

func TestMissionAdapter_Complete_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Complete(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "marked as complete") {
		t.Errorf("expected completion message, got '%s'", buf.String())
	}
}

func TestMissionAdapter_Complete_GuardError(t *testing.T) {
	mock := &mockMissionService{
		completeMissionFn: func(ctx context.Context, missionID string) error {
			return errors.New("cannot complete pinned mission")
		},
	}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Complete(context.Background(), "MISSION-001")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "pinned") {
		t.Errorf("expected guard error, got '%s'", err.Error())
	}
}

// ============================================================================
// Archive Tests
// ============================================================================

func TestMissionAdapter_Archive_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Archive(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "archived") {
		t.Errorf("expected archive message, got '%s'", buf.String())
	}
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestMissionAdapter_Delete_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Delete(context.Background(), "MISSION-001", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "Deleted mission") {
		t.Errorf("expected delete message, got '%s'", buf.String())
	}
}

func TestMissionAdapter_Delete_WithForce(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	_ = adapter.Delete(context.Background(), "MISSION-001", true)

	if !mock.lastDeleteReq.Force {
		t.Error("expected force flag to be true")
	}
}

// ============================================================================
// Pin/Unpin Tests
// ============================================================================

func TestMissionAdapter_Pin_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Pin(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "pinned") {
		t.Errorf("expected pin message, got '%s'", buf.String())
	}
}

func TestMissionAdapter_Unpin_Success(t *testing.T) {
	mock := &mockMissionService{}
	var buf bytes.Buffer
	adapter := NewMissionAdapter(mock, &buf)

	err := adapter.Unpin(context.Background(), "MISSION-001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "unpinned") {
		t.Errorf("expected unpin message, got '%s'", buf.String())
	}
}
