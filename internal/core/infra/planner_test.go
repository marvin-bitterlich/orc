package infra

import (
	"testing"
)

func TestGeneratePlan_Empty(t *testing.T) {
	input := PlanInput{
		WorkshopID:    "WORK-001",
		WorkshopName:  "Test Workshop",
		FactoryID:     "FACT-001",
		FactoryName:   "default",
		GatehouseID:   "GATE-001",
		GatehousePath: "/home/user/.orc/ws/WORK-001-test-workshop",
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
	if plan.Gatehouse == nil {
		t.Fatal("expected Gatehouse to be set")
	}
	if plan.Gatehouse.ID != "GATE-001" {
		t.Errorf("expected Gatehouse.ID 'GATE-001', got %q", plan.Gatehouse.ID)
	}
	if len(plan.Workbenches) != 0 {
		t.Errorf("expected 0 workbenches, got %d", len(plan.Workbenches))
	}
}

func TestGeneratePlan_WithWorkbenches(t *testing.T) {
	input := PlanInput{
		WorkshopID:            "WORK-001",
		WorkshopName:          "Test Workshop",
		FactoryID:             "FACT-001",
		FactoryName:           "default",
		GatehouseID:           "GATE-001",
		GatehousePath:         "/home/user/.orc/ws/WORK-001-test",
		GatehousePathExists:   true,
		GatehouseConfigExists: true,
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

func TestGeneratePlan_GatehouseState(t *testing.T) {
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
				WorkshopID:            "WORK-001",
				WorkshopName:          "Test",
				FactoryID:             "FACT-001",
				FactoryName:           "default",
				GatehouseID:           "GATE-001",
				GatehousePath:         "/home/user/.orc/ws/WORK-001-test",
				GatehousePathExists:   tt.pathExists,
				GatehouseConfigExists: tt.configExists,
			}

			plan := GeneratePlan(input)

			if plan.Gatehouse.Exists != tt.expectPath {
				t.Errorf("expected Gatehouse.Exists=%v, got %v", tt.expectPath, plan.Gatehouse.Exists)
			}
			if plan.Gatehouse.ConfigExists != tt.expectConfig {
				t.Errorf("expected Gatehouse.ConfigExists=%v, got %v", tt.expectConfig, plan.Gatehouse.ConfigExists)
			}
		})
	}
}

func TestGeneratePlan_WorkbenchMetadata(t *testing.T) {
	input := PlanInput{
		WorkshopID:    "WORK-001",
		WorkshopName:  "Test",
		FactoryID:     "FACT-001",
		FactoryName:   "default",
		GatehouseID:   "GATE-001",
		GatehousePath: "/home/.orc/ws/WORK-001",
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
