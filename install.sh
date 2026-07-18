#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}🛡️  Starting Sentinel v2.0 Automated Setup...${NC}\n"

ENV_FILE="secrets/.env"

if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}✗ Error: Configuration file not found at $ENV_FILE${NC}"
    echo -e "Please create the file and add: AI_API_KEY=\"your_key\""
    exit 1
fi

API_KEY_CHECK=$(grep -v '^#' "$ENV_FILE" | grep 'AI_API_KEY' | cut -d '=' -f2 | tr -d '"' | tr -d "'")

if [ -z "$API_KEY_CHECK" ]; then
    echo -e "${RED}✗ Error: AI_API_KEY is defined but empty in $ENV_FILE.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Secret file $ENV_FILE validated successfully.${NC}\n"

SOCKET_DIR="/run/sentinel"

if [ ! -d "$SOCKET_DIR" ]; then
    echo "Creating directory $SOCKET_DIR (requires sudo)..."
    sudo mkdir -p "$SOCKET_DIR"
fi

sudo chown "$USER:$USER" "$SOCKET_DIR"
echo -e "${GREEN}✓ IPC directory ready.${NC}\n"

mkdir -p bin
go build -o bin/sentinel ./cmd/sentinel
echo -e "${GREEN}✓ CLI compiled at ./bin/sentinel${NC}\n"

# NUEVO: Instalación global automática creando un symlink en /usr/local/bin
echo "📦 Installing CLI globally (requires sudo)..."
sudo ln -sf "$(pwd)/bin/sentinel" /usr/local/bin/sentinel
echo -e "${GREEN}✓ CLI available globally at /usr/local/bin/sentinel${NC}\n"

docker compose --env-file secrets/.env up --build -d
echo -e "${GREEN}✓ Infrastructure running in background.${NC}\n"

sleep 1
if [ -S "$SOCKET_DIR/sentinel.sock" ]; then
    echo -e "${GREEN}✓ Active UNIX Socket detected at $SOCKET_DIR/sentinel.sock${NC}\n"
else
    echo -e "${RED}⚠️  Warning: Socket file not detected yet. Container might still be initializing.${NC}\n"
fi

echo -e "${GREEN}${BOLD}🚀 SENTINEL v2.0 IS LIVE!${NC}"
echo -e "──────────────────────────────────────────────────"
echo -e "• Test your local CLI running from ANYWHERE:"
echo -e "  ${BOLD}sentinel explain docker \"Why is my container restarting?\"${NC}"
echo -e "• Open Grafana Dashboard instantly at:"
echo -e "  ${BOLD}http://localhost:3000 (admin/admin)${NC}"
echo -e "──────────────────────────────────────────────────\n"
