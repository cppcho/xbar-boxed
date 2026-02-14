"""Tests for state read/write/clear and atomic_write."""

import json


class TestAtomicWrite:
    def test_writes_valid_json(self, boxed_env):
        path = boxed_env.config_dir / "test.json"
        boxed_env.mod.atomic_write(path, {"key": "value"})
        assert json.loads(path.read_text()) == {"key": "value"}

    def test_overwrites_existing(self, boxed_env):
        path = boxed_env.config_dir / "test.json"
        boxed_env.mod.atomic_write(path, {"a": 1})
        boxed_env.mod.atomic_write(path, {"b": 2})
        assert json.loads(path.read_text()) == {"b": 2}

    def test_no_leftover_tmp_files(self, boxed_env):
        path = boxed_env.config_dir / "test.json"
        boxed_env.mod.atomic_write(path, {"x": 1})
        tmp_files = list(boxed_env.config_dir.glob("*.tmp"))
        assert tmp_files == []

    def test_nested_data(self, boxed_env):
        path = boxed_env.config_dir / "test.json"
        data = {"nested": {"list": [1, 2, 3]}}
        boxed_env.mod.atomic_write(path, data)
        assert json.loads(path.read_text()) == data


class TestWriteState:
    def test_writes_state_file(self, boxed_env):
        boxed_env.mod.write_state(task="test", started_epoch=1000, duration=300)
        state = json.loads(boxed_env.state_file.read_text())
        assert state["task"] == "test"
        assert state["started_epoch"] == 1000
        assert state["duration"] == 300

    def test_creates_config_dir(self, boxed_env, tmp_path):
        # Use a fresh subdir that doesn't exist yet
        new_dir = tmp_path / "new" / "config"
        boxed_env.mod.CONFIG_DIR = new_dir
        boxed_env.mod.STATE_FILE = new_dir / "state.json"
        boxed_env.mod.write_state(task="t", started_epoch=1, duration=60)
        assert (new_dir / "state.json").exists()


class TestReadState:
    def test_no_file_returns_none(self, boxed_env):
        assert boxed_env.mod.read_state() is None

    def test_reads_valid_state(self, boxed_env):
        boxed_env.mod.write_state(task="hello", started_epoch=500, duration=120)
        state = boxed_env.mod.read_state()
        assert state["task"] == "hello"
        assert state["duration"] == 120

    def test_corrupt_json_returns_none(self, boxed_env):
        boxed_env.state_file.write_text("{invalid json")
        assert boxed_env.mod.read_state() is None

    def test_empty_file_returns_none(self, boxed_env):
        boxed_env.state_file.write_text("")
        assert boxed_env.mod.read_state() is None


class TestClearState:
    def test_removes_state_file(self, boxed_env):
        boxed_env.mod.write_state(task="t", started_epoch=1, duration=60)
        assert boxed_env.state_file.exists()
        boxed_env.mod.clear_state()
        assert not boxed_env.state_file.exists()

    def test_clear_when_no_file(self, boxed_env):
        # Should not raise
        boxed_env.mod.clear_state()
        assert not boxed_env.state_file.exists()


class TestXbarReadState:
    def test_no_file_returns_none(self, xbar_env):
        assert xbar_env.mod.read_state() is None

    def test_reads_valid_state(self, xbar_env):
        data = {"task": "work", "started_epoch": 100, "duration": 60}
        xbar_env.state_file.write_text(json.dumps(data))
        state = xbar_env.mod.read_state()
        assert state["task"] == "work"

    def test_corrupt_json_returns_none(self, xbar_env):
        xbar_env.state_file.write_text("not json")
        assert xbar_env.mod.read_state() is None
