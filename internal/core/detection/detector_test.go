package detection

import (
	"testing"
)

func TestDetectOutcome_Menu(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "menu with proceed question",
			content: `Would you like to proceed?
❯ 1. Yes, clear context (shift+tab)
  2. Yes, auto-accept edits`,
			want: OutcomeMenu,
		},
		{
			name: "menu with numbered options",
			content: `Select an option:
❯ 1. Option A
  2. Option B`,
			want: OutcomeMenu,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOutcome(tt.content)
			if got != tt.want {
				t.Errorf("DetectOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcome_Working(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "thundering spinner",
			content: "✶ Thundering… (esc to interrupt)",
			want:    OutcomeWorking,
		},
		{
			name:    "kneading message",
			content: "Kneading your request...",
			want:    OutcomeWorking,
		},
		{
			name:    "contemplating",
			content: "✶ Contemplating… (esc to interrupt)",
			want:    OutcomeWorking,
		},
		{
			name:    "spinner character",
			content: "⠋ Processing...",
			want:    OutcomeWorking,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOutcome(tt.content)
			if got != tt.want {
				t.Errorf("DetectOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcome_Idle(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "cogitated completion",
			content: `✻ Cogitated for 5m 49s
───────────────────────────────────────
❯`,
			want: OutcomeIdle,
		},
		{
			name:    "empty prompt",
			content: "❯",
			want:    OutcomeIdle,
		},
		{
			name: "completion with empty prompt",
			content: `Done!
$`,
			want: OutcomeIdle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOutcome(tt.content)
			if got != tt.want {
				t.Errorf("DetectOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcome_Typed(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "typed command",
			content: `Previous output here
❯ /imp-plan-create`,
			want: OutcomeTyped,
		},
		{
			name:    "typed in shell",
			content: `$ git status`,
			want:    OutcomeTyped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOutcome(tt.content)
			if got != tt.want {
				t.Errorf("DetectOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcome_Error(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "go panic",
			content: "panic: runtime error: invalid memory address",
			want:    OutcomeError,
		},
		{
			name:    "generic error",
			content: "Error: failed to connect to database",
			want:    OutcomeError,
		},
		{
			name: "python traceback",
			content: `Traceback (most recent call last):
  File "test.py", line 1, in <module>
    raise Exception("test")`,
			want: OutcomeError,
		},
		{
			name:    "test failure",
			content: "--- FAILED: TestSomething (0.00s)",
			want:    OutcomeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOutcome(tt.content)
			if got != tt.want {
				t.Errorf("DetectOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcome_Priority(t *testing.T) {
	// Error should take priority over other patterns
	t.Run("error over working", func(t *testing.T) {
		content := "✶ Thundering… Error: something went wrong"
		got := DetectOutcome(content)
		if got != OutcomeError {
			t.Errorf("DetectOutcome() = %q, want %q (error should take priority)", got, OutcomeError)
		}
	})

	// Menu should take priority over working
	t.Run("menu over working", func(t *testing.T) {
		content := "Would you like to proceed? ✶ processing"
		got := DetectOutcome(content)
		if got != OutcomeMenu {
			t.Errorf("DetectOutcome() = %q, want %q (menu should take priority)", got, OutcomeMenu)
		}
	})
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no ansi",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "red text",
			input: "\x1b[31mred\x1b[0m",
			want:  "red",
		},
		{
			name:  "bold green",
			input: "\x1b[1;32mbold green\x1b[0m",
			want:  "bold green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			if got != tt.want {
				t.Errorf("stripANSI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectOutcomeWithContext(t *testing.T) {
	t.Run("consecutive typed", func(t *testing.T) {
		content := "❯ /imp-plan"
		prevContent := "❯ /imp-plan"
		prevOutcome := OutcomeTyped

		got := DetectOutcomeWithContext(content, prevContent, prevOutcome)
		if got != OutcomeTyped {
			t.Errorf("DetectOutcomeWithContext() = %q, want %q", got, OutcomeTyped)
		}
	})
}
