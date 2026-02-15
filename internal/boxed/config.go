package boxed

import (
	"os"
	"strconv"
	"strings"
)

// Config holds typed configuration values.
type Config struct {
	NotifySound  bool // default: true
	TickInterval int  // minutes, default: 0 (disabled)
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{NotifySound: true, TickInterval: 0}
}

const defaultConfigContent = `# Boxed configuration
# notify_sound = true
# tick_interval = 5
`

// ReadConfig reads a key=value config file into a typed Config struct.
func ReadConfig(configFile string) Config {
	config := DefaultConfig()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			switch key {
			case "notify_sound":
				switch value {
				case "true":
					config.NotifySound = true
				case "false":
					config.NotifySound = false
				}
			case "tick_interval":
				num, err := strconv.Atoi(value)
				if err == nil {
					config.TickInterval = num
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
