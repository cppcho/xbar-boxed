package boxed

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// StateKey is a typed key for state.json entries.
type StateKey string

// Valid state keys.
const (
	StateKeyCurrent StateKey = "current"
	StateKeyLast    StateKey = "last"
)

// CurrentTimer represents an active timer's fields.
type CurrentTimer struct {
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

// readStateMap reads the state file as a raw key-value map.
func readStateMap(stateFile string) map[string]json.RawMessage {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// ReadStateKey reads a specific key from state.json and unmarshals into dest.
// Returns false if the file is missing, corrupt, or the key doesn't exist.
func ReadStateKey(stateFile string, key StateKey, dest any) bool {
	m := readStateMap(stateFile)
	if m == nil {
		return false
	}
	raw, ok := m[string(key)]
	if !ok {
		return false
	}
	return json.Unmarshal(raw, dest) == nil
}

// WriteStateKey writes a value under the given key in state.json (read-modify-write).
func WriteStateKey(p Paths, key StateKey, value any) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	m := readStateMap(p.StateFile)
	if m == nil {
		m = make(map[string]json.RawMessage)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m[string(key)] = raw
	return AtomicWriteJSON(p.StateFile, m)
}

// ClearStateKey removes a key from state.json. Removes the file if no keys remain.
func ClearStateKey(p Paths, key StateKey) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	m := readStateMap(p.StateFile)
	if m == nil {
		return nil
	}
	delete(m, string(key))
	if len(m) == 0 {
		err := os.Remove(p.StateFile)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return AtomicWriteJSON(p.StateFile, m)
}
