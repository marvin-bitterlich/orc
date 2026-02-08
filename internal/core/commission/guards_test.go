package commission

import (
	"testing"
)

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

func TestCanCompleteCommission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CommissionStateContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned commission",
			ctx: CommissionStateContext{
				CommissionID: "COMM-001",
				IsPinned:     false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot complete pinned commission",
			ctx: CommissionStateContext{
				CommissionID: "COMM-002",
				IsPinned:     true,
			},
			wantAllowed: false,
			wantReason:  "Cannot complete pinned commission COMM-002. Unpin first with: orc commission unpin COMM-002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteCommission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCompleteCommission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCompleteCommission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanArchiveCommission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CommissionStateContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can archive unpinned commission",
			ctx: CommissionStateContext{
				CommissionID: "COMM-001",
				IsPinned:     false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot archive pinned commission",
			ctx: CommissionStateContext{
				CommissionID: "COMM-003",
				IsPinned:     true,
			},
			wantAllowed: false,
			wantReason:  "Cannot archive pinned commission COMM-003. Unpin first with: orc commission unpin COMM-003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanArchiveCommission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanArchiveCommission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanArchiveCommission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanDeleteCommission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete commission with no dependents",
			ctx: DeleteContext{
				CommissionID:   "COMM-001",
				ShipmentCount:  0,
				WorkbenchCount: 0,
				ForceDelete:    false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot delete commission with shipments without force",
			ctx: DeleteContext{
				CommissionID:   "COMM-002",
				ShipmentCount:  3,
				WorkbenchCount: 0,
				ForceDelete:    false,
			},
			wantAllowed: false,
			wantReason:  "Commission COMM-002 has 3 shipments and 0 workbenches. Use --force to delete anyway",
		},
		{
			name: "cannot delete commission with workbenches without force",
			ctx: DeleteContext{
				CommissionID:   "COMM-003",
				ShipmentCount:  0,
				WorkbenchCount: 2,
				ForceDelete:    false,
			},
			wantAllowed: false,
			wantReason:  "Commission COMM-003 has 0 shipments and 2 workbenches. Use --force to delete anyway",
		},
		{
			name: "cannot delete commission with both shipments and workbenches without force",
			ctx: DeleteContext{
				CommissionID:   "COMM-004",
				ShipmentCount:  5,
				WorkbenchCount: 3,
				ForceDelete:    false,
			},
			wantAllowed: false,
			wantReason:  "Commission COMM-004 has 5 shipments and 3 workbenches. Use --force to delete anyway",
		},
		{
			name: "can force delete commission with dependents",
			ctx: DeleteContext{
				CommissionID:   "COMM-005",
				ShipmentCount:  5,
				WorkbenchCount: 3,
				ForceDelete:    true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteCommission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanDeleteCommission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanDeleteCommission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPinCommission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         PinContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pin existing commission",
			ctx: PinContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				IsPinned:         false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot pin non-existent commission",
			ctx: PinContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
				IsPinned:         false,
			},
			wantAllowed: false,
			wantReason:  "Commission COMM-999 not found",
		},
		{
			name: "pin already pinned commission is allowed (no-op)",
			ctx: PinContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				IsPinned:         true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPinCommission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanPinCommission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanPinCommission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanUnpinCommission(t *testing.T) {
	tests := []struct {
		name        string
		ctx         PinContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can unpin existing pinned commission",
			ctx: PinContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				IsPinned:         true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot unpin non-existent commission",
			ctx: PinContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
				IsPinned:         false,
			},
			wantAllowed: false,
			wantReason:  "Commission COMM-999 not found",
		},
		{
			name: "unpin already unpinned commission is allowed (no-op)",
			ctx: PinContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				IsPinned:         false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanUnpinCommission(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanUnpinCommission() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanUnpinCommission() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}
