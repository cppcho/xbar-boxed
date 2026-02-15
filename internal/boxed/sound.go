package boxed

import "fmt"

// PlaySoundByName plays a macOS system sound by name (e.g. "Glass", "Sosumi").
func PlaySoundByName(runner CommandRunner, name string) {
	path := fmt.Sprintf("/System/Library/Sounds/%s.aiff", name)
	runner.Start("afplay", path)
}

// PlaySoundFile plays a custom sound file via afplay.
func PlaySoundFile(runner CommandRunner, filePath string) {
	runner.Start("afplay", filePath)
}
