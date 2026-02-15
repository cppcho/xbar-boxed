package boxed

import (
	"fmt"
	"io"
	"path/filepath"
)

// XbarApp holds dependencies for the xbar plugin.
type XbarApp struct {
	Paths    Paths
	Runner   CommandRunner
	NowFunc  func() int64
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
		duration := state.Duration
		started := state.StartedEpoch
		now := x.NowFunc()
		remaining := started + int64(duration) - now

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
				intervalSecs := int64(config.TickInterval * 60)
				lastTick := state.LastTickEpoch
				if lastTick == 0 {
					lastTick = started
				}
				if now-lastTick >= intervalSecs {
					PlaySoundFile(x.Runner, filepath.Join(x.Paths.SoundsDir, "PeonYes3.ogg"))
					state.LastTickEpoch = now
					WriteStateFull(x.Paths, state)
				}
			}
			x.out(fmt.Sprintf("%s (%s)", task, FormatDuration(int(remaining))))
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
