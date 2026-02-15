package boxed

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testApp(t *testing.T) (*App, *RecordingRunner, *bytes.Buffer, *bytes.Buffer) {
	p := testPaths(t)
	runner := &RecordingRunner{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app := &App{
		Paths:   p,
		Runner:  runner,
		NowFunc: func() time.Time { return time.Unix(1700000000, 0) },
		Stdout:  stdout,
		Stderr:  stderr,
	}
	return app, runner, stdout, stderr
}

// --- TestCmdStart ---

func TestCmdStart_Basic(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "my", "task"})

	data, _ := os.ReadFile(app.Paths.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if s.Task != "my task" {
		t.Errorf("expected task='my task', got %q", s.Task)
	}
	if s.StartedEpoch != 1700000000 {
		t.Errorf("expected started_epoch=1700000000, got %d", s.StartedEpoch)
	}
	if s.Duration != 1500 {
		t.Errorf("expected duration=1500, got %d", s.Duration)
	}
}

func TestCmdStart_SavesLastFile(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.CmdStart([]string{"10", "quick"})

	data, _ := os.ReadFile(app.Paths.LastFile)
	var lt LastTimer
	json.Unmarshal(data, &lt)
	if lt.Duration != 10 {
		t.Errorf("expected duration=10, got %d", lt.Duration)
	}
	if lt.Task != "quick" {
		t.Errorf("expected task='quick', got %q", lt.Task)
	}
}

func TestCmdStart_SendsNotification(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.CmdStart([]string{"5", "notify"})

	if len(runner.Runs) == 0 {
		t.Fatal("expected osascript call")
	}
	if !strings.HasPrefix(runner.Runs[0], "osascript") {
		t.Errorf("expected osascript command, got %q", runner.Runs[0])
	}
}

func TestCmdStart_PlaysSoundWhenEnabled(t *testing.T) {
	app, runner, _, _ := testApp(t)
	os.WriteFile(app.Paths.ConfigFile, []byte("notify_sound = true\n"), 0o644)
	app.CmdStart([]string{"5", "sound"})

	if len(runner.Starts) != 1 {
		t.Fatalf("expected 1 Start call, got %d", len(runner.Starts))
	}
}

func TestCmdStart_NoSoundWhenDisabled(t *testing.T) {
	app, runner, _, _ := testApp(t)
	os.WriteFile(app.Paths.ConfigFile, []byte("notify_sound = false\n"), 0o644)
	app.CmdStart([]string{"5", "quiet"})

	if len(runner.Starts) != 0 {
		t.Errorf("expected no Start calls, got %d", len(runner.Starts))
	}
}

func TestCmdStart_ReplacesRunningTimer(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "first"})

	app.NowFunc = func() time.Time { return time.Unix(1700000060, 0) }
	app.CmdStart([]string{"10", "second"})

	data, _ := os.ReadFile(app.Paths.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if s.Task != "second" {
		t.Errorf("expected task='second', got %q", s.Task)
	}
	if s.Duration != 600 {
		t.Errorf("expected duration=600, got %d", s.Duration)
	}
}

func TestCmdStart_ReplacesExpiredTimer(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "short"})

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdStart([]string{"10", "next"})

	data, _ := os.ReadFile(app.Paths.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if s.Task != "next" {
		t.Errorf("expected task='next', got %q", s.Task)
	}
}

func TestCmdStart_CreatesLogEntry(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.CmdStart([]string{"5", "logged"})

	data, _ := os.ReadFile(app.Paths.LogFile)
	if !strings.Contains(string(data), "logged") {
		t.Error("expected 'logged' in log file")
	}
}

func TestCmdStart_PrintsMessage(t *testing.T) {
	app, _, stdout, _ := testApp(t)
	app.CmdStart([]string{"5", "hello"})

	out := stdout.String()
	if !strings.Contains(out, "Timer started") {
		t.Error("expected 'Timer started' in output")
	}
	if !strings.Contains(out, "hello") {
		t.Error("expected 'hello' in output")
	}
}

func TestCmdStart_MissingArgs(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStart([]string{})
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestCmdStart_MissingTask(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStart([]string{"25"})
	if err == nil {
		t.Error("expected error for missing task")
	}
}

func TestCmdStart_InvalidDuration(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStart([]string{"abc", "task"})
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestCmdStart_ZeroDuration(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStart([]string{"0", "task"})
	if err == nil {
		t.Error("expected error for zero duration")
	}
}

func TestCmdStart_NegativeDuration(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStart([]string{"-5", "task"})
	if err == nil {
		t.Error("expected error for negative duration")
	}
}

// --- TestCmdStop ---

func TestCmdStop_RunningTimer(t *testing.T) {
	app, _, stdout, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "work"})
	stdout.Reset()

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	if _, err := os.Stat(app.Paths.StateFile); !os.IsNotExist(err) {
		t.Error("state file should be removed")
	}
	if !strings.Contains(stdout.String(), "stopped") {
		t.Error("expected 'stopped' in output")
	}
}

