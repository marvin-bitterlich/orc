package workbench

import (
	"testing"
)

func TestCanCreateWorkbench(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateWorkbenchContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create workbench when workshop exists",
			ctx: CreateWorkbenchContext{
				GuardContext: GuardContext{
					AgentType:  AgentTypeORC,
					AgentID:    "ORC",
					WorkshopID: "WORK-001",
				},
				WorkshopExists: true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot create workbench when workshop does not exist",
			ctx: CreateWorkbenchContext{
				GuardContext: GuardContext{
					AgentType:  AgentTypeORC,
					AgentID:    "ORC",
					WorkshopID: "WORK-999",
				},
				WorkshopExists: false,
			},
			wantAllowed: false,
			wantReason:  "cannot create workbench: workshop WORK-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateWorkbench(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanCreateWorkbench() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanCreateWorkbench() Reason = %q, want %q", result.Reason, tt.wantReason)
			}

			// Test Error() method
			err := result.Error()
			if tt.wantAllowed && err != nil {
				t.Errorf("CanCreateWorkbench().Error() = %v, want nil", err)
			}
			if !tt.wantAllowed && err == nil {
				t.Error("CanCreateWorkbench().Error() = nil, want error")
			}
		})
	}
}

func TestCanDeleteWorkbench(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteWorkbenchContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete workbench with no active tasks",
			ctx: DeleteWorkbenchContext{
				WorkbenchID:     "BENCH-001",
				ActiveTaskCount: 0,
				ForceDelete:     false,
			},
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name: "cannot delete workbench with active tasks without force",
			ctx: DeleteWorkbenchContext{
				WorkbenchID:     "BENCH-002",
				ActiveTaskCount: 3,
				ForceDelete:     false,
			},
			wantAllowed: false,
			wantReason:  "workbench BENCH-002 has 3 active tasks. Use --force to delete anyway",
		},
		{
			name: "can force delete workbench with active tasks",
			ctx: DeleteWorkbenchContext{
				WorkbenchID:     "BENCH-003",
				ActiveTaskCount: 5,
				ForceDelete:     true,
			},
			wantAllowed: true,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteWorkbench(tt.ctx)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanDeleteWorkbench() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanDeleteWorkbench() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanRenameWorkbench(t *testing.T) {
	tests := []struct {
		name            string
		workbenchExists bool
		workbenchID     string
		wantAllowed     bool
		wantReason      string
	}{
		{
			name:            "can rename existing workbench",
			workbenchExists: true,
			workbenchID:     "BENCH-001",
			wantAllowed:     true,
			wantReason:      "",
		},
		{
			name:            "cannot rename non-existent workbench",
			workbenchExists: false,
			workbenchID:     "BENCH-999",
			wantAllowed:     false,
			wantReason:      "workbench BENCH-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanRenameWorkbench(tt.workbenchExists, tt.workbenchID)

			if result.Allowed != tt.wantAllowed {
				t.Errorf("CanRenameWorkbench() Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}

			if result.Reason != tt.wantReason {
				t.Errorf("CanRenameWorkbench() Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

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
