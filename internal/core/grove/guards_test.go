package grove

import "testing"

func TestCanCreateGrove(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateGroveContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "ORC can create grove with existing mission",
			ctx: CreateGroveContext{
				GuardContext: GuardContext{
					AgentType: AgentTypeORC,
					AgentID:   "ORC",
					MissionID: "MISSION-001",
				},
				MissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "IMP cannot create groves",
			ctx: CreateGroveContext{
				GuardContext: GuardContext{
					AgentType: AgentTypeIMP,
					AgentID:   "IMP-GROVE-001",
					MissionID: "MISSION-001",
				},
				MissionExists: true,
			},
			wantAllowed: false,
			wantReason:  "IMPs cannot create groves - only ORC can create groves (agent: IMP-GROVE-001)",
		},
		{
			name: "cannot create grove for non-existent mission",
			ctx: CreateGroveContext{
				GuardContext: GuardContext{
					AgentType: AgentTypeORC,
					AgentID:   "ORC",
					MissionID: "MISSION-999",
				},
				MissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "cannot create grove: mission MISSION-999 not found",
		},
		{
			name: "IMP blocked even with non-existent mission",
			ctx: CreateGroveContext{
				GuardContext: GuardContext{
					AgentType: AgentTypeIMP,
					AgentID:   "IMP-GROVE-002",
					MissionID: "MISSION-999",
				},
				MissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "IMPs cannot create groves - only ORC can create groves (agent: IMP-GROVE-002)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateGrove(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanOpenGrove(t *testing.T) {
	tests := []struct {
		name        string
		ctx         OpenGroveContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can open grove when all conditions met",
			ctx: OpenGroveContext{
				GroveID:       "GROVE-001",
				GroveExists:   true,
				PathExists:    true,
				InTMuxSession: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot open non-existent grove",
			ctx: OpenGroveContext{
				GroveID:       "GROVE-999",
				GroveExists:   false,
				PathExists:    true,
				InTMuxSession: true,
			},
			wantAllowed: false,
			wantReason:  "grove GROVE-999 not found",
		},
		{
			name: "cannot open grove without path",
			ctx: OpenGroveContext{
				GroveID:       "GROVE-001",
				GroveExists:   true,
				PathExists:    false,
				InTMuxSession: true,
			},
			wantAllowed: false,
			wantReason:  "grove worktree not found - run 'orc grove create' to materialize",
		},
		{
			name: "cannot open grove outside TMux",
			ctx: OpenGroveContext{
				GroveID:       "GROVE-001",
				GroveExists:   true,
				PathExists:    true,
				InTMuxSession: false,
			},
			wantAllowed: false,
			wantReason:  "not in a TMux session - run this command from within a TMux session",
		},
		{
			name: "grove not found checked first",
			ctx: OpenGroveContext{
				GroveID:       "GROVE-999",
				GroveExists:   false,
				PathExists:    false,
				InTMuxSession: false,
			},
			wantAllowed: false,
			wantReason:  "grove GROVE-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanOpenGrove(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanDeleteGrove(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteGroveContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete grove with no active tasks",
			ctx: DeleteGroveContext{
				GroveID:         "GROVE-001",
				ActiveTaskCount: 0,
				ForceDelete:     false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot delete grove with active tasks without force",
			ctx: DeleteGroveContext{
				GroveID:         "GROVE-001",
				ActiveTaskCount: 3,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "grove GROVE-001 has 3 active tasks. Use --force to delete anyway",
		},
		{
			name: "can force delete grove with active tasks",
			ctx: DeleteGroveContext{
				GroveID:         "GROVE-001",
				ActiveTaskCount: 3,
				ForceDelete:     true,
			},
			wantAllowed: true,
		},
		{
			name: "force flag works with many tasks",
			ctx: DeleteGroveContext{
				GroveID:         "GROVE-002",
				ActiveTaskCount: 100,
				ForceDelete:     true,
			},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteGrove(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanRenameGrove(t *testing.T) {
	tests := []struct {
		name        string
		groveExists bool
		groveID     string
		wantAllowed bool
		wantReason  string
	}{
		{
			name:        "can rename existing grove",
			groveExists: true,
			groveID:     "GROVE-001",
			wantAllowed: true,
		},
		{
			name:        "cannot rename non-existent grove",
			groveExists: false,
			groveID:     "GROVE-999",
			wantAllowed: false,
			wantReason:  "grove GROVE-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanRenameGrove(tt.groveExists, tt.groveID)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGuardResult_Error(t *testing.T) {
	t.Run("allowed result returns nil error", func(t *testing.T) {
		result := GuardResult{Allowed: true}
		if err := result.Error(); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("not allowed result returns error with reason", func(t *testing.T) {
		result := GuardResult{Allowed: false, Reason: "test reason"}
		err := result.Error()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "test reason" {
			t.Errorf("error = %q, want %q", err.Error(), "test reason")
		}
	})
}
