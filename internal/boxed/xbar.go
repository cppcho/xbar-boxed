package boxed

import (
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// XbarApp holds dependencies for the xbar plugin.
type XbarApp struct {
	Paths    Paths
	Runner   CommandRunner
	NowFunc  func() time.Time
	Stdout   io.Writer
	BoxedBin string // path to the boxed CLI binary
}

// Run renders the xbar menu bar output.
func (x *XbarApp) Run() {
	state := ReadState(x.Paths.StateFile)

	if state != nil {
		task := state.Task
		if task == "" {
			task = "Untitled"
		}
		now := x.NowFunc()
		started := time.Unix(state.StartedEpoch, 0)
		duration := time.Duration(state.Duration) * time.Second
		remaining := duration - now.Sub(started)

		if remaining <= 0 {
			if !state.Notified {
				x.Runner.Run(x.BoxedBin, "complete")
				// Re-read state after complete
				state = ReadState(x.Paths.StateFile)
			}
			x.out("📦")
		} else {
			config := ReadConfig(x.Paths.ConfigFile)
			if config.TickInterval > 0 {
				intervalDur := time.Duration(config.TickInterval) * time.Minute
				lastTick := time.Unix(state.LastTickEpoch, 0)
				if state.LastTickEpoch == 0 {
					lastTick = started
				}
				if now.Sub(lastTick) >= intervalDur {
					PlaySoundFile(x.Runner, filepath.Join(x.Paths.SoundsDir, "PeonYes3.ogg"))
					state.LastTickEpoch = now.Unix()
					WriteStateFull(x.Paths, state)
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
