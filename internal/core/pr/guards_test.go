package pr

import "testing"

func TestCanCreatePR(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreatePRContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create PR for active shipment",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-001",
				RepoID:         "REPO-001",
				ShipmentExists: true,
				ShipmentStatus: "active",
				ShipmentHasPR:  false,
				RepoExists:     true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create PR for non-existent shipment",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-999",
				RepoID:         "REPO-001",
				ShipmentExists: false,
				ShipmentStatus: "",
				ShipmentHasPR:  false,
				RepoExists:     true,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-999 not found",
		},
		{
			name: "cannot create PR for paused shipment",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-001",
				RepoID:         "REPO-001",
				ShipmentExists: true,
				ShipmentStatus: "paused",
				ShipmentHasPR:  false,
				RepoExists:     true,
			},
			wantAllowed: false,
			wantReason:  "can only create PR for active shipments (current status: paused)",
		},
		{
			name: "cannot create PR for complete shipment",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-001",
				RepoID:         "REPO-001",
				ShipmentExists: true,
				ShipmentStatus: "complete",
				ShipmentHasPR:  false,
				RepoExists:     true,
			},
			wantAllowed: false,
			wantReason:  "can only create PR for active shipments (current status: complete)",
		},
		{
			name: "cannot create PR when shipment already has PR",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-001",
				RepoID:         "REPO-001",
				ShipmentExists: true,
				ShipmentStatus: "active",
				ShipmentHasPR:  true,
				RepoExists:     true,
			},
			wantAllowed: false,
			wantReason:  "shipment SHIP-001 already has a PR",
		},
		{
			name: "cannot create PR for non-existent repo",
			ctx: CreatePRContext{
				ShipmentID:     "SHIP-001",
				RepoID:         "REPO-999",
				ShipmentExists: true,
				ShipmentStatus: "active",
				ShipmentHasPR:  false,
				RepoExists:     false,
			},
			wantAllowed: false,
			wantReason:  "repository REPO-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreatePR(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanOpenPR(t *testing.T) {
	tests := []struct {
		name        string
		ctx         OpenPRContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can open draft PR",
			ctx: OpenPRContext{
				PRID:   "PR-001",
				Status: "draft",
			},
			wantAllowed: true,
		},
		{
			name: "cannot open already open PR",
			ctx: OpenPRContext{
				PRID:   "PR-001",
				Status: "open",
			},
			wantAllowed: false,
			wantReason:  "can only open draft PRs (current status: open)",
		},
		{
			name: "cannot open merged PR",
			ctx: OpenPRContext{
				PRID:   "PR-001",
				Status: "merged",
			},
			wantAllowed: false,
			wantReason:  "can only open draft PRs (current status: merged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanOpenPR(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanApprovePR(t *testing.T) {
	tests := []struct {
		name        string
		ctx         ApprovePRContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can approve open PR",
			ctx: ApprovePRContext{
				PRID:   "PR-001",
				Status: "open",
			},
			wantAllowed: true,
		},
		{
			name: "cannot approve draft PR",
			ctx: ApprovePRContext{
				PRID:   "PR-001",
				Status: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only approve open PRs (current status: draft)",
		},
		{
			name: "cannot approve merged PR",
			ctx: ApprovePRContext{
				PRID:   "PR-001",
				Status: "merged",
			},
			wantAllowed: false,
			wantReason:  "can only approve open PRs (current status: merged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanApprovePR(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanMergePR(t *testing.T) {
	tests := []struct {
		name        string
		ctx         MergePRContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can merge open PR",
			ctx: MergePRContext{
				PRID:   "PR-001",
				Status: "open",
			},
			wantAllowed: true,
		},
		{
			name: "can merge approved PR",
			ctx: MergePRContext{
				PRID:   "PR-001",
				Status: "approved",
			},
			wantAllowed: true,
		},
		{
			name: "cannot merge draft PR",
			ctx: MergePRContext{
				PRID:   "PR-001",
				Status: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only merge open or approved PRs (current status: draft)",
		},
		{
			name: "cannot merge closed PR",
			ctx: MergePRContext{
				PRID:   "PR-001",
				Status: "closed",
			},
			wantAllowed: false,
			wantReason:  "can only merge open or approved PRs (current status: closed)",
		},
		{
			name: "cannot merge already merged PR",
			ctx: MergePRContext{
				PRID:   "PR-001",
				Status: "merged",
			},
			wantAllowed: false,
			wantReason:  "can only merge open or approved PRs (current status: merged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanMergePR(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanClosePR(t *testing.T) {
	tests := []struct {
		name        string
		ctx         ClosePRContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can close open PR",
			ctx: ClosePRContext{
				PRID:   "PR-001",
				Status: "open",
			},
			wantAllowed: true,
		},
		{
			name: "can close approved PR",
			ctx: ClosePRContext{
				PRID:   "PR-001",
				Status: "approved",
			},
			wantAllowed: true,
		},
		{
			name: "cannot close draft PR",
			ctx: ClosePRContext{
				PRID:   "PR-001",
				Status: "draft",
			},
			wantAllowed: false,
			wantReason:  "can only close open or approved PRs (current status: draft)",
		},
		{
			name: "cannot close merged PR",
			ctx: ClosePRContext{
				PRID:   "PR-001",
				Status: "merged",
			},
			wantAllowed: false,
			wantReason:  "can only close open or approved PRs (current status: merged)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanClosePR(tt.ctx)
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
