### Kitty-Resurrecter

Auto-saves the current Kitty session when windows/tabs change.

Setup
- Add a listener socket in `kitty.conf`:
  - `listen_on unix:@mykitty`
- Enable the watcher in `kitty.conf`:
  - `watcher ~/.config/kitty/watcher.py`
- Create `~/.config/kitty/watcher.py` to ping the daemon:

```python
import socket
from typing import Any, Dict

def notify_daemon() -> None:
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.sendto(b"update", ("127.0.0.1", 11223))
    sock.close()

def on_window_created(boss: Any, window: Any, data: Dict[str, Any]) -> None:
    notify_daemon()

def on_window_closed(boss: Any, window: Any, data: Dict[str, Any]) -> None:
    notify_daemon()

def on_focus_change(boss: Any, window: Any, data: Dict[str, Any]) -> None:
    notify_daemon()

def on_set_tab_title(boss: Any, tab: Any, data: Dict[str, Any]) -> None:
    notify_daemon()
```

Run
- `KITTY_LISTEN_ON=unix:'socket_name' (or don't set the variable if the socket is @mykitty, and just run) go run ./cmd/kitty-resurrect`
