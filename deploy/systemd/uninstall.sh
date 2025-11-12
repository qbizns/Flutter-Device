#!/bin/bash
# Device Bridge v2 - systemd Uninstallation Script

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SERVICE_NAME="device-bridge"
SERVICE_USER="devicebridge"

echo -e "${YELLOW}Device Bridge v2 - Uninstallation${NC}"
echo "===================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    exit 1
fi

# Confirm uninstallation
read -p "Are you sure you want to uninstall Device Bridge? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Uninstallation cancelled"
    exit 0
fi

echo ""
echo -e "${YELLOW}Stopping and disabling service...${NC}"
systemctl stop $SERVICE_NAME 2>/dev/null || true
systemctl disable $SERVICE_NAME 2>/dev/null || true

echo ""
echo -e "${YELLOW}Removing systemd service...${NC}"
rm -f /etc/systemd/system/device-bridge.service
systemctl daemon-reload

echo ""
echo -e "${YELLOW}Removing binary...${NC}"
rm -f /usr/local/bin/device-bridge

echo ""
echo -e "${YELLOW}Removing udev rules...${NC}"
rm -f /etc/udev/rules.d/99-device-bridge.rules
udevadm control --reload-rules 2>/dev/null || true

echo ""
read -p "Remove configuration and data? (yes/no): " remove_data
if [ "$remove_data" == "yes" ]; then
    echo -e "${YELLOW}Removing directories...${NC}"
    rm -rf /opt/device-bridge
    rm -rf /etc/device-bridge
    rm -rf /var/log/device-bridge
    rm -rf /var/lib/device-bridge
    echo "Removed all data and configuration"
else
    echo "Configuration and data preserved in:"
    echo "  - /etc/device-bridge"
    echo "  - /var/lib/device-bridge"
    echo "  - /var/log/device-bridge"
fi

echo ""
read -p "Remove service user? (yes/no): " remove_user
if [ "$remove_user" == "yes" ]; then
    userdel $SERVICE_USER 2>/dev/null || true
    echo "Removed user: $SERVICE_USER"
fi

echo ""
echo -e "${GREEN}Uninstallation complete!${NC}"
