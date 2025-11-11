#!/bin/bash
#
# USB HID Scanner Validation Script
# Tests USB HID scanner functionality on Linux/macOS
#
# Usage: ./test-usb-scanner.sh [vendor_id] [product_id]
# Example: ./test-usb-scanner.sh 05e0 1200

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default scanner IDs (Symbol/Zebra LS2208)
VENDOR_ID="${1:-05e0}"
PRODUCT_ID="${2:-1200}"

echo "========================================="
echo "USB HID Scanner Validation"
echo "========================================="
echo ""
echo "Target Scanner: VID=$VENDOR_ID PID=$PRODUCT_ID"
echo ""

# Step 1: Check OS
echo "Step 1: Detecting platform..."
OS="$(uname -s)"
case "${OS}" in
    Linux*)     PLATFORM=Linux;;
    Darwin*)    PLATFORM=macOS;;
    *)          PLATFORM="UNKNOWN:${OS}"
esac
echo -e "${GREEN}✓${NC} Platform: $PLATFORM"
echo ""

# Step 2: Check Go version
echo "Step 2: Checking Go version..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗${NC} Go not installed"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓${NC} Go version: $GO_VERSION"
echo ""

# Step 3: Check libusb (Linux/macOS)
echo "Step 3: Checking libusb..."
if [ "$PLATFORM" = "Linux" ]; then
    if pkg-config --exists libusb-1.0; then
        LIBUSB_VERSION=$(pkg-config --modversion libusb-1.0)
        echo -e "${GREEN}✓${NC} libusb version: $LIBUSB_VERSION"
    else
        echo -e "${RED}✗${NC} libusb not found"
        echo "Install: sudo apt-get install libusb-1.0-0 libusb-1.0-0-dev"
        exit 1
    fi
elif [ "$PLATFORM" = "macOS" ]; then
    if brew list libusb &> /dev/null; then
        LIBUSB_VERSION=$(brew list --versions libusb | awk '{print $2}')
        echo -e "${GREEN}✓${NC} libusb version: $LIBUSB_VERSION (via Homebrew)"
    else
        echo -e "${RED}✗${NC} libusb not found"
        echo "Install: brew install libusb"
        exit 1
    fi
fi
echo ""

# Step 4: Check for USB scanner
echo "Step 4: Checking for USB scanner..."
if [ "$PLATFORM" = "Linux" ]; then
    if lsusb | grep -q "$VENDOR_ID:$PRODUCT_ID"; then
        SCANNER_INFO=$(lsusb | grep "$VENDOR_ID:$PRODUCT_ID")
        echo -e "${GREEN}✓${NC} Scanner found: $SCANNER_INFO"
    else
        echo -e "${YELLOW}⚠${NC} Scanner not found (VID=$VENDOR_ID PID=$PRODUCT_ID)"
        echo "Available USB devices:"
        lsusb
        exit 1
    fi
elif [ "$PLATFORM" = "macOS" ]; then
    if system_profiler SPUSBDataType | grep -A 5 "Vendor ID: 0x$VENDOR_ID" | grep -q "Product ID: 0x$PRODUCT_ID"; then
        echo -e "${GREEN}✓${NC} Scanner found"
    else
        echo -e "${YELLOW}⚠${NC} Scanner not found (VID=$VENDOR_ID PID=$PRODUCT_ID)"
        echo "Run: system_profiler SPUSBDataType"
        exit 1
    fi
fi
echo ""

# Step 5: Check permissions (Linux)
if [ "$PLATFORM" = "Linux" ]; then
    echo "Step 5: Checking USB permissions..."
    USB_BUS=$(lsusb | grep "$VENDOR_ID:$PRODUCT_ID" | awk '{print $2}')
    USB_DEV=$(lsusb | grep "$VENDOR_ID:$PRODUCT_ID" | awk '{print $4}' | tr -d ':')
    USB_PATH="/dev/bus/usb/$USB_BUS/$USB_DEV"

    if [ -r "$USB_PATH" ] && [ -w "$USB_PATH" ]; then
        echo -e "${GREEN}✓${NC} Permissions OK: $USB_PATH"
    else
        echo -e "${RED}✗${NC} No permissions for: $USB_PATH"
        echo "Solution 1: Run as sudo (temporary)"
        echo "Solution 2: Create udev rules (recommended)"
        echo "  See: docs/USB_SCANNER_VALIDATION.md"
        exit 1
    fi
    echo ""
fi

# Step 6: Build Device Bridge
echo "Step 6: Building Device Bridge..."
cd "$(dirname "$0")/../.."
if make build &> /dev/null; then
    echo -e "${GREEN}✓${NC} Build successful"
else
    echo -e "${RED}✗${NC} Build failed"
    echo "Run: make build"
    exit 1
fi
echo ""

# Step 7: Create test configuration
echo "Step 7: Creating test configuration..."
cat > /tmp/test-scanner-config.yaml << EOF
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090

devices:
  - id: scanner-test
    name: "USB HID Scanner Test"
    kind: scanner.hid
    vendor_id: 0x$VENDOR_ID
    product_id: 0x$PRODUCT_ID
    buffer_size: 10
    reconnect_delay: 5s
    read_timeout: 1s

logging:
  level: debug
  format: text

telemetry:
  metrics: true
  tracing: false
EOF
echo -e "${GREEN}✓${NC} Config created: /tmp/test-scanner-config.yaml"
echo ""

# Step 8: Run Device Bridge
echo "Step 8: Starting Device Bridge (press Ctrl+C to stop)..."
echo "----------------------------------------"
echo ""
./bin/device-bridge -config /tmp/test-scanner-config.yaml &
BRIDGE_PID=$!

# Wait for startup
sleep 2

# Check if bridge is running
if ps -p $BRIDGE_PID > /dev/null; then
    echo ""
    echo "========================================="
    echo -e "${GREEN}✓ Device Bridge started successfully${NC}"
    echo "========================================="
    echo ""
    echo "Test Instructions:"
    echo "1. Scan a barcode with your scanner"
    echo "2. Check the output above for scan events"
    echo "3. Press Ctrl+C to stop"
    echo ""
    echo "Or in another terminal:"
    echo "  grpcurl -plaintext -d '{\"device_id\": \"scanner-test\"}' \\"
    echo "    localhost:50051 devicebridge.v1.DeviceBridge/SubscribeScanner"
    echo ""

    # Wait for user interrupt
    wait $BRIDGE_PID
else
    echo -e "${RED}✗ Device Bridge failed to start${NC}"
    echo "Check logs above for errors"
    exit 1
fi

# Cleanup
trap "kill $BRIDGE_PID 2>/dev/null" EXIT
