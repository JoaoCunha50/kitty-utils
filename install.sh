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

# Detect kitty binary location
KITTY_BIN=""
if command -v kitty &>/dev/null; then
    KITTY_BIN="$(command -v kitty)"
elif [[ -x "/Applications/kitty.app/Contents/MacOS/kitty" ]]; then
    KITTY_BIN="/Applications/kitty.app/Contents/MacOS/kitty"
fi

if [ -z "$KITTY_BIN" ]; then
    echo "ERROR: kitty binary not found. Install kitty first or ensure it is in your PATH."
    exit 1
fi
echo "Kitty binary: $KITTY_BIN"

if [[ "$(uname -s)" == "Darwin" ]]; then
    LISTEN_ON="listen_on unix:/tmp/mykitty"
    IS_MACOS=true
else
    LISTEN_ON="listen_on unix:@mykitty"
    IS_MACOS=false
fi

echo "Building kitty-resurrect..."
go build -o "$KITTY_UTILS_DIR/kitty-resurrect" "$SCRIPT_DIR/cmd/kitty-resurrect"

echo "Copying watcher.py..."
cp "$SCRIPT_DIR/watcher.py" "$KITTY_UTILS_DIR/watcher.py"

if [[ "$(uname -s)" == "Darwin" ]]; then
    LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
    mkdir -p "$LAUNCH_AGENTS_DIR"

    echo "👾 A configurar o daemon no launchd..."
    cat <<EOF > "$LAUNCH_AGENTS_DIR/kitty-resurrect.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>kitty-resurrect</string>
    <key>ProgramArguments</key>
    <array>
        <string>$KITTY_UTILS_DIR/kitty-resurrect</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:/Applications/kitty.app/Contents/MacOS:/usr/bin:/bin</string>
        <key>KITTY_PATH</key>
        <string>$KITTY_BIN</string>
        <key>HOME</key>
        <string>$HOME</string>
        <key>XDG_CONFIG_HOME</key>
        <string>$XDG_CONFIG_HOME</string>
        <key>KITTY_CONFIG_DIRECTORY</key>
        <string>$KITTY_CONFIG_DIR</string>
    </dict>
</dict>
</plist>
EOF

    launchctl load -w "$LAUNCH_AGENTS_DIR/kitty-resurrect.plist"
    launchctl start kitty-resurrect

    SERVICE_INFO="LaunchAgent: $LAUNCH_AGENTS_DIR/kitty-resurrect.plist"
else
    SYSTEMD_DIR="$HOME/.config/systemd/user"
    mkdir -p "$SYSTEMD_DIR"

    echo "👾 A configurar o daemon no Systemd..."
    mkdir -p "$SYSTEMD_DIR"

    cat <<EOF > "$SYSTEMD_DIR/kitty-resurrect.service"
[Unit]
Description=Kitty Auto Sessionizer Daemon
After=network.target

[Service]
Type=simple
Environment="PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"
Environment="HOME=$HOME"
Environment="XDG_CONFIG_HOME=$XDG_CONFIG_HOME"
Environment="KITTY_CONFIG_DIRECTORY=$KITTY_CONFIG_DIR"
Environment="KITTY_PATH=$KITTY_BIN"
ExecStart=$KITTY_UTILS_DIR/kitty-resurrect
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
EOF

    systemctl --user daemon-reload
    systemctl --user enable --now kitty-resurrect.service

    SERVICE_INFO="Systemd service: $SYSTEMD_DIR/kitty-resurrect.service"
fi

echo "Updating kitty.conf..."
KITTY_CONF="$KITTY_CONFIG_DIR/kitty.conf"

REQUIRED_LINES=(
    "allow_remote_control yes"
    "$LISTEN_ON"
    "watcher $KITTY_UTILS_DIR/watcher.py"
    "startup_session $KITTY_CONFIG_DIR/kitty-session.conf"
)

for line in "${REQUIRED_LINES[@]}"; do
    if ! grep -qF "$line" "$KITTY_CONF" 2>/dev/null; then
        echo "$line" >> "$KITTY_CONF"
    fi
done

if grep -qF "listen_on" "$KITTY_CONF" 2>/dev/null && ! grep -qF "$LISTEN_ON" "$KITTY_CONF" 2>/dev/null; then
    if $IS_MACOS && grep -qE '^listen_on\s+unix:@' "$KITTY_CONF" 2>/dev/null; then
        echo "Migrating 'listen_on' from Linux abstract socket to macOS file-based socket..."
        sed -i.bak 's|^listen_on\s\+unix:@.*|'"$LISTEN_ON"'|' "$KITTY_CONF"
        echo "  Replaced with: $LISTEN_ON"
    else
        echo "WARNING: Found 'listen_on' in kitty.conf but not set to '$LISTEN_ON'."
        echo "         The watcher needs this specific socket. Consider updating your kitty.conf."
    fi
fi

echo ""
echo "Installation complete!"
echo ""
echo "Files installed to: $KITTY_UTILS_DIR"
echo "$SERVICE_INFO"
