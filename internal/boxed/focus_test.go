package boxed

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestSetFocus_On(t *testing.T) {
	runner := &RecordingRunner{}
	SetFocus(runner, true)

	if len(runner.Starts) != 1 {
		t.Fatalf("expected 1 Start call, got %d", len(runner.Starts))
	}
	if runner.Starts[0] != "shortcuts run Focus On" {
		t.Errorf("expected 'shortcuts run Focus On', got %q", runner.Starts[0])
	}
}

func TestSetFocus_Off(t *testing.T) {
	runner := &RecordingRunner{}
	SetFocus(runner, false)

	if len(runner.Starts) != 1 {
		t.Fatalf("expected 1 Start call, got %d", len(runner.Starts))
	}
	if runner.Starts[0] != "shortcuts run Focus Off" {
		t.Errorf("expected 'shortcuts run Focus Off', got %q", runner.Starts[0])
	}
}

func TestCmdStart_SetsFocusOn(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.CmdStart([]string{"5", "focus"})

	found := false
	for _, s := range runner.Starts {
		if s == "shortcuts run Focus On" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'shortcuts run Focus On' Start call")
	}
}

func TestCmdStart_SetsFocusOnWhenSoundDisabled(t *testing.T) {
	app, runner, _, _ := testApp(t)
	os.WriteFile(app.Paths.ConfigFile, []byte("notify_sound = false\n"), 0o644)
	app.CmdStart([]string{"5", "focus"})

	found := false
	for _, s := range runner.Starts {
		if s == "shortcuts run Focus On" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'shortcuts run Focus On' even when sound disabled")
	}
}

func TestCmdStop_SetsFocusOff(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "work"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdStop([]string{})

	found := false
	for _, s := range runner.Starts {
		if s == "shortcuts run Focus Off" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'shortcuts run Focus Off' Start call")
	}
}

func TestCmdStop_ExpiredTimer_SetsFocusOff(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "short"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdStop([]string{})

	found := false
	for _, s := range runner.Starts {
		if s == "shortcuts run Focus Off" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'shortcuts run Focus Off' for expired timer stop")
	}
}

func TestCmdComplete_SetsFocusOff(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"1", "done"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000120, 0) }
	app.CmdComplete([]string{})

	found := false
	for _, s := range runner.Starts {
		if s == "shortcuts run Focus Off" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'shortcuts run Focus Off' Start call")
	}
}

func TestCmdComplete_NotExpired_NoFocusOff(t *testing.T) {
	app, runner, _, _ := testApp(t)
	app.NowFunc = func() time.Time { return time.Unix(1700000000, 0) }
	app.CmdStart([]string{"25", "running"})
	runner.Starts = nil

	app.NowFunc = func() time.Time { return time.Unix(1700000300, 0) }
	app.CmdComplete([]string{})

	for _, s := range runner.Starts {
		if strings.HasPrefix(s, "shortcuts") {
			t.Error("should not call shortcuts for non-expired timer")
		}
	}
}
