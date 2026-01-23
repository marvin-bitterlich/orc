package commission

import (
	"testing"

	"github.com/example/orc/internal/core/effects"
)

func TestGenerateLaunchPlan_BasicCommission(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-001",
		CommissionTitle: "Test Commission",
		WorkspacePath:   "/home/user/commissions/COMM-001",
		CreateTMux:      false,
		Workbenches:     []WorkbenchPlanInput{},
	}

	plan := GenerateLaunchPlan(input)

	// Should have 2 filesystem ops: workspace dir + workbenches dir
	if len(plan.FilesystemOps) != 2 {
		t.Errorf("FilesystemOps count = %d, want 2", len(plan.FilesystemOps))
	}

	// First op should be workspace mkdir
	if plan.FilesystemOps[0].Operation != "mkdir" {
		t.Errorf("First op = %q, want mkdir", plan.FilesystemOps[0].Operation)
	}
	if plan.FilesystemOps[0].Path != "/home/user/commissions/COMM-001" {
		t.Errorf("First op path = %q, want workspace path", plan.FilesystemOps[0].Path)
	}

	// Second op should be workbenches mkdir
	if plan.FilesystemOps[1].Path != "/home/user/commissions/COMM-001/groves" {
		t.Errorf("Second op path = %q, want workbenches path", plan.FilesystemOps[1].Path)
	}

	// No TMux ops
	if len(plan.TMuxOps) != 0 {
		t.Errorf("TMuxOps count = %d, want 0", len(plan.TMuxOps))
	}

	// No DB ops
	if len(plan.DatabaseOps) != 0 {
		t.Errorf("DatabaseOps count = %d, want 0", len(plan.DatabaseOps))
	}
}

func TestGenerateLaunchPlan_WithWorkbench(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-002",
		CommissionTitle: "Test Commission",
		WorkspacePath:   "/home/user/commissions/COMM-002",
		CreateTMux:      false,
		Workbenches: []WorkbenchPlanInput{
			{
				ID:          "BENCH-001",
				Name:        "api-workbench",
				CurrentPath: "/old/path/api-workbench",
				Repos:       []string{"https://github.com/example/api"},
				PathExists:  true,
			},
		},
	}

	plan := GenerateLaunchPlan(input)

	// Should have 4 filesystem ops: workspace + workbenches + .orc dir + config.json
	if len(plan.FilesystemOps) != 4 {
		t.Errorf("FilesystemOps count = %d, want 4", len(plan.FilesystemOps))
	}

	// Should have 1 DB op for path update
	if len(plan.DatabaseOps) != 1 {
		t.Errorf("DatabaseOps count = %d, want 1", len(plan.DatabaseOps))
	}

	// Verify DB op updates path
	dbOp := plan.DatabaseOps[0]
	if dbOp.Entity != "grove" || dbOp.Operation != "update" {
		t.Errorf("DB op = %s/%s, want grove/update", dbOp.Entity, dbOp.Operation)
	}
}

func TestGenerateLaunchPlan_WorkbenchPathUnchanged(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-003",
		CommissionTitle: "Test Commission",
		WorkspacePath:   "/home/user/commissions/COMM-003",
		CreateTMux:      false,
		Workbenches: []WorkbenchPlanInput{
			{
				ID:          "BENCH-001",
				Name:        "web-workbench",
				CurrentPath: "/home/user/commissions/COMM-003/groves/web-workbench", // Already correct
				Repos:       []string{},
				PathExists:  true,
			},
		},
	}

	plan := GenerateLaunchPlan(input)

	// No DB ops when path is already correct
	if len(plan.DatabaseOps) != 0 {
		t.Errorf("DatabaseOps count = %d, want 0 (path unchanged)", len(plan.DatabaseOps))
	}
}

func TestGenerateLaunchPlan_WithTMux(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-004",
		CommissionTitle: "Test Commission",
		WorkspacePath:   "/home/user/commissions/COMM-004",
		CreateTMux:      true,
		Workbenches: []WorkbenchPlanInput{
			{
				ID:          "BENCH-001",
				Name:        "backend",
				CurrentPath: "/home/user/commissions/COMM-004/groves/backend",
				PathExists:  true,
			},
			{
				ID:          "BENCH-002",
				Name:        "frontend",
				CurrentPath: "/home/user/commissions/COMM-004/groves/frontend",
				PathExists:  true,
			},
		},
	}

	plan := GenerateLaunchPlan(input)

	// Should have TMux ops: 1 session + 2 windows
	if len(plan.TMuxOps) != 3 {
		t.Errorf("TMuxOps count = %d, want 3", len(plan.TMuxOps))
	}

	// First TMux op should be new_session
	if plan.TMuxOps[0].Operation != "new_session" {
		t.Errorf("First TMux op = %q, want new_session", plan.TMuxOps[0].Operation)
	}
	if plan.TMuxOps[0].SessionName != "orc-COMM-004" {
		t.Errorf("Session name = %q, want orc-COMM-004", plan.TMuxOps[0].SessionName)
	}

	// Second and third should be new_window
	if plan.TMuxOps[1].Operation != "new_window" || plan.TMuxOps[2].Operation != "new_window" {
		t.Errorf("Window ops not new_window")
	}
}

