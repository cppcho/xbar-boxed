package boxed

import (
	"os"
	"path/filepath"
)

// Paths holds all filesystem paths used by Boxed.
// Tests substitute these with temp directory paths.
type Paths struct {
	ConfigDir string
	StateFile string
	ConfigFile string
	LogFile   string
	LastFile  string
	SoundsDir string
}

// DefaultPaths returns production paths under ~/.config/boxed.
func DefaultPaths() Paths {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "boxed")
	// SoundsDir is relative to the binary's location at build time,
	// but for installed binaries we use a fixed path.
	soundsDir := filepath.Join(home, ".local", "lib", "boxed", "sounds")
	return Paths{
		ConfigDir:  configDir,
		StateFile:  filepath.Join(configDir, "state.json"),
		ConfigFile: filepath.Join(configDir, "config"),
		LogFile:    filepath.Join(configDir, "log"),
		LastFile:   filepath.Join(configDir, "last.json"),
		SoundsDir:  soundsDir,
	}
}

// EnsureDirs creates the config directory if it doesn't exist.
func (p Paths) EnsureDirs() error {
	return os.MkdirAll(p.ConfigDir, 0o755)
}
