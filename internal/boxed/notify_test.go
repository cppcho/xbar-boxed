package boxed_test

import (
	"testing"

	"github.com/cppcho/boxed/internal/boxed"
)

func TestNotify_Basic(t *testing.T) {
	runner := &boxed.RecordingRunner{}
	boxed.Notify(runner, "Boxed", "Timer started")

	if len(runner.Runs) != 1 {
		t.Fatalf("expected 1 Run call, got %d", len(runner.Runs))
	}
	want := `osascript -e display notification "Timer started" with title "Boxed"`
	if runner.Runs[0] != want {
		t.Errorf("got %q, want %q", runner.Runs[0], want)
	}
}

func TestNotify_EscapesSpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		message     string
		wantCommand string
	}{
		{
			name:        "message with double quotes",
			title:       "Boxed",
			message:     `Say "hello"`,
			wantCommand: `osascript -e display notification "Say \"hello\"" with title "Boxed"`,
		},
		{
			name:        "title with backslashes",
			title:       `Back\slash`,
			message:     "Timer done",
			wantCommand: `osascript -e display notification "Timer done" with title "Back\\slash"`,
		},
		{
			name:        "both special characters combined",
			title:       `Te"st`,
			message:     `He said "go\" now`,
			wantCommand: `osascript -e display notification "He said \"go\\\" now" with title "Te\"st"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &boxed.RecordingRunner{}
			boxed.Notify(runner, tt.title, tt.message)
			if len(runner.Runs) != 1 {
				t.Fatalf("expected 1 Run call, got %d", len(runner.Runs))
			}
			if runner.Runs[0] != tt.wantCommand {
				t.Errorf("got %q, want %q", runner.Runs[0], tt.wantCommand)
			}
		})
	}
}
