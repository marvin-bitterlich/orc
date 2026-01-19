package grove

import (
	"strings"
	"testing"

	"github.com/example/orc/internal/core/effects"
)

func TestGenerateCreateGrovePlan_Basic(t *testing.T) {
	input := CreateGrovePlanInput{
		GroveID:   "GROVE-001",
		GroveName: "auth-backend",
		MissionID: "MISSION-001",
		BasePath:  "/home/user/src/worktrees",
		Repos:     []string{},
	}

	plan := GenerateCreateGrovePlan(input)

	// Verify path format
	expectedPath := "/home/user/src/worktrees/MISSION-001-auth-backend"
	if plan.GrovePath != expectedPath {
		t.Errorf("GrovePath = %q, want %q", plan.GrovePath, expectedPath)
	}

	// Should have 3 filesystem ops: grove dir + .orc dir + config.json
	if len(plan.FilesystemOps) != 3 {
		t.Errorf("FilesystemOps count = %d, want 3", len(plan.FilesystemOps))
	}

	// First op should be mkdir for grove
	if plan.FilesystemOps[0].Operation != "mkdir" {
		t.Errorf("First op = %q, want mkdir", plan.FilesystemOps[0].Operation)
	}
	if plan.FilesystemOps[0].Path != expectedPath {
		t.Errorf("First op path = %q, want %q", plan.FilesystemOps[0].Path, expectedPath)
	}

	// Second op should be mkdir for .orc
	if plan.FilesystemOps[1].Operation != "mkdir" {
		t.Errorf("Second op = %q, want mkdir", plan.FilesystemOps[1].Operation)
	}
	if !strings.HasSuffix(plan.FilesystemOps[1].Path, ".orc") {
		t.Errorf("Second op path should end with .orc, got %q", plan.FilesystemOps[1].Path)
	}

	// Third op should be write for config.json
	if plan.FilesystemOps[2].Operation != "write" {
		t.Errorf("Third op = %q, want write", plan.FilesystemOps[2].Operation)
	}
	if !strings.HasSuffix(plan.FilesystemOps[2].Path, "config.json") {
		t.Errorf("Third op path should end with config.json, got %q", plan.FilesystemOps[2].Path)
	}

	// Should have 1 database op
	if len(plan.DatabaseOps) != 1 {
		t.Errorf("DatabaseOps count = %d, want 1", len(plan.DatabaseOps))
	}

	// No git ops without repos
	if len(plan.GitOps) != 0 {
		t.Errorf("GitOps count = %d, want 0 (no repos)", len(plan.GitOps))
	}
}

func TestGenerateCreateGrovePlan_WithRepos(t *testing.T) {
	input := CreateGrovePlanInput{
		GroveID:   "GROVE-002",
		GroveName: "frontend",
		MissionID: "MISSION-002",
		BasePath:  "/home/user/src/worktrees",
		Repos:     []string{"main-app", "api-service"},
	}

	plan := GenerateCreateGrovePlan(input)

	// Should have 2 git ops for repos
	if len(plan.GitOps) != 2 {
		t.Errorf("GitOps count = %d, want 2", len(plan.GitOps))
	}

	for i, op := range plan.GitOps {
		if op.Operation != "worktree_add" {
			t.Errorf("Git op %d = %q, want worktree_add", i, op.Operation)
		}
	}

	// Verify repo paths
	if plan.GitOps[0].RepoPath != "main-app" {
		t.Errorf("First git op repo = %q, want main-app", plan.GitOps[0].RepoPath)
	}
	if plan.GitOps[1].RepoPath != "api-service" {
		t.Errorf("Second git op repo = %q, want api-service", plan.GitOps[1].RepoPath)
	}
}

func TestGenerateCreateGrovePlan_ConfigContent(t *testing.T) {
	input := CreateGrovePlanInput{
		GroveID:   "GROVE-001",
		GroveName: "test-grove",
		MissionID: "MISSION-001",
		BasePath:  "/tmp",
		Repos:     []string{"repo1"},
	}

	plan := GenerateCreateGrovePlan(input)

	// Find the config.json write op
	var configOp *effects.FileEffect
	for _, op := range plan.FilesystemOps {
		if op.Operation == "write" && strings.HasSuffix(op.Path, "config.json") {
			configOp = &op
			break
		}
	}

	if configOp == nil {
		t.Fatal("config.json write op not found")
	}

	// Verify config content includes expected fields
	content := string(configOp.Content)
	if !strings.Contains(content, "GROVE-001") {
		t.Error("config content should contain grove ID")
	}
	if !strings.Contains(content, "MISSION-001") {
		t.Error("config content should contain mission ID")
	}
	if !strings.Contains(content, "test-grove") {
		t.Error("config content should contain grove name")
	}
	if !strings.Contains(content, `"type": "grove"`) {
		t.Error("config content should contain type: grove")
	}
}

