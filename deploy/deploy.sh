#!/bin/bash
set -e

# Configuration
# SERVER can be set via environment variable, or defaults to SSH config alias "do-regserver"
# Users should configure their SSH alias in ~/.ssh/config:
#   Host do-regserver
#       HostName <your-server-ip>
#       User root
#       IdentityFile ~/.ssh/id_ed25519
SERVER="${REGSERVER_HOST:-do-regserver}"
REMOTE_PATH="/opt/regserver"
BINARY_NAME="regserver"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building...${NC}"
cd "$(dirname "$0")/.."
go build -o ./dist/${BINARY_NAME} ./cmd/regserver

echo -e "${YELLOW}Uploading...${NC}"
scp ./dist/${BINARY_NAME} ${SERVER}:${REMOTE_PATH}/${BINARY_NAME}.new

echo -e "${YELLOW}Deploying...${NC}"
ssh ${SERVER} << 'EOF'
    sudo systemctl stop regserver
    sudo mv /opt/regserver/regserver.new /opt/regserver/regserver
    sudo chmod +x /opt/regserver/regserver
    sudo chown regserver:regserver /opt/regserver/regserver
    sudo systemctl start regserver
    sleep 2
    sudo systemctl status regserver --no-pager
EOF

echo -e "${GREEN}Done!${NC}"
