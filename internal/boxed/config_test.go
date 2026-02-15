package boxed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfig_MissingFileReturnsDefaults(t *testing.T) {
	config := ReadConfig("/nonexistent/path")
	if !config.NotifySound {
		t.Error("expected NotifySound=true")
	}
	if config.TickInterval != 0 {
		t.Errorf("expected TickInterval=0, got %d", config.TickInterval)
	}
}

func TestReadConfig_EmptyFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte(""), 0o644)
	config := ReadConfig(f)
	if !config.NotifySound {
		t.Error("expected NotifySound=true")
	}
}

func TestReadConfig_CommentsIgnored(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte("# this is a comment\n"), 0o644)
	config := ReadConfig(f)
	if !config.NotifySound {
		t.Error("expected NotifySound=true")
	}
}

func TestReadConfig_KeyValueParsed(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte("notify_sound = false\n"), 0o644)
	config := ReadConfig(f)
	if config.NotifySound {
		t.Error("expected NotifySound=false")
	}
}

func TestReadConfig_WhitespaceTrimmed(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte("  notify_sound  =  false  \n"), 0o644)
	config := ReadConfig(f)
	if config.NotifySound {
		t.Error("expected NotifySound=false")
	}
}

func TestReadConfig_MultipleKeys(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte("notify_sound = false\ntick_interval = 5\n"), 0o644)
	config := ReadConfig(f)
	if config.NotifySound {
		t.Error("expected NotifySound=false")
	}
	if config.TickInterval != 5 {
		t.Errorf("expected TickInterval=5, got %d", config.TickInterval)
	}
}

func TestReadConfig_BlankLinesSkipped(t *testing.T) {
	f := filepath.Join(t.TempDir(), "config")
	os.WriteFile(f, []byte("\n\nnotify_sound = false\n\n"), 0o644)
	config := ReadConfig(f)
	if config.NotifySound {
		t.Error("expected NotifySound=false")
	}
}


func TestEnsureConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	p := Paths{ConfigDir: dir, ConfigFile: filepath.Join(dir, "config")}
	if err := EnsureConfig(p); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty config file")
	}
}

func TestEnsureConfig_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	p := Paths{ConfigDir: dir, ConfigFile: filepath.Join(dir, "config")}
	os.WriteFile(p.ConfigFile, []byte("custom"), 0o644)
	EnsureConfig(p)
	data, _ := os.ReadFile(p.ConfigFile)
	if string(data) != "custom" {
		t.Errorf("expected existing content preserved, got %q", string(data))
	}
}
