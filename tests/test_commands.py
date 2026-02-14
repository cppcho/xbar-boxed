"""Tests for cmd_start, cmd_stop, cmd_complete, cmd_again."""

import json
import time

import pytest


class TestCmdStart:
    def test_basic_start(self, boxed_env, monkeypatch):
        now = 1700000000
        monkeypatch.setattr(time, "time", lambda: now)
        boxed_env.mod.cmd_start(["25", "my", "task"])
        state = json.loads(boxed_env.state_file.read_text())
        assert state["task"] == "my task"
        assert state["started_epoch"] == now
        assert state["duration"] == 1500

    def test_start_saves_last_file(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["10", "quick"])
        last = json.loads(boxed_env.last_file.read_text())
        assert last["duration"] == 10
        assert last["task"] == "quick"

    def test_start_sends_notification(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["5", "notify"])
        boxed_env.subprocess_mock.run.assert_called_once()
        call_args = boxed_env.subprocess_mock.run.call_args
        assert call_args[0][0][0] == "osascript"

    def test_start_plays_sound_when_enabled(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.config_file.write_text("notify_sound = true\n")
        boxed_env.mod.cmd_start(["5", "sound"])
        boxed_env.subprocess_mock.Popen.assert_called_once()

    def test_start_no_sound_when_disabled(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.config_file.write_text("notify_sound = false\n")
        boxed_env.mod.cmd_start(["5", "quiet"])
        boxed_env.subprocess_mock.Popen.assert_not_called()

    def test_start_replaces_running_timer(self, boxed_env, monkeypatch):
        # Start first timer
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "first"])

        # Start second timer (replaces first, which was still running)
        monkeypatch.setattr(time, "time", lambda: 1700000060)
        boxed_env.mod.cmd_start(["10", "second"])
        state = json.loads(boxed_env.state_file.read_text())
        assert state["task"] == "second"
        assert state["duration"] == 600

    def test_start_replaces_expired_timer(self, boxed_env, monkeypatch):
        # Start first timer
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "short"])

        # Start second after first expired
        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_start(["10", "next"])
        state = json.loads(boxed_env.state_file.read_text())
        assert state["task"] == "next"

    def test_start_creates_log_entry(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["5", "logged"])
        assert boxed_env.log_file.exists()
        content = boxed_env.log_file.read_text()
        assert "logged" in content

    def test_start_prints_message(self, boxed_env, monkeypatch, capsys):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["5", "hello"])
        captured = capsys.readouterr()
        assert "Timer started" in captured.out
        assert "hello" in captured.out

    def test_start_missing_args(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_start([])

    def test_start_missing_task(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_start(["25"])

    def test_start_invalid_duration(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_start(["abc", "task"])

    def test_start_zero_duration(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_start(["0", "task"])

    def test_start_negative_duration(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_start(["-5", "task"])


class TestCmdStop:
    def test_stop_running_timer(self, boxed_env, monkeypatch, capsys):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "work"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_stop([])
        assert not boxed_env.state_file.exists()
        captured = capsys.readouterr()
        assert "stopped" in captured.out

    def test_stop_sends_notification(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "work"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_stop([])
        boxed_env.subprocess_mock.run.assert_called_once()

    def test_stop_expired_timer_clears_state(self, boxed_env, monkeypatch, capsys):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "short"])
        boxed_env.subprocess_mock.reset_mock()

        # Timer has expired (60s timer, 120s elapsed)
        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_stop([])
        assert not boxed_env.state_file.exists()
        captured = capsys.readouterr()
        assert "Cleared" in captured.out

    def test_stop_no_timer(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_stop([])

    def test_stop_logs_cross_marker(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "stopped"])
        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_stop([])
        content = boxed_env.log_file.read_text()
        assert "✕" in content

    def test_stop_plays_sound_when_enabled(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.config_file.write_text("notify_sound = true\n")
        boxed_env.mod.cmd_start(["25", "work"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_stop([])
        boxed_env.subprocess_mock.Popen.assert_called_once()

    def test_stop_no_sound_when_disabled(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.config_file.write_text("notify_sound = false\n")
        boxed_env.mod.cmd_start(["25", "quiet"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_stop([])
        boxed_env.subprocess_mock.Popen.assert_not_called()


class TestCmdComplete:
    def test_complete_marks_notified(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "done"])
        boxed_env.subprocess_mock.reset_mock()

        # Timer expired
        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_complete([])
        state = json.loads(boxed_env.state_file.read_text())
        assert state["notified"] is True

    def test_complete_sends_notification(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "done"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_complete([])
        boxed_env.subprocess_mock.run.assert_called_once()

    def test_complete_idempotent(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "done"])

        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_complete([])
        boxed_env.subprocess_mock.reset_mock()

        # Second call should be a no-op (already notified)
        boxed_env.mod.cmd_complete([])
        boxed_env.subprocess_mock.run.assert_not_called()

    def test_complete_not_expired_yet(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "running"])
        boxed_env.subprocess_mock.reset_mock()

        # Only 5 minutes in, timer not expired
        monkeypatch.setattr(time, "time", lambda: 1700000300)
        boxed_env.mod.cmd_complete([])
        # Should not send notification
        boxed_env.subprocess_mock.run.assert_not_called()

    def test_complete_no_state(self, boxed_env):
        # Should not raise
        boxed_env.mod.cmd_complete([])

    def test_complete_logs_checkmark(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["1", "finished"])

        monkeypatch.setattr(time, "time", lambda: 1700000120)
        boxed_env.mod.cmd_complete([])
        content = boxed_env.log_file.read_text()
        assert "✓" in content


class TestCmdAgain:
    def test_again_repeats_last(self, boxed_env, monkeypatch):
        monkeypatch.setattr(time, "time", lambda: 1700000000)
        boxed_env.mod.cmd_start(["25", "repeated"])
        boxed_env.subprocess_mock.reset_mock()

        monkeypatch.setattr(time, "time", lambda: 1700002000)
        boxed_env.mod.cmd_again([])
        state = json.loads(boxed_env.state_file.read_text())
        assert state["task"] == "repeated"
        assert state["duration"] == 1500

    def test_again_no_previous(self, boxed_env):
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_again([])

    def test_again_corrupt_last_file(self, boxed_env):
        boxed_env.last_file.write_text("{bad json")
        with pytest.raises(SystemExit):
            boxed_env.mod.cmd_again([])
