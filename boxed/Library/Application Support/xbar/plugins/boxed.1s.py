#!/usr/bin/env python3

# https://github.com/matryer/xbar-plugins/blob/main/CONTRIBUTING.md#metadata

# <xbar.title>Boxed</xbar.title>
# <xbar.version>v1.0</xbar.version>
# <xbar.author>cppcho</xbar.author>
# <xbar.author.github>cppcho</xbar.author.github>
# <xbar.desc>Timeboxing countdown timer</xbar.desc>
# <xbar.dependencies>python</xbar.dependencies>

import json
import os
import subprocess
import sys
import tempfile
import time
from pathlib import Path

CONFIG_DIR = Path.home() / ".config" / "boxed"
STATE_FILE = CONFIG_DIR / "state.json"
CONFIG_FILE = CONFIG_DIR / "config"
LOG_FILE = CONFIG_DIR / "log"

BOXED_PY = Path(__file__).resolve().parent / "boxed.py"
PYTHON = sys.executable or "/usr/bin/python3"


def main():
    state = read_state()

    if state:
        task = state.get("task", "Untitled")
        duration = int(state.get("duration", 0))
        started = int(state.get("started_epoch", 0))

        now = int(time.time())
        remaining = started + duration - now

        if remaining <= 0:
            if not state.get("notified"):
                notify_done(task)
                mark_notified(state)
            out("📦")
        else:
            out(f"{task} ({format_remaining(remaining)})")
        out("---")
    else:
        out("📦")
        out("---")
    out(f"Open Config | bash=/usr/bin/open param1={CONFIG_FILE} terminal=false")
    out(f"Open Log | bash=/usr/bin/open param1={LOG_FILE} terminal=false")
    out("---")
    out(f"Open Config Directory | bash=/usr/bin/open param1={CONFIG_DIR} terminal=false")


def read_config():
    config = {"notify_sound": "true"}
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
    if not STATE_FILE.exists():
        return None
    try:
        with open(STATE_FILE, "r") as f:
            state = json.load(f)
    except (json.JSONDecodeError, OSError):
        return None
    return state


def format_remaining(seconds):
    """Menu bar display: '1h', '10m', '53s'."""
    if seconds >= 3600:
        h = seconds // 3600
        m = (seconds % 3600) // 60
        if m > 0:
            return f"{h}h{m}m"
        return f"{h}h"
    elif seconds >= 60:
        return f"{seconds // 60}m"
    else:
        return f"{seconds}s"


def notify_done(task):
    safe_task = task.replace("\\", "\\\\").replace('"', '\\"')
    script = f'display notification "Time\'s up!" with title "Boxed — {safe_task}"'
    config = read_config()
    if config.get("notify_sound", "true") == "true":
        script += ' sound name "default"'
    subprocess.run(["osascript", "-e", script], capture_output=True)


def mark_notified(state):
    state["notified"] = True
    dir_name = os.path.dirname(STATE_FILE) or "."
    with tempfile.NamedTemporaryFile(
        "w", dir=dir_name, delete=False, suffix=".tmp"
    ) as tmp:
        json.dump(state, tmp, indent=2)
        tmp.flush()
        os.fsync(tmp.fileno())
        tmp_path = tmp.name
    os.replace(tmp_path, STATE_FILE)


def out(text):
    print(text, flush=True)


if __name__ == "__main__":
    main()
