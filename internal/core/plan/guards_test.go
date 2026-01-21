package plan

import "testing"

func TestCanCreatePlan(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreatePlanContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create plan when mission exists and no shipment",
			ctx: CreatePlanContext{
				MissionID:     "MISSION-001",
				MissionExists: true,
				ShipmentID:    "",
			},
			wantAllowed: true,
		},
		{
			name: "can create plan when mission exists with shipment and no active plan",
			ctx: CreatePlanContext{
				MissionID:             "MISSION-001",
				MissionExists:         true,
				ShipmentID:            "SHIP-001",
				ShipmentExists:        true,
				ShipmentHasActivePlan: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create plan when mission not found",
			ctx: CreatePlanContext{
				MissionID:     "MISSION-999",
				MissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "mission MISSION-999 not found",
		},
		{
			name: "cannot create plan when shipment not found",
			ctx: CreatePlanContext{
				MissionID:      "MISSION-001",
				MissionExists:  true,
				ShipmentID:     "SHIP-999",
				ShipmentExists: false,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
		{
			name: "cannot create plan when shipment already has active plan",
			ctx: CreatePlanContext{
				MissionID:             "MISSION-001",
				MissionExists:         true,
				ShipmentID:            "SHIP-001",
				ShipmentExists:        true,
				ShipmentHasActivePlan: true,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-001 already has an active plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreatePlan(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanApprovePlan(t *testing.T) {
	tests := []struct {
		name        string
		ctx         ApprovePlanContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can approve draft unpinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "draft",
				IsPinned: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot approve draft pinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "draft",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "cannot approve pinned plan PLAN-001. Unpin first with: orc plan unpin PLAN-001",
		},
		{
			name: "cannot approve already approved plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "approved",
				IsPinned: false,
			},
			wantAllowed: false,
			wantReason:  "can only approve draft plans (current status: approved)",
		},
		{
			name: "cannot approve already approved and pinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "approved",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "can only approve draft plans (current status: approved)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanApprovePlan(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanDeletePlan(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeletePlanContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete unpinned plan",
			ctx: DeletePlanContext{
				PlanID:   "PLAN-001",
				IsPinned: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot delete pinned plan",
			ctx: DeletePlanContext{
				PlanID:   "PLAN-001",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "cannot delete pinned plan PLAN-001. Unpin first with: orc plan unpin PLAN-001",
		},
		{
			name: "can delete approved unpinned plan",
			ctx: DeletePlanContext{
				PlanID:   "PLAN-001",
				IsPinned: false,
			},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeletePlan(tt.ctx)
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
