#!/usr/bin/env python3

# https://github.com/matryer/xbar-plugins/blob/main/CONTRIBUTING.md#metadata

# <xbar.title>Boxed</xbar.title>
# <xbar.version>v1.0</xbar.version>
# <xbar.author>cppcho</xbar.author>
# <xbar.author.github>cppcho</xbar.author.github>
# <xbar.desc>Timeboxing countdown timer</xbar.desc>
# <xbar.dependencies>python</xbar.dependencies>

import subprocess
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path.home() / ".local" / "lib" / "boxed"))
from boxed_common import (
    CONFIG_DIR,
    STATE_FILE,
    CONFIG_FILE,
    LOG_FILE,
    SOUNDS_DIR,
    atomic_write,
    format_duration,
    play_sound,
    read_config,
    read_state,
)

BOXED_PY = Path.home() / "bin" / "boxed.py"
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
                subprocess.run([PYTHON, str(BOXED_PY), "complete"], capture_output=True)
                state = read_state()
            out("📦")
        else:
            config = read_config()
            tick_interval = int(config.get("tick_interval", "0"))
            if tick_interval > 0:
                interval_secs = tick_interval * 60
                last_tick = state.get("last_tick_epoch") or started
                if now - last_tick >= interval_secs:
                    play_sound(sound_file=SOUNDS_DIR / "PeonYes3.ogg")
                    state["last_tick_epoch"] = now
                    atomic_write(STATE_FILE, state)
            out(f"{task} ({format_duration(remaining)})")
        out("---")
    else:
        out("📦")
        out("---")
    out(f"Open Config | bash=/usr/bin/open param1={CONFIG_FILE} terminal=false")
    out(f"Open Log | bash=/usr/bin/open param1={LOG_FILE} terminal=false")
    out("---")
    out(
        f"Open Config Directory | bash=/usr/bin/open param1={CONFIG_DIR} terminal=false"
    )


def out(text):
    print(text, flush=True)


if __name__ == "__main__":
    main()