func TestGenerateLaunchPlan_TMuxSkipsNonExistentPaths(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-005",
		CommissionTitle: "Test Commission",
		WorkspacePath:   "/home/user/commissions/COMM-005",
		CreateTMux:      true,
		Workbenches: []WorkbenchPlanInput{
			{
				ID:          "BENCH-001",
				Name:        "existing",
				CurrentPath: "/some/path",
				PathExists:  true,
			},
			{
				ID:          "BENCH-002",
				Name:        "not-existing",
				CurrentPath: "/some/other/path",
				PathExists:  false, // Does not exist
			},
		},
	}

	plan := GenerateLaunchPlan(input)

	// Should only have 2 TMux ops: 1 session + 1 window (skips non-existent)
	if len(plan.TMuxOps) != 2 {
		t.Errorf("TMuxOps count = %d, want 2 (should skip non-existent workbench)", len(plan.TMuxOps))
	}
}

func TestLaunchPlan_Effects(t *testing.T) {
	input := LaunchPlanInput{
		CommissionID:    "COMM-006",
		CommissionTitle: "Test",
		WorkspacePath:   "/test",
		CreateTMux:      true,
		Workbenches: []WorkbenchPlanInput{
			{ID: "BENCH-001", Name: "test", CurrentPath: "/old", PathExists: true},
		},
	}

	plan := GenerateLaunchPlan(input)
	allEffects := plan.Effects()

	// Verify Effects() returns all effects
	expectedCount := len(plan.FilesystemOps) + len(plan.DatabaseOps) + len(plan.TMuxOps)
	if len(allEffects) != expectedCount {
		t.Errorf("Effects() returned %d, want %d", len(allEffects), expectedCount)
	}

	// Verify effect types are valid
	for _, e := range allEffects {
		switch e.(type) {
		case effects.FileEffect, effects.PersistEffect, effects.TMuxEffect:
			// OK
		default:
			t.Errorf("Unexpected effect type: %T", e)
		}
	}
}

func TestGenerateStartPlan_Basic(t *testing.T) {
	input := StartPlanInput{
		CommissionID:  "COMM-007",
		WorkspacePath: "/home/user/commissions/COMM-007",
		Workbenches: []WorkbenchPlanInput{
			{
				ID:          "BENCH-001",
				Name:        "main",
				CurrentPath: "/home/user/commissions/COMM-007/groves/main",
				PathExists:  true,
			},
		},
	}

	plan := GenerateStartPlan(input)

	// Should have 2 TMux ops: session + window
	if len(plan.TMuxOps) != 2 {
		t.Errorf("TMuxOps count = %d, want 2", len(plan.TMuxOps))
	}

	if plan.TMuxOps[0].Operation != "new_session" {
		t.Errorf("First op = %q, want new_session", plan.TMuxOps[0].Operation)
	}

	if plan.TMuxOps[0].SessionName != "orc-COMM-007" {
		t.Errorf("Session name = %q, want orc-COMM-007", plan.TMuxOps[0].SessionName)
	}
}

func TestGenerateStartPlan_SkipsNonExistent(t *testing.T) {
	input := StartPlanInput{
		CommissionID:  "COMM-008",
		WorkspacePath: "/home/user/commissions/COMM-008",
		Workbenches: []WorkbenchPlanInput{
			{ID: "BENCH-001", Name: "exists", PathExists: true},
			{ID: "BENCH-002", Name: "missing", PathExists: false},
		},
	}

	plan := GenerateStartPlan(input)

	// Should have 2 TMux ops: session + 1 window (skips missing)
	if len(plan.TMuxOps) != 2 {
		t.Errorf("TMuxOps count = %d, want 2", len(plan.TMuxOps))
	}
}

func TestStartPlan_Effects(t *testing.T) {
	input := StartPlanInput{
		CommissionID:  "COMM-009",
		WorkspacePath: "/test",
		Workbenches: []WorkbenchPlanInput{
			{ID: "BENCH-001", Name: "test", PathExists: true},
		},
	}

	plan := GenerateStartPlan(input)
	allEffects := plan.Effects()

	if len(allEffects) != len(plan.TMuxOps) {
		t.Errorf("Effects() returned %d, want %d", len(allEffects), len(plan.TMuxOps))
	}

	for _, e := range allEffects {
		if _, ok := e.(effects.TMuxEffect); !ok {
			t.Errorf("Expected TMuxEffect, got %T", e)
		}
	}
}

func TestGenerateWorkbenchConfig(t *testing.T) {
	content := generateWorkbenchConfig("BENCH-001", "COMM-001", "api-workbench", []string{"https://github.com/example/api"})

	if len(content) == 0 {
		t.Error("generateWorkbenchConfig returned empty content")
	}

	// Verify it's valid JSON by checking it contains expected fields
	contentStr := string(content)
	expectedFields := []string{
		`"version": "1.0"`,
		`"type": "grove"`,
		`"grove_id": "BENCH-001"`,
		`"commission_id": "COMM-001"`,
		`"name": "api-workbench"`,
	}

	for _, field := range expectedFields {
		if !contains(contentStr, field) {
			t.Errorf("Config missing expected field: %s", field)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
