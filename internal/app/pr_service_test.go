package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockPRRepository implements secondary.PRRepository for testing.
type mockPRRepository struct {
	prs            map[string]*secondary.PRRecord
	prsByShipment  map[string]*secondary.PRRecord
	nextID         int
	shipmentExists map[string]bool
	shipmentStatus map[string]string
	repoExists     map[string]bool
	shipmentHasPR  map[string]bool
}

func newMockPRRepository() *mockPRRepository {
	return &mockPRRepository{
		prs:            make(map[string]*secondary.PRRecord),
		prsByShipment:  make(map[string]*secondary.PRRecord),
		nextID:         1,
		shipmentExists: make(map[string]bool),
		shipmentStatus: make(map[string]string),
		repoExists:     make(map[string]bool),
		shipmentHasPR:  make(map[string]bool),
	}
}

func (m *mockPRRepository) Create(ctx context.Context, pr *secondary.PRRecord) error {
	if pr.Status == "" {
		pr.Status = "open"
	}
	m.prs[pr.ID] = pr
	m.prsByShipment[pr.ShipmentID] = pr
	m.shipmentHasPR[pr.ShipmentID] = true
	return nil
}

func (m *mockPRRepository) GetByID(ctx context.Context, id string) (*secondary.PRRecord, error) {
	if r, ok := m.prs[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("PR %s not found", id)
}

func (m *mockPRRepository) GetByShipment(ctx context.Context, shipmentID string) (*secondary.PRRecord, error) {
	if r, ok := m.prsByShipment[shipmentID]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *mockPRRepository) List(ctx context.Context, filters secondary.PRFilters) ([]*secondary.PRRecord, error) {
	var result []*secondary.PRRecord
	for _, r := range m.prs {
		if filters.Status == "" || r.Status == filters.Status {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockPRRepository) Update(ctx context.Context, pr *secondary.PRRecord) error {
	if _, ok := m.prs[pr.ID]; !ok {
		return fmt.Errorf("PR %s not found", pr.ID)
	}
	existing := m.prs[pr.ID]
	if pr.Title != "" {
		existing.Title = pr.Title
	}
	if pr.Description != "" {
		existing.Description = pr.Description
	}
	return nil
}

func (m *mockPRRepository) Delete(ctx context.Context, id string) error {
	if r, ok := m.prs[id]; ok {
		delete(m.prsByShipment, r.ShipmentID)
		delete(m.shipmentHasPR, r.ShipmentID)
		delete(m.prs, id)
		return nil
	}
	return fmt.Errorf("PR %s not found", id)
}

func (m *mockPRRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("PR-%03d", id), nil
}

func (m *mockPRRepository) UpdateStatus(ctx context.Context, id, status string, setMerged, setClosed bool) error {
	if r, ok := m.prs[id]; ok {
		r.Status = status
		if setMerged {
			r.MergedAt = "2026-01-21T00:00:00Z"
		}
		if setClosed {
			r.ClosedAt = "2026-01-21T00:00:00Z"
		}
		return nil
	}
	return fmt.Errorf("PR %s not found", id)
}

func (m *mockPRRepository) ShipmentExists(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentExists[shipmentID], nil
}

func (m *mockPRRepository) RepoExists(ctx context.Context, repoID string) (bool, error) {
	return m.repoExists[repoID], nil
}

func (m *mockPRRepository) ShipmentHasPR(ctx context.Context, shipmentID string) (bool, error) {
	return m.shipmentHasPR[shipmentID], nil
}

func (m *mockPRRepository) GetShipmentStatus(ctx context.Context, shipmentID string) (string, error) {
	if s, ok := m.shipmentStatus[shipmentID]; ok {
		return s, nil
	}
	return "", fmt.Errorf("shipment %s not found", shipmentID)
}

// mockShipmentServiceForPR implements primary.ShipmentService for testing PR service.
type mockShipmentServiceForPR struct {
	shipments map[string]*primary.Shipment
	completed map[string]bool
}

func newMockShipmentServiceForPR() *mockShipmentServiceForPR {
	return &mockShipmentServiceForPR{
		shipments: make(map[string]*primary.Shipment),
		completed: make(map[string]bool),
	}
}

func (m *mockShipmentServiceForPR) CreateShipment(ctx context.Context, req primary.CreateShipmentRequest) (*primary.CreateShipmentResponse, error) {
	return nil, nil
}

func (m *mockShipmentServiceForPR) GetShipment(ctx context.Context, shipmentID string) (*primary.Shipment, error) {
	if s, ok := m.shipments[shipmentID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("shipment %s not found", shipmentID)
}

func (m *mockShipmentServiceForPR) ListShipments(ctx context.Context, filters primary.ShipmentFilters) ([]*primary.Shipment, error) {
	return nil, nil
}

func (m *mockShipmentServiceForPR) UpdateShipment(ctx context.Context, req primary.UpdateShipmentRequest) error {
	return nil
}

func (m *mockShipmentServiceForPR) DeleteShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) PinShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) UnpinShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) PauseShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) ResumeShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) CompleteShipment(ctx context.Context, shipmentID string, force bool) error {
	m.completed[shipmentID] = true
	return nil
}

