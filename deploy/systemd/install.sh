#!/bin/bash
# Device Bridge v2 - systemd Installation Script
# Installs Device Bridge as a systemd service on Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="device-bridge"
SERVICE_USER="devicebridge"
SERVICE_GROUP="devicebridge"
INSTALL_DIR="/opt/device-bridge"
CONFIG_DIR="/etc/device-bridge"
LOG_DIR="/var/log/device-bridge"
DATA_DIR="/var/lib/device-bridge"
BIN_DIR="/usr/local/bin"

echo -e "${GREEN}Device Bridge v2 - systemd Installation${NC}"
echo "=========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if binary exists
if [ ! -f "./device-bridge" ]; then
    echo -e "${RED}Error: device-bridge binary not found${NC}"
    echo "Please build the binary first: make build"
    exit 1
fi

echo -e "${YELLOW}Step 1: Creating service user and group${NC}"
if ! id -u $SERVICE_USER > /dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false $SERVICE_USER
    echo "Created user: $SERVICE_USER"
else
    echo "User already exists: $SERVICE_USER"
fi

echo ""
echo -e "${YELLOW}Step 2: Creating directories${NC}"
mkdir -p $INSTALL_DIR
mkdir -p $CONFIG_DIR
mkdir -p $LOG_DIR
mkdir -p $DATA_DIR

echo "Created directories:"
echo "  - $INSTALL_DIR"
echo "  - $CONFIG_DIR"
echo "  - $LOG_DIR"
echo "  - $DATA_DIR"

echo ""
echo -e "${YELLOW}Step 3: Installing binary${NC}"
cp ./device-bridge $BIN_DIR/device-bridge
chmod 755 $BIN_DIR/device-bridge
chown root:root $BIN_DIR/device-bridge
echo "Installed binary to $BIN_DIR/device-bridge"

echo ""
echo -e "${YELLOW}Step 4: Installing configuration${NC}"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    if [ -f "./configs/payment/config.production.yaml" ]; then
        cp ./configs/payment/config.production.yaml $CONFIG_DIR/config.yaml
        echo "Installed default configuration"
    else
        echo -e "${YELLOW}Warning: No default config found, you'll need to create $CONFIG_DIR/config.yaml${NC}"
    fi
else
    echo "Configuration already exists, not overwriting"
fi

# Create environment file template
cat > $CONFIG_DIR/environment << 'EOF'
# Device Bridge Environment Variables
# Uncomment and configure as needed

# Payment Terminal Settings
# MADA_TERMINAL_ID=
# MADA_MERCHANT_ID=
# MADA_AUTH_KEY=
# KNET_TERMINAL_ID=
# KNET_MERCHANT_ID=
# KNET_AUTH_KEY=

# API Settings
# API_KEY=
# API_PORT=8080

# Monitoring
# PROMETHEUS_PORT=9090
# JAEGER_ENDPOINT=
# LOGSTASH_ENDPOINT=

# Security
# ENCRYPTION_KEY_ID=
EOF

echo "Created environment file template: $CONFIG_DIR/environment"

echo ""
echo -e "${YELLOW}Step 5: Setting permissions${NC}"
chown -R $SERVICE_USER:$SERVICE_GROUP $INSTALL_DIR
chown -R $SERVICE_USER:$SERVICE_GROUP $CONFIG_DIR
chown -R $SERVICE_USER:$SERVICE_GROUP $LOG_DIR
chown -R $SERVICE_USER:$SERVICE_GROUP $DATA_DIR
chmod 755 $INSTALL_DIR
chmod 750 $CONFIG_DIR
chmod 750 $LOG_DIR
chmod 750 $DATA_DIR
chmod 640 $CONFIG_DIR/config.yaml 2>/dev/null || true
echo "Permissions set"

echo ""
echo -e "${YELLOW}Step 6: Installing systemd service${NC}"
cp ./deploy/systemd/device-bridge.service /etc/systemd/system/
chmod 644 /etc/systemd/system/device-bridge.service
systemctl daemon-reload
echo "Installed systemd service"

echo ""
echo -e "${YELLOW}Step 7: Configuring USB device access (optional)${NC}"
# Create udev rule for USB devices
cat > /etc/udev/rules.d/99-device-bridge.rules << 'EOF'
# Device Bridge USB Device Access
# Allow devicebridge user to access USB devices

# USB HID Scanners (adjust vendor:product IDs as needed)
SUBSYSTEM=="usb", ATTR{idVendor}=="05f9", ATTR{idProduct}=="4204", MODE="0660", GROUP="devicebridge"

# USB Printers (adjust as needed)
SUBSYSTEM=="usb", ATTR{bInterfaceClass}=="07", MODE="0660", GROUP="devicebridge"

# USB Serial devices
SUBSYSTEM=="tty", ATTRS{idVendor}=="067b", MODE="0660", GROUP="devicebridge"
EOF

udevadm control --reload-rules
echo "Created udev rules for USB access"
echo -e "${YELLOW}Note: You may need to customize /etc/udev/rules.d/99-device-bridge.rules for your devices${NC}"

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: sudo nano $CONFIG_DIR/config.yaml"
echo "  2. Edit environment: sudo nano $CONFIG_DIR/environment"
echo "  3. Start service:     sudo systemctl start $SERVICE_NAME"
echo "  4. Enable on boot:    sudo systemctl enable $SERVICE_NAME"
echo "  5. Check status:      sudo systemctl status $SERVICE_NAME"
echo "  6. View logs:         sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo -e "${YELLOW}Security reminder:${NC}"
echo "  - Configure API keys and credentials in $CONFIG_DIR/environment"
echo "  - Review and adjust firewall rules"
echo "  - Enable TLS/SSL in production"
echo ""
