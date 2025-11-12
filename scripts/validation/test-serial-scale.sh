#!/bin/bash
#
# Serial Scale Hardware Validation Script
# Tests serial scale functionality on Linux/macOS
#
# Usage: ./test-serial-scale.sh [port] [protocol]
# Example: ./test-serial-scale.sh /dev/ttyUSB0 mtsics

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PORT="${1:-/dev/ttyUSB0}"
PROTOCOL="${2:-auto}"

echo "========================================="
echo "Serial Scale Hardware Validation"
echo "========================================="
echo ""
echo "Target Port: $PORT"
echo "Protocol: $PROTOCOL"
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

# Step 3: Check serial port
echo "Step 3: Checking serial port..."
if [ "$PLATFORM" = "Linux" ]; then
    if [ -e "$PORT" ]; then
        echo -e "${GREEN}✓${NC} Port exists: $PORT"

        # Check permissions
        if [ -r "$PORT" ] && [ -w "$PORT" ]; then
            echo -e "${GREEN}✓${NC} Port permissions OK"
        else
            echo -e "${RED}✗${NC} No read/write permissions for: $PORT"
            echo "Solution 1: Run with sudo (temporary)"
            echo "Solution 2: Add user to dialout group:"
            echo "  sudo usermod -a -G dialout \$USER"
            echo "  (logout and login again)"
            exit 1
        fi
    else
        echo -e "${RED}✗${NC} Port not found: $PORT"
        echo "Available serial ports:"
        ls /dev/tty{USB,S,ACM}* 2>/dev/null || echo "No serial ports found"
        exit 1
    fi
elif [ "$PLATFORM" = "macOS" ]; then
    if [ -e "$PORT" ]; then
        echo -e "${GREEN}✓${NC} Port exists: $PORT"
    else
        echo -e "${RED}✗${NC} Port not found: $PORT"
        echo "Available serial ports:"
        ls /dev/{tty,cu}.* 2>/dev/null || echo "No serial ports found"
        exit 1
    fi
fi
echo ""

# Step 4: Build Device Bridge
echo "Step 4: Building Device Bridge..."
cd "$(dirname "$0")/../.."
if make build &> /dev/null; then
    echo -e "${GREEN}✓${NC} Build successful"
else
    echo -e "${RED}✗${NC} Build failed"
    echo "Run: make build"
    exit 1
fi
echo ""

# Step 5: Create test configuration
echo "Step 5: Creating test configuration..."
cat > /tmp/test-scale-config.yaml << EOF
server:
  grpc_port: 50051
  http_port: 8080
  metrics_port: 9090

devices:
  - id: scale-test
    name: "Serial Scale Hardware Test"
    kind: scale.serial
    enabled: true
    metadata:
      port: "$PORT"
      baud_rate: 9600
      data_bits: 8
      parity: "none"
      stop_bits: 1
      protocol: "$PROTOCOL"
      read_timeout: 1000
      retry_attempts: 3
      retry_delay: 100
      preferred_unit: "kg"

logging:
  level: debug
  format: text

telemetry:
  metrics: true
  tracing: false
EOF
echo -e "${GREEN}✓${NC} Config created: /tmp/test-scale-config.yaml"
echo ""

# Step 6: Test scale connectivity
echo "Step 6: Testing scale connectivity..."
echo "This will attempt to read from the scale..."
echo ""

# Start Device Bridge in background
echo "Starting Device Bridge..."
./bin/device-bridge -config /tmp/test-scale-config.yaml > /tmp/scale-test.log 2>&1 &
BRIDGE_PID=$!

# Wait for startup
sleep 3

# Check if bridge is running
if ! ps -p $BRIDGE_PID > /dev/null; then
    echo -e "${RED}✗${NC} Device Bridge failed to start"
    echo "Log output:"
    cat /tmp/scale-test.log
    exit 1
fi

echo -e "${GREEN}✓${NC} Device Bridge started (PID: $BRIDGE_PID)"
echo ""

# Step 7: Run validation tests
echo "========================================="
echo "Step 7: Running Validation Tests"
echo "========================================="
echo ""

# Function to make gRPC request
grpc_request() {
    local method=$1
    local data=$2
    grpcurl -plaintext -d "$data" localhost:50051 "devicebridge.v1.DeviceBridge/$method" 2>&1
}

# Test 1: Check device status
echo -e "${BLUE}Test 1: Device Status${NC}"
STATUS=$(grpc_request "GetDeviceStatus" '{"device_id": "scale-test"}')
if echo "$STATUS" | grep -q "READY\|ONLINE"; then
    echo -e "${GREEN}✓${NC} Device is online"
else
    echo -e "${YELLOW}⚠${NC} Device status: $STATUS"
fi
echo ""

# Test 2: Read weight (immediate)
echo -e "${BLUE}Test 2: Read Weight (Immediate)${NC}"
echo "Reading current weight..."
WEIGHT=$(grpc_request "ReadWeight" '{"device_id": "scale-test", "stable_only": false}')
if echo "$WEIGHT" | grep -q "value"; then
    echo -e "${GREEN}✓${NC} Weight reading successful"
    echo "$WEIGHT" | grep -E "value|unit|stable"
