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
			name: "can create plan when commission exists and no shipment",
			ctx: CreatePlanContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				ShipmentID:       "",
			},
			wantAllowed: true,
		},
		{
			name: "can create plan when commission exists with shipment and no active plan",
			ctx: CreatePlanContext{
				CommissionID:          "COMM-001",
				CommissionExists:      true,
				ShipmentID:            "SHIP-001",
				ShipmentExists:        true,
				ShipmentHasActivePlan: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create plan when commission not found",
			ctx: CreatePlanContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
		{
			name: "cannot create plan when shipment not found",
			ctx: CreatePlanContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				ShipmentID:       "SHIP-999",
				ShipmentExists:   false,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
		{
			name: "cannot create plan when shipment already has active plan",
			ctx: CreatePlanContext{
				CommissionID:          "COMM-001",
				CommissionExists:      true,
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

func TestCanSubmitPlan(t *testing.T) {
	tests := []struct {
		name        string
		ctx         SubmitPlanContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can submit draft plan with content",
			ctx: SubmitPlanContext{
				PlanID:     "PLAN-001",
				Status:     "draft",
				HasContent: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot submit draft plan without content",
			ctx: SubmitPlanContext{
				PlanID:     "PLAN-001",
				Status:     "draft",
				HasContent: false,
			},
			wantAllowed: false,
			wantReason:  "cannot submit plan without content",
		},
		{
			name: "cannot submit pending_review plan",
			ctx: SubmitPlanContext{
				PlanID:     "PLAN-001",
				Status:     "pending_review",
				HasContent: true,
			},
			wantAllowed: false,
			wantReason:  "can only submit draft plans (current status: pending_review)",
		},
		{
			name: "cannot submit approved plan",
			ctx: SubmitPlanContext{
				PlanID:     "PLAN-001",
				Status:     "approved",
				HasContent: true,
			},
			wantAllowed: false,
			wantReason:  "can only submit draft plans (current status: approved)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanSubmitPlan(tt.ctx)
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
			name: "can approve pending_review unpinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "pending_review",
				IsPinned: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot approve pending_review pinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "pending_review",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "cannot approve pinned plan PLAN-001. Unpin first with: orc plan unpin PLAN-001",
		},
		{
			name: "cannot approve draft plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "draft",
				IsPinned: false,
			},
			wantAllowed: false,
			wantReason:  "can only approve plans pending review (current status: draft)",
		},
		{
			name: "cannot approve already approved plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "approved",
				IsPinned: false,
			},
			wantAllowed: false,
			wantReason:  "can only approve plans pending review (current status: approved)",
		},
		{
			name: "cannot approve already approved and pinned plan",
			ctx: ApprovePlanContext{
				PlanID:   "PLAN-001",
				Status:   "approved",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "can only approve plans pending review (current status: approved)",
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

func TestCanEscalatePlan(t *testing.T) {
	tests := []struct {
		name        string
		ctx         EscalatePlanContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can escalate draft plan with content and reason",
			ctx: EscalatePlanContext{
				PlanID:     "PLAN-001",
				Status:     "draft",
				HasContent: true,
				HasReason:  true,
			},
			wantAllowed: true,
		},
		{
			name: "can escalate pending_review plan with content and reason",
			ctx: EscalatePlanContext{
				PlanID:     "PLAN-001",
				Status:     "pending_review",
				HasContent: true,
				HasReason:  true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot escalate approved plan",
			ctx: EscalatePlanContext{
				PlanID:     "PLAN-001",
				Status:     "approved",
				HasContent: true,
				HasReason:  true,
			},
			wantAllowed: false,
			wantReason:  "can only escalate draft or pending_review plans (current status: approved)",
		},
		{
			name: "cannot escalate plan without content",
			ctx: EscalatePlanContext{
				PlanID:     "PLAN-001",
				Status:     "draft",
				HasContent: false,
				HasReason:  true,
			},
			wantAllowed: false,
			wantReason:  "cannot escalate plan without content",
		},
		{
			name: "cannot escalate plan without reason",
			ctx: EscalatePlanContext{
				PlanID:     "PLAN-001",
				Status:     "draft",
				HasContent: true,
				HasReason:  false,
			},
			wantAllowed: false,
			wantReason:  "escalation reason is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanEscalatePlan(tt.ctx)
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
