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
			name: "can create conclave when commission exists",
			ctx: CreateConclaveContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create conclave when commission not found",
			ctx: CreateConclaveContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
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
			name: "can pause open conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "open",
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
			wantReason:  "can only pause open conclaves (current status: paused)",
		},
		{
			name: "cannot pause closed conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "closed",
			},
			wantAllowed: false,
			wantReason:  "can only pause open conclaves (current status: closed)",
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
			name: "cannot resume open conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "open",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused conclaves (current status: open)",
		},
		{
			name: "cannot resume closed conclave",
			ctx: StatusTransitionContext{
				ConclaveID: "CON-001",
				Status:     "closed",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused conclaves (current status: closed)",
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
