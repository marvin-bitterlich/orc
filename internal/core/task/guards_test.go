package task

import "testing"

func TestCanCreateTask(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateTaskContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create task when commission exists (no shipment)",
			ctx: CreateTaskContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				ShipmentID:       "",
			},
			wantAllowed: true,
		},
		{
			name: "can create task when commission and shipment exist",
			ctx: CreateTaskContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				ShipmentID:       "SHIP-001",
				ShipmentExists:   true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create task when commission not found",
			ctx: CreateTaskContext{
				CommissionID:     "COMM-999",
				CommissionExists: false,
				ShipmentID:       "",
			},
			wantAllowed: false,
			wantReason:  "commission COMM-999 not found",
		},
		{
			name: "cannot create task when shipment not found",
			ctx: CreateTaskContext{
				CommissionID:     "COMM-001",
				CommissionExists: true,
				ShipmentID:       "SHIP-999",
				ShipmentExists:   false,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateTask(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanCompleteTask(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CompleteTaskContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can complete unpinned task",
			ctx: CompleteTaskContext{
				TaskID:   "TASK-001",
				IsPinned: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot complete pinned task",
			ctx: CompleteTaskContext{
				TaskID:   "TASK-001",
				IsPinned: true,
			},
			wantAllowed: false,
			wantReason:  "cannot complete pinned task TASK-001. Unpin first with: orc task unpin TASK-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCompleteTask(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanPauseTask(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can pause in_progress task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "in_progress",
			},
			wantAllowed: true,
		},
		{
			name: "cannot pause ready task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "ready",
			},
			wantAllowed: false,
			wantReason:  "can only pause in_progress tasks (current status: ready)",
		},
		{
			name: "cannot pause paused task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "paused",
			},
			wantAllowed: false,
			wantReason:  "can only pause in_progress tasks (current status: paused)",
		},
		{
			name: "cannot pause complete task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "complete",
			},
			wantAllowed: false,
			wantReason:  "can only pause in_progress tasks (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanPauseTask(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanResumeTask(t *testing.T) {
	tests := []struct {
		name        string
		ctx         StatusTransitionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can resume paused task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "paused",
			},
			wantAllowed: true,
		},
		{
			name: "cannot resume ready task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "ready",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused tasks (current status: ready)",
		},
		{
			name: "cannot resume in_progress task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "in_progress",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused tasks (current status: in_progress)",
		},
		{
			name: "cannot resume complete task",
			ctx: StatusTransitionContext{
				TaskID: "TASK-001",
				Status: "complete",
			},
			wantAllowed: false,
			wantReason:  "can only resume paused tasks (current status: complete)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanResumeTask(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanTagTask(t *testing.T) {
	tests := []struct {
		name        string
		ctx         TagTaskContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can tag task with no existing tag",
			ctx: TagTaskContext{
				TaskID:          "TASK-001",
				ExistingTagID:   "",
				ExistingTagName: "",
			},
			wantAllowed: true,
		},
		{
			name: "cannot tag task that already has a tag",
			ctx: TagTaskContext{
				TaskID:          "TASK-001",
				ExistingTagID:   "TAG-001",
				ExistingTagName: "bug",
			},
			wantAllowed: false,
			wantReason:  "task TASK-001 already has tag 'bug' (one tag per task limit)\nRemove existing tag first with: orc task untag TASK-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTagTask(tt.ctx)
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
