"""Tests for the xbar plugin main() output."""

import json
import time


class TestXbarNoTimer:
    def test_shows_box_emoji(self, xbar_env, capsys):
        xbar_env.mod.main()
        out = capsys.readouterr().out
        lines = out.strip().splitlines()
        assert lines[0] == "\U0001f4e6"  # 📦

    def test_shows_menu_items(self, xbar_env, capsys):
        xbar_env.mod.main()
        out = capsys.readouterr().out
        assert "Open Config |" in out
        assert "Open Log |" in out
        assert "Open Config Directory |" in out


class TestXbarRunningTimer:
    def test_shows_task_and_remaining(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {
            "task": "focus work",
            "started_epoch": now,
            "duration": 1500,
        }
        xbar_env.state_file.write_text(json.dumps(state))
        monkeypatch.setattr(time, "time", lambda: now + 300)
        xbar_env.mod.main()
        out = capsys.readouterr().out
        first_line = out.strip().splitlines()[0]
        assert "focus work" in first_line
        assert "20m" in first_line

    def test_separator_present(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {"task": "t", "started_epoch": now, "duration": 1500}
        xbar_env.state_file.write_text(json.dumps(state))
        monkeypatch.setattr(time, "time", lambda: now + 60)
        xbar_env.mod.main()
        out = capsys.readouterr().out
        assert "---" in out


class TestXbarExpiredTimer:
    def test_shows_box_emoji_when_expired(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {
            "task": "done",
            "started_epoch": now,
            "duration": 60,
            "notified": True,
        }
        xbar_env.state_file.write_text(json.dumps(state))
        monkeypatch.setattr(time, "time", lambda: now + 120)
        xbar_env.mod.main()
        out = capsys.readouterr().out
        first_line = out.strip().splitlines()[0]
        assert first_line == "\U0001f4e6"  # 📦

    def test_expired_not_notified_calls_complete(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {
            "task": "expired",
            "started_epoch": now,
            "duration": 60,
        }
        xbar_env.state_file.write_text(json.dumps(state))
        monkeypatch.setattr(time, "time", lambda: now + 120)
        xbar_env.mod.main()
        # Should have called subprocess.run to invoke boxed.py complete
        xbar_env.subprocess_mock.run.assert_called_once()
        call_args = xbar_env.subprocess_mock.run.call_args
        cmd = call_args[0][0]
        assert "complete" in cmd


class TestXbarTickSound:
    def test_tick_plays_when_interval_reached(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {
            "task": "tick test",
            "started_epoch": now,
            "duration": 1500,
        }
        xbar_env.state_file.write_text(json.dumps(state))
        xbar_env.config_file.write_text("tick_interval = 5\n")
        # 5 minutes later
        monkeypatch.setattr(time, "time", lambda: now + 300)
        xbar_env.mod.main()
        xbar_env.subprocess_mock.Popen.assert_called_once()

    def test_no_tick_before_interval(self, xbar_env, monkeypatch, capsys):
        now = 1700000000
        state = {
            "task": "tick test",
            "started_epoch": now,
            "duration": 1500,
        }
        xbar_env.state_file.write_text(json.dumps(state))
        xbar_env.config_file.write_text("tick_interval = 5\n")
        # Only 1 minute in
        monkeypatch.setattr(time, "time", lambda: now + 60)
        xbar_env.mod.main()
        xbar_env.subprocess_mock.Popen.assert_not_called()