func TestCmdStop_SendsNotification(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "work"})
	runner.Runs = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	found := false
	for _, r := range runner.Runs {
		if strings.HasPrefix(r, "osascript") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected osascript notification call")
	}
}

func TestCmdStop_ExpiredTimerClearsState(t *testing.T) {
	app, _, stdout, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "short"})
	stdout.Reset()

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdStop([]string{})

	if _, err := os.Stat(app.Paths.StateFile); !os.IsNotExist(err) {
		t.Error("state file should be removed")
	}
	if !strings.Contains(stdout.String(), "Cleared") {
		t.Error("expected 'Cleared' in output")
	}
}

func TestCmdStop_NoTimer(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdStop([]string{})
	if err == nil {
		t.Error("expected error for no timer")
	}
}

func TestCmdStop_LogsCrossMarker(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "stopped"})

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	data, _ := os.ReadFile(app.Paths.LogFile)
	if !strings.Contains(string(data), "✕") {
		t.Error("expected cross marker in log")
	}
}

func TestCmdStop_PlaysSoundWhenEnabled(t *testing.T) {
	app, runner, _, _ := testApp(t)
	os.WriteFile(app.Paths.ConfigFile, []byte("notify_sound = true\n"), 0o644)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "work"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	if len(runner.Starts) != 1 {
		t.Errorf("expected 1 sound Start call, got %d", len(runner.Starts))
	}
}

func TestCmdStop_NoSoundWhenDisabled(t *testing.T) {
	app, runner, _, _ := testApp(t)
	os.WriteFile(app.Paths.ConfigFile, []byte("notify_sound = false\n"), 0o644)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "quiet"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	if len(runner.Starts) != 0 {
		t.Errorf("expected no Start calls, got %d", len(runner.Starts))
	}
}

// --- TestCmdComplete ---

func TestCmdComplete_MarksNotified(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "done"})

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdComplete([]string{})

	data, _ := os.ReadFile(app.Paths.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if !s.Notified {
		t.Error("expected notified=true")
	}
}

func TestCmdComplete_SendsNotification(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "done"})
	runner.Runs = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdComplete([]string{})

	found := false
	for _, r := range runner.Runs {
		if strings.HasPrefix(r, "osascript") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected osascript notification call")
	}
}

func TestCmdComplete_Idempotent(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "done"})

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdComplete([]string{})
	runner.Runs = nil

	// Second call should be a no-op
	app.CmdComplete([]string{})
	for _, r := range runner.Runs {
		if strings.HasPrefix(r, "osascript") {
			t.Error("should not send notification twice")
		}
	}
}

func TestCmdComplete_NotExpiredYet(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "running"})
	runner.Runs = nil

	// Only 5 minutes in
	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdComplete([]string{})

	for _, r := range runner.Runs {
		if strings.HasPrefix(r, "osascript") {
			t.Error("should not send notification for non-expired timer")
		}
	}
}

func TestCmdComplete_NoState(t *testing.T) {
	app, _, _, _ := testApp(t)
	// Should not error
	if err := app.CmdComplete([]string{}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCmdComplete_LogsCheckmark(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "finished"})

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdComplete([]string{})

	data, _ := os.ReadFile(app.Paths.LogFile)
	if !strings.Contains(string(data), "✓") {
		t.Error("expected checkmark in log")
	}
}

// --- TestCmdAgain ---

func TestCmdAgain_RepeatsLast(t *testing.T) {
	app, _, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "repeated"})

	app.NowFunc = func() time.Time { return time.Unix(1700002000, 0) }
	app.CmdAgain([]string{})

	data, _ := os.ReadFile(app.Paths.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if s.Task != "repeated" {
		t.Errorf("expected task='repeated', got %q", s.Task)
	}
	if s.Duration != 1500 {
		t.Errorf("expected duration=1500, got %d", s.Duration)
	}
}

func TestCmdAgain_NoPrevious(t *testing.T) {
	app, _, _, _ := testApp(t)
	err := app.CmdAgain([]string{})
	if err == nil {
		t.Error("expected error for no previous timer")
	}
}

func TestCmdAgain_CorruptLastFile(t *testing.T) {
	app, _, _, _ := testApp(t)
	os.WriteFile(filepath.Join(app.Paths.ConfigDir, "last.json"), []byte("{bad json"), 0o644)
	err := app.CmdAgain([]string{})
	if err == nil {
		t.Error("expected error for corrupt last file")
	}
}
