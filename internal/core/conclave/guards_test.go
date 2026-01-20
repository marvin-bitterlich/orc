package conclave

import "testing"

func TestCanCreateConclave(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateConclaveContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create conclave when mission exists",
			ctx: CreateConclaveContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create conclave when mission not found",
			ctx: CreateConclaveContext{
				MissionID:     "MISSION-999",
				MissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "mission MISSION-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateConclave(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCompleteConclave(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CompleteConclaveContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned conclave",
			ctx: CompleteConclaveContext{
				ConclaveID: "CON-001",
				IsPinned:   false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete pinned conclave",
			ctx: CompleteConclaveContext{
				ConclaveID: "CON-001",
				IsPinned:   true,
			},
			wantAllowed: false,
			wantReason:  "cannot complete pinned conclave CON-001. Unpin first with: orc conclave unpin CON-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteConclave(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPauseConclave(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pause active conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "active",
			},
			wantAllowed: true,
		},
		{
			name: "cannot pause paused conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "paused",
			},
			wantAllowed: false,
			wantReason:  "can only pause active conclaves (current status: paused)",
		},
		{
			name: "cannot pause complete conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only pause active conclaves (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPauseConclave(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanResumeConclave(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can resume paused conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "paused",
			},
			wantAllowed: true,
		},
		{
			name: "cannot resume active conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "active",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused conclaves (current status: active)",
		},
		{
			name: "cannot resume complete conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused conclaves (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanResumeConclave(tt.ctx)
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
