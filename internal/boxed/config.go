package boxed

import (
	"os"
	"strings"
	"time"
)

// Config holds typed configuration values.
type Config struct {
	NotifySound  bool          // default: true
	TickInterval time.Duration // minutes, default: 0 (disabled)
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{NotifySound: true, TickInterval: 0}
}

const defaultConfigContent = `# Boxed configuration
# notify_sound = true
# tick_interval = 5m
`

// ReadConfig reads a key=value config file into a typed Config struct.
func ReadConfig(configFile string) Config {
	config := DefaultConfig()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, found := strings.Cut(line, "="); found {
			key := strings.TrimSpace(key)
			value := strings.TrimSpace(value)
			switch key {
			case "notify_sound":
				switch value {
				case "true":
					config.NotifySound = true
				case "false":
					config.NotifySound = false
				}
			case "tick_interval":
				if tickInterval, err := time.ParseDuration(value); err == nil {
					config.TickInterval = tickInterval
				}
			}
		}
	}
	return config
}

// EnsureConfig creates the default config file if it doesn't exist.
func EnsureConfig(p Paths) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	if _, err := os.Stat(p.ConfigFile); os.IsNotExist(err) {
		return os.WriteFile(p.ConfigFile, []byte(defaultConfigContent), 0o644)
	}
	return nil
}
