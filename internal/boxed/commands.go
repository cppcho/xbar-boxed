package boxed

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// App holds dependencies for CLI commands.
type App struct {
	Paths   Paths
	Runner  CommandRunner
	NowFunc func() time.Time
	Stdout  io.Writer
	Stderr  io.Writer
}

// CmdStart starts a new timer.
func (a *App) CmdStart(args []string) error {
	if len(args) < 2 {
		fmt.Fprintln(a.Stderr, "Usage: boxed start <duration in minutes> <task name...>")
		return fmt.Errorf("exit 1")
	}

	durationMins, err := strconv.Atoi(args[0])
	if err != nil || durationMins <= 0 {
		fmt.Fprintf(a.Stderr, "Invalid duration: %s\n", args[0])
		return fmt.Errorf("exit 1")
	}

	task := strings.Join(args[1:], " ")

	// If timer already running, stop it first
	state := ReadState(a.Paths.StateFile)
	now := a.NowFunc()
	if state != nil {
		oldTask := state.Task
		if oldTask == "" {
			oldTask = "Untitled"
		}
		oldStarted := time.Unix(state.StartedEpoch, 0)
		oldDuration := time.Duration(state.Duration) * time.Second
		oldElapsed := now.Sub(oldStarted)
		if oldElapsed < oldDuration {
			LogEnd(a.Paths, oldStarted, oldDuration, oldTask, false, a.NowFunc)
		} else {
			LogEnd(a.Paths, oldStarted, oldDuration, oldTask, true, a.NowFunc)
		}
	}

	WriteLastTimer(a.Paths, durationMins, task)

	durationSecs := durationMins * 60
	duration := time.Duration(durationSecs) * time.Second
	WriteState(a.Paths, task, now.Unix(), durationSecs)

	config := ReadConfig(a.Paths.ConfigFile)
	LogStart(a.Paths, now, duration, task)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Timer started: %dm — %s", durationMins, task))
	if config.NotifySound {
		PlaySoundFile(a.Runner, filepath.Join(a.Paths.SoundsDir, "PeonReady1.ogg"))
	}
	fmt.Fprintf(a.Stdout, "Timer started: %dm — %s\n", durationMins, task)
	return nil
}

// CmdComplete is called by the xbar plugin when a timer expires.
func (a *App) CmdComplete(args []string) error {
	state := ReadState(a.Paths.StateFile)
	if state == nil {
		return nil
	}
	if state.Notified {
		return nil
	}

	now := a.NowFunc()
	started := time.Unix(state.StartedEpoch, 0)
	duration := time.Duration(state.Duration) * time.Second
	elapsed := now.Sub(started)
	if elapsed < duration {
		return nil
	}

	task := state.Task
	if task == "" {
		task = "Untitled"
	}
	config := ReadConfig(a.Paths.ConfigFile)

	LogEnd(a.Paths, started, duration, task, true, a.NowFunc)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Time's up! — %s", task))
	if config.NotifySound {
		PlaySoundByName(a.Runner, "Glass")
	}

	state.Notified = true
	WriteStateFull(a.Paths, state)
	return nil
}

// CmdAgain repeats the last started timer.
func (a *App) CmdAgain(args []string) error {
	lt, err := ReadLastTimer(a.Paths.LastFile)
	if err != nil {
		fmt.Fprintln(a.Stderr, "No previous timer to repeat.")
		return fmt.Errorf("exit 1")
	}
	return a.CmdStart([]string{strconv.Itoa(lt.Duration), lt.Task})
}

// CmdStop stops the current timer.
func (a *App) CmdStop(args []string) error {
	state := ReadState(a.Paths.StateFile)
	if state == nil {
		fmt.Fprintln(a.Stderr, "No timer running.")
		return fmt.Errorf("exit 1")
	}

	now := a.NowFunc()
	task := state.Task
	if task == "" {
		task = "Untitled"
	}
	started := time.Unix(state.StartedEpoch, 0)
	duration := time.Duration(state.Duration) * time.Second
	elapsed := now.Sub(started)

	// Timer already expired — finalize and clean up
	if elapsed >= duration {
		LogEnd(a.Paths, started, duration, task, true, a.NowFunc)
		ClearState(a.Paths)
		fmt.Fprintf(a.Stdout, "Cleared ended timer: %s\n", task)
		return nil
	}

	config := ReadConfig(a.Paths.ConfigFile)
	ClearState(a.Paths)
	LogEnd(a.Paths, started, duration, task, false, a.NowFunc)
	elapsedStr := FormatDuration(elapsed)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Timer stopped: %s (%s elapsed)", task, elapsedStr))
	if config.NotifySound {
		PlaySoundByName(a.Runner, "Sosumi")
	}
	fmt.Fprintf(a.Stdout, "Timer stopped: %s (%s elapsed)\n", task, elapsedStr)
	return nil
}
