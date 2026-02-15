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
	var old Timer
	now := a.NowFunc()
	if ReadStateKey(a.Paths.StateFile, "current", &old) {
		oldTask := old.Task
		if oldTask == "" {
			oldTask = "Untitled"
		}
		oldStarted := time.Unix(old.StartedEpoch, 0)
		oldDuration := time.Duration(old.Duration) * time.Second
		oldElapsed := now.Sub(oldStarted)
		if oldElapsed < oldDuration {
			LogEnd(a.Paths, oldStarted, oldDuration, oldTask, false, a.NowFunc)
		} else {
			LogEnd(a.Paths, oldStarted, oldDuration, oldTask, true, a.NowFunc)
		}
	}

	durationSecs := durationMins * 60
	duration := time.Duration(durationSecs) * time.Second
	WriteStateKey(a.Paths, "current", &Timer{
		Task:         task,
		StartedEpoch: now.Unix(),
		Duration:     durationSecs,
	})
	WriteStateKey(a.Paths, "last", &LastTimer{
		Duration: durationMins,
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
	var timer Timer
	if !ReadStateKey(a.Paths.StateFile, "current", &timer) {
		return nil
	}
	if timer.Notified {
		return nil
	}

	now := a.NowFunc()
	started := time.Unix(timer.StartedEpoch, 0)
	duration := time.Duration(timer.Duration) * time.Second
	elapsed := now.Sub(started)
	if elapsed < duration {
		return nil
	}

	task := timer.Task
	if task == "" {
		task = "Untitled"
	}
	config := ReadConfig(a.Paths.ConfigFile)

	LogEnd(a.Paths, started, duration, task, true, a.NowFunc)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Time's up! — %s", task))
	if config.NotifySound {
		PlaySoundByName(a.Runner, "Glass")
	}

	timer.Notified = true
	WriteStateKey(a.Paths, "current", &timer)
	return nil
}

// CmdAgain repeats the last started timer.
func (a *App) CmdAgain(args []string) error {
	var lt LastTimer
	if !ReadStateKey(a.Paths.StateFile, "last", &lt) {
		fmt.Fprintln(a.Stderr, "No previous timer to repeat.")
		return fmt.Errorf("exit 1")
	}
	return a.CmdStart([]string{strconv.Itoa(lt.Duration), lt.Task})
}

// CmdStop stops the current timer.
func (a *App) CmdStop(args []string) error {
	var timer Timer
	if !ReadStateKey(a.Paths.StateFile, "current", &timer) {
		fmt.Fprintln(a.Stderr, "No timer running.")
		return fmt.Errorf("exit 1")
	}

	now := a.NowFunc()
	task := timer.Task
	if task == "" {
		task = "Untitled"
	}
	started := time.Unix(timer.StartedEpoch, 0)
	duration := time.Duration(timer.Duration) * time.Second
	elapsed := now.Sub(started)

	// Timer already expired — finalize and clean up
	if elapsed >= duration {
		LogEnd(a.Paths, started, duration, task, true, a.NowFunc)
		ClearStateKey(a.Paths, "current")
		fmt.Fprintf(a.Stdout, "Cleared ended timer: %s\n", task)
		return nil
	}

	config := ReadConfig(a.Paths.ConfigFile)
	ClearStateKey(a.Paths, "current")
	LogEnd(a.Paths, started, duration, task, false, a.NowFunc)
	elapsedStr := FormatDuration(elapsed)
	Notify(a.Runner, "Boxed", fmt.Sprintf("Timer stopped: %s (%s elapsed)", task, elapsedStr))
	if config.NotifySound {
		PlaySoundByName(a.Runner, "Sosumi")
	}
	fmt.Fprintf(a.Stdout, "Timer stopped: %s (%s elapsed)\n", task, elapsedStr)
	return nil
}
