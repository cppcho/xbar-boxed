"""Shared fixtures for boxed test suite."""

import importlib.util
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from unittest.mock import MagicMock

import pytest

COMMON_PY = (
    Path(__file__).resolve().parent.parent
    / "boxed"
    / ".local"
    / "lib"
    / "boxed"
    / "boxed_common.py"
)
BOXED_PY = Path(__file__).resolve().parent.parent / "boxed" / "bin" / "boxed.py"
XBAR_PY = (
    Path(__file__).resolve().parent.parent
    / "boxed"
    / "Library"
    / "Application Support"
    / "xbar"
    / "plugins"
    / "boxed.1s.py"
)


def _load_module(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    mod = importlib.util.module_from_spec(spec)
    # Prevent the module from executing main() on import
    # by temporarily setting __name__ to the module name
    spec.loader.exec_module(mod)
    return mod


@pytest.fixture(scope="session")
def common_module():
    mod = _load_module("boxed_common", COMMON_PY)
    sys.modules["boxed_common"] = mod
    return mod


@pytest.fixture(scope="session")
def boxed_module(common_module):
    return _load_module("boxed_cli", BOXED_PY)


@pytest.fixture(scope="session")
def xbar_module(common_module):
    return _load_module("boxed_1s", XBAR_PY)


@dataclass
class Env:
    mod: object
    config_dir: Path
    state_file: Path
    log_file: Path
    config_file: Path
    subprocess_mock: MagicMock

    @property
    def last_file(self):
        return self.config_dir / "last.json"


@pytest.fixture()
def boxed_env(boxed_module, common_module, tmp_path, monkeypatch):
    config_dir = tmp_path / ".config" / "boxed"
    config_dir.mkdir(parents=True)

    mod = boxed_module
    for m in (mod, common_module):
        monkeypatch.setattr(m, "CONFIG_DIR", config_dir)
        monkeypatch.setattr(m, "STATE_FILE", config_dir / "state.json")
        monkeypatch.setattr(m, "CONFIG_FILE", config_dir / "config")
        monkeypatch.setattr(m, "LOG_FILE", config_dir / "log")
    monkeypatch.setattr(mod, "LAST_FILE", config_dir / "last.json")
    monkeypatch.setattr(common_module, "LAST_FILE", config_dir / "last.json")

    mock_sub = MagicMock(spec=subprocess)
    mock_sub.DEVNULL = subprocess.DEVNULL
    monkeypatch.setattr(mod, "subprocess", mock_sub)
    monkeypatch.setattr(common_module, "subprocess", mock_sub)

    return Env(
        mod=mod,
        config_dir=config_dir,
        state_file=config_dir / "state.json",
        log_file=config_dir / "log",
        config_file=config_dir / "config",
        subprocess_mock=mock_sub,
    )


@pytest.fixture()
def xbar_env(xbar_module, common_module, tmp_path, monkeypatch):
    config_dir = tmp_path / ".config" / "boxed"
    config_dir.mkdir(parents=True)

    mod = xbar_module
    for m in (mod, common_module):
        monkeypatch.setattr(m, "CONFIG_DIR", config_dir)
        monkeypatch.setattr(m, "STATE_FILE", config_dir / "state.json")
        monkeypatch.setattr(m, "CONFIG_FILE", config_dir / "config")
        monkeypatch.setattr(m, "LOG_FILE", config_dir / "log")

    mock_sub = MagicMock(spec=subprocess)
    mock_sub.DEVNULL = subprocess.DEVNULL
    monkeypatch.setattr(mod, "subprocess", mock_sub)
    monkeypatch.setattr(common_module, "subprocess", mock_sub)

    return Env(
        mod=mod,
        config_dir=config_dir,
        state_file=config_dir / "state.json",
        log_file=config_dir / "log",
        config_file=config_dir / "config",
        subprocess_mock=mock_sub,
    )
