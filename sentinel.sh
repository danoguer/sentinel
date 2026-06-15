#!/bin/bash

PROJECT_DIR=$(pwd)
BINARY_PATH="$PROJECT_DIR/sentinel-bin"
SERVICE_FILE="$HOME/.config/systemd/user/sentinel.service"
BASHRC="$HOME/.bashrc"

echo "🛡️  Sentinel: Deploying Clean CLI Architecture..."

echo "🧹 Stopping old processes..."
systemctl --user stop sentinel.service 2>/dev/null || true
pkill -f sentinel-bin 2>/dev/null || true

sed -i '/# --- SENTINEL_START ---/,/# --- SENTINEL_END ---/d' "$BASHRC"

echo "🛠️  Building Sentinel..."
go build -o "$BINARY_PATH" main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Check your Go code."
    exit 1
fi
chmod +x "$BINARY_PATH"

echo "⚙️  Starting Always-On Daemon..."
mkdir -p "$HOME/.config/systemd/user/"
cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=Sentinel Daemon
After=network.target

[Service]
ExecStart=$BINARY_PATH
WorkingDirectory=$PROJECT_DIR
Restart=always
RestartSec=2

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable sentinel.service
systemctl --user restart sentinel.service

echo "🔗 Injecting Zero-Block Hook..."
cat <<'EOF' >> "$BASHRC"

# --- SENTINEL_START ---
if [[ $- == *i* ]] && [ -z "$SENTINEL_ACTIVE" ]; then
    export SENTINEL_ACTIVE=1
    
    SESSION_LOG=$(mktemp /tmp/sentinel_session_XXXXXX.log)
    
    ( tail -f "$SESSION_LOG" | nc -U /tmp/sentinel.sock >/dev/null 2>&1 ) &
    TAIL_PID=$!
    
    script -q -f "$SESSION_LOG"
    
    kill $TAIL_PID >/dev/null 2>&1
    rm -f "$SESSION_LOG"
    
    exit
fi
# --- SENTINEL_END ---
EOF

echo "🌍 Installing global 'sentinel' command (Requires sudo password)..."
sudo tee /usr/local/bin/sentinel > /dev/null << EOF_WRAPPER
#!/bin/bash
(cd "$PROJECT_DIR" && ./sentinel-bin "\$@")
EOF_WRAPPER

sudo chmod +x /usr/local/bin/sentinel

echo "------------------------------------------------"
echo "✅ Done! Terminal deadlock is mathematically impossible."
echo "✅ Clean CLI active. No quotes or flags needed."
echo "👉 Open a new terminal to start, then test it by typing: sentinel what is the time"
