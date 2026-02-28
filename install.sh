#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "$KITTY_CONFIG_DIRECTORY" ]; then
    KITTY_CONFIG_DIR="$KITTY_CONFIG_DIRECTORY"
elif [ -n "$XDG_CONFIG_HOME" ]; then
    KITTY_CONFIG_DIR="$XDG_CONFIG_HOME/kitty"
else
    KITTY_CONFIG_DIR="$HOME/.config/kitty"
fi

echo "Kitty config directory: $KITTY_CONFIG_DIR"
echo "Kitty logs directory: $KITTY_CONFIG_DIR/kitty-utils/logs"

KITTY_UTILS_DIR="$KITTY_CONFIG_DIR/kitty-utils"
KITTY_LOGS_DIR="$KITTY_CONFIG_DIR/kitty-utils/logs"
mkdir -p "$KITTY_UTILS_DIR"
mkdir -p "$KITTY_LOGS_DIR"

echo "Building kitty-resurrect..."
go build -o "$KITTY_UTILS_DIR/kitty-resurrect" "$SCRIPT_DIR/cmd/kitty-resurrect"

echo "Copying watcher.py..."
cp "$SCRIPT_DIR/watcher.py" "$KITTY_UTILS_DIR/watcher.py"

SYSTEMD_DIR="$HOME/.config/systemd/user"
mkdir -p "$SYSTEMD_DIR"

echo "👾 A configurar o daemon no Systemd..."
mkdir -p ~/.config/systemd/user/

cat <<EOF > ~/.config/systemd/user/kitty-resurrect.service
[Unit]
Description=Kitty Auto Sessionizer Daemon
After=network.target

[Service]
Type=simple
Environment="PATH=$HOME/.local/bin:/usr/local/bin:/usr/bin:/bin"
Environment="HOME=$HOME"
Environment="XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
Environment="KITTY_CONFIG_DIRECTORY=$KITTY_CONFIG_DIR"
ExecStart=$KITTY_UTILS_DIR/kitty-resurrect
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable --now kitty-resurrect.service

echo "Updating kitty.conf..."
KITTY_CONF="$KITTY_CONFIG_DIR/kitty.conf"

REQUIRED_LINES=(
    "allow_remote_control yes"
    "listen_on unix:@mykitty"
    "watcher $KITTY_UTILS_DIR/watcher.py"
)

for line in "${REQUIRED_LINES[@]}"; do
    if ! grep -qF "$line" "$KITTY_CONF" 2>/dev/null; then
        echo "$line" >> "$KITTY_CONF"
    fi
done

echo ""
echo "Installation complete!"
echo ""
echo "Files installed to: $KITTY_UTILS_DIR"
echo "Systemd service: $SYSTEMD_DIR/kitty-resurrect.service"
