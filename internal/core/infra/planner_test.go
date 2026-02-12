package infra

import (
	"testing"
)

func TestGeneratePlan_Empty(t *testing.T) {
	input := PlanInput{
		WorkshopID:      "WORK-001",
		WorkshopName:    "Test Workshop",
		FactoryID:       "FACT-001",
		FactoryName:     "default",
		WorkshopDirID:   "WORK-001",
		WorkshopDirPath: "/home/user/.orc/ws/WORK-001-test-workshop",
	}

	plan := GeneratePlan(input)

	if plan.WorkshopID != "WORK-001" {
		t.Errorf("expected WorkshopID 'WORK-001', got %q", plan.WorkshopID)
	}
	if plan.WorkshopName != "Test Workshop" {
		t.Errorf("expected WorkshopName 'Test Workshop', got %q", plan.WorkshopName)
	}
	if plan.FactoryID != "FACT-001" {
		t.Errorf("expected FactoryID 'FACT-001', got %q", plan.FactoryID)
	}
	if plan.WorkshopDir == nil {
		t.Fatal("expected WorkshopDir to be set")
	}
	if plan.WorkshopDir.ID != "WORK-001" {
		t.Errorf("expected WorkshopDir.ID 'WORK-001', got %q", plan.WorkshopDir.ID)
	}
	if len(plan.Workbenches) != 0 {
		t.Errorf("expected 0 workbenches, got %d", len(plan.Workbenches))
	}
}

func TestGeneratePlan_WithWorkbenches(t *testing.T) {
	input := PlanInput{
		WorkshopID:              "WORK-001",
		WorkshopName:            "Test Workshop",
		FactoryID:               "FACT-001",
		FactoryName:             "default",
		WorkshopDirID:           "WORK-001",
		WorkshopDirPath:         "/home/user/.orc/ws/WORK-001-test",
		WorkshopDirPathExists:   true,
		WorkshopDirConfigExists: true,
		Workbenches: []WorkbenchPlanInput{
			{
				ID:             "BENCH-001",
				Name:           "test-bench",
				WorktreePath:   "/home/user/wb/test-bench",
				RepoName:       "my-repo",
				HomeBranch:     "main",
				WorktreeExists: true,
				ConfigExists:   true,
			},
			{
				ID:             "BENCH-002",
				Name:           "other-bench",
				WorktreePath:   "/home/user/wb/other-bench",
				RepoName:       "my-repo",
				HomeBranch:     "feature",
				WorktreeExists: false,
				ConfigExists:   false,
			},
		},
	}

	plan := GeneratePlan(input)

	if len(plan.Workbenches) != 2 {
		t.Fatalf("expected 2 workbenches, got %d", len(plan.Workbenches))
	}

	// First workbench should exist
	if plan.Workbenches[0].ID != "BENCH-001" {
		t.Errorf("expected first workbench ID 'BENCH-001', got %q", plan.Workbenches[0].ID)
	}
	if !plan.Workbenches[0].Exists {
		t.Error("expected first workbench to exist")
	}
	if !plan.Workbenches[0].ConfigExists {
		t.Error("expected first workbench config to exist")
	}

	// Second workbench should not exist
	if plan.Workbenches[1].ID != "BENCH-002" {
		t.Errorf("expected second workbench ID 'BENCH-002', got %q", plan.Workbenches[1].ID)
	}
	if plan.Workbenches[1].Exists {
		t.Error("expected second workbench to not exist")
	}
	if plan.Workbenches[1].ConfigExists {
		t.Error("expected second workbench config to not exist")
	}
}

func TestGeneratePlan_WorkshopDirState(t *testing.T) {
	tests := []struct {
		name         string
		pathExists   bool
		configExists bool
		expectPath   bool
		expectConfig bool
	}{
		{"both exist", true, true, true, true},
		{"neither exist", false, false, false, false},
		{"path only", true, false, true, false},
		{"config only", false, true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := PlanInput{
				WorkshopID:              "WORK-001",
				WorkshopName:            "Test",
				FactoryID:               "FACT-001",
				FactoryName:             "default",
				WorkshopDirID:           "WORK-001",
				WorkshopDirPath:         "/home/user/.orc/ws/WORK-001-test",
				WorkshopDirPathExists:   tt.pathExists,
				WorkshopDirConfigExists: tt.configExists,
			}

			plan := GeneratePlan(input)

			if plan.WorkshopDir.Exists != tt.expectPath {
				t.Errorf("expected WorkshopDir.Exists=%v, got %v", tt.expectPath, plan.WorkshopDir.Exists)
			}
			if plan.WorkshopDir.ConfigExists != tt.expectConfig {
				t.Errorf("expected WorkshopDir.ConfigExists=%v, got %v", tt.expectConfig, plan.WorkshopDir.ConfigExists)
			}
		})
	}
}

func TestGeneratePlan_WorkbenchMetadata(t *testing.T) {
	input := PlanInput{
		WorkshopID:      "WORK-001",
		WorkshopName:    "Test",
		FactoryID:       "FACT-001",
		FactoryName:     "default",
		WorkshopDirID:   "WORK-001",
		WorkshopDirPath: "/home/.orc/ws/WORK-001",
		Workbenches: []WorkbenchPlanInput{
			{
				ID:           "BENCH-001",
				Name:         "feature-bench",
				WorktreePath: "/home/wb/feature-bench",
				RepoName:     "intercom",
				HomeBranch:   "ml/feature-branch",
			},
		},
	}

	plan := GeneratePlan(input)

	if len(plan.Workbenches) != 1 {
		t.Fatalf("expected 1 workbench, got %d", len(plan.Workbenches))
	}

	wb := plan.Workbenches[0]
	if wb.Name != "feature-bench" {
		t.Errorf("expected Name 'feature-bench', got %q", wb.Name)
	}
	if wb.Path != "/home/wb/feature-bench" {
		t.Errorf("expected Path '/home/wb/feature-bench', got %q", wb.Path)
	}
	if wb.RepoName != "intercom" {
		t.Errorf("expected RepoName 'intercom', got %q", wb.RepoName)
	}
	if wb.Branch != "ml/feature-branch" {
		t.Errorf("expected Branch 'ml/feature-branch', got %q", wb.Branch)
	}
}

