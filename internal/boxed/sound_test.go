package boxed_test

import (
	"testing"

	"github.com/cppcho/boxed/internal/boxed"
)

func TestPlaySoundByName(t *testing.T) {
	tests := []struct {
		name    string
		sound   string
		wantCmd string
	}{
		{
			name:    "Glass sound",
			sound:   "Glass",
			wantCmd: "afplay /System/Library/Sounds/Glass.aiff",
		},
		{
			name:    "Sosumi sound",
			sound:   "Sosumi",
			wantCmd: "afplay /System/Library/Sounds/Sosumi.aiff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &boxed.RecordingRunner{}
			boxed.PlaySoundByName(runner, tt.sound)
			if len(runner.Starts) != 1 {
				t.Fatalf("expected 1 Start call, got %d", len(runner.Starts))
			}
			if runner.Starts[0] != tt.wantCmd {
				t.Errorf("got %q, want %q", runner.Starts[0], tt.wantCmd)
			}
		})
	}
}

func TestPlaySoundFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantCmd  string
	}{
		{
			name:     "absolute path",
			filePath: "/usr/local/share/sounds/alert.aiff",
			wantCmd:  "afplay /usr/local/share/sounds/alert.aiff",
		},
		{
			name:     "path with spaces",
			filePath: "/Users/me/My Sounds/ding.aiff",
			wantCmd:  "afplay /Users/me/My Sounds/ding.aiff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &boxed.RecordingRunner{}
			boxed.PlaySoundFile(runner, tt.filePath)
			if len(runner.Starts) != 1 {
				t.Fatalf("expected 1 Start call, got %d", len(runner.Starts))
			}
			if runner.Starts[0] != tt.wantCmd {
				t.Errorf("got %q, want %q", runner.Starts[0], tt.wantCmd)
			}
		})
	}
}
