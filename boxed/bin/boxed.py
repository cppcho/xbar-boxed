#!/usr/bin/env python3
"""Boxed - A timeboxing timer for your menu bar.

Usage:
    boxed start <duration> <task name...>
    boxed stop
    boxed again
"""

import os
import subprocess
import sys
import time
import json
import tempfile
from datetime import datetime
from pathlib import Path

CONFIG_DIR = Path.home() / ".config" / "boxed"
STATE_FILE = CONFIG_DIR / "state.json"
LOG_FILE = CONFIG_DIR / "log"
LAST_FILE = CONFIG_DIR / "last.json"
CONFIG_FILE = CONFIG_DIR / "config"

DEFAULT_CONFIG = """\
# Boxed configuration
# notify_sound = true
# tick_interval = 5
"""


def ensure_dirs():
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)


def ensure_config():
    ensure_dirs()
    if not CONFIG_FILE.exists():
        CONFIG_FILE.write_text(DEFAULT_CONFIG)


def read_config():
    config = {
        "notify_sound": "true",
    }
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
    """Read current timer state. Returns dict or None if no timer running."""
    if not STATE_FILE.exists():
        return None
    state = {}
    try:
        with open(STATE_FILE, "r") as f:
            state = json.load(f)
    except (json.JSONDecodeError, OSError) as e:
        print(f"Error reading state: {e}", file=sys.stderr)
        return None
    return state


def atomic_write(filepath, data):
    # Write to temp file in the same directory
    dir_name = os.path.dirname(filepath) or "."
    with tempfile.NamedTemporaryFile(
        "w", dir=dir_name, delete=False, suffix=".tmp"
    ) as tmp:
        json.dump(data, tmp, indent=2)
        tmp.flush()
        os.fsync(tmp.fileno())  # ensure data hits disk
        tmp_path = tmp.name
    os.replace(tmp_path, filepath)  # atomic rename (cross-platform)


def write_state(*, task=None, started_epoch=None, duration=None):
    """Write timer state atomically."""
    ensure_dirs()
    state = {
        "task": task,
        "started_epoch": started_epoch,
        "duration": duration,
    }
    atomic_write(STATE_FILE, state)


def clear_state():
    ensure_dirs()
    STATE_FILE.unlink(missing_ok=True)


def log_event(event, duration_str=None, task=None, extra=None):
    ensure_dirs()
    ts = datetime.now().strftime("%Y-%m-%dT%H:%M:%S")
    parts = [ts, event]
    if duration_str:
        parts.append(duration_str)
    if task:
        parts.append(f'"{task}"')
    if extra:
        parts.append(extra)
    with open(LOG_FILE, "a") as f:
        f.write(" ".join(parts) + "\n")


def notify(title, message, sound=True):
    # Escape backslashes and double quotes for AppleScript string literals
    safe_title = title.replace("\\", "\\\\").replace('"', '\\"')
    safe_message = message.replace("\\", "\\\\").replace('"', '\\"')
    script = f'display notification "{safe_message}" with title "{safe_title}"'
    if sound:
        script += ' sound name "default"'
    subprocess.run(
        ["osascript", "-e", script],
        capture_output=True,
    )


def format_elapsed(seconds):
    if seconds >= 3600:
        h = seconds // 3600
        m = (seconds % 3600) // 60
        return f"{h}h{m}m" if m > 0 else f"{h}h"
    elif seconds >= 60:
        return f"{seconds // 60}m"
    else:
        return f"{seconds}s"


def cmd_start(args):
    if len(args) < 2:
        print(
            "Usage: boxed start <duration in minutes> <task name...>", file=sys.stderr
        )
        sys.exit(1)

    duration_mins = args[0]
    task = " ".join(args[1:])

    try:
        duration_mins = int(duration_mins)
        if duration_mins <= 0:
            raise ValueError("must be positive")
    except (ValueError, IndexError):
        print(f"Invalid duration: {args[0]}", file=sys.stderr)
        sys.exit(1)

    # If timer already running, stop it first
    state = read_state()
    now = int(time.time())
    if state:
        old_task = state.get("task", "Untitled")
        old_started = int(state.get("started_epoch", 0))
        old_duration = int(state.get("duration", 0))
        old_elapsed = now - old_started
        # Only log STOP if the old timer hadn't already expired
        if old_elapsed < old_duration:
            old_duration_mins = old_duration // 60
            log_event(
                "STOP",
                str(old_duration_mins),
                old_task,
            )

    atomic_write(LAST_FILE, {"duration": duration_mins, "task": task})

    duration_secs = duration_mins * 60
    write_state(
        task=task,
        started_epoch=now,
        duration=duration_secs,
    )

    config = read_config()
    log_event("START", str(duration_mins), task)
    notify("Boxed", f"Timer started: {duration_mins}m — {task}", sound=(config["notify_sound"] == "true"))
    print(f"Timer started: {duration_mins}m — {task}")


def cmd_complete(args):
    """Called by the xbar plugin when a timer expires."""
    state = read_state()
    if not state:
        return
    if state.get("notified"):
        return

    now = int(time.time())
    started = int(state.get("started_epoch", 0))
    duration = int(state.get("duration", 0))
    elapsed = now - started

    if elapsed < duration:
        return

    task = state.get("task", "Untitled")
    duration_mins = duration // 60
    config = read_config()

    log_event("COMPLETE", str(duration_mins), task)
    notify("Boxed", f"Time's up! — {task}", sound=(config["notify_sound"] == "true"))

    state["notified"] = True
    atomic_write(STATE_FILE, state)


def cmd_again(args):
    """Repeat the last started timer."""
    if not LAST_FILE.exists():
        print("No previous timer to repeat.", file=sys.stderr)
        sys.exit(1)
    try:
        with open(LAST_FILE, "r") as f:
            last = json.load(f)
    except (json.JSONDecodeError, OSError) as e:
        print(f"Error reading last timer: {e}", file=sys.stderr)
        sys.exit(1)
    cmd_start([str(last["duration"]), last["task"]])


def cmd_stop(args):
    state = read_state()
    if not state:
        print("No timer running.", file=sys.stderr)
        sys.exit(1)

    now = int(time.time())
    task = state.get("task", "Untitled")
    started = int(state.get("started_epoch", 0))
    duration = int(state.get("duration", 0))
    elapsed = now - started
    duration_mins = duration // 60

    # Timer already expired — just clean up, no log/notification
    if elapsed >= duration:
        clear_state()
        print(f"Cleared ended timer: {task}")
        return

    config = read_config()
    clear_state()
    log_event(
        "STOP",
        str(duration_mins),
        task,
    )
    elapsed_str = format_elapsed(elapsed)
    notify("Boxed", f"Timer stopped: {task} ({elapsed_str} elapsed)", sound=(config["notify_sound"] == "true"))
    print(f"Timer stopped: {task} ({elapsed_str} elapsed)")


def main():
    ensure_config()

    if len(sys.argv) < 2:
        print(__doc__.strip(), file=sys.stderr)
        sys.exit(1)

    cmd = sys.argv[1]
    if cmd == "start":
        cmd_start(sys.argv[2:])
    elif cmd == "stop":
        cmd_stop(sys.argv[2:])
    elif cmd == "again":
        cmd_again(sys.argv[2:])
    elif cmd == "complete":
        cmd_complete(sys.argv[2:])
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
