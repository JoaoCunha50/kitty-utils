# kitty-utils

Auto-saves the current Kitty session when windows/tabs change.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│  Kitty Terminal                                                 │
│  ┌──────────────────────┐                                       │
│  │ watcher.py           │                                       │
│  │ - on_window_created  │                                       │
│  │ - on_window_closed   │                                       │
│  │ - on_focus_change    │                                       │
│  │ - on_set_tab_title   │                                       │
│  └──────────────────────┘                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                     (UDP): localhost:11223
                              │
┌─────────────────────────────────────────────────────────────────┐
│  kitty-resurrect (Go Daemon)                                    │
│  - Listens to UDP on port 11223                                 │
│  - 1 Second Debounce                                            │
│  - Salva sessão para ~/.config/kitty/kitty-session.conf         │
└─────────────────────────────────────────────────────────────────┘
```

## Dependencies

- **Go** 1.25.0 or compatible
- **Kitty** terminal with remote control enabled
- **Systemd**

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
   - `watcher ~/.config/kitty/kitty-utils/watcher.py`

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
   watcher ~/.config/kitty/kitty-utils/watcher.py
   ```

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
# Logs do daemon
journalctl --user -u kitty-resurrect -f

# Logs do watcher (em)
tail -f ~/.config/kitty/kitty-utils/logs/watcher.log

# Logs do resurrecter
tail -f ~/.config/kitty/kitty-utils/logs/kitty-resurrecter.log
```

### Restore a session

(RECOMMENDED) Add to the `kitty.conf`:
```kitty
startup_session ~/.config/kitty/kitty-session.conf
```

To restore the saved session:
```bash
kitty @ load-session ~/.config/kitty/kitty-session.conf
```

## Ficheiros Gerados

- **Sessão guardada**: `~/.config/kitty/kitty-session.conf`
- **Binário**: `~/.config/kitty/kitty-utils/kitty-resurrect`
- **Watcher**: `~/.config/kitty/kitty-utils/watcher.py`
- **Logs**: `~/.config/kitty/kitty-utils/logs/`

## Callbacks do Watcher

O `watcher.py` atualmente suporta os seguintes callbacks:

| Callback | Descrição |
|----------|-----------|
| `on_window_created` | Quando uma janela é criada |
| `on_window_closed` | Quando uma janela é fechada |
| `on_focus_change` | Quando o foco muda |
| `on_set_tab_title` | Quando o título de um tab é alterado |
