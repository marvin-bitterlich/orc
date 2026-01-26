package operation

import "testing"

func TestCanCreateOperation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateOperationContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create operation when commission exists",
			ctx: CreateOperationContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create operation when commission not found",
			ctx: CreateOperationContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateOperation(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCompleteOperation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CompleteOperationContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete operation with ready status",
			ctx: CompleteOperationContext{
				OperationID: "OP-001",
				Status:      "ready",
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete already completed operation",
			ctx: CompleteOperationContext{
				OperationID: "OP-001",
				Status:      "complete",
			},
			wantAllowed: false,
			wantReason:  "can only complete ready operations (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteOperation(tt.ctx)
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
