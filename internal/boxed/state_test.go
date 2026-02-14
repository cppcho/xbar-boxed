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
		LastFile:   filepath.Join(dir, "last.json"),
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

func TestWriteState_WritesStateFile(t *testing.T) {
	p := testPaths(t)
	if err := WriteState(p, "test", 1000, 300); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p.StateFile)
	var s State
	json.Unmarshal(data, &s)
	if s.Task != "test" || s.StartedEpoch != 1000 || s.Duration != 300 {
		t.Errorf("unexpected state: %+v", s)
	}
}

func TestWriteState_CreatesConfigDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new", "config")
	p := Paths{ConfigDir: dir, StateFile: filepath.Join(dir, "state.json")}
	WriteState(p, "t", 1, 60)
	if _, err := os.Stat(filepath.Join(dir, "state.json")); err != nil {
		t.Errorf("state file not created: %v", err)
	}
}

func TestReadState_NoFileReturnsNil(t *testing.T) {
	if s := ReadState("/nonexistent"); s != nil {
		t.Errorf("expected nil, got %+v", s)
	}
}

func TestReadState_ReadsValidState(t *testing.T) {
	p := testPaths(t)
	WriteState(p, "hello", 500, 120)
	s := ReadState(p.StateFile)
	if s == nil {
		t.Fatal("expected non-nil state")
	}
	if s.Task != "hello" || s.Duration != 120 {
		t.Errorf("unexpected state: %+v", s)
	}
}

func TestReadState_CorruptJSONReturnsNil(t *testing.T) {
	p := testPaths(t)
	os.WriteFile(p.StateFile, []byte("{invalid json"), 0o644)
	if s := ReadState(p.StateFile); s != nil {
		t.Errorf("expected nil for corrupt JSON, got %+v", s)
	}
}

func TestReadState_EmptyFileReturnsNil(t *testing.T) {
	p := testPaths(t)
	os.WriteFile(p.StateFile, []byte(""), 0o644)
	if s := ReadState(p.StateFile); s != nil {
		t.Errorf("expected nil for empty file, got %+v", s)
	}
}

func TestClearState_RemovesFile(t *testing.T) {
	p := testPaths(t)
	WriteState(p, "t", 1, 60)
	if err := ClearState(p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.StateFile); !os.IsNotExist(err) {
		t.Error("state file still exists after clear")
	}
}

func TestClearState_NoFile(t *testing.T) {
	p := testPaths(t)
	if err := ClearState(p); err != nil {
		t.Errorf("clear with no file should not error: %v", err)
	}
}

func TestReadLastTimer(t *testing.T) {
	p := testPaths(t)
	WriteLastTimer(p, 25, "my task")
	lt, err := ReadLastTimer(p.LastFile)
	if err != nil {
		t.Fatal(err)
	}
	if lt.Duration != 25 || lt.Task != "my task" {
		t.Errorf("unexpected last timer: %+v", lt)
	}
}

func TestReadLastTimer_CorruptFile(t *testing.T) {
	p := testPaths(t)
	os.WriteFile(p.LastFile, []byte("{bad json"), 0o644)
	_, err := ReadLastTimer(p.LastFile)
	if err == nil {
		t.Error("expected error for corrupt last file")
	}
}
