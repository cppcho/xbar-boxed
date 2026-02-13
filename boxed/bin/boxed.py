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


def _atomic_write_text(filepath, text):
    """Write text to a file atomically via temp file + fsync + rename."""
    dir_name = os.path.dirname(filepath) or "."
    with tempfile.NamedTemporaryFile(
        "w", dir=dir_name, delete=False, suffix=".tmp"
    ) as tmp:
        tmp.write(text)
        tmp.flush()
        os.fsync(tmp.fileno())
        tmp_path = tmp.name
    os.replace(tmp_path, filepath)


def _read_log():
    """Read the log file contents, returning empty string if missing."""
    if LOG_FILE.exists():
        return LOG_FILE.read_text()
    return ""


def _write_log(content):
    """Write the log file atomically."""
    ensure_dirs()
    _atomic_write_text(LOG_FILE, content)


def _insert_entry_in_log(content, date_str, entry_line):
    """Insert an entry under the correct date section.

    Dates are ordered newest-first; entries within a date are chronological.
    """
    header = f"# {date_str}"
    lines = content.split("\n") if content else []

    # Find if this date section already exists
    header_idx = None
    for i, line in enumerate(lines):
        if line.strip() == header:
            header_idx = i
            break

    if header_idx is not None:
        # Find end of this date section (next header or end of file)
        insert_idx = header_idx + 1
        while insert_idx < len(lines):
            if lines[insert_idx].startswith("# ") and lines[insert_idx] != header:
                break
            insert_idx += 1
        # Back up past trailing blank lines to insert before them
        while insert_idx > header_idx + 1 and lines[insert_idx - 1].strip() == "":
            insert_idx -= 1
        lines.insert(insert_idx, entry_line)
    else:
        # Need to create a new date section — newest dates first
        # Find where to insert based on date ordering
        insert_before = None
        for i, line in enumerate(lines):
            if line.startswith("# "):
                existing_date = line[2:].strip()
                if date_str > existing_date:
                    insert_before = i
                    break
        if insert_before is not None:
            block = [header, "", entry_line, ""]
            for item in reversed(block):
                lines.insert(insert_before, item)
        else:
            # Append at end (oldest date or first entry)
            if lines and lines[-1].strip() != "":
                lines.append("")
            lines.append(header)
            lines.append("")
            lines.append(entry_line)
            lines.append("")

    # Clean up: ensure single trailing newline
    result = "\n".join(lines).rstrip("\n") + "\n"
    return result


def _migrate_log_if_needed():
    """If the log is in old format, rename to log.old and start fresh."""
    if not LOG_FILE.exists():
        return
    content = _read_log()
    if not content or content.startswith("#"):
        return
    # Old format detected — rename
    old_path = CONFIG_DIR / "log.old"
    os.replace(LOG_FILE, old_path)
    # If a timer is currently running, write its partial entry
    state = read_state()
    if state and state.get("started_epoch"):
        started = int(state["started_epoch"])
        duration = int(state.get("duration", 0))
        task = state.get("task", "Untitled")
        dt = datetime.fromtimestamp(started)
        date_str = dt.strftime("%Y-%m-%d")
        time_str = dt.strftime("%H:%M:%S")
        entry = f"{time_str} - ... {task} ({format_elapsed(duration)})"
        new_content = _insert_entry_in_log("", date_str, entry)
        _write_log(new_content)


def log_start(started_epoch, duration_secs, task):
    """Log a partial entry for a timer that just started."""
    ensure_dirs()
    _migrate_log_if_needed()
    dt = datetime.fromtimestamp(started_epoch)
    date_str = dt.strftime("%Y-%m-%d")
    time_str = dt.strftime("%H:%M:%S")
    entry = f"{time_str} - ... {task} ({format_elapsed(duration_secs)})"
    content = _read_log()
    content = _insert_entry_in_log(content, date_str, entry)
    _write_log(content)


def log_end(started_epoch, duration_secs, task, completed):
    """Update a partial log entry to its final form, or append if not found."""
    ensure_dirs()
    _migrate_log_if_needed()
    dt = datetime.fromtimestamp(started_epoch)
    date_str = dt.strftime("%Y-%m-%d")
    start_time = dt.strftime("%H:%M:%S")
    now = int(time.time())
    end_time = datetime.fromtimestamp(now).strftime("%H:%M:%S")
    symbol = "✓" if completed else "✕"
    elapsed = now - int(started_epoch)
    configured_dur = format_elapsed(duration_secs)
    elapsed_dur = format_elapsed(elapsed)

    partial = f"{start_time} - ... {task} ({configured_dur})"
    final = f"{start_time} - {end_time} {task} ({elapsed_dur}) {symbol}"

    content = _read_log()
    if partial in content:
        content = content.replace(partial, final, 1)
        _write_log(content)
    else:
        # Fallback: insert a complete entry
        content = _insert_entry_in_log(content, date_str, final)
        _write_log(content)


def play_sound(sound_name=None, sound_file=None):
    """Play a macOS system sound by name or a custom sound file."""
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


def notify(title, message):
    # Escape backslashes and double quotes for AppleScript string literals
    safe_title = title.replace("\\", "\\\\").replace('"', '\\"')
    safe_message = message.replace("\\", "\\\\").replace('"', '\\"')
    script = f'display notification "{safe_message}" with title "{safe_title}"'
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
        # If old timer hadn't expired, mark as stopped; otherwise mark complete
        if old_elapsed < old_duration:
            log_end(old_started, old_duration, old_task, completed=False)
        else:
            log_end(old_started, old_duration, old_task, completed=True)

    atomic_write(LAST_FILE, {"duration": duration_mins, "task": task})

    duration_secs = duration_mins * 60
    write_state(
        task=task,
        started_epoch=now,
        duration=duration_secs,
    )

    config = read_config()
    log_start(now, duration_secs, task)
    notify("Boxed", f"Timer started: {duration_mins}m — {task}")
    if config["notify_sound"] == "true":
        play_sound(sound_file=CONFIG_DIR / "sounds" / "PeonReady1.ogg")
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
    config = read_config()

    log_end(started, duration, task, completed=True)
    notify("Boxed", f"Time's up! — {task}")
    if config["notify_sound"] == "true":
        play_sound(sound_name="Glass")

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
    # Timer already expired — finalize the partial log entry and clean up
    if elapsed >= duration:
        log_end(started, duration, task, completed=True)
        clear_state()
        print(f"Cleared ended timer: {task}")
        return

    config = read_config()
    clear_state()
    log_end(started, duration, task, completed=False)
    elapsed_str = format_elapsed(elapsed)
    notify("Boxed", f"Timer stopped: {task} ({elapsed_str} elapsed)")
    if config["notify_sound"] == "true":
        play_sound(sound_name="Sosumi")
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
