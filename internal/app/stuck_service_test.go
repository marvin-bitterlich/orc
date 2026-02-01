package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/example/orc/internal/core/detection"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
)

// mockStuckRepository implements secondary.StuckRepository for testing.
type mockStuckRepository struct {
	stucks       map[string]*secondary.StuckRecord
	openByPatrol map[string]*secondary.StuckRecord
	nextID       int
}

func newMockStuckRepository() *mockStuckRepository {
	return &mockStuckRepository{
		stucks:       make(map[string]*secondary.StuckRecord),
		openByPatrol: make(map[string]*secondary.StuckRecord),
		nextID:       1,
	}
}

func (m *mockStuckRepository) Create(ctx context.Context, stuck *secondary.StuckRecord) error {
	m.stucks[stuck.ID] = stuck
	if stuck.Status == primary.StuckStatusOpen {
		m.openByPatrol[stuck.PatrolID] = stuck
	}
	return nil
}

func (m *mockStuckRepository) GetByID(ctx context.Context, id string) (*secondary.StuckRecord, error) {
	if s, ok := m.stucks[id]; ok {
		return s, nil
	}
	return nil, errors.New("not found")
}

func (m *mockStuckRepository) GetByPatrol(ctx context.Context, patrolID string) ([]*secondary.StuckRecord, error) {
	var result []*secondary.StuckRecord
	for _, s := range m.stucks {
		if s.PatrolID == patrolID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockStuckRepository) GetOpenByPatrol(ctx context.Context, patrolID string) (*secondary.StuckRecord, error) {
	if s, ok := m.openByPatrol[patrolID]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockStuckRepository) List(ctx context.Context, filters secondary.StuckFilters) ([]*secondary.StuckRecord, error) {
	var result []*secondary.StuckRecord
	for _, s := range m.stucks {
		if filters.PatrolID != "" && s.PatrolID != filters.PatrolID {
			continue
		}
		if filters.Status != "" && s.Status != filters.Status {
			continue
		}
		result = append(result, s)
	}
	return result, nil
}

func (m *mockStuckRepository) Update(ctx context.Context, stuck *secondary.StuckRecord) error {
	if _, ok := m.stucks[stuck.ID]; !ok {
		return errors.New("not found")
	}
	m.stucks[stuck.ID] = stuck
	return nil
}

func (m *mockStuckRepository) IncrementCount(ctx context.Context, id string) error {
	if s, ok := m.stucks[id]; ok {
		s.CheckCount++
		return nil
	}
	return errors.New("not found")
}

func (m *mockStuckRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if s, ok := m.stucks[id]; ok {
		s.Status = status
		if status != primary.StuckStatusOpen {
			delete(m.openByPatrol, s.PatrolID)
		}
		return nil
	}
	return errors.New("not found")
}

func (m *mockStuckRepository) GetNextID(ctx context.Context) (string, error) {
	id := m.nextID
	m.nextID++
	return fmt.Sprintf("STUCK-%03d", id), nil
}

func (m *mockStuckRepository) PatrolExists(ctx context.Context, patrolID string) (bool, error) {
	return true, nil
}

func newTestStuckService() (*StuckServiceImpl, *mockStuckRepository) {
	repo := newMockStuckRepository()
	service := NewStuckServiceWithThreshold(repo, 3) // Use threshold of 3 for tests
	return service, repo
}

func TestStuckService_ProcessOutcome_FirstFailure(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should create a new stuck
	if result.StuckID != "STUCK-001" {
		t.Errorf("expected StuckID 'STUCK-001', got %q", result.StuckID)
	}
	if result.CheckCount != 1 {
		t.Errorf("expected CheckCount 1, got %d", result.CheckCount)
	}
	if result.Action != detection.ActionNudge {
		t.Errorf("expected ActionNudge, got %q", result.Action)
	}
	if result.NeedsEscalation {
		t.Error("should not need escalation on first failure")
	}

	// Verify stuck was created
	if _, ok := repo.stucks["STUCK-001"]; !ok {
		t.Error("stuck record not created")
	}
}

func TestStuckService_ProcessOutcome_ConsecutiveFailures(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	// First failure
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeError)

	// Second failure
	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-002", detection.OutcomeError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CheckCount != 2 {
		t.Errorf("expected CheckCount 2, got %d", result.CheckCount)
	}
	if result.Action != detection.ActionNudge {
		t.Errorf("expected ActionNudge, got %q", result.Action)
	}
	if result.NeedsEscalation {
		t.Error("should not need escalation yet")
	}

	// Verify count was incremented
	if repo.stucks["STUCK-001"].CheckCount != 2 {
		t.Errorf("expected stuck CheckCount 2, got %d", repo.stucks["STUCK-001"].CheckCount)
	}
}

func TestStuckService_ProcessOutcome_EscalationAtThreshold(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	// First two failures
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeError)
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-002", detection.OutcomeError)

	// Third failure (at threshold of 3)
	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-003", detection.OutcomeError)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CheckCount != 3 {
		t.Errorf("expected CheckCount 3, got %d", result.CheckCount)
	}
	if result.Action != detection.ActionEscalate {
		t.Errorf("expected ActionEscalate, got %q", result.Action)
	}
	if !result.NeedsEscalation {
		t.Error("should need escalation at threshold")
	}
	if result.Message == "" {
		t.Error("expected escalation message")
	}

	// Verify stuck count
	if repo.stucks["STUCK-001"].CheckCount != 3 {
		t.Errorf("expected stuck CheckCount 3, got %d", repo.stucks["STUCK-001"].CheckCount)
	}
}

