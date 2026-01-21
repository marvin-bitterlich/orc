package question

import "testing"

func TestCanCreateQuestion(t *testing.T) {
	tests := []struct {
		name        string
		ctx         CreateQuestionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can create question when mission exists and no investigation",
			ctx: CreateQuestionContext{
				MissionID:       "MISSION-001",
				MissionExists:   true,
				InvestigationID: "",
			},
			wantAllowed: true,
		},
		{
			name: "can create question when mission exists with investigation",
			ctx: CreateQuestionContext{
				MissionID:           "MISSION-001",
				MissionExists:       true,
				InvestigationID:     "INV-001",
				InvestigationExists: true,
			},
			wantAllowed: true,
		},
		{
			name: "cannot create question when mission not found",
			ctx: CreateQuestionContext{
				MissionID:     "MISSION-999",
				MissionExists: false,
			},
			wantAllowed: false,
			wantReason:  "mission MISSION-999 not found",
		},
		{
			name: "cannot create question when investigation not found",
			ctx: CreateQuestionContext{
				MissionID:           "MISSION-001",
				MissionExists:       true,
				InvestigationID:     "INV-999",
				InvestigationExists: false,
			},
			wantAllowed: false,
			wantReason:  "investigation INV-999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanCreateQuestion(tt.ctx)
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestCanAnswerQuestion(t *testing.T) {
	tests := []struct {
		name        string
		ctx         AnswerQuestionContext
		wantAllowed bool
		wantReason  string
	}{
		{
			name: "can answer open unpinned question",
			ctx: AnswerQuestionContext{
				QuestionID: "Q-001",
				Status:     "open",
				IsPinned:   false,
			},
			wantAllowed: true,
		},
		{
			name: "cannot answer open pinned question",
			ctx: AnswerQuestionContext{
				QuestionID: "Q-001",
				Status:     "open",
				IsPinned:   true,
			},
			wantAllowed: false,
			wantReason:  "cannot answer pinned question Q-001. Unpin first with: orc question unpin Q-001",
		},
		{
			name: "cannot answer already answered question",
			ctx: AnswerQuestionContext{
				QuestionID: "Q-001",
				Status:     "answered",
				IsPinned:   false,
			},
			wantAllowed: false,
			wantReason:  "can only answer open questions (current status: answered)",
		},
		{
			name: "cannot answer already answered and pinned question",
			ctx: AnswerQuestionContext{
				QuestionID: "Q-001",
				Status:     "answered",
				IsPinned:   true,
			},
			wantAllowed: false,
			wantReason:  "can only answer open questions (current status: answered)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAnswerQuestion(tt.ctx)
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
