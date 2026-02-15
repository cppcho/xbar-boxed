package boxed

import (
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// XbarApp holds dependencies for the xbar plugin.
type XbarApp struct {
	Paths   Paths
	Runner  CommandRunner
	NowFunc func() time.Time
	Stdout  io.Writer
}

// Run renders the xbar menu bar output.
func (x *XbarApp) Run() {
	var timer CurrentTimer
	hasTimer := ReadStateKey(x.Paths.StateFile, StateKeyCurrent, &timer)

	if hasTimer {
		task := timer.Task
		if task == "" {
			task = "Untitled"
		}
		now := x.NowFunc()
		remaining := timer.Duration - now.Sub(timer.StartedAt)

		if remaining <= 0 {
			if !timer.Notified {
				CompleteTimer(x.Paths, x.Runner, x.NowFunc)
			}
			x.out("📦")
		} else {
			config := ReadConfig(x.Paths.ConfigFile)
			if config.TickInterval > 0 {
				lastTick := timer.LastTickAt
				if timer.LastTickAt.IsZero() {
					lastTick = timer.StartedAt
				}
				if now.Sub(lastTick) >= config.TickInterval {
					// TODO: should i change sound?
					PlaySoundFile(x.Runner, filepath.Join(x.Paths.SoundsDir, "PeonYes3.ogg"))
					timer.LastTickAt = now
					WriteStateKey(x.Paths, StateKeyCurrent, &timer)
				}
			}
			x.out(fmt.Sprintf("%s (%s)", task, FormatDuration(remaining)))
		}
		x.out("---")
	} else {
		x.out("📦")
		x.out("---")
	}
	x.out(fmt.Sprintf("Open Config | bash=/usr/bin/open param1=%s terminal=false", x.Paths.ConfigFile))
	x.out(fmt.Sprintf("Open Log | bash=/usr/bin/open param1=%s terminal=false", x.Paths.LogFile))
	x.out("---")
	x.out(fmt.Sprintf("Open Config Directory | bash=/usr/bin/open param1=%s terminal=false", x.Paths.ConfigDir))
}

func (x *XbarApp) out(text string) {
	fmt.Fprintln(x.Stdout, text)
}