func TestStuckService_ProcessOutcome_RecoveryResolvesStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	// Create stuck with failures
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeError)
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-002", detection.OutcomeError)

	// Recovery
	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-003", detection.OutcomeWorking)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != detection.ActionNone {
		t.Errorf("expected ActionNone, got %q", result.Action)
	}
	if result.StuckID != "" {
		t.Errorf("expected empty StuckID on recovery, got %q", result.StuckID)
	}

	// Verify stuck was resolved
	if repo.stucks["STUCK-001"].Status != primary.StuckStatusResolved {
		t.Errorf("expected stuck status 'resolved', got %q", repo.stucks["STUCK-001"].Status)
	}

	// No open stuck anymore
	if _, ok := repo.openByPatrol["PATROL-001"]; ok {
		t.Error("open stuck should be removed on recovery")
	}
}

func TestStuckService_ProcessOutcome_WorkingWithNoStuck(t *testing.T) {
	service, _ := newTestStuckService()
	ctx := context.Background()

	// Working with no prior stuck
	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeWorking)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != detection.ActionNone {
		t.Errorf("expected ActionNone, got %q", result.Action)
	}
}

func TestStuckService_ProcessOutcome_IdleTriggersNudge(t *testing.T) {
	service, _ := newTestStuckService()
	ctx := context.Background()

	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeIdle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != detection.ActionNudge {
		t.Errorf("expected ActionNudge, got %q", result.Action)
	}
	if result.StuckID != "" {
		t.Error("idle should not create a stuck")
	}
	if result.Message == "" {
		t.Error("expected nudge message for idle")
	}
}

func TestStuckService_ProcessOutcome_MenuTriggersBTab(t *testing.T) {
	service, _ := newTestStuckService()
	ctx := context.Background()

	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeMenu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != detection.ActionBTab {
		t.Errorf("expected ActionBTab, got %q", result.Action)
	}
}

func TestStuckService_ProcessOutcome_TypedTriggersEnter(t *testing.T) {
	service, _ := newTestStuckService()
	ctx := context.Background()

	result, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeTyped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != detection.ActionEnter {
		t.Errorf("expected ActionEnter, got %q", result.Action)
	}
}

func TestStuckService_ProcessOutcome_MenuResolvesOpenStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	// Create stuck
	service.ProcessOutcome(ctx, "PATROL-001", "CHECK-001", detection.OutcomeError)

	// Menu action
	_, err := service.ProcessOutcome(ctx, "PATROL-001", "CHECK-002", detection.OutcomeMenu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify stuck was resolved
	if repo.stucks["STUCK-001"].Status != primary.StuckStatusResolved {
		t.Errorf("expected stuck status 'resolved', got %q", repo.stucks["STUCK-001"].Status)
	}
}

func TestStuckService_GetStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	repo.stucks["STUCK-001"] = &secondary.StuckRecord{
		ID:         "STUCK-001",
		PatrolID:   "PATROL-001",
		CheckCount: 3,
		Status:     "open",
	}

	stuck, err := service.GetStuck(ctx, "STUCK-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stuck.CheckCount != 3 {
		t.Errorf("expected CheckCount 3, got %d", stuck.CheckCount)
	}
}

func TestStuckService_GetOpenStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	repo.stucks["STUCK-001"] = &secondary.StuckRecord{
		ID:         "STUCK-001",
		PatrolID:   "PATROL-001",
		CheckCount: 2,
		Status:     "open",
	}
	repo.openByPatrol["PATROL-001"] = repo.stucks["STUCK-001"]

	stuck, err := service.GetOpenStuck(ctx, "PATROL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stuck.ID != "STUCK-001" {
		t.Errorf("expected STUCK-001, got %q", stuck.ID)
	}
}

func TestStuckService_GetOpenStuck_NoOpenStuck(t *testing.T) {
	service, _ := newTestStuckService()
	ctx := context.Background()

	stuck, err := service.GetOpenStuck(ctx, "PATROL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stuck != nil {
		t.Error("expected nil for no open stuck")
	}
}

func TestStuckService_ResolveStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	repo.stucks["STUCK-001"] = &secondary.StuckRecord{
		ID:       "STUCK-001",
		PatrolID: "PATROL-001",
		Status:   "open",
	}
	repo.openByPatrol["PATROL-001"] = repo.stucks["STUCK-001"]

	err := service.ResolveStuck(ctx, "STUCK-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.stucks["STUCK-001"].Status != primary.StuckStatusResolved {
		t.Errorf("expected status 'resolved', got %q", repo.stucks["STUCK-001"].Status)
	}
}

func TestStuckService_EscalateStuck(t *testing.T) {
	service, repo := newTestStuckService()
	ctx := context.Background()

	repo.stucks["STUCK-001"] = &secondary.StuckRecord{
		ID:       "STUCK-001",
		PatrolID: "PATROL-001",
		Status:   "open",
	}
	repo.openByPatrol["PATROL-001"] = repo.stucks["STUCK-001"]

	err := service.EscalateStuck(ctx, "STUCK-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.stucks["STUCK-001"].Status != primary.StuckStatusEscalated {
		t.Errorf("expected status 'escalated', got %q", repo.stucks["STUCK-001"].Status)
	}
}
