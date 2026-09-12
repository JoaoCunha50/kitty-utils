# kitty-utils

Auto-saves the current Kitty session when windows/tabs change.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Kitty Terminal                                                 │
│  ┌──────────────────────┐                                       │
│  │ watcher.py           │                                       │
│  │ - on_load            │                                       │
│  │ - on_close           │                                       │
│  │ - on_focus_change    │                                       │
│  │ - on_title_change    │                                       │
│  └──────────────────────┘                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                     (UDP): localhost:11223
                              │
┌─────────────────────────────────────────────────────────────────┐
│  kitty-resurrect (Go Daemon)                                    │
│  - Listens to UDP on port 11223                                 │
│  - 1 Second Debounce                                            │
│  - Asks kitty to serialize itself:                              │
│    kitty @ action save_as_session --save-only <file>            │
└─────────────────────────────────────────────────────────────────┘
```

## Platforms

- Linux
- macOS

## Dependencies

- **Go** 1.25.0 or compatible
- **Kitty** 0.48 or newer, with remote control enabled. The daemon relies on
  kitty's `save_as_session` action; older versions may not have it, and the
  daemon will log `Unknown action: save_as_session` if that is the case.
- **Linux**: Systemd
- **macOS**: launchd

## Installation

### Option 1: Script Installation

```bash
./install.sh
```

This script:
1. Compiles the `kitty-resurrect` binary to `~/.config/kitty/kitty-utils/`
2. Copies the `watcher.py` to `~/.config/kitty/kitty-utils/`
3. Creates the Systemd service
4. Adds the necessary lines to the `kitty.conf`:
   - `allow_remote_control yes`
   - `listen_on unix:@mykitty`
   - `single_instance yes`
   - `watcher ~/.config/kitty/kitty-utils/watcher.py`

`single_instance yes` is required, not cosmetic. `startup_session` is applied by
every kitty instance, so without it each terminal you open restores the whole
saved session on top of itself and the session file grows every time. With it,
all terminals share one process: only the first one restores, the rest open
clean, and the daemon sees a single socket. Use `kitty --instance-group <name>`
if you deliberately want an isolated kitty.

### Option 2: Manual

1. **Compile the daemon:**
   ```bash
   go build -o ~/.config/kitty/kitty-utils/kitty-resurrect ./cmd/kitty-resurrect
   ```

2. **Copy the watcher:**
   ```bash
   cp watcher.py ~/.config/kitty/kitty-utils/watcher.py
   ```

3. **Configure the kitty.conf:**
   ```kitty
   allow_remote_control yes
   listen_on unix:@mykitty
   single_instance yes
   watcher ~/.config/kitty/kitty-utils/watcher.py
   ```

   > **Note:** On Linux, use `listen_on unix:@mykitty` (abstract socket). On macOS, use `listen_on unix:/tmp/mykitty` (file-based socket).

4. **Start the daemon:**
   ```bash
   systemctl --user start kitty-resurrect
   ```

## Usage

### Start the daemon manually

```bash
~/.config/kitty/kitty-utils/kitty-resurrect
```

### Check the service status

```bash
systemctl --user status kitty-resurrect
```

### View logs

```bash
# Daemon logs
journalctl --user -u kitty-resurrect -f

# Watcher logs
tail -f ~/.config/kitty/kitty-utils/logs/watcher.log

# Resurrecter logs
tail -f ~/.config/kitty/kitty-utils/logs/kitty-resurrecter.log
```

### Restore a session

(RECOMMENDED) Add to the `kitty.conf` for it to be automatic:
```kitty
startup_session ~/.config/kitty/kitty-session.conf
```

To manually restore the saved session:
```bash
kitty @ load-session ~/.config/kitty/kitty-session.conf
```

## Generaed Files

- **Loaded Session**: `~/.config/kitty/kitty-session.conf`
- **Binary**: `~/.config/kitty/kitty-utils/kitty-resurrect`
- **Watcher**: `~/.config/kitty/kitty-utils/watcher.py`
- **Logs**: `~/.config/kitty/kitty-utils/logs/`

## Watcher Callbacks

These are kitty's real watcher hook names. Note that kitty has no
window-created hook — a new window taking focus fires `on_focus_change`, which
covers it.

| Callback | Descrição |
|----------|-----------|
| `on_load` | Once per kitty process, when the watcher is loaded. Registers the socket with the daemon. Takes `(boss, data)`, not `(boss, window, data)` |
| `on_close` | When a window is closed |
| `on_focus_change` | When the focus changes between windows |
| `on_title_change` | When the kitty window title is changed |
