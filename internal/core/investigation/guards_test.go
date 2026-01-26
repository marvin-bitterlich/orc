package investigation

import "testing"

func TestCanCreateInvestigation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateInvestigationContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create investigation when commission exists",
			ctx: CreateInvestigationContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create investigation when commission not found",
			ctx: CreateInvestigationContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateInvestigation(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCompleteInvestigation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CompleteInvestigationContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned investigation",
			ctx: CompleteInvestigationContext{
				InvestigationID: "INV-001",
				IsPinned:        false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete pinned investigation",
			ctx: CompleteInvestigationContext{
				InvestigationID: "INV-001",
				IsPinned:        true,
			},
			wantAllowed: false,
			wantReason:  "cannot complete pinned investigation INV-001. Unpin first with: orc investigation unpin INV-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteInvestigation(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPauseInvestigation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pause active investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "active",
			},
			wantAllowed: true,
		},
		{
			name: "cannot pause paused investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "paused",
			},
			wantAllowed: false,
			wantReason:  "can only pause active investigations (current status: paused)",
		},
		{
			name: "cannot pause complete investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "complete",
			},
			wantAllowed: false,
			wantReason:  "can only pause active investigations (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPauseInvestigation(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanResumeInvestigation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can resume paused investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "paused",
			},
			wantAllowed: true,
		},
		{
			name: "cannot resume active investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "active",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused investigations (current status: active)",
		},
		{
			name: "cannot resume complete investigation",
			ctx: StatusTransitionContext{
				InvestigationID: "INV-001",
				Status:          "complete",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused investigations (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanResumeInvestigation(tt.ctx)
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
