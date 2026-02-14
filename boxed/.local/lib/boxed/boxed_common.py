"""Shared constants and utilities for Boxed timer scripts."""

import json
import os
import subprocess
import tempfile
from pathlib import Path

# Config paths
CONFIG_DIR = Path.home() / ".config" / "boxed"
STATE_FILE = CONFIG_DIR / "state.json"
CONFIG_FILE = CONFIG_DIR / "config"
LOG_FILE = CONFIG_DIR / "log"
LAST_FILE = CONFIG_DIR / "last.json"
SOUNDS_DIR = Path(__file__).resolve().parent / "sounds"


def read_config(defaults=None):
    """Read key=value config file, merged over optional defaults."""
    config = dict(defaults) if defaults else {}
    if CONFIG_FILE.exists():
        for line in CONFIG_FILE.read_text().splitlines():
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, value = line.split("=", 1)
                config[key.strip()] = value.strip()
    return config


def read_state():
    """Read timer state JSON. Returns dict or None."""
    if not STATE_FILE.exists():
        return None
    try:
        with open(STATE_FILE, "r") as f:
            state = json.load(f)
    except (json.JSONDecodeError, OSError):
        return None
    return state


def atomic_write(filepath, data):
    """Write JSON atomically via temp file + fsync + rename."""
    dir_name = os.path.dirname(filepath) or "."
    with tempfile.NamedTemporaryFile(
        "w", dir=dir_name, delete=False, suffix=".tmp"
    ) as tmp:
        json.dump(data, tmp, indent=2)
        tmp.flush()
        os.fsync(tmp.fileno())
        tmp_path = tmp.name
    os.replace(tmp_path, filepath)


def format_duration(seconds):
    """Format seconds as '1h30m', '25m', or '45s'."""
    if seconds >= 3600:
        h = seconds // 3600
        m = (seconds % 3600) // 60
        return f"{h}h{m}m" if m > 0 else f"{h}h"
    elif seconds >= 60:
        return f"{seconds // 60}m"
    else:
        return f"{seconds}s"


def play_sound(sound_name=None, sound_file=None):
    """Play a macOS system sound by name or custom file."""
    if sound_name:
        path = f"/System/Library/Sounds/{sound_name}.aiff"
    elif sound_file:
        path = str(sound_file)
    else:
        return
    subprocess.Popen(
        ["afplay", path],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
