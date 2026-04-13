package boxed

import (
	"fmt"
	"time"
)

// FormatTimer formats a time.Duration as a clock-style string for the menu bar.
// >= 1 hour: "H:MM:SS" (e.g., "1:30:00"), < 1 hour: "M:SS" (e.g., "25:00").
func FormatTimer(d time.Duration) string {
	total := int(d.Seconds())
	if total >= 3600 {
		h := total / 3600
		m := (total % 3600) / 60
		s := total % 60
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	m := total / 60
	s := total % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

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
