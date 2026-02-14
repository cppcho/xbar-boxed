"""Tests for format_duration (shared via boxed_common)."""


class TestFormatDurationBoxed:
    """Tests for format_duration() via boxed.py."""

    def test_zero_seconds(self, boxed_module):
        assert boxed_module.format_duration(0) == "0s"

    def test_one_second(self, boxed_module):
        assert boxed_module.format_duration(1) == "1s"

    def test_59_seconds(self, boxed_module):
        assert boxed_module.format_duration(59) == "59s"

    def test_60_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_duration(60) == "1m"

    def test_90_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_duration(90) == "1m"

    def test_119_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_duration(119) == "1m"

    def test_120_seconds_is_2m(self, boxed_module):
        assert boxed_module.format_duration(120) == "2m"

    def test_3599_seconds_is_59m(self, boxed_module):
        assert boxed_module.format_duration(3599) == "59m"

    def test_3600_seconds_is_1h(self, boxed_module):
        assert boxed_module.format_duration(3600) == "1h"

    def test_3660_seconds_is_1h1m(self, boxed_module):
        assert boxed_module.format_duration(3660) == "1h1m"

    def test_3661_seconds_is_1h1m(self, boxed_module):
        assert boxed_module.format_duration(3661) == "1h1m"

    def test_7200_seconds_is_2h(self, boxed_module):
        assert boxed_module.format_duration(7200) == "2h"

    def test_7261_seconds_is_2h1m(self, boxed_module):
        assert boxed_module.format_duration(7261) == "2h1m"


class TestFormatDurationXbar:
    """Tests for format_duration() via xbar plugin."""

    def test_zero_seconds(self, xbar_module):
        assert xbar_module.format_duration(0) == "0s"

    def test_59_seconds(self, xbar_module):
        assert xbar_module.format_duration(59) == "59s"

    def test_60_seconds(self, xbar_module):
        assert xbar_module.format_duration(60) == "1m"

    def test_3600_seconds(self, xbar_module):
        assert xbar_module.format_duration(3600) == "1h"
