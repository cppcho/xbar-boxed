package boxed

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// InsertEntryInLog inserts an entry under the correct date section.
// Dates are ordered newest-first; entries within a date are chronological.
func InsertEntryInLog(content, dateStr, entryLine string) string {
	header := "# " + dateStr
	var lines []string
	if content != "" {
		lines = strings.Split(content, "\n")
	}

	// Find if this date section already exists
	headerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			headerIdx = i
			break
		}
	}

	if headerIdx >= 0 {
		// Find end of this date section (next header or end of file)
		insertIdx := headerIdx + 1
		for insertIdx < len(lines) {
			if strings.HasPrefix(lines[insertIdx], "# ") && lines[insertIdx] != header {
				break
			}
			insertIdx++
		}
		// Back up past trailing blank lines
		for insertIdx > headerIdx+1 && strings.TrimSpace(lines[insertIdx-1]) == "" {
			insertIdx--
		}
		// Insert the entry
		lines = append(lines[:insertIdx], append([]string{entryLine}, lines[insertIdx:]...)...)
	} else {
		// Need to create a new date section — newest dates first
		insertBefore := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "# ") {
				existingDate := strings.TrimSpace(line[2:])
				if dateStr > existingDate {
					insertBefore = i
					break
				}
			}
		}
		if insertBefore >= 0 {
			block := []string{header, "", entryLine, ""}
			lines = append(lines[:insertBefore], append(block, lines[insertBefore:]...)...)
		} else {
			// Append at end (oldest date or first entry)
			if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
				lines = append(lines, "")
			}
			lines = append(lines, header, "", entryLine, "")
		}
	}

	// Clean up: ensure single trailing newline
	result := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	return result
}

// readLog reads the log file, returning empty string if missing.
func readLog(p Paths) string {
	data, err := os.ReadFile(p.LogFile)
	if err != nil {
		return ""
	}
	return string(data)
}

// writeLog writes the log file atomically.
func writeLog(p Paths, content string) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	return atomicWriteText(p.LogFile, content)
}

// atomicWriteText writes text atomically via temp file + fsync + rename.
func atomicWriteText(filePath, text string) error {
	dir := dirOf(filePath)
	tmp, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(text); err != nil {
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
	return os.Rename(tmpName, filePath)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}

// LogStart logs a partial entry for a timer that just started.
func LogStart(p Paths, startedEpoch int64, durationSecs int, task string) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	dt := time.Unix(startedEpoch, 0)
	dateStr := dt.Format("2006-01-02")
	timeStr := dt.Format("15:04:05")
	entry := fmt.Sprintf("%s - ... %s (%s)", timeStr, task, FormatDuration(durationSecs))
	content := readLog(p)
	content = InsertEntryInLog(content, dateStr, entry)
	return writeLog(p, content)
}

// LogEnd updates a partial log entry to its final form, or appends if not found.
func LogEnd(p Paths, startedEpoch int64, durationSecs int, task string, completed bool, nowFunc func() int64) error {
	if err := p.EnsureDirs(); err != nil {
		return err
	}
	dt := time.Unix(startedEpoch, 0)
	dateStr := dt.Format("2006-01-02")
	startTime := dt.Format("15:04:05")
	now := nowFunc()
	endTime := time.Unix(now, 0).Format("15:04:05")

	symbol := "✕"
	if completed {
		symbol = "✓"
	}
	elapsed := int(now - startedEpoch)
	configuredDur := FormatDuration(durationSecs)
	elapsedDur := FormatDuration(elapsed)

	partial := fmt.Sprintf("%s - ... %s (%s)", startTime, task, configuredDur)
	final := fmt.Sprintf("%s - %s %s (%s) %s", startTime, endTime, task, elapsedDur, symbol)

	content := readLog(p)
	if strings.Contains(content, partial) {
		content = strings.Replace(content, partial, final, 1)
		return writeLog(p, content)
	}
	// Fallback: insert a complete entry
	content = InsertEntryInLog(content, dateStr, final)
	return writeLog(p, content)
}
