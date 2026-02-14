package boxed

import "fmt"

// FormatDuration formats seconds as "1h30m", "25m", or "45s".
func FormatDuration(seconds int) string {
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