func TestGeneratePlan_TMuxSession(t *testing.T) {
	input := PlanInput{
		WorkshopID:            "WORK-001",
		WorkshopName:          "Test Workshop",
		FactoryID:             "FACT-001",
		FactoryName:           "default",
		WorkshopDirID:         "WORK-001",
		WorkshopDirPath:       "/home/user/.orc/ws/WORK-001-test",
		TMuxSessionExists:     true,
		TMuxActualSessionName: "test-workshop",
		TMuxExistingWindows:   []string{"bench-1", "bench-2"},
		TMuxExpectedWindows: []TMuxWindowInput{
			{Name: "bench-1", Path: "/home/wb/bench-1"},
			{Name: "bench-2", Path: "/home/wb/bench-2"},
		},
	}

	plan := GeneratePlan(input)

	if plan.TMuxSession == nil {
		t.Fatal("expected TMuxSession to be set")
	}
	if plan.TMuxSession.SessionName != "test-workshop" {
		t.Errorf("expected SessionName 'test-workshop', got %q", plan.TMuxSession.SessionName)
	}
	if !plan.TMuxSession.Exists {
		t.Error("expected TMuxSession.Exists to be true")
	}
	if len(plan.TMuxSession.Windows) != 2 {
		t.Errorf("expected 2 windows, got %d", len(plan.TMuxSession.Windows))
	}
	if len(plan.TMuxSession.OrphanWindows) != 0 {
		t.Errorf("expected 0 orphan windows, got %d", len(plan.TMuxSession.OrphanWindows))
	}
}

func TestGeneratePlan_TMuxSession_WithMissingWindows(t *testing.T) {
	input := PlanInput{
		WorkshopID:            "WORK-001",
		WorkshopName:          "Test",
		FactoryID:             "FACT-001",
		FactoryName:           "default",
		WorkshopDirID:         "WORK-001",
		WorkshopDirPath:       "/home/.orc/ws/WORK-001",
		TMuxSessionExists:     true,
		TMuxActualSessionName: "test",
		TMuxExistingWindows:   []string{"bench-1"}, // Only bench-1 exists
		TMuxExpectedWindows: []TMuxWindowInput{
			{Name: "bench-1", Path: "/home/wb/bench-1"},
			{Name: "bench-2", Path: "/home/wb/bench-2"}, // bench-2 expected but missing
		},
	}

	plan := GeneratePlan(input)

	if len(plan.TMuxSession.Windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(plan.TMuxSession.Windows))
	}

	// bench-1 should exist
	if !plan.TMuxSession.Windows[0].Exists {
		t.Error("expected bench-1 to exist")
	}
	// bench-2 should not exist
	if plan.TMuxSession.Windows[1].Exists {
		t.Error("expected bench-2 to not exist")
	}
}

func TestGeneratePlan_TMuxSession_WithOrphanWindows(t *testing.T) {
	input := PlanInput{
		WorkshopID:            "WORK-001",
		WorkshopName:          "Test",
		FactoryID:             "FACT-001",
		FactoryName:           "default",
		WorkshopDirID:         "WORK-001",
		WorkshopDirPath:       "/home/.orc/ws/WORK-001",
		TMuxSessionExists:     true,
		TMuxActualSessionName: "test",
		TMuxExistingWindows:   []string{"bench-1", "old-bench"}, // old-bench exists but not expected
		TMuxExpectedWindows: []TMuxWindowInput{
			{Name: "bench-1", Path: "/home/wb/bench-1"},
		},
	}

	plan := GeneratePlan(input)

	if len(plan.TMuxSession.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(plan.TMuxSession.Windows))
	}
	if len(plan.TMuxSession.OrphanWindows) != 1 {
		t.Fatalf("expected 1 orphan window, got %d", len(plan.TMuxSession.OrphanWindows))
	}
	if plan.TMuxSession.OrphanWindows[0].Name != "old-bench" {
		t.Errorf("expected orphan 'old-bench', got %q", plan.TMuxSession.OrphanWindows[0].Name)
	}
}

func TestGeneratePlan_TMuxSession_NoSession(t *testing.T) {
	input := PlanInput{
		WorkshopID:            "WORK-001",
		WorkshopName:          "Test",
		FactoryID:             "FACT-001",
		FactoryName:           "default",
		WorkshopDirID:         "WORK-001",
		WorkshopDirPath:       "/home/.orc/ws/WORK-001",
		TMuxSessionExists:     false,
		TMuxActualSessionName: "",
		TMuxExpectedWindows: []TMuxWindowInput{
			{Name: "bench-1", Path: "/home/wb/bench-1"},
		},
	}

	plan := GeneratePlan(input)

	if plan.TMuxSession == nil {
		t.Fatal("expected TMuxSession to be set")
	}
	if plan.TMuxSession.Exists {
		t.Error("expected TMuxSession.Exists to be false")
	}
	// Windows should still be listed but marked as not existing
	if len(plan.TMuxSession.Windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(plan.TMuxSession.Windows))
	}
	if plan.TMuxSession.Windows[0].Exists {
		t.Error("expected window to not exist when session doesn't exist")
	}
}
