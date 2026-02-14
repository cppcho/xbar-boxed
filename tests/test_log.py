"""Tests for log insertion, log_start, and log_end."""

import time
from datetime import datetime


class TestInsertEntryInLog:
    def test_empty_log_creates_section(self, boxed_env):
        result = boxed_env.mod._insert_entry_in_log(
            "", "2025-01-15", "09:00:00 - ... task (25m)"
        )
        assert "# 2025-01-15" in result
        assert "09:00:00 - ... task (25m)" in result

    def test_existing_date_appends_entry(self, boxed_env):
        existing = "# 2025-01-15\n\n09:00:00 - 09:25:00 task1 (25m) ✓\n"
        result = boxed_env.mod._insert_entry_in_log(
            existing, "2025-01-15", "10:00:00 - ... task2 (30m)"
        )
        lines = result.strip().splitlines()
        # Both entries should be under the same date header
        assert lines.count("# 2025-01-15") == 1
        assert "09:00:00 - 09:25:00 task1 (25m) ✓" in result
        assert "10:00:00 - ... task2 (30m)" in result

    def test_newer_date_inserted_before_older(self, boxed_env):
        existing = "# 2025-01-14\n\n09:00:00 - 09:25:00 task (25m) ✓\n"
        result = boxed_env.mod._insert_entry_in_log(
            existing, "2025-01-15", "10:00:00 - ... new (25m)"
        )
        # Newer date should appear first
        idx_new = result.index("# 2025-01-15")
        idx_old = result.index("# 2025-01-14")
        assert idx_new < idx_old

    def test_older_date_appended_after_newer(self, boxed_env):
        existing = "# 2025-01-15\n\n10:00:00 - 10:25:00 task (25m) ✓\n"
        result = boxed_env.mod._insert_entry_in_log(
            existing, "2025-01-14", "09:00:00 - ... old (25m)"
        )
        idx_new = result.index("# 2025-01-15")
        idx_old = result.index("# 2025-01-14")
        assert idx_new < idx_old

    def test_entries_within_date_are_chronological(self, boxed_env):
        existing = "# 2025-01-15\n\n09:00:00 - 09:25:00 task1 (25m) ✓\n"
        result = boxed_env.mod._insert_entry_in_log(
            existing, "2025-01-15", "10:00:00 - ... task2 (30m)"
        )
        idx1 = result.index("task1")
        idx2 = result.index("task2")
        assert idx1 < idx2

    def test_result_ends_with_newline(self, boxed_env):
        result = boxed_env.mod._insert_entry_in_log(
            "", "2025-01-15", "09:00:00 - ... t (5m)"
        )
        assert result.endswith("\n")


class TestLogStart:
    def test_creates_log_file(self, boxed_env):
        epoch = 1705312800  # some fixed timestamp
        boxed_env.mod.log_start(epoch, 1500, "my task")
        assert boxed_env.log_file.exists()
        content = boxed_env.log_file.read_text()
        dt = datetime.fromtimestamp(epoch)
        date_str = dt.strftime("%Y-%m-%d")
        time_str = dt.strftime("%H:%M:%S")
        assert f"# {date_str}" in content
        assert f"{time_str} - ... my task (25m)" in content

    def test_partial_entry_format(self, boxed_env):
        epoch = 1705312800
        boxed_env.mod.log_start(epoch, 300, "quick")
        content = boxed_env.log_file.read_text()
        dt = datetime.fromtimestamp(epoch)
        time_str = dt.strftime("%H:%M:%S")
        assert f"{time_str} - ... quick (5m)" in content


class TestLogEnd:
    def test_replaces_partial_entry(self, boxed_env, monkeypatch):
        started = 1705312800
        duration = 1500  # 25m
        end_time = started + duration
        monkeypatch.setattr(time, "time", lambda: end_time)

        # First log_start to create partial entry
        boxed_env.mod.log_start(started, duration, "my task")
        content_before = boxed_env.log_file.read_text()
        assert "..." in content_before

        # Now log_end to replace it
        boxed_env.mod.log_end(started, duration, "my task", completed=True)
        content = boxed_env.log_file.read_text()
        assert "..." not in content
        assert "✓" in content

    def test_stopped_marker(self, boxed_env, monkeypatch):
        started = 1705312800
        duration = 1500
        end_time = started + 600  # stopped early
        monkeypatch.setattr(time, "time", lambda: end_time)

        boxed_env.mod.log_start(started, duration, "aborted")
        boxed_env.mod.log_end(started, duration, "aborted", completed=False)
        content = boxed_env.log_file.read_text()
        assert "✕" in content

    def test_fallback_when_no_partial(self, boxed_env, monkeypatch):
        started = 1705312800
        end_time = started + 300
        monkeypatch.setattr(time, "time", lambda: end_time)

        # Log end without a preceding log_start
        boxed_env.mod.log_end(started, 300, "orphan", completed=True)
        content = boxed_env.log_file.read_text()
        assert "orphan" in content
        assert "✓" in content
