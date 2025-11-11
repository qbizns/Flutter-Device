#!/bin/bash
# Device Bridge v2 - macOS Installation Script
# Installs Device Bridge as a launchd service

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
SERVICE_LABEL="com.devicebridge.service"
SERVICE_USER="_devicebridge"
SERVICE_GROUP="_devicebridge"
INSTALL_DIR="/usr/local/var/device-bridge"
CONFIG_DIR="/usr/local/etc/device-bridge"
LOG_DIR="/usr/local/var/log/device-bridge"
BIN_DIR="/usr/local/bin"
PLIST_DIR="/Library/LaunchDaemons"

echo -e "${GREEN}Device Bridge v2 - macOS Installation${NC}"
echo "========================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run with sudo${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if binary exists
if [ ! -f "./device-bridge" ]; then
    echo -e "${RED}Error: device-bridge binary not found${NC}"
    echo "Please build the macOS binary first: make build-macos"
    exit 1
fi

echo -e "${YELLOW}Step 1: Creating service user and group${NC}"
# Check if user exists
if ! dscl . -read /Users/$SERVICE_USER &>/dev/null; then
    # Find next available UID
    MAXID=$(dscl . -list /Users UniqueID | awk '{print $2}' | sort -ug | tail -1)
    USERID=$((MAXID+1))

    # Create group
    dscl . -create /Groups/$SERVICE_GROUP
    dscl . -create /Groups/$SERVICE_GROUP PrimaryGroupID $USERID
    dscl . -create /Groups/$SERVICE_GROUP RealName "Device Bridge Service"

    # Create user
    dscl . -create /Users/$SERVICE_USER
    dscl . -create /Users/$SERVICE_USER UniqueID $USERID
    dscl . -create /Users/$SERVICE_USER PrimaryGroupID $USERID
    dscl . -create /Users/$SERVICE_USER UserShell /usr/bin/false
    dscl . -create /Users/$SERVICE_USER NFSHomeDirectory /var/empty
    dscl . -create /Users/$SERVICE_USER RealName "Device Bridge Service"

    echo "Created user: $SERVICE_USER (UID: $USERID)"
else
    echo "User already exists: $SERVICE_USER"
fi

echo ""
echo -e "${YELLOW}Step 2: Creating directories${NC}"
mkdir -p $INSTALL_DIR
mkdir -p $CONFIG_DIR
mkdir -p $LOG_DIR

echo "Created directories:"
echo "  - $INSTALL_DIR"
echo "  - $CONFIG_DIR"
echo "  - $LOG_DIR"

echo ""
echo -e "${YELLOW}Step 3: Installing binary${NC}"
cp ./device-bridge $BIN_DIR/device-bridge
chmod 755 $BIN_DIR/device-bridge
chown root:wheel $BIN_DIR/device-bridge
echo "Installed binary to $BIN_DIR/device-bridge"

echo ""
echo -e "${YELLOW}Step 4: Installing configuration${NC}"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    if [ -f "./configs/payment/config.production.yaml" ]; then
        cp ./configs/payment/config.production.yaml $CONFIG_DIR/config.yaml
        echo "Installed default configuration"
    else
        echo -e "${YELLOW}Warning: No default config found${NC}"
    fi
else
    echo "Configuration already exists"
fi

echo ""
echo -e "${YELLOW}Step 5: Setting permissions${NC}"
chown -R $SERVICE_USER:$SERVICE_GROUP $INSTALL_DIR
chown -R $SERVICE_USER:$SERVICE_GROUP $CONFIG_DIR
chown -R $SERVICE_USER:$SERVICE_GROUP $LOG_DIR
chmod 755 $INSTALL_DIR
chmod 750 $CONFIG_DIR
chmod 750 $LOG_DIR
chmod 640 $CONFIG_DIR/config.yaml 2>/dev/null || true
echo "Permissions set"

echo ""
echo -e "${YELLOW}Step 6: Installing launchd service${NC}"
cp ./deploy/macos/com.devicebridge.service.plist $PLIST_DIR/
chmod 644 $PLIST_DIR/com.devicebridge.service.plist
chown root:wheel $PLIST_DIR/com.devicebridge.service.plist
echo "Installed launchd plist"

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: sudo nano $CONFIG_DIR/config.yaml"
echo "  2. Load service:       sudo launchctl load $PLIST_DIR/com.devicebridge.service.plist"
echo "  3. Start service:      sudo launchctl start $SERVICE_LABEL"
echo "  4. Check status:       sudo launchctl list | grep devicebridge"
echo "  5. View logs:          tail -f $LOG_DIR/stdout.log"
echo ""
echo "Service Management:"
echo "  Start:   sudo launchctl start $SERVICE_LABEL"
echo "  Stop:    sudo launchctl stop $SERVICE_LABEL"
echo "  Restart: sudo launchctl stop $SERVICE_LABEL && sudo launchctl start $SERVICE_LABEL"
echo "  Unload:  sudo launchctl unload $PLIST_DIR/com.devicebridge.service.plist"
echo ""
