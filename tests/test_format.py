"""Tests for format_elapsed (boxed.py) and format_remaining (xbar plugin)."""


class TestFormatElapsed:
    """Tests for boxed.py format_elapsed()."""

    def test_zero_seconds(self, boxed_module):
        assert boxed_module.format_elapsed(0) == "0s"

    def test_one_second(self, boxed_module):
        assert boxed_module.format_elapsed(1) == "1s"

    def test_59_seconds(self, boxed_module):
        assert boxed_module.format_elapsed(59) == "59s"

    def test_60_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_elapsed(60) == "1m"

    def test_90_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_elapsed(90) == "1m"

    def test_119_seconds_is_1m(self, boxed_module):
        assert boxed_module.format_elapsed(119) == "1m"

    def test_120_seconds_is_2m(self, boxed_module):
        assert boxed_module.format_elapsed(120) == "2m"

    def test_3599_seconds_is_59m(self, boxed_module):
        assert boxed_module.format_elapsed(3599) == "59m"

    def test_3600_seconds_is_1h(self, boxed_module):
        assert boxed_module.format_elapsed(3600) == "1h"

    def test_3660_seconds_is_1h1m(self, boxed_module):
        assert boxed_module.format_elapsed(3660) == "1h1m"

    def test_3661_seconds_is_1h1m(self, boxed_module):
        assert boxed_module.format_elapsed(3661) == "1h1m"

    def test_7200_seconds_is_2h(self, boxed_module):
        assert boxed_module.format_elapsed(7200) == "2h"

    def test_7261_seconds_is_2h1m(self, boxed_module):
        assert boxed_module.format_elapsed(7261) == "2h1m"


class TestFormatRemaining:
    """Tests for xbar plugin format_remaining()."""

    def test_zero_seconds(self, xbar_module):
        assert xbar_module.format_remaining(0) == "0s"

    def test_59_seconds(self, xbar_module):
        assert xbar_module.format_remaining(59) == "59s"

    def test_60_seconds(self, xbar_module):
        assert xbar_module.format_remaining(60) == "1m"

    def test_3600_seconds(self, xbar_module):
        assert xbar_module.format_remaining(3600) == "1h"
