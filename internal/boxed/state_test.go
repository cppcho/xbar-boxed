package boxed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testPaths(t *testing.T) Paths {
	dir := t.TempDir()
	return Paths{
		ConfigDir:  dir,
		StateFile:  filepath.Join(dir, "state.json"),
		ConfigFile: filepath.Join(dir, "config"),
		LogFile:    filepath.Join(dir, "log"),
		SoundsDir:  filepath.Join(dir, "sounds"),
	}
}

func TestAtomicWriteJSON_WritesValidJSON(t *testing.T) {
	p := testPaths(t)
	path := filepath.Join(p.ConfigDir, "test.json")
	if err := AtomicWriteJSON(path, map[string]string{"key": "value"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["key"] != "value" {
		t.Errorf("expected key=value, got %v", m)
	}
}

func TestAtomicWriteJSON_OverwritesExisting(t *testing.T) {
	p := testPaths(t)
	path := filepath.Join(p.ConfigDir, "test.json")
	AtomicWriteJSON(path, map[string]int{"a": 1})
	AtomicWriteJSON(path, map[string]int{"b": 2})
	data, _ := os.ReadFile(path)
	var m map[string]int
	json.Unmarshal(data, &m)
	if m["b"] != 2 {
		t.Errorf("expected b=2, got %v", m)
	}
	if _, ok := m["a"]; ok {
		t.Error("expected key 'a' to be gone")
	}
}

func TestAtomicWriteJSON_NoLeftoverTmpFiles(t *testing.T) {
	p := testPaths(t)
	path := filepath.Join(p.ConfigDir, "test.json")
	AtomicWriteJSON(path, map[string]int{"x": 1})
	entries, _ := os.ReadDir(p.ConfigDir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("found leftover tmp file: %s", e.Name())
		}
	}
}

func TestWriteStateKey_WritesKey(t *testing.T) {
	p := testPaths(t)
	timer := &Timer{Task: "test", StartedEpoch: 1000, Duration: 300}
	if err := WriteStateKey(p, StateKeyCurrent, timer); err != nil {
		t.Fatal(err)
	}
	var got Timer
	if !ReadStateKey(p.StateFile, StateKeyCurrent, &got) {
		t.Fatal("expected to read current key")
	}
	if got.Task != "test" || got.StartedEpoch != 1000 || got.Duration != 300 {
		t.Errorf("unexpected timer: %+v", got)
	}
}

func TestWriteStateKey_PreservesOtherKeys(t *testing.T) {
	p := testPaths(t)
	WriteStateKey(p, StateKeyCurrent, &Timer{Task: "work", StartedEpoch: 1, Duration: 60})
	WriteStateKey(p, StateKeyLast, &LastTimer{Duration: 1, Task: "work"})

	var timer Timer
	if !ReadStateKey(p.StateFile, StateKeyCurrent, &timer) {
		t.Fatal("expected current key to be preserved")
	}
	if timer.Task != "work" {
		t.Errorf("expected task='work', got %q", timer.Task)
	}
}

func TestWriteStateKey_CreatesConfigDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new", "config")
	p := Paths{ConfigDir: dir, StateFile: filepath.Join(dir, "state.json")}
	WriteStateKey(p, StateKeyCurrent, &Timer{Task: "t", StartedEpoch: 1, Duration: 60})
	if _, err := os.Stat(filepath.Join(dir, "state.json")); err != nil {
		t.Errorf("state file not created: %v", err)
	}
}

func TestReadStateKey_NoFileReturnsFalse(t *testing.T) {
	var timer Timer
	if ReadStateKey("/nonexistent", StateKeyCurrent, &timer) {
		t.Error("expected false for missing file")
	}
}

func TestReadStateKey_MissingKeyReturnsFalse(t *testing.T) {
	p := testPaths(t)
	WriteStateKey(p, StateKeyCurrent, &Timer{Task: "t", StartedEpoch: 1, Duration: 60})
	var lt LastTimer
	if ReadStateKey(p.StateFile, StateKeyLast, &lt) {
		t.Error("expected false for missing key")
	}
}

func TestReadStateKey_CorruptJSONReturnsFalse(t *testing.T) {
	p := testPaths(t)
	os.WriteFile(p.StateFile, []byte("{invalid json"), 0o644)
	var timer Timer
	if ReadStateKey(p.StateFile, StateKeyCurrent, &timer) {
		t.Error("expected false for corrupt JSON")
	}
}

func TestReadStateKey_EmptyFileReturnsFalse(t *testing.T) {
	p := testPaths(t)
	os.WriteFile(p.StateFile, []byte(""), 0o644)
	var timer Timer
	if ReadStateKey(p.StateFile, StateKeyCurrent, &timer) {
		t.Error("expected false for empty file")
	}
}

func TestClearStateKey_PreservesOtherKeys(t *testing.T) {
	p := testPaths(t)
	WriteStateKey(p, StateKeyCurrent, &Timer{Task: "t", StartedEpoch: 1, Duration: 60})
	WriteStateKey(p, StateKeyLast, &LastTimer{Duration: 1, Task: "t"})

	if err := ClearStateKey(p, StateKeyCurrent); err != nil {
		t.Fatal(err)
	}
	var timer Timer
	if ReadStateKey(p.StateFile, StateKeyCurrent, &timer) {
		t.Error("expected current to be gone")
	}
	var lt LastTimer
	if !ReadStateKey(p.StateFile, StateKeyLast, &lt) {
		t.Fatal("expected last to be preserved")
	}
	if lt.Task != "t" {
		t.Errorf("expected task='t', got %q", lt.Task)
	}
}

func TestClearStateKey_RemovesFileWhenEmpty(t *testing.T) {
	p := testPaths(t)
	WriteStateKey(p, StateKeyCurrent, &Timer{Task: "t", StartedEpoch: 1, Duration: 60})

	if err := ClearStateKey(p, StateKeyCurrent); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.StateFile); !os.IsNotExist(err) {
		t.Error("state file should be removed when no keys remain")
	}
}

func TestClearStateKey_NoFile(t *testing.T) {
	p := testPaths(t)
	if err := ClearStateKey(p, StateKeyCurrent); err != nil {
		t.Errorf("clear with no file should not error: %v", err)
	}
}
