package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockEscalationRepository implements secondary.EscalationRepository for testing.
type mockEscalationRepository struct {
	escalations    map[string]*secondary.EscalationRecord
	planExists     map[string]bool
	taskExists     map[string]bool
	approvalExists map[string]bool
	nextID         int
}

func newMockEscalationRepository() *mockEscalationRepository {
	return &mockEscalationRepository{
		escalations:    make(map[string]*secondary.EscalationRecord),
		planExists:     make(map[string]bool),
		taskExists:     make(map[string]bool),
		approvalExists: make(map[string]bool),
		nextID:         1,
	}
}

func (m *mockEscalationRepository) Create(ctx context.Context, escalation *secondary.EscalationRecord) error {
	m.escalations[escalation.ID] = escalation
	return nil
}

func (m *mockEscalationRepository) GetByID(ctx context.Context, id string) (*secondary.EscalationRecord, error) {
	if e, ok := m.escalations[id]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *mockEscalationRepository) List(ctx context.Context, filters secondary.EscalationFilters) ([]*secondary.EscalationRecord, error) {
	var result []*secondary.EscalationRecord
	for _, e := range m.escalations {
		if filters.TaskID != "" && e.TaskID != filters.TaskID {
			continue
		}
		if filters.Status != "" && e.Status != filters.Status {
			continue
		}
		if filters.TargetActorID != "" && e.TargetActorID != filters.TargetActorID {
			continue
		}
		result = append(result, e)
	}
	return result, nil
}

func (m *mockEscalationRepository) Update(ctx context.Context, escalation *secondary.EscalationRecord) error {
	if _, ok := m.escalations[escalation.ID]; !ok {
		return errors.New("not found")
	}
	return nil
}

func (m *mockEscalationRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.escalations[id]; !ok {
		return errors.New("not found")
	}
	delete(m.escalations, id)
	return nil
}

func (m *mockEscalationRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("ESC-%03d", id), nil
}

func (m *mockEscalationRepository) UpdateStatus(ctx context.Context, id, status string, setResolved bool) error {
	if e, ok := m.escalations[id]; ok {
		e.Status = status
		return nil
	}
	return errors.New("not found")
}

func (m *mockEscalationRepository) Resolve(ctx context.Context, id, resolution, resolvedBy string) error {
	if e, ok := m.escalations[id]; ok {
		e.Status = "resolved"
		e.Resolution = resolution
		e.ResolvedBy = resolvedBy
		return nil
	}
	return errors.New("not found")
}

func (m *mockEscalationRepository) PlanExists(ctx context.Context, planID string) (bool, error) {
	return m.planExists[planID], nil
}

func (m *mockEscalationRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	return m.taskExists[taskID], nil
}

func (m *mockEscalationRepository) ApprovalExists(ctx context.Context, approvalID string) (bool, error) {
	return m.approvalExists[approvalID], nil
}

func newTestEscalationService() (*EscalationServiceImpl, *mockEscalationRepository) {
	repo := newMockEscalationRepository()
	service := NewEscalationService(repo)
	return service, repo
}

func TestEscalationService_GetEscalation(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Complexity exceeded",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "IMP-BENCH-001",
	}

	escalation, err := service.GetEscalation(ctx, "ESC-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if escalation.PlanID != "PLAN-001" {
		t.Errorf("expected planID 'PLAN-001', got %q", escalation.PlanID)
	}
}

func TestEscalationService_GetEscalation_NotFound(t *testing.T) {
	service, _ := newTestEscalationService()
	ctx := context.Background()

	_, err := service.GetEscalation(ctx, "ESC-999")
	if err == nil {
		t.Error("expected error for non-existent escalation")
	}
}

