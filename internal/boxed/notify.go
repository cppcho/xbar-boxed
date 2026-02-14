package boxed

import (
	"fmt"
	"strings"
)

// Notify sends a macOS notification via osascript.
func Notify(runner CommandRunner, title, message string) {
	safeTitle := escapeAppleScript(title)
	safeMessage := escapeAppleScript(message)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, safeMessage, safeTitle)
	runner.Run("osascript", "-e", script)
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
