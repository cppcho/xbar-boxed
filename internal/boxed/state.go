package boxed

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State represents the current timer state persisted in state.json.
// JSON field names match the Python version for compatibility.
type State struct {
	Task          string `json:"task"`
	StartedEpoch  int64  `json:"started_epoch"`
	Duration      int    `json:"duration"`
	Notified      bool   `json:"notified,omitempty"`
	LastTickEpoch int64  `json:"last_tick_epoch,omitempty"`
}

// LastTimer stores the most recent timer params for "again" command.
type LastTimer struct {
	Duration int    `json:"duration"`
	Task     string `json:"task"`
}

// AtomicWriteJSON writes data as JSON atomically via temp file + fsync + rename.
func AtomicWriteJSON(filepath_ string, data any) error {
	dir := filepath.Dir(filepath_)

	tmp, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, filepath_)
}

// ReadState reads the timer state from disk. Returns nil if missing or corrupt.
func ReadState(stateFile string) *State {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

// WriteState writes timer state atomically.
func WriteState(p Paths, task string, startedEpoch int64, duration int) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	s := State{
		Task:         task,
		StartedEpoch: startedEpoch,
		Duration:     duration,
	}
	return AtomicWriteJSON(p.StateFile, s)
}

// WriteStateFull writes a complete State struct atomically.
func WriteStateFull(p Paths, s *State) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	return AtomicWriteJSON(p.StateFile, s)
}

// ClearState removes the state file.
func ClearState(p Paths) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	err := os.Remove(p.StateFile)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ReadLastTimer reads the last timer params from disk.
func ReadLastTimer(lastFile string) (*LastTimer, error) {
	data, err := os.ReadFile(lastFile)
	if err != nil {
		return nil, err
	}
	var lt LastTimer
	if err := json.Unmarshal(data, &lt); err != nil {
		return nil, err
	}
	return &lt, nil
}

// WriteLastTimer writes last timer params atomically.
func WriteLastTimer(p Paths, duration int, task string) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	lt := LastTimer{Duration: duration, Task: task}
	return AtomicWriteJSON(p.LastFile, lt)
}
