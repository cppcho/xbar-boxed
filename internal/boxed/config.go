package boxed

import (
	"os"
	"strings"
)

const defaultConfigContent = `# Boxed configuration
# notify_sound = true
# tick_interval = 5
`

// ReadConfig reads a key=value config file, merged over defaults.
func ReadConfig(configFile string, defaults map[string]string) map[string]string {
	config := make(map[string]string)
	for k, v := range defaults {
		config[k] = v
	}

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
			config[key] = value
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
