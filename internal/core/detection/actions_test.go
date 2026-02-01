package detection

import (
	"context"
	"testing"
)

func TestSelectAction_Menu(t *testing.T) {
	action := SelectAction(OutcomeMenu, "session:window.2", 0, DefaultStuckThreshold)

	if action.Type != ActionBTab {
		t.Errorf("expected ActionBTab, got %q", action.Type)
	}
	if action.Target != "session:window.2" {
		t.Errorf("expected target 'session:window.2', got %q", action.Target)
	}
}

func TestSelectAction_Typed(t *testing.T) {
	action := SelectAction(OutcomeTyped, "session:window.2", 0, DefaultStuckThreshold)

	if action.Type != ActionEnter {
		t.Errorf("expected ActionEnter, got %q", action.Type)
	}
}

func TestSelectAction_Idle(t *testing.T) {
	action := SelectAction(OutcomeIdle, "session:window.2", 0, DefaultStuckThreshold)

	if action.Type != ActionNudge {
		t.Errorf("expected ActionNudge, got %q", action.Type)
	}
	if action.Message == "" {
		t.Error("expected nudge message to be set")
	}
}

func TestSelectAction_Working(t *testing.T) {
	action := SelectAction(OutcomeWorking, "session:window.2", 0, DefaultStuckThreshold)

	if action.Type != ActionNone {
		t.Errorf("expected ActionNone, got %q", action.Type)
	}
}

func TestSelectAction_Error_BelowThreshold(t *testing.T) {
	action := SelectAction(OutcomeError, "session:window.2", 2, DefaultStuckThreshold)

	if action.Type != ActionNudge {
		t.Errorf("expected ActionNudge below threshold, got %q", action.Type)
	}
}

func TestSelectAction_Error_AtThreshold(t *testing.T) {
	action := SelectAction(OutcomeError, "session:window.2", 5, DefaultStuckThreshold)

	if action.Type != ActionEscalate {
		t.Errorf("expected ActionEscalate at threshold, got %q", action.Type)
	}
	if action.Message == "" {
		t.Error("expected escalation message to be set")
	}
}

func TestSelectAction_Error_AboveThreshold(t *testing.T) {
	action := SelectAction(OutcomeError, "session:window.2", 10, DefaultStuckThreshold)

	if action.Type != ActionEscalate {
		t.Errorf("expected ActionEscalate above threshold, got %q", action.Type)
	}
}

// mockExecutor implements ActionExecutor for testing.
type mockExecutor struct {
	sentKeys   []string
	sentNudges []string
	lastTarget string
}

func (m *mockExecutor) SendKeys(ctx context.Context, target, keys string) error {
	m.lastTarget = target
	m.sentKeys = append(m.sentKeys, keys)
	return nil
}

func (m *mockExecutor) NudgeSession(ctx context.Context, target, message string) error {
	m.lastTarget = target
	m.sentNudges = append(m.sentNudges, message)
	return nil
}

func TestExecuteAction_BTab(t *testing.T) {
	executor := &mockExecutor{}
	action := Action{Type: ActionBTab, Target: "session:window.2"}

	err := ExecuteAction(context.Background(), executor, action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(executor.sentKeys) != 1 || executor.sentKeys[0] != "BTab" {
		t.Errorf("expected BTab to be sent, got %v", executor.sentKeys)
	}
	if executor.lastTarget != "session:window.2" {
		t.Errorf("expected target 'session:window.2', got %q", executor.lastTarget)
	}
}

func TestExecuteAction_Enter(t *testing.T) {
	executor := &mockExecutor{}
	action := Action{Type: ActionEnter, Target: "session:window.2"}

	err := ExecuteAction(context.Background(), executor, action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(executor.sentKeys) != 1 || executor.sentKeys[0] != "Enter" {
		t.Errorf("expected Enter to be sent, got %v", executor.sentKeys)
	}
}

func TestExecuteAction_Nudge(t *testing.T) {
	executor := &mockExecutor{}
	action := Action{Type: ActionNudge, Target: "session:window.2", Message: "test nudge"}

	err := ExecuteAction(context.Background(), executor, action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(executor.sentNudges) != 1 || executor.sentNudges[0] != "test nudge" {
		t.Errorf("expected nudge message to be sent, got %v", executor.sentNudges)
	}
}

func TestExecuteAction_None(t *testing.T) {
	executor := &mockExecutor{}
	action := Action{Type: ActionNone, Target: "session:window.2"}

	err := ExecuteAction(context.Background(), executor, action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(executor.sentKeys) != 0 || len(executor.sentNudges) != 0 {
		t.Error("expected no actions for ActionNone")
	}
}

func TestExecuteAction_Escalate(t *testing.T) {
	executor := &mockExecutor{}
	action := Action{Type: ActionEscalate, Target: "session:window.2", Message: "needs help"}

	err := ExecuteAction(context.Background(), executor, action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Escalate doesn't execute anything directly - it signals to caller
	if len(executor.sentKeys) != 0 || len(executor.sentNudges) != 0 {
		t.Error("expected no direct actions for ActionEscalate")
	}
}
