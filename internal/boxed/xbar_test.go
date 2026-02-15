package boxed

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func testXbarApp(t *testing.T) (*XbarApp, *RecordingRunner, *bytes.Buffer) {
	p := testPaths(t)
	runner := &RecordingRunner{}
	stdout := &bytes.Buffer{}
	return &XbarApp{
		Paths:    p,
		Runner:   runner,
		NowFunc:  func() time.Time { return time.Unix(1700000000, 0) },
		Stdout:   stdout,
		BoxedBin: "/usr/local/bin/boxed",
	}, runner, stdout
}

func writeStateJSON(t *testing.T, path string, s map[string]any) {
	data, _ := json.Marshal(s)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- No Timer ---

func TestXbar_NoTimer_ShowsBoxEmoji(t *testing.T) {
	x, _, stdout := testXbarApp(t)
	x.Run()
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if lines[0] != "📦" {
		t.Errorf("expected 📦, got %q", lines[0])
	}
}

func TestXbar_NoTimer_ShowsMenuItems(t *testing.T) {
	x, _, stdout := testXbarApp(t)
	x.Run()
	out := stdout.String()
	if !strings.Contains(out, "Open Config |") {
		t.Error("expected 'Open Config' menu item")
	}
	if !strings.Contains(out, "Open Log |") {
		t.Error("expected 'Open Log' menu item")
	}
	if !strings.Contains(out, "Open Config Directory |") {
		t.Error("expected 'Open Config Directory' menu item")
	}
}

// --- Running Timer ---

func TestXbar_RunningTimer_ShowsTaskAndRemaining(t *testing.T) {
	x, _, stdout := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "focus work",
		"started_epoch": now.Unix(),
		"duration":      1500,
	})
	x.NowFunc = func() time.Time { return now.Add(300 * time.Second) }
	x.Run()

	firstLine := strings.Split(strings.TrimSpace(stdout.String()), "\n")[0]
	if !strings.Contains(firstLine, "focus work") {
		t.Errorf("expected 'focus work' in first line, got %q", firstLine)
	}
	if !strings.Contains(firstLine, "20m") {
		t.Errorf("expected '20m' in first line, got %q", firstLine)
	}
}

func TestXbar_RunningTimer_SeparatorPresent(t *testing.T) {
	x, _, stdout := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "t",
		"started_epoch": now.Unix(),
		"duration":      1500,
	})
	x.NowFunc = func() time.Time { return now.Add(60 * time.Second) }
	x.Run()

	if !strings.Contains(stdout.String(), "---") {
		t.Error("expected separator")
	}
}

// --- Expired Timer ---

func TestXbar_ExpiredTimer_ShowsBoxEmoji(t *testing.T) {
	x, _, stdout := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "done",
		"started_epoch": now.Unix(),
		"duration":      60,
		"notified":      true,
	})
	x.NowFunc = func() time.Time { return now.Add(120 * time.Second) }
	x.Run()

	firstLine := strings.Split(strings.TrimSpace(stdout.String()), "\n")[0]
	if firstLine != "📦" {
		t.Errorf("expected 📦, got %q", firstLine)
	}
}

func TestXbar_ExpiredNotNotified_CallsComplete(t *testing.T) {
	x, runner, _ := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "expired",
		"started_epoch": now.Unix(),
		"duration":      60,
	})
	x.NowFunc = func() time.Time { return now.Add(120 * time.Second) }
	x.Run()

	if len(runner.Runs) == 0 {
		t.Fatal("expected Run call for 'complete'")
	}
	found := false
	for _, r := range runner.Runs {
		if strings.Contains(r, "complete") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'complete' command, got %v", runner.Runs)
	}
}

// --- Tick Sound ---

func TestXbar_TickPlaysWhenIntervalReached(t *testing.T) {
	x, runner, _ := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "tick test",
		"started_epoch": now.Unix(),
		"duration":      1500,
	})
	os.WriteFile(x.Paths.ConfigFile, []byte("tick_interval = 5\n"), 0o644)
	x.NowFunc = func() time.Time { return now.Add(300 * time.Second) }
	x.Run()

	if len(runner.Starts) != 1 {
		t.Errorf("expected 1 sound Start call, got %d", len(runner.Starts))
	}
}

func TestXbar_NoTickBeforeInterval(t *testing.T) {
	x, runner, _ := testXbarApp(t)
	now := time.Unix(1700000000, 0)
	writeStateJSON(t, x.Paths.StateFile, map[string]any{
		"task":          "tick test",
		"started_epoch": now.Unix(),
		"duration":      1500,
	})
	os.WriteFile(x.Paths.ConfigFile, []byte("tick_interval = 5\n"), 0o644)
	x.NowFunc = func() time.Time { return now.Add(60 * time.Second) }
	x.Run()

	if len(runner.Starts) != 0 {
		t.Errorf("expected no Start calls, got %d", len(runner.Starts))
	}
}