func (m *mockShipmentServiceForPR) AssignShipmentToWorkbench(ctx context.Context, shipmentID, workbenchID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) GetShipmentsByWorkbench(ctx context.Context, workbenchID string) ([]*primary.Shipment, error) {
	return nil, nil
}

func (m *mockShipmentServiceForPR) GetShipmentTasks(ctx context.Context, shipmentID string) ([]*primary.Task, error) {
	return nil, nil
}

func (m *mockShipmentServiceForPR) ParkShipment(ctx context.Context, shipmentID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) UnparkShipment(ctx context.Context, shipmentID, conclaveID string) error {
	return nil
}

func (m *mockShipmentServiceForPR) ListShipyardQueue(ctx context.Context, commissionID string) ([]*primary.ShipyardQueueEntry, error) {
	return nil, nil
}

func (m *mockShipmentServiceForPR) SetShipmentPriority(ctx context.Context, shipmentID string, priority *int) error {
	return nil
}

func TestPRService_CreatePR(t *testing.T) {
	ctx := context.Background()

	t.Run("creates PR for active shipment", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.shipmentExists["SHIP-001"] = true
		prRepo.shipmentStatus["SHIP-001"] = "active"
		prRepo.repoExists["REPO-001"] = true

		shipmentSvc := newMockShipmentServiceForPR()
		shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
			ID:           "SHIP-001",
			CommissionID: "COMM-001",
			Status:       "active",
		}

		svc := NewPRService(prRepo, shipmentSvc)

		resp, err := svc.CreatePR(ctx, primary.CreatePRRequest{
			ShipmentID: "SHIP-001",
			RepoID:     "REPO-001",
			Title:      "Test PR",
			Branch:     "feature/test",
		})

		if err != nil {
			t.Fatalf("CreatePR failed: %v", err)
		}
		if resp.PR.Title != "Test PR" {
			t.Errorf("Title = %q, want %q", resp.PR.Title, "Test PR")
		}
		if resp.PR.Status != "open" {
			t.Errorf("Status = %q, want %q", resp.PR.Status, "open")
		}
	})

	t.Run("creates draft PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.shipmentExists["SHIP-001"] = true
		prRepo.shipmentStatus["SHIP-001"] = "active"
		prRepo.repoExists["REPO-001"] = true

		shipmentSvc := newMockShipmentServiceForPR()
		shipmentSvc.shipments["SHIP-001"] = &primary.Shipment{
			ID:           "SHIP-001",
			CommissionID: "COMM-001",
			Status:       "active",
		}

		svc := NewPRService(prRepo, shipmentSvc)

		resp, err := svc.CreatePR(ctx, primary.CreatePRRequest{
			ShipmentID: "SHIP-001",
			RepoID:     "REPO-001",
			Title:      "Draft PR",
			Branch:     "feature/draft",
			Draft:      true,
		})

		if err != nil {
			t.Fatalf("CreatePR failed: %v", err)
		}
		if resp.PR.Status != "draft" {
			t.Errorf("Status = %q, want %q", resp.PR.Status, "draft")
		}
	})

	t.Run("fails for non-existent shipment", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.shipmentExists["SHIP-001"] = false

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		_, err := svc.CreatePR(ctx, primary.CreatePRRequest{
			ShipmentID: "SHIP-001",
			RepoID:     "REPO-001",
			Title:      "Test PR",
			Branch:     "feature/test",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("fails for paused shipment", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.shipmentExists["SHIP-001"] = true
		prRepo.shipmentStatus["SHIP-001"] = "paused"
		prRepo.repoExists["REPO-001"] = true

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		_, err := svc.CreatePR(ctx, primary.CreatePRRequest{
			ShipmentID: "SHIP-001",
			RepoID:     "REPO-001",
			Title:      "Test PR",
			Branch:     "feature/test",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("fails when shipment already has PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.shipmentExists["SHIP-001"] = true
		prRepo.shipmentStatus["SHIP-001"] = "active"
		prRepo.repoExists["REPO-001"] = true
		prRepo.shipmentHasPR["SHIP-001"] = true

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		_, err := svc.CreatePR(ctx, primary.CreatePRRequest{
			ShipmentID: "SHIP-001",
			RepoID:     "REPO-001",
			Title:      "Test PR",
			Branch:     "feature/test",
		})

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestPRService_MergePR(t *testing.T) {
	ctx := context.Background()

	t.Run("merges PR and completes shipment", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:         "PR-001",
			ShipmentID: "SHIP-001",
			Status:     "open",
		}

		shipmentSvc := newMockShipmentServiceForPR()
		svc := NewPRService(prRepo, shipmentSvc)

		err := svc.MergePR(ctx, "PR-001")
		if err != nil {
			t.Fatalf("MergePR failed: %v", err)
		}

		// Verify PR is merged
		pr, _ := prRepo.GetByID(ctx, "PR-001")
		if pr.Status != "merged" {
			t.Errorf("PR Status = %q, want %q", pr.Status, "merged")
		}

		// Verify shipment was completed
		if !shipmentSvc.completed["SHIP-001"] {
			t.Error("Shipment should have been completed")
		}
	})

	t.Run("merges approved PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:         "PR-001",
			ShipmentID: "SHIP-001",
			Status:     "approved",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.MergePR(ctx, "PR-001")
		if err != nil {
			t.Fatalf("MergePR failed: %v", err)
		}

		pr, _ := prRepo.GetByID(ctx, "PR-001")
		if pr.Status != "merged" {
			t.Errorf("Status = %q, want %q", pr.Status, "merged")
		}
	})

	t.Run("fails to merge draft PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:     "PR-001",
			Status: "draft",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.MergePR(ctx, "PR-001")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestPRService_ClosePR(t *testing.T) {
	ctx := context.Background()

	t.Run("closes open PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:     "PR-001",
			Status: "open",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.ClosePR(ctx, "PR-001")
		if err != nil {
			t.Fatalf("ClosePR failed: %v", err)
		}

		pr, _ := prRepo.GetByID(ctx, "PR-001")
		if pr.Status != "closed" {
			t.Errorf("Status = %q, want %q", pr.Status, "closed")
		}
	})

	t.Run("fails to close merged PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:     "PR-001",
			Status: "merged",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.ClosePR(ctx, "PR-001")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestPRService_OpenPR(t *testing.T) {
	ctx := context.Background()

	t.Run("opens draft PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:     "PR-001",
			Status: "draft",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.OpenPR(ctx, "PR-001")
		if err != nil {
			t.Fatalf("OpenPR failed: %v", err)
		}

		pr, _ := prRepo.GetByID(ctx, "PR-001")
		if pr.Status != "open" {
			t.Errorf("Status = %q, want %q", pr.Status, "open")
		}
	})

	t.Run("fails to open already open PR", func(t *testing.T) {
		prRepo := newMockPRRepository()
		prRepo.prs["PR-001"] = &secondary.PRRecord{
			ID:     "PR-001",
			Status: "open",
		}

		svc := NewPRService(prRepo, newMockShipmentServiceForPR())

		err := svc.OpenPR(ctx, "PR-001")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
