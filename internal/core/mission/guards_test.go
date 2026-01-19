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
