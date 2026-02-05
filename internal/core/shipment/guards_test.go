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
			name: "can complete unpinned shipment with all tasks complete",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "complete"},
					{ID: "TASK-002", Status: "complete"},
				},
			},
			wantAllowed: true,
		},
		{
			name: "can complete shipment with no tasks",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks:      []TaskSummary{},
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete pinned shipment",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   true,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "complete"},
				},
			},
			wantAllowed: false,
			wantReason:  "cannot complete pinned shipment SHIP-001. Unpin first with: orc shipment unpin SHIP-001",
		},
		{
			name: "cannot complete shipment with incomplete tasks",
			ctx: CompleteShipmentContext{
				ShipmentID: "SHIP-001",
				IsPinned:   false,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "complete"},
					{ID: "TASK-002", Status: "ready"},
					{ID: "TASK-003", Status: "in_progress"},
				},
			},
			wantAllowed: false,
			wantReason:  "cannot complete shipment: 2 task(s) incomplete (TASK-002, TASK-003). Use --force to complete anyway",
		},
		{
			name: "can force complete shipment with incomplete tasks",
			ctx: CompleteShipmentContext{
				ShipmentID:      "SHIP-001",
				IsPinned:        false,
				ForceCompletion: true,
				Tasks: []TaskSummary{
					{ID: "TASK-001", Status: "complete"},
					{ID: "TASK-002", Status: "ready"},
				},
			},
			wantAllowed: true,
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
			name: "can pause implementing shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "implementing",
			},
			wantAllowed: true,
		},
		{
			name: "can pause auto_implementing shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "auto_implementing",
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
			wantReason:  "can only pause implementing shipments (current status: paused)",
		},
		{
			name: "cannot pause complete shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "complete",
			},
			wantAllowed: false,
			wantReason:  "can only pause implementing shipments (current status: complete)",
		},
		{
			name: "cannot pause draft shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "draft",
			},
			wantAllowed: false,
			wantReason:  "can only pause implementing shipments (current status: draft)",
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
			name: "cannot resume implementing shipment",
			ctx: StatusTransitionContext{
				ShipmentID: "SHIP-001",
				Status:     "implementing",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused shipments (current status: implementing)",
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

func TestCanDeployShipment(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can deploy implemented shipment with no open tasks",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "implemented",
				OpenTaskCount: 0,
			},
			wantAllowed: true,
		},
		{
			name: "can deploy implementing shipment with no open tasks",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "implementing",
				OpenTaskCount: 0,
			},
			wantAllowed: true,
		},
		{
			name: "can deploy auto_implementing shipment with no open tasks",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "auto_implementing",
				OpenTaskCount: 0,
			},
			wantAllowed: true,
		},
		{
			name: "can deploy complete shipment with no open tasks",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "complete",
				OpenTaskCount: 0,
			},
			wantAllowed: true,
		},
		{
			name: "cannot deploy shipment with open tasks",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "implemented",
				OpenTaskCount: 2,
			},
			wantAllowed: false,
			wantReason:  "cannot deploy: 2 task(s) still open",
		},
		{
			name: "cannot deploy exploring shipment",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "exploring",
				OpenTaskCount: 0,
			},
			wantAllowed: false,
			wantReason:  "can only deploy implementing/implemented shipments (current status: exploring)",
		},
		{
			name: "cannot deploy tasked shipment",
			ctx: StatusTransitionContext{
				ShipmentID:    "SHIP-001",
				Status:        "tasked",
				OpenTaskCount: 0,
			},
			wantAllowed: false,
			wantReason:  "can only deploy implementing/implemented shipments (current status: tasked)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeployShipment(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGetAutoTransitionStatus(t *testing.T) {
	tests := []struct {
		name       string
		ctx        AutoTransitionContext
		wantStatus string
	}{
		// Focus transitions
		{
			name: "focus on draft shipment transitions to exploring",
			ctx: AutoTransitionContext{
				CurrentStatus: "draft",
				TriggerEvent:  "focus",
			},
			wantStatus: "exploring",
		},
		{
			name: "focus on exploring shipment does not transition",
			ctx: AutoTransitionContext{
				CurrentStatus: "exploring",
				TriggerEvent:  "focus",
			},
			wantStatus: "",
		},
		// Task created transitions
		{
			name: "first task created on exploring shipment transitions to tasked",
			ctx: AutoTransitionContext{
				CurrentStatus: "exploring",
				TriggerEvent:  "task_created",
			},
			wantStatus: "tasked",
		},
		{
			name: "task created on tasked shipment does not transition",
			ctx: AutoTransitionContext{
				CurrentStatus: "tasked",
				TriggerEvent:  "task_created",
			},
			wantStatus: "",
		},
		// Task claimed transitions
		{
			name: "task claimed on tasked shipment transitions to implementing",
			ctx: AutoTransitionContext{
				CurrentStatus: "tasked",
				TriggerEvent:  "task_claimed",
			},
			wantStatus: "implementing",
		},
		{
			name: "task claimed on ready_for_imp shipment transitions to implementing",
			ctx: AutoTransitionContext{
				CurrentStatus: "ready_for_imp",
				TriggerEvent:  "task_claimed",
			},
			wantStatus: "implementing",
		},
		{
			name: "task claimed on implementing shipment does not transition",
			ctx: AutoTransitionContext{
				CurrentStatus: "implementing",
				TriggerEvent:  "task_claimed",
			},
			wantStatus: "",
		},
		// Deploy transitions
		{
			name: "deploy from implementing transitions to deployed",
			ctx: AutoTransitionContext{
				CurrentStatus: "implementing",
				TriggerEvent:  "deploy",
			},
			wantStatus: "deployed",
		},
		{
			name: "deploy from auto_implementing transitions to deployed",
			ctx: AutoTransitionContext{
				CurrentStatus: "auto_implementing",
				TriggerEvent:  "deploy",
			},
			wantStatus: "deployed",
		},
		{
			name: "deploy from implemented transitions to deployed",
			ctx: AutoTransitionContext{
				CurrentStatus: "implemented",
				TriggerEvent:  "deploy",
			},
			wantStatus: "deployed",
		},
		{
			name: "deploy from complete (legacy) transitions to deployed",
			ctx: AutoTransitionContext{
				CurrentStatus: "complete",
				TriggerEvent:  "deploy",
			},
			wantStatus: "deployed",
		},
		// Verify transitions
		{
			name: "verify from deployed transitions to verified",
			ctx: AutoTransitionContext{
				CurrentStatus: "deployed",
				TriggerEvent:  "verify",
			},
			wantStatus: "verified",
		},
		{
			name: "verify from implemented does not transition",
			ctx: AutoTransitionContext{
				CurrentStatus: "implemented",
				TriggerEvent:  "verify",
			},
			wantStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAutoTransitionStatus(tt.ctx)
			if got != tt.wantStatus {
				t.Errorf("GetAutoTransitionStatus() = %q, want %q", got, tt.wantStatus)
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
