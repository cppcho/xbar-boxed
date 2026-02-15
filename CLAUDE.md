# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Boxed is a macOS menu bar timeboxing timer built as an [xbar](https://xbarapp.com/) plugin. It consists of two Go binaries with no external dependencies (stdlib only).

## Architecture

- **`cmd/boxed/main.go`** — CLI entrypoint (`boxed start <minutes> <task>`, `boxed stop`, `boxed again`, `boxed complete`). Dispatches to `App` struct in `internal/boxed/commands.go`.
- **`cmd/boxed-xbar/main.go`** — xbar plugin entrypoint. Creates `XbarApp` and calls `Run()`. The installed binary is named `boxed.1s.cgo` so xbar runs it every 1 second.
- **`internal/boxed/`** — All shared logic: paths, config, state, logging, notifications, sound, commands, and xbar rendering.

Both binaries share the same config directory (`~/.config/boxed/`) with:
- `state.json` — current timer state (task, started_epoch, duration, notified)
- `config` — simple `key = value` config file (e.g., `notify_sound = true`)
- `log` — human-readable event log, grouped by date with time ranges and outcome symbols (✓ completed, ✕ stopped)

State writes use atomic file operations (temp file + fsync + os.Rename) to avoid corruption.

Testability is achieved via dependency injection: `Paths` struct (file paths), `CommandRunner` interface (subprocess calls), and `NowFunc` (clock) are injected into `App` and `XbarApp` structs.

## Installation

```
make install    # builds binaries, copies to ~/bin/ and xbar plugins dir
```

## Testing

```
make test
```

## Linting

```
make lint       # runs go vet ./...
```
