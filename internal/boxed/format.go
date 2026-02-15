package boxed

import (
	"fmt"
	"time"
)

// FormatDuration formats a time.Duration as "1h30m", "25m", or "45s".
func FormatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds >= 3600 {
		h := seconds / 3600
		m := (seconds % 3600) / 60
		if m > 0 {
			return fmt.Sprintf("%dh%dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	} else if seconds >= 60 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%ds", seconds)
}
