package boxed

const (
	focusOnShortcut  = "Focus On"
	focusOffShortcut = "Focus Off"
)

// SetFocus activates or deactivates macOS Focus Mode by running a Shortcut.
// If the shortcut doesn't exist, the operation is silently skipped.
func SetFocus(runner CommandRunner, on bool) {
	shortcut := focusOffShortcut
	if on {
		shortcut = focusOnShortcut
	}
	runner.Start("shortcuts", "run", shortcut)
}
