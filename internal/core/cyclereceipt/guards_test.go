package cyclereceipt

import "testing"

func TestCanCreateCREC(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateCRECContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create CREC when CWO exists and has no CREC",
			ctx: CreateCRECContext{
				CWOID:            "CWO-001",
				CWOExists:        true,
				CWOHasCREC:       false,
				DeliveredOutcome: "Implemented feature X",
			},
			wantAllowed: true,
		},
		{
			name: "cannot create CREC when CWO not found",
			ctx: CreateCRECContext{
				CWOID:            "CWO-999",
				CWOExists:        false,
				CWOHasCREC:       false,
				DeliveredOutcome: "Implemented feature X",
			},
			wantAllowed: false,
			wantReason:  "CWO CWO-999 not found",
		},
		{
			name: "cannot create CREC when CWO already has CREC",
			ctx: CreateCRECContext{
				CWOID:            "CWO-001",
				CWOExists:        true,
				CWOHasCREC:       true,
				DeliveredOutcome: "Implemented feature X",
			},
			wantAllowed: false,
			wantReason:  "CWO CWO-001 already has a CREC",
		},
		{
			name: "cannot create CREC with empty delivered outcome",
			ctx: CreateCRECContext{
				CWOID:            "CWO-001",
				CWOExists:        true,
				CWOHasCREC:       false,
				DeliveredOutcome: "",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
		{
			name: "cannot create CREC with whitespace-only delivered outcome",
			ctx: CreateCRECContext{
				CWOID:            "CWO-001",
				CWOExists:        true,
				CWOHasCREC:       false,
				DeliveredOutcome: "   ",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateCREC(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanSubmit(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can submit draft CREC when CWO is complete",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "draft",
				CWOExists:     true,
				CWOStatus:     "complete",
			},
			wantAllowed: true,
		},
		{
			name: "cannot submit submitted CREC",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "submitted",
				CWOExists:     true,
				CWOStatus:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only submit draft CRECs (current status: submitted)",
		},
		{
			name: "cannot submit verified CREC",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "verified",
				CWOExists:     true,
				CWOStatus:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only submit draft CRECs (current status: verified)",
		},
		{
			name: "cannot submit CREC when CWO is active",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "draft",
				CWOExists:     true,
				CWOStatus:     "active",
			},
			wantAllowed: false,
			wantReason:  "cannot submit CREC: parent CWO is not complete (status: active)",
		},
		{
			name: "cannot submit CREC when CWO is draft",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "draft",
				CWOExists:     true,
				CWOStatus:     "draft",
			},
			wantAllowed: false,
			wantReason:  "cannot submit CREC: parent CWO is not complete (status: draft)",
		},
		{
			name: "cannot submit CREC when CWO does not exist",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "draft",
				CWOExists:     false,
				CWOStatus:     "",
			},
			wantAllowed: false,
			wantReason:  "cannot submit CREC: parent CWO no longer exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanSubmit(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanVerify(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can verify submitted CREC",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "submitted",
			},
			wantAllowed: true,
		},
		{
			name: "cannot verify draft CREC",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted CRECs (current status: draft)",
		},
		{
			name: "cannot verify already verified CREC",
			ctx: StatusTransitionContext{
				CRECID:        "CREC-001",
				CurrentStatus: "verified",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted CRECs (current status: verified)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanVerify(tt.ctx)
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
