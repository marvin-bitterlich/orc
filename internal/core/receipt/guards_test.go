package receipt

import "testing"

func TestCanCreateREC(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateRECContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create REC when shipment exists and has no REC",
			ctx: CreateRECContext{
				ShipmentID:       "SHIP-001",
				ShipmentExists:   true,
				ShipmentHasREC:   false,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: true,
		},
		{
			name: "cannot create REC when shipment not found",
			ctx: CreateRECContext{
				ShipmentID:       "SHIP-999",
				ShipmentExists:   false,
				ShipmentHasREC:   false,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
		{
			name: "cannot create REC when shipment already has REC",
			ctx: CreateRECContext{
				ShipmentID:       "SHIP-001",
				ShipmentExists:   true,
				ShipmentHasREC:   true,
				DeliveredOutcome: "Delivered complete feature set",
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-001 already has a REC",
		},
		{
			name: "cannot create REC with empty delivered outcome",
			ctx: CreateRECContext{
				ShipmentID:       "SHIP-001",
				ShipmentExists:   true,
				ShipmentHasREC:   false,
				DeliveredOutcome: "",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
		{
			name: "cannot create REC with whitespace-only delivered outcome",
			ctx: CreateRECContext{
				ShipmentID:       "SHIP-001",
				ShipmentExists:   true,
				ShipmentHasREC:   false,
				DeliveredOutcome: "   ",
			},
			wantAllowed: false,
			wantReason:  "delivered outcome cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateREC(tt.ctx)
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
			name: "can submit draft REC when WO is complete and all CRECs verified",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "draft",
				WOExists:         true,
				WOStatus:         "complete",
				AllCRECsVerified: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot submit submitted REC",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "submitted",
				WOExists:         true,
				WOStatus:         "complete",
				AllCRECsVerified: true,
			},
			wantAllowed: false,
			wantReason:  "can only submit draft RECs (current status: submitted)",
		},
		{
			name: "cannot submit verified REC",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "verified",
				WOExists:         true,
				WOStatus:         "complete",
				AllCRECsVerified: true,
			},
			wantAllowed: false,
			wantReason:  "can only submit draft RECs (current status: verified)",
		},
		{
			name: "cannot submit REC when WO is active",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "draft",
				WOExists:         true,
				WOStatus:         "active",
				AllCRECsVerified: true,
			},
			wantAllowed: false,
			wantReason:  "cannot submit REC: Work Order is not complete (status: active)",
		},
		{
			name: "cannot submit REC when WO is draft",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "draft",
				WOExists:         true,
				WOStatus:         "draft",
				AllCRECsVerified: true,
			},
			wantAllowed: false,
			wantReason:  "cannot submit REC: Work Order is not complete (status: draft)",
		},
		{
			name: "cannot submit REC when WO does not exist",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "draft",
				WOExists:         false,
				WOStatus:         "",
				AllCRECsVerified: true,
			},
			wantAllowed: false,
			wantReason:  "cannot submit REC: Work Order not found",
		},
		{
			name: "cannot submit REC when some CRECs not verified",
			ctx: StatusTransitionContext{
				RECID:            "REC-001",
				CurrentStatus:    "draft",
				WOExists:         true,
				WOStatus:         "complete",
				AllCRECsVerified: false,
			},
			wantAllowed: false,
			wantReason:  "cannot submit REC: not all CRECs are verified",
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
			name: "can verify submitted REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "submitted",
			},
			wantAllowed: true,
		},
		{
			name: "cannot verify draft REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted RECs (current status: draft)",
		},
		{
			name: "cannot verify already verified REC",
			ctx: StatusTransitionContext{
				RECID:         "REC-001",
				CurrentStatus: "verified",
			},
			wantAllowed: false,
			wantReason:  "can only verify submitted RECs (current status: verified)",
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