else
    echo -e "${YELLOW}⚠${NC} Weight reading failed or no response"
    echo "$WEIGHT"
fi
echo ""

# Test 3: Read stable weight
echo -e "${BLUE}Test 3: Read Stable Weight${NC}"
echo "Waiting for stable weight (put weight on scale)..."
echo "This may take a few seconds..."
STABLE=$(grpc_request "ReadWeight" '{"device_id": "scale-test", "stable_only": true}')
if echo "$STABLE" | grep -q "stable.*true"; then
    echo -e "${GREEN}✓${NC} Stable weight reading successful"
    echo "$STABLE" | grep -E "value|unit|stable"
else
    echo -e "${YELLOW}⚠${NC} No stable weight detected"
    echo "$STABLE"
fi
echo ""

# Test 4: Zero operation
echo -e "${BLUE}Test 4: Zero Scale${NC}"
echo "Sending zero command..."
ZERO=$(grpc_request "ZeroScale" '{"device_id": "scale-test"}')
if echo "$ZERO" | grep -q "success\|ok" || [ -z "$ZERO" ]; then
    echo -e "${GREEN}✓${NC} Zero command successful"
else
    echo -e "${YELLOW}⚠${NC} Zero command response: $ZERO"
fi
sleep 1
echo ""

# Test 5: Tare operation
echo -e "${BLUE}Test 5: Tare Scale${NC}"
echo "Place container on scale and press Enter..."
read -p ""
echo "Sending tare command..."
TARE=$(grpc_request "TareScale" '{"device_id": "scale-test"}')
if echo "$TARE" | grep -q "success\|ok" || [ -z "$TARE" ]; then
    echo -e "${GREEN}✓${NC} Tare command successful"
else
    echo -e "${YELLOW}⚠${NC} Tare command response: $TARE"
fi
sleep 1
echo ""

# Test 6: Continuous reading
echo -e "${BLUE}Test 6: Continuous Reading (10 samples)${NC}"
echo "Reading weight 10 times..."
SUCCESS_COUNT=0
for i in {1..10}; do
    READING=$(grpc_request "ReadWeight" '{"device_id": "scale-test", "stable_only": false}')
    if echo "$READING" | grep -q "value"; then
        VALUE=$(echo "$READING" | grep -o '"value":[0-9.]*' | cut -d: -f2)
        UNIT=$(echo "$READING" | grep -o '"unit":"[^"]*"' | cut -d'"' -f4)
        echo "  Reading $i: $VALUE $UNIT"
        ((SUCCESS_COUNT++))
    else
        echo -e "  ${YELLOW}Reading $i: Failed${NC}"
    fi
    sleep 0.5
done
echo -e "${GREEN}✓${NC} Successful readings: $SUCCESS_COUNT/10"
echo ""

# Test 7: Protocol validation
echo -e "${BLUE}Test 7: Protocol Validation${NC}"
if [ "$PROTOCOL" = "auto" ]; then
    echo "Checking detected protocol..."
    # Check logs for protocol detection
    if grep -q "protocol.*detected\|using protocol" /tmp/scale-test.log; then
        DETECTED=$(grep -o "protocol.*detected\|using protocol.*" /tmp/scale-test.log | head -1)
        echo -e "${GREEN}✓${NC} $DETECTED"
    else
        echo -e "${YELLOW}⚠${NC} Auto-detection status unknown"
    fi
else
    echo -e "${GREEN}✓${NC} Using protocol: $PROTOCOL"
fi
echo ""

# Cleanup
echo "========================================="
echo "Cleaning up..."
echo "========================================="
kill $BRIDGE_PID 2>/dev/null || true
wait $BRIDGE_PID 2>/dev/null || true
echo -e "${GREEN}✓${NC} Device Bridge stopped"
echo ""

# Summary
echo "========================================="
echo "VALIDATION SUMMARY"
echo "========================================="
echo ""
echo "Port: $PORT"
echo "Protocol: $PROTOCOL"
echo "Successful readings: $SUCCESS_COUNT/10"
echo ""

if [ $SUCCESS_COUNT -ge 8 ]; then
    echo -e "${GREEN}✓ VALIDATION PASSED${NC}"
    echo ""
    echo "The scale is working correctly!"
    echo ""
    echo "Next steps:"
    echo "1. Test with different protocols (if supported)"
    echo "2. Test continuous mode over extended period"
    echo "3. Test auto-reconnection (unplug/replug scale)"
    echo "4. Update hardware compatibility matrix"
else
    echo -e "${YELLOW}⚠ VALIDATION INCOMPLETE${NC}"
    echo ""
    echo "Some tests failed. Check the following:"
    echo "1. Scale is powered on and connected"
    echo "2. Correct port specified"
    echo "3. Correct protocol for your scale model"
    echo "4. Serial cable is working"
    echo "5. Baud rate matches scale configuration"
    echo ""
    echo "View detailed logs:"
    echo "  cat /tmp/scale-test.log"
fi
echo ""

# Offer to show logs
read -p "View detailed logs? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    less /tmp/scale-test.log
fi
