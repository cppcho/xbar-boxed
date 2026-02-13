# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Boxed is a macOS menu bar timeboxing timer built as an [xbar](https://xbarapp.com/) plugin. It consists of two Python scripts with no external dependencies.

## Architecture

- **`boxed/bin/boxed.py`** — CLI for starting/stopping timers (`boxed start <minutes> <task>`, `boxed stop`). Writes state, logs events, sends macOS notifications via osascript.
- **`boxed/Library/Application Support/xbar/plugins/boxed.1s.py`** — xbar plugin that reads state and renders the menu bar countdown. The `1s` in the filename means xbar runs it every 1 second. Sends a "time's up" notification when a timer expires.

Both scripts share the same config directory (`~/.config/boxed/`) with:
- `state.json` — current timer state (task, started_epoch, duration, notified)
- `config` — simple `key = value` config file (e.g., `notify_sound = true`)
- `log` — human-readable event log, grouped by date with time ranges and outcome symbols (✓ completed, ✕ stopped)

State writes use atomic file operations (temp file + fsync + os.replace) to avoid corruption.

## Installation

Uses GNU Stow to symlink the `boxed/` directory tree into `~`:
```
./install.sh    # runs: stow -v -R -t ~ "boxed"
```

## Linting

```
uvx ruff check .
```
