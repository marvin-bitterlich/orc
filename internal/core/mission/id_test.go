package mission

import "testing"

func TestGenerateMissionID(t *testing.T) {
	tests := []struct {
		name       string
		currentMax int
		want       string
	}{
		{
			name:       "first mission (max=0)",
			currentMax: 0,
			want:       "MISSION-001",
		},
		{
			name:       "second mission (max=1)",
			currentMax: 1,
			want:       "MISSION-002",
		},
		{
			name:       "tenth mission (max=9)",
			currentMax: 9,
			want:       "MISSION-010",
		},
		{
			name:       "hundredth mission (max=99)",
			currentMax: 99,
			want:       "MISSION-100",
		},
		{
			name:       "three-digit boundary (max=999)",
			currentMax: 999,
			want:       "MISSION-1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMissionID(tt.currentMax)
			if got != tt.want {
				t.Errorf("GenerateMissionID(%d) = %q, want %q", tt.currentMax, got, tt.want)
			}
		})
	}
}

func TestParseMissionNumber(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want int
	}{
		{
			name: "valid single digit",
			id:   "MISSION-001",
			want: 1,
		},
		{
			name: "valid double digit",
			id:   "MISSION-042",
			want: 42,
		},
		{
			name: "valid triple digit",
			id:   "MISSION-123",
			want: 123,
		},
		{
			name: "valid four digit",
			id:   "MISSION-1000",
			want: 1000,
		},
		{
			name: "invalid format - no dash",
			id:   "MISSION001",
			want: -1,
		},
		{
			name: "invalid format - wrong prefix",
			id:   "GROVE-001",
			want: -1,
		},
		{
			name: "invalid format - empty",
			id:   "",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMissionNumber(tt.id)
			if got != tt.want {
				t.Errorf("ParseMissionNumber(%q) = %d, want %d", tt.id, got, tt.want)
			}
		})
	}
}
