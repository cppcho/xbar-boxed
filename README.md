# Boxed

A timeboxing countdown timer that lives in your macOS menu bar, built as an [xbar](https://xbarapp.com/) plugin.

## Usage

```bash
# Start a 25-minute timer
boxed start 25 Write the report

# Stop the current timer
boxed stop
```

The menu bar shows the task name and remaining time while a timer is running. When time's up, you get a macOS notification.

## Install

Requires [xbar](https://xbarapp.com/) and Python 3.

```bash
# Uses GNU Stow to symlink files into place
./install.sh
```

## Configuration

Config lives at `~/.config/boxed/config`:

```
# Disable notification sound
notify_sound = false
```
