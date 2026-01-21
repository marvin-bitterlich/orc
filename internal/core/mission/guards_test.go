package mission

import (
	"testing"
)

func TestCanCreateMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         GuardContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "ORC can create missions",
			ctx: GuardContext{
				AgentType: AgentTypeORC,
				AgentID:   "ORC",
				MissionID: "",
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "ORC in mission context can create missions",
			ctx: GuardContext{
				AgentType: AgentTypeORC,
				AgentID:   "ORC",
				MissionID: "MISSION-001",
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "IMP cannot create missions",
			ctx: GuardContext{
				AgentType: AgentTypeIMP,
				AgentID:   "IMP-GROVE-001",
				MissionID: "MISSION-001",
			},
			wantAllowed: false,
			wantReason:  "IMPs cannot create missions - only ORC can create missions (agent: IMP-GROVE-001)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCreateMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCreateMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			// Test Error() method
			err := result.Error()
			if tt.wantAllowed && err != nil {
				t.Errorf("CanCreateMission().Error() = %v, want nil", err)
			}
			if !tt.wantAllowed && err == nil {
				t.Error("CanCreateMission().Error() = nil, want error")
			}
		})
	}
}

func TestCanStartMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         GuardContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "ORC can start missions",
			ctx: GuardContext{
				AgentType: AgentTypeORC,
				AgentID:   "ORC",
				MissionID: "",
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "IMP cannot start missions",
			ctx: GuardContext{
				AgentType: AgentTypeIMP,
				AgentID:   "IMP-GROVE-002",
				MissionID: "MISSION-001",
			},
			wantAllowed: false,
			wantReason:  "IMPs cannot start missions - only ORC can start missions (agent: IMP-GROVE-002)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanStartMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanStartMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanStartMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanLaunchMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         GuardContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "ORC can launch missions",
			ctx: GuardContext{
				AgentType: AgentTypeORC,
				AgentID:   "ORC",
				MissionID: "",
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "IMP cannot launch missions",
			ctx: GuardContext{
				AgentType: AgentTypeIMP,
				AgentID:   "IMP-GROVE-003",
				MissionID: "MISSION-002",
			},
			wantAllowed: false,
			wantReason:  "IMPs cannot launch missions - only ORC can launch missions (agent: IMP-GROVE-003)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanLaunchMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanLaunchMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanLaunchMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGuardResult_Error(t *testing.T) {
	tests := []struct {
		name      string
		result    GuardResult
		wantError bool
	}{
		{
			name:      "allowed result returns nil error",
			result:    GuardResult{Allowed: true, Reason: ""},
			wantError: false,
		},
		{
			name:      "disallowed result returns error",
			result:    GuardResult{Allowed: false, Reason: "not allowed"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Error()
			if (err != nil) != tt.wantError {
				t.Errorf("GuardResult.Error() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCanCompleteMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         MissionStateContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned mission",
			ctx: MissionStateContext{
				MissionID: "MISSION-001",
				IsPinned:  false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot complete pinned mission",
			ctx: MissionStateContext{
				MissionID: "MISSION-002",
				IsPinned:  true,
			},
			wantAllowed: false,
			wantReason:  "Cannot complete pinned mission MISSION-002. Unpin first with: orc mission unpin MISSION-002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCompleteMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCompleteMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanArchiveMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         MissionStateContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can archive unpinned mission",
			ctx: MissionStateContext{
				MissionID: "MISSION-001",
				IsPinned:  false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot archive pinned mission",
			ctx: MissionStateContext{
				MissionID: "MISSION-003",
				IsPinned:  true,
			},
			wantAllowed: false,
			wantReason:  "Cannot archive pinned mission MISSION-003. Unpin first with: orc mission unpin MISSION-003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanArchiveMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanArchiveMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanArchiveMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanDeleteMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete mission with no dependents",
			ctx: DeleteContext{
				MissionID:     "MISSION-001",
				ShipmentCount: 0,
				GroveCount:    0,
				ForceDelete:   false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot delete mission with shipments without force",
			ctx: DeleteContext{
				MissionID:     "MISSION-002",
				ShipmentCount: 3,
				GroveCount:    0,
				ForceDelete:   false,
			},
			wantAllowed: false,
			wantReason:  "Mission MISSION-002 has 3 shipments and 0 groves. Use --force to delete anyway",
		},
		{
			name: "cannot delete mission with groves without force",
			ctx: DeleteContext{
				MissionID:     "MISSION-003",
				ShipmentCount: 0,
				GroveCount:    2,
				ForceDelete:   false,
			},
			wantAllowed: false,
			wantReason:  "Mission MISSION-003 has 0 shipments and 2 groves. Use --force to delete anyway",
		},
		{
			name: "cannot delete mission with both shipments and groves without force",
			ctx: DeleteContext{
				MissionID:     "MISSION-004",
				ShipmentCount: 5,
				GroveCount:    3,
				ForceDelete:   false,
			},
			wantAllowed: false,
			wantReason:  "Mission MISSION-004 has 5 shipments and 3 groves. Use --force to delete anyway",
		},
		{
			name: "can force delete mission with dependents",
			ctx: DeleteContext{
				MissionID:     "MISSION-005",
				ShipmentCount: 5,
				GroveCount:    3,
				ForceDelete:   true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanDeleteMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanDeleteMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPinMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         PinContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pin existing mission",
			ctx: PinContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
				IsPinned:      false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot pin non-existent mission",
			ctx: PinContext{
				MissionID:     "MISSION-999",
				MissionExists: false,
				IsPinned:      false,
			},
			wantAllowed: false,
			wantReason:  "Mission MISSION-999 not found",
		},
		{
			name: "pin already pinned mission is allowed (no-op)",
			ctx: PinContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
				IsPinned:      true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPinMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanPinMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanPinMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanUnpinMission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         PinContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can unpin existing pinned mission",
			ctx: PinContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
				IsPinned:      true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot unpin non-existent mission",
			ctx: PinContext{
				MissionID:     "MISSION-999",
				MissionExists: false,
				IsPinned:      false,
			},
			wantAllowed: false,
			wantReason:  "Mission MISSION-999 not found",
		},
		{
			name: "unpin already unpinned mission is allowed (no-op)",
			ctx: PinContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
				IsPinned:      false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanUnpinMission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanUnpinMission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanUnpinMission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}
