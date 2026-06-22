# timeon

A lightweight macOS time tracker written in Go. It polls the active application every 5 seconds, handles idle/sleep/lock edge cases, and writes structured daily Markdown reports.

## Features

- **Active app tracking** — polls the frontmost application every 1 second via native Cocoa APIs (no heavy dependencies)
- **Idle detection** — stops counting after 5 minutes without keyboard or mouse input
- **Sleep & lock handling** — detects system sleep (wall-clock gaps) and screen lock; neither counts as active time
- **Midnight rollover** — seamlessly starts a new daily report when the date changes
- **Markdown reports** — updated every 10 minutes in `~/git/timeon/reports/YYYY-MM-DD.md`

## Requirements

- macOS (Apple Silicon or Intel)
- Go 1.22+
- Xcode Command Line Tools (`xcode-select --install`)

## Build

```bash
cd ~/git/timeon
make build
```

Binaries are written to `bin/timeon` and `bin/ontime`.

## Install & Start on Login

This installs binaries to `/usr/local/bin`, creates the reports directory, and registers a `launchd` agent:

```bash
make launchd
```

To remove:

```bash
make uninstall
```

## Manual Usage

Run the tracker in the foreground (useful for debugging):

```bash
./bin/timeon daemon
```

Print today's report:

```bash
./bin/timeon report
# or
./bin/ontime
```

Print a report for a day in the past:

```bash
./bin/timeon day 3
```

Print a weekly summary:

```bash
./bin/timeon week
./bin/timeon week 1  # previous calendar week
```

## CLI Alias (`ontime`)

After `make install`, the `ontime` binary is on your PATH. To add a shell alias instead:

```bash
echo 'alias ontime="cat ~/git/timeon/reports/$(date +%Y-%m-%d).md"' >> ~/.zshrc
source ~/.zshrc
```

## Report Format

Each daily file (`~/git/timeon/reports/YYYY-MM-DD.md`) contains:

1. **Total Active Time** for the day
2. **Active Time by Application** — sorted most to least used
3. **Timeline** — active minutes per 10-minute block from midnight through the current time

## Configuration

| Variable | Default | Description |
|---|---|---|
| `TIMEON_REPORTS_DIR` | `~/git/timeon/reports` | Where state and reports are stored |

## launchd Details

The plist is installed to `~/Library/LaunchAgents/com.duhd.timeon.plist`. Logs:

- `~/git/timeon/reports/timeon.log`
- `~/git/timeon/reports/timeon.err`

View logs:

```bash
make logs
```

Check agent status:

```bash
launchctl print gui/$(id -u)/com.duhd.timeon
```

## How Edge Cases Are Handled

| Scenario | Behavior |
|---|---|
| **Idle (5+ min)** | Uses `CGEventSourceSecondsSinceLastEventType`; no time accumulated |
| **Screen locked** | Queries `CGSSessionScreenIsLocked` via Core Graphics session dictionary |
| **System sleep** | Detects wall-clock gaps > 5s between polls |
| **Midnight** | Finalizes previous day's report, resets state for the new date |
| **Shutdown** | Writes a final report on SIGTERM/SIGINT (launchd unload, logout) |

## Project Layout

```
timeon/
├── cmd/timeon/          # Main CLI + daemon
├── cmd/ontime/          # Quick report printer
├── internal/
│   ├── config/          # Paths and intervals
│   ├── macos/           # CGO bindings (Cocoa / CoreGraphics)
│   ├── report/          # Markdown generation
│   ├── state/           # JSON persistence
│   └── tracker/         # Core polling loop
├── launchd/             # launchd plist template
├── reports/             # Daily .md reports and state
├── Makefile
└── README.md
```

## Permissions

The tracker uses **`lsappinfo`** (Launch Services) as the primary way to detect the frontmost application. This works from a background `launchd` agent without extra permissions.

For best results, also grant **Accessibility** to `timeon` (enables an additional AX-based check):

1. Run `timeon daemon` once in a terminal, or restart the agent — macOS may prompt you
2. Or open **System Settings → Privacy & Security → Accessibility** and add `/usr/local/bin/timeon`
3. Restart: `launchctl kickstart -k gui/$(id -u)/com.duhd.timeon`

Verify detection while switching between apps (e.g. Slack, Cursor, Safari):

```bash
timeon diagnose
```

All methods should show the app you are actually using. The `selected` line is what the tracker records.
