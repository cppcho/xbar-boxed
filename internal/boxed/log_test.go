package boxed

import (
	"strings"
	"testing"
	"time"
)

func TestInsertEntryInLog_EmptyLog(t *testing.T) {
	result := InsertEntryInLog("", "2025-01-15", "09:00:00 - ... task (25m)")
	if !strings.Contains(result, "# 2025-01-15") {
		t.Error("expected date header")
	}
	if !strings.Contains(result, "09:00:00 - ... task (25m)") {
		t.Error("expected entry")
	}
}

func TestInsertEntryInLog_ExistingDateAppendsEntry(t *testing.T) {
	existing := "# 2025-01-15\n\n09:00:00 - 09:25:00 task1 (25m) ✓\n"
	result := InsertEntryInLog(existing, "2025-01-15", "10:00:00 - ... task2 (30m)")
	if strings.Count(result, "# 2025-01-15") != 1 {
		t.Error("expected exactly one date header")
	}
	if !strings.Contains(result, "09:00:00 - 09:25:00 task1 (25m) ✓") {
		t.Error("expected existing entry preserved")
	}
	if !strings.Contains(result, "10:00:00 - ... task2 (30m)") {
		t.Error("expected new entry")
	}
}

func TestInsertEntryInLog_NewerDateInsertedBefore(t *testing.T) {
	existing := "# 2025-01-14\n\n09:00:00 - 09:25:00 task (25m) ✓\n"
	result := InsertEntryInLog(existing, "2025-01-15", "10:00:00 - ... new (25m)")
	idxNew := strings.Index(result, "# 2025-01-15")
	idxOld := strings.Index(result, "# 2025-01-14")
	if idxNew >= idxOld {
		t.Error("newer date should appear before older date")
	}
}

func TestInsertEntryInLog_OlderDateAppendedAfter(t *testing.T) {
	existing := "# 2025-01-15\n\n10:00:00 - 10:25:00 task (25m) ✓\n"
	result := InsertEntryInLog(existing, "2025-01-14", "09:00:00 - ... old (25m)")
	idxNew := strings.Index(result, "# 2025-01-15")
	idxOld := strings.Index(result, "# 2025-01-14")
	if idxNew >= idxOld {
		t.Error("newer date should appear before older date")
	}
}

func TestInsertEntryInLog_EntriesChronological(t *testing.T) {
	existing := "# 2025-01-15\n\n09:00:00 - 09:25:00 task1 (25m) ✓\n"
	result := InsertEntryInLog(existing, "2025-01-15", "10:00:00 - ... task2 (30m)")
	idx1 := strings.Index(result, "task1")
	idx2 := strings.Index(result, "task2")
	if idx1 >= idx2 {
		t.Error("task1 should appear before task2")
	}
}

func TestInsertEntryInLog_EndsWithNewline(t *testing.T) {
	result := InsertEntryInLog("", "2025-01-15", "09:00:00 - ... t (5m)")
	if !strings.HasSuffix(result, "\n") {
		t.Error("result should end with newline")
	}
}

func TestLogStart_CreatesLogFile(t *testing.T) {
	p := testPaths(t)
	started := time.Unix(1705312800, 0)
	if err := LogStart(p, started, 25*time.Minute, "my task"); err != nil {
		t.Fatal(err)
	}
	content := readLog(p)
	dateStr := started.Format("2006-01-02")
	timeStr := started.Format("15:04:05")
	if !strings.Contains(content, "# "+dateStr) {
		t.Error("expected date header in log")
	}
	if !strings.Contains(content, timeStr+" - ... my task (25m)") {
		t.Errorf("expected partial entry, got:\n%s", content)
	}
}

func TestLogStart_PartialEntryFormat(t *testing.T) {
	p := testPaths(t)
	started := time.Unix(1705312800, 0)
	LogStart(p, started, 5*time.Minute, "quick")
	content := readLog(p)
	timeStr := started.Format("15:04:05")
	if !strings.Contains(content, timeStr+" - ... quick (5m)") {
		t.Errorf("expected partial entry format, got:\n%s", content)
	}
}

func TestLogEnd_ReplacesPartialEntry(t *testing.T) {
	p := testPaths(t)
	started := time.Unix(1705312800, 0)
	duration := 25 * time.Minute

	LogStart(p, started, duration, "my task")
	contentBefore := readLog(p)
	if !strings.Contains(contentBefore, "...") {
		t.Error("expected partial entry with ...")
	}

	LogEnd(p, started, duration, "my task", true, func() time.Time { return started.Add(duration) })
	content := readLog(p)
	if strings.Contains(content, "...") {
		t.Error("partial entry should be replaced")
	}
	if !strings.Contains(content, "✓") {
		t.Error("expected checkmark")
	}
}

func TestLogEnd_StoppedMarker(t *testing.T) {
	p := testPaths(t)
	started := time.Unix(1705312800, 0)
	duration := 25 * time.Minute

	LogStart(p, started, duration, "aborted")
	LogEnd(p, started, duration, "aborted", false, func() time.Time { return started.Add(10 * time.Minute) })
	content := readLog(p)
	if !strings.Contains(content, "✕") {
		t.Error("expected cross marker")
	}
}

func TestLogEnd_FallbackWhenNoPartial(t *testing.T) {
	p := testPaths(t)
	started := time.Unix(1705312800, 0)
	duration := 5 * time.Minute

	// Log end without a preceding log_start
	LogEnd(p, started, duration, "orphan", true, func() time.Time { return started.Add(duration) })
	content := readLog(p)
	if !strings.Contains(content, "orphan") {
		t.Error("expected orphan entry")
	}
	if !strings.Contains(content, "✓") {
		t.Error("expected checkmark")
	}
}
