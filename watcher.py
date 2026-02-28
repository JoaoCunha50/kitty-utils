import socket
import logging
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
LOG_DIR = SCRIPT_DIR / "logs"
LOG_DIR.mkdir(exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.FileHandler(LOG_DIR / "watcher.log")],
)
logger = logging.getLogger(__name__)


def notify_go_daemon(boss):
    try:
        socket_addr = getattr(boss, 'listening_on', '')
        if not socket_addr:
            logger.warning("No socket address found on boss")
            return
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.sendto(socket_addr.encode(), ("127.0.0.1", 11223))
        sock.close()
    except Exception as e:
        logger.exception(f"Failed to notify daemon: {e}")


def on_window_created(boss, window):
    notify_go_daemon(boss)

def on_window_closed(boss, window):
    notify_go_daemon(boss)

def on_focus_change(boss, window, data):
    notify_go_daemon(boss)

def on_title_change(boss, window, data):
    notify_go_daemon(boss)