func TestGenerateOpenGrovePlan(t *testing.T) {
	input := OpenGrovePlanInput{
		GroveID:         "GROVE-001",
		GroveName:       "backend",
		GrovePath:       "/home/user/src/worktrees/MISSION-001-backend",
		SessionName:     "orc-MISSION-001",
		NextWindowIndex: 2,
	}

	plan := GenerateOpenGrovePlan(input)

	// Should have TMux ops for: new_window, split_vertical, split_horizontal, send_keys x2
	if len(plan.TMuxOps) < 5 {
		t.Errorf("TMuxOps count = %d, want at least 5", len(plan.TMuxOps))
	}

	// First op should be new_window
	if plan.TMuxOps[0].Operation != "new_window" {
		t.Errorf("First TMux op = %q, want new_window", plan.TMuxOps[0].Operation)
	}
	if plan.TMuxOps[0].WindowName != "backend" {
		t.Errorf("Window name = %q, want backend", plan.TMuxOps[0].WindowName)
	}
	if plan.TMuxOps[0].SessionName != "orc-MISSION-001" {
		t.Errorf("Session name = %q, want orc-MISSION-001", plan.TMuxOps[0].SessionName)
	}

	// Second op should be split_vertical
	if plan.TMuxOps[1].Operation != "split_vertical" {
		t.Errorf("Second TMux op = %q, want split_vertical", plan.TMuxOps[1].Operation)
	}

	// Third op should be split_horizontal
	if plan.TMuxOps[2].Operation != "split_horizontal" {
		t.Errorf("Third TMux op = %q, want split_horizontal", plan.TMuxOps[2].Operation)
	}

	// Should have send_keys operations for vim and claude
	hasVim := false
	hasClaude := false
	for _, op := range plan.TMuxOps {
		if op.Operation == "send_keys" {
			if op.Command == "vim" {
				hasVim = true
			}
			if strings.Contains(op.Command, "claude") {
				hasClaude = true
			}
		}
	}
	if !hasVim {
		t.Error("expected send_keys for vim")
	}
	if !hasClaude {
		t.Error("expected send_keys for claude")
	}
}

func TestCreateGrovePlan_Effects(t *testing.T) {
	input := CreateGrovePlanInput{
		GroveID:   "GROVE-001",
		GroveName: "test",
		MissionID: "MISSION-001",
		BasePath:  "/tmp",
		Repos:     []string{"repo1"},
	}

	plan := GenerateCreateGrovePlan(input)
	allEffects := plan.Effects()

	expectedCount := len(plan.FilesystemOps) + len(plan.GitOps) + len(plan.DatabaseOps)
	if len(allEffects) != expectedCount {
		t.Errorf("Effects() returned %d, want %d", len(allEffects), expectedCount)
	}

	// Verify effect types
	for _, e := range allEffects {
		switch e.(type) {
		case effects.FileEffect, effects.GitEffect, effects.PersistEffect:
			// OK
		default:
			t.Errorf("Unexpected effect type: %T", e)
		}
	}
}

func TestOpenGrovePlan_Effects(t *testing.T) {
	input := OpenGrovePlanInput{
		GroveID:         "GROVE-001",
		GroveName:       "test",
		GrovePath:       "/tmp/test",
		SessionName:     "test-session",
		NextWindowIndex: 1,
	}

	plan := GenerateOpenGrovePlan(input)
	allEffects := plan.Effects()

	if len(allEffects) != len(plan.TMuxOps) {
		t.Errorf("Effects() returned %d, want %d", len(allEffects), len(plan.TMuxOps))
	}

	// All effects should be TMuxEffect
	for i, e := range allEffects {
		if _, ok := e.(effects.TMuxEffect); !ok {
			t.Errorf("Effect %d is %T, want TMuxEffect", i, e)
		}
	}
}

func TestGenerateCreateGrovePlan_PathFormatting(t *testing.T) {
	tests := []struct {
		name         string
		missionID    string
		groveName    string
		basePath     string
		expectedPath string
	}{
		{
			name:         "simple names",
			missionID:    "MISSION-001",
			groveName:    "backend",
			basePath:     "/home/user/worktrees",
			expectedPath: "/home/user/worktrees/MISSION-001-backend",
		},
		{
			name:         "longer mission ID",
			missionID:    "MISSION-123",
			groveName:    "frontend-service",
			basePath:     "/tmp",
			expectedPath: "/tmp/MISSION-123-frontend-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := CreateGrovePlanInput{
				GroveID:   "GROVE-001",
				GroveName: tt.groveName,
				MissionID: tt.missionID,
				BasePath:  tt.basePath,
				Repos:     nil,
			}
			plan := GenerateCreateGrovePlan(input)
			if plan.GrovePath != tt.expectedPath {
				t.Errorf("GrovePath = %q, want %q", plan.GrovePath, tt.expectedPath)
			}
		})
	}
}
