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
	var old CurrentTimer
	now := a.NowFunc()
	if ReadStateKey(a.Paths.StateFile, StateKeyCurrent, &old) {
		oldTask := old.Task
		if oldTask == "" {
			oldTask = "Untitled"
		}
		oldElapsed := now.Sub(old.StartedAt)
		if oldElapsed < old.Duration {
			LogEnd(a.Paths, old.StartedAt, old.Duration, oldTask, false, a.NowFunc)
		} else {
			LogEnd(a.Paths, old.StartedAt, old.Duration, oldTask, true, a.NowFunc)
		}
	}

	duration := time.Duration(durationMins) * time.Minute
	WriteStateKey(a.Paths, StateKeyCurrent, &CurrentTimer{
		Task:      task,
		StartedAt: now,
		Duration:  duration,
	})
	WriteStateKey(a.Paths, StateKeyLast, &LastTimer{
		Duration: duration,
		Task:     task,
	})

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
	return CompleteTimer(a.Paths, a.Runner, a.NowFunc)
}

// CompleteTimer marks an expired timer as notified, logs it, and sends a notification.
func CompleteTimer(paths Paths, runner CommandRunner, nowFunc func() time.Time) error {
	var timer CurrentTimer
	if !ReadStateKey(paths.StateFile, StateKeyCurrent, &timer) {
		return nil
	}
	if timer.Notified {
		return nil
	}

	now := nowFunc()
	elapsed := now.Sub(timer.StartedAt)
	if elapsed < timer.Duration {
		return nil
	}

	task := timer.Task
	if task == "" {
		task = "Untitled"
	}
	config := ReadConfig(paths.ConfigFile)

	LogEnd(paths, timer.StartedAt, timer.Duration, task, true, nowFunc)
	Notify(runner, "Boxed", fmt.Sprintf("Time's up! — %s", task))
	if config.NotifySound {
		PlaySoundByName(runner, "Glass")
	}

	timer.Notified = true
	WriteStateKey(paths, StateKeyCurrent, &timer)
	return nil
}

// CmdAgain repeats the last started timer.
func (a *App) CmdAgain(args []string) error {
	var lt LastTimer
	if !ReadStateKey(a.Paths.StateFile, StateKeyLast, &lt) {
		fmt.Fprintln(a.Stderr, "No previous timer to repeat.")
		return fmt.Errorf("exit 1")
	}
	return a.CmdStart([]string{strconv.Itoa(int(lt.Duration.Minutes())), lt.Task})
}

// CmdStop stops the current timer.
func (a *App) CmdStop(args []string) error {
	var timer CurrentTimer
	if !ReadStateKey(a.Paths.StateFile, StateKeyCurrent, &timer) {
		fmt.Fprintln(a.Stderr, "No timer running.")
		return fmt.Errorf("exit 1")
	}

	now := a.NowFunc()
	task := timer.Task
	if task == "" {
		task = "Untitled"
	}
	elapsed := now.Sub(timer.StartedAt)

	// Timer already expired — finalize and clean up
	if elapsed >= timer.Duration {
		LogEnd(a.Paths, timer.StartedAt, timer.Duration, task, true, a.NowFunc)
		ClearStateKey(a.Paths, StateKeyCurrent)
		fmt.Fprintf(a.Stdout, "Cleared ended timer: %s\n", task)
		return nil
	}

	config := ReadConfig(a.Paths.ConfigFile)
	ClearStateKey(a.Paths, StateKeyCurrent)
	LogEnd(a.Paths, timer.StartedAt, timer.Duration, task, false, a.NowFunc)
	elapsedStr := FormatDuration(elapsed)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Timer stopped: %s (%s elapsed)", task, elapsedStr))
	if config.NotifySound {
		PlaySoundByName(a.Runner, "Sosumi")
	}
	fmt.Fprintf(a.Stdout, "Timer stopped: %s (%s elapsed)\n", task, elapsedStr)
	return nil
}
