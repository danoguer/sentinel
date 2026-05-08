#!/bin/bash

PROJECT_DIR=$(pwd)
BINARY_PATH="$PROJECT_DIR/sentinel-bin"
SERVICE_FILE="$HOME/.config/systemd/user/sentinel.service"
BASHRC="$HOME/.bashrc"

echo "🛡️  Sentinel: Deploying Clean CLI Architecture..."

# 0. SRE SAFEGUARD: Stop everything before compiling!
# We must stop the systemd service and kill any ghost processes, 
# otherwise Go cannot overwrite the binary (Text File Busy error).
echo "🧹 Stopping old processes..."
systemctl --user stop sentinel.service 2>/dev/null || true
pkill -f sentinel-bin 2>/dev/null || true

# 1. EMERGENCY CLEANUP: Wipe out the old FIFO hooks
sed -i '/# --- SENTINEL_START ---/,/# --- SENTINEL_END ---/d' "$BASHRC"

# 2. BUILD THE BINARY
echo "🛠️  Building Sentinel..."
go build -o "$BINARY_PATH" main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Check your Go code."
    exit 1
fi
chmod +x "$BINARY_PATH"

# 3. SETUP SYSTEMD SERVICE
# Notice: No more '-daemon' flag. If there are no arguments, 
# your Go code now knows it should be the background listener.
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

# 4. THE ZERO-BLOCK HOOK
echo "🔗 Injecting Zero-Block Hook..."
cat <<'EOF' >> "$BASHRC"

# --- SENTINEL_START ---
if [[ $- == *i* ]] && [ -z "$SENTINEL_ACTIVE" ]; then
    export SENTINEL_ACTIVE=1
    
    # 1. Create a NORMAL text file, NOT a FIFO pipe. This guarantees no freezing.
    SESSION_LOG=$(mktemp /tmp/sentinel_session_XXXXXX.log)
    
    # 2. Background Courier: Tail the file and send to the socket.
    # If the daemon crashes, this simply dies silently. The file keeps working.
    ( tail -f "$SESSION_LOG" | nc -U /tmp/sentinel.sock >/dev/null 2>&1 ) &
    TAIL_PID=$!
    
    # 3. Start terminal monitoring. It writes to the file. Will NEVER block.
    script -q -f "$SESSION_LOG"
    
    # 4. Cleanup when the user types 'exit' to close the terminal
    kill $TAIL_PID >/dev/null 2>&1
    rm -f "$SESSION_LOG"
    
    # 5. Prevent double-exiting issues
    exit
fi
# --- SENTINEL_END ---
EOF

# 5. GLOBAL COMMAND WRAPPER
echo "🌍 Installing global 'sentinel' command (Requires sudo password)..."
# We use sudo here because /usr/local/bin is a system directory
sudo tee /usr/local/bin/sentinel > /dev/null << EOF_WRAPPER
#!/bin/bash
# Run the binary in a subshell so it finds its files, but keep the user in their current directory.
(cd "$PROJECT_DIR" && ./sentinel-bin "\$@")
EOF_WRAPPER

# Make it executable
sudo chmod +x /usr/local/bin/sentinel

echo "------------------------------------------------"
echo "✅ Done! Terminal deadlock is mathematically impossible."
echo "✅ Clean CLI active. No quotes or flags needed."
echo "👉 Open a new terminal to start, then test it by typing: sentinel what is the time"
