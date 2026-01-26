package shipment

import "testing"

func TestCanCreateShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateShipmentContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create shipment when commission exists",
			ctx: CreateShipmentContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create shipment when commission not found",
			ctx: CreateShipmentContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCompleteShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CompleteShipmentContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned shipment",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete pinned shipment",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   true,
			},
			wantAllowed: false,
			wantReason:  "cannot complete pinned shipment SHIP-001. Unpin first with: orc shipment unpin SHIP-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPauseShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pause active shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "active",
			},
			wantAllowed: true,
		},
		{
			name: "cannot pause paused shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "paused",
			},
			wantAllowed: false,
			wantReason:  "can only pause active shipments (current status: paused)",
		},
		{
			name: "cannot pause complete shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only pause active shipments (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPauseShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanResumeShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can resume paused shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "paused",
			},
			wantAllowed: true,
		},
		{
			name: "cannot resume active shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "active",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused shipments (current status: active)",
		},
		{
			name: "cannot resume complete shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused shipments (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanResumeShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanAssignWorkbench(t *testing.T) {
	tests := []struct {
		name        string
		ctx         AssignWorkbenchContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can assign unassigned workbench",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "",
			},
			wantAllowed: true,
		},
		{
			name: "can assign workbench already assigned to same shipment (idempotent)",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "SHIP-001",
			},
			wantAllowed: true,
		},
		{
			name: "cannot assign workbench assigned to another shipment",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-001",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        true,
				WorkbenchAssignedToID: "SHIP-002",
			},
			wantAllowed: false,
			wantReason:  "workbench already assigned to shipment SHIP-002",
		},
		{
			name: "cannot assign workbench to non-existent shipment",
			ctx: AssignWorkbenchContext{
				ShipmentID:            "SHIP-999",
				WorkbenchID:           "BENCH-001",
				ShipmentExists:        false,
				WorkbenchAssignedToID: "",
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAssignWorkbench(tt.ctx)
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