func TestEscalationService_ListEscalations(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{ID: "ESC-001", PlanID: "PLAN-001", TaskID: "TASK-001", Reason: "Test 1", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001"}
	repo.escalations["ESC-002"] = &secondary.EscalationRecord{ID: "ESC-002", PlanID: "PLAN-002", TaskID: "TASK-002", Reason: "Test 2", Status: "resolved", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-002"}

	escalations, err := service.ListEscalations(ctx, primary.EscalationFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(escalations) != 2 {
		t.Errorf("expected 2 escalations, got %d", len(escalations))
	}
}

func TestEscalationService_ListEscalations_FilterByStatus(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{ID: "ESC-001", PlanID: "PLAN-001", TaskID: "TASK-001", Reason: "Test 1", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001"}
	repo.escalations["ESC-002"] = &secondary.EscalationRecord{ID: "ESC-002", PlanID: "PLAN-002", TaskID: "TASK-002", Reason: "Test 2", Status: "resolved", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-002"}

	escalations, err := service.ListEscalations(ctx, primary.EscalationFilters{Status: "pending"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(escalations) != 1 {
		t.Errorf("expected 1 escalation, got %d", len(escalations))
	}
}

func TestEscalationService_ListEscalations_FilterByTaskID(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{ID: "ESC-001", PlanID: "PLAN-001", TaskID: "TASK-001", Reason: "Test 1", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-001"}
	repo.escalations["ESC-002"] = &secondary.EscalationRecord{ID: "ESC-002", PlanID: "PLAN-002", TaskID: "TASK-002", Reason: "Test 2", Status: "pending", RoutingRule: "workshop_gatehouse", OriginActorID: "IMP-BENCH-002"}

	escalations, err := service.ListEscalations(ctx, primary.EscalationFilters{TaskID: "TASK-001"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(escalations) != 1 {
		t.Errorf("expected 1 escalation, got %d", len(escalations))
	}
	if escalations[0].TaskID != "TASK-001" {
		t.Errorf("expected TaskID 'TASK-001', got %q", escalations[0].TaskID)
	}
}

func TestEscalationService_ResolveEscalation_Approved(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Complexity exceeded",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "BENCH-001",
	}

	err := service.ResolveEscalation(ctx, primary.ResolveEscalationRequest{
		EscalationID: "ESC-001",
		Outcome:      "approved",
		Resolution:   "Looks good, proceed",
		ResolvedBy:   "GATE-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.escalations["ESC-001"].Status != "resolved" {
		t.Errorf("expected status 'resolved', got %q", repo.escalations["ESC-001"].Status)
	}
	if repo.escalations["ESC-001"].Resolution != "Looks good, proceed" {
		t.Errorf("expected resolution to be set")
	}
}

func TestEscalationService_ResolveEscalation_Rejected(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Uncertain approach",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "BENCH-001",
	}

	err := service.ResolveEscalation(ctx, primary.ResolveEscalationRequest{
		EscalationID: "ESC-001",
		Outcome:      "rejected",
		Resolution:   "Please reconsider the approach",
		ResolvedBy:   "GATE-001",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.escalations["ESC-001"].Status != "dismissed" {
		t.Errorf("expected status 'dismissed', got %q", repo.escalations["ESC-001"].Status)
	}
}

func TestEscalationService_ResolveEscalation_NotPending(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Already resolved",
		Status:        "resolved",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "BENCH-001",
	}

	err := service.ResolveEscalation(ctx, primary.ResolveEscalationRequest{
		EscalationID: "ESC-001",
		Outcome:      "approved",
	})

	if err == nil {
		t.Fatal("expected error for non-pending escalation")
	}
}

func TestEscalationService_ResolveEscalation_InvalidOutcome(t *testing.T) {
	service, repo := newTestEscalationService()
	ctx := context.Background()

	repo.escalations["ESC-001"] = &secondary.EscalationRecord{
		ID:            "ESC-001",
		PlanID:        "PLAN-001",
		TaskID:        "TASK-001",
		Reason:        "Test",
		Status:        "pending",
		RoutingRule:   "workshop_gatehouse",
		OriginActorID: "BENCH-001",
	}

	err := service.ResolveEscalation(ctx, primary.ResolveEscalationRequest{
		EscalationID: "ESC-001",
		Outcome:      "invalid",
	})

	if err == nil {
		t.Fatal("expected error for invalid outcome")
	}
}

func TestEscalationService_ResolveEscalation_NotFound(t *testing.T) {
	service, _ := newTestEscalationService()
	ctx := context.Background()

	err := service.ResolveEscalation(ctx, primary.ResolveEscalationRequest{
		EscalationID: "ESC-999",
		Outcome:      "approved",
	})

	if err == nil {
		t.Fatal("expected error for non-existent escalation")
	}
}
