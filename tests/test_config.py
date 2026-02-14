"""Tests for read_config in both modules."""


class TestBoxedReadConfig:
    """Tests for boxed.py read_config()."""

    def test_missing_file_returns_defaults(self, boxed_env):
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "true"

    def test_empty_file(self, boxed_env):
        boxed_env.config_file.write_text("")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "true"

    def test_comments_ignored(self, boxed_env):
        boxed_env.config_file.write_text("# this is a comment\n")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "true"

    def test_key_value_parsed(self, boxed_env):
        boxed_env.config_file.write_text("notify_sound = false\n")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "false"

    def test_whitespace_trimmed(self, boxed_env):
        boxed_env.config_file.write_text("  notify_sound  =  false  \n")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "false"

    def test_multiple_keys(self, boxed_env):
        boxed_env.config_file.write_text("notify_sound = false\ntick_interval = 5\n")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "false"
        assert config["tick_interval"] == "5"

    def test_blank_lines_skipped(self, boxed_env):
        boxed_env.config_file.write_text("\n\nnotify_sound = false\n\n")
        config = boxed_env.mod.read_config()
        assert config["notify_sound"] == "false"


class TestXbarReadConfig:
    """Tests for xbar plugin read_config()."""

    def test_missing_file_returns_empty(self, xbar_env):
        config = xbar_env.mod.read_config()
        assert config == {}

    def test_key_value_parsed(self, xbar_env):
        xbar_env.config_file.write_text("tick_interval = 5\n")
        config = xbar_env.mod.read_config()
        assert config["tick_interval"] == "5"

    def test_comments_ignored(self, xbar_env):
        xbar_env.config_file.write_text("# comment\ntick_interval = 10\n")
        config = xbar_env.mod.read_config()
        assert "comment" not in str(config.keys())
        assert config["tick_interval"] == "10"

    def test_empty_file(self, xbar_env):
        xbar_env.config_file.write_text("")
        config = xbar_env.mod.read_config()
        assert config == {}
