#!/bin/bash
# Device Bridge v2 - macOS Uninstallation Script

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SERVICE_LABEL="com.devicebridge.service"
SERVICE_USER="_devicebridge"
PLIST_PATH="/Library/LaunchDaemons/com.devicebridge.service.plist"

echo -e "${YELLOW}Device Bridge v2 - Uninstallation${NC}"
echo "===================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run with sudo${NC}"
    exit 1
fi

# Confirm
read -p "Are you sure you want to uninstall Device Bridge? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Uninstallation cancelled"
    exit 0
fi

echo ""
echo -e "${YELLOW}Stopping and unloading service...${NC}"
launchctl stop $SERVICE_LABEL 2>/dev/null || true
launchctl unload $PLIST_PATH 2>/dev/null || true

echo ""
echo -e "${YELLOW}Removing launchd plist...${NC}"
rm -f $PLIST_PATH

echo ""
echo -e "${YELLOW}Removing binary...${NC}"
rm -f /usr/local/bin/device-bridge

echo ""
read -p "Remove configuration and data? (yes/no): " remove_data
if [ "$remove_data" == "yes" ]; then
    echo -e "${YELLOW}Removing directories...${NC}"
    rm -rf /usr/local/var/device-bridge
    rm -rf /usr/local/etc/device-bridge
    rm -rf /usr/local/var/log/device-bridge
fi

echo ""
read -p "Remove service user? (yes/no): " remove_user
if [ "$remove_user" == "yes" ]; then
    dscl . -delete /Users/$SERVICE_USER 2>/dev/null || true
    dscl . -delete /Groups/$SERVICE_USER 2>/dev/null || true
    echo "Removed user: $SERVICE_USER"
fi

echo ""
echo -e "${GREEN}Uninstallation complete!${NC}"
