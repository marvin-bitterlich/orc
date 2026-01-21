package repo

import "testing"

func TestCanCreateRepo(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateRepoContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create repo with unique name",
			ctx: CreateRepoContext{
				Name:       "my-repo",
				NameExists: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create repo with empty name",
			ctx: CreateRepoContext{
				Name:       "",
				NameExists: false,
			},
			wantAllowed: false,
			wantReason:  "repository name cannot be empty",
		},
		{
			name: "cannot create repo with whitespace-only name",
			ctx: CreateRepoContext{
				Name:       "   ",
				NameExists: false,
			},
			wantAllowed: false,
			wantReason:  "repository name cannot be empty",
		},
		{
			name: "cannot create repo with duplicate name",
			ctx: CreateRepoContext{
				Name:       "existing-repo",
				NameExists: true,
			},
			wantAllowed: false,
			wantReason:  `repository with name "existing-repo" already exists`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateRepo(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanArchiveRepo(t *testing.T) {
	tests := []struct {
		name        string
		ctx         ArchiveRepoContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can archive active repo",
			ctx: ArchiveRepoContext{
				RepoID: "REPO-001",
				Status: "active",
			},
			wantAllowed: true,
		},
		{
			name: "cannot archive archived repo",
			ctx: ArchiveRepoContext{
				RepoID: "REPO-001",
				Status: "archived",
			},
			wantAllowed: false,
			wantReason:  "can only archive active repositories (current status: archived)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanArchiveRepo(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanRestoreRepo(t *testing.T) {
	tests := []struct {
		name        string
		ctx         RestoreRepoContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can restore archived repo",
			ctx: RestoreRepoContext{
				RepoID: "REPO-001",
				Status: "archived",
			},
			wantAllowed: true,
		},
		{
			name: "cannot restore active repo",
			ctx: RestoreRepoContext{
				RepoID: "REPO-001",
				Status: "active",
			},
			wantAllowed: false,
			wantReason:  "can only restore archived repositories (current status: active)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanRestoreRepo(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanDeleteRepo(t *testing.T) {
	tests := []struct {
		name        string
		ctx         DeleteRepoContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can delete repo with no active PRs",
			ctx: DeleteRepoContext{
				RepoID:       "REPO-001",
				HasActivePRs: false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot delete repo with active PRs",
			ctx: DeleteRepoContext{
				RepoID:       "REPO-001",
				HasActivePRs: true,
			},
			wantAllowed: false,
			wantReason:  "cannot delete repository REPO-001 with active pull requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteRepo(tt.ctx)
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
