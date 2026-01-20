package mission

import (
	"testing"
	"time"
)

func TestApplyStatusTransition(t *testing.T) {
	fixedTime := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		newStatus       MissionStatus
		wantStatus      MissionStatus
		wantCompletedAt bool
	}{
		{
			name:            "transition to active",
			newStatus:       StatusActive,
			wantStatus:      StatusActive,
			wantCompletedAt: false,
		},
		{
			name:            "transition to paused",
			newStatus:       StatusPaused,
			wantStatus:      StatusPaused,
			wantCompletedAt: false,
		},
		{
			name:            "transition to complete sets CompletedAt",
			newStatus:       StatusComplete,
			wantStatus:      StatusComplete,
			wantCompletedAt: true,
		},
		{
			name:            "transition to archived",
			newStatus:       StatusArchived,
			wantStatus:      StatusArchived,
			wantCompletedAt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyStatusTransition(tt.newStatus, fixedTime)

			if result.NewStatus != tt.wantStatus {
				t.Errorf("ApplyStatusTransition().NewStatus = %q, want %q", result.NewStatus, tt.wantStatus)
			}

			if tt.wantCompletedAt {
				if result.CompletedAt == nil {
					t.Error("ApplyStatusTransition().CompletedAt = nil, want non-nil")
				} else if !result.CompletedAt.Equal(fixedTime) {
					t.Errorf("ApplyStatusTransition().CompletedAt = %v, want %v", result.CompletedAt, fixedTime)
				}
			} else {
				if result.CompletedAt != nil {
					t.Errorf("ApplyStatusTransition().CompletedAt = %v, want nil", result.CompletedAt)
				}
			}
		})
	}
}

func TestInitialStatus(t *testing.T) {
	got := InitialStatus()
	want := StatusActive

	if got != want {
		t.Errorf("InitialStatus() = %q, want %q", got, want)
	}
}
