# Serial Scale Setup Guide

This guide provides comprehensive instructions for setting up and using serial scales with Device Bridge v2.

## Table of Contents

- [Overview](#overview)
- [Supported Protocols](#supported-protocols)
- [Hardware Requirements](#hardware-requirements)
- [Platform-Specific Setup](#platform-specific-setup)
  - [Linux](#linux)
  - [macOS](#macos)
  - [Windows](#windows)
- [Configuration](#configuration)
- [Testing Your Scale](#testing-your-scale)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)
- [Advanced Topics](#advanced-topics)

## Overview

Device Bridge v2 supports serial scales from various manufacturers through a protocol-based driver architecture. The driver communicates with scales over RS-232/RS-485 serial connections (typically via USB-to-Serial adapters).

### Key Features

- **Multiple Protocol Support**: MT-SICS (Mettler Toledo), CAS (Korean manufacturer), Dibal (Spanish retail), Toledo 8217 (legacy), and generic ASCII
- **Auto-Protocol Detection**: Automatically detect which protocol your scale uses
- **Auto-retry Logic**: Configurable retry attempts for reliable communication
- **Unit Conversion**: Automatic conversion between g, kg, lb, oz, and more
- **Stable Weight Detection**: Wait for stable readings before returning values
- **Zero and Tare Operations**: Full control over scale calibration

## Supported Protocols

### MT-SICS (Mettler Toledo Standard Interface Command Set)

**Manufacturers**: Mettler Toledo, various compatible scales

**Configuration**:
```yaml
protocol: "mtsics"  # or "mt-sics" or "mettler"
baud_rate: 9600
data_bits: 8
parity: "none"
stop_bits: 1
```

**Features**:
- Text-based ASCII protocol
- Wide industry adoption
- Stable/unstable weight indication
- Overload/underload detection
- Scale information query

**Common Models**:
- Mettler Toledo ICS685 (Industrial)
- Mettler Toledo PS60 (Retail)
- Mettler Toledo XP205 (Laboratory)
- Mettler Toledo JL602 (Jewelry)

### CAS Protocol

**Manufacturers**: CAS Corporation (Korea), various Asian brands

**Configuration**:
```yaml
protocol: "cas"
baud_rate: 9600
data_bits: 8
parity: "none"
stop_bits: 1
```

**Features**:
- Binary protocol with STX/ETX delimiters
- Fixed 6-digit weight format
- Optional checksum validation
- Configurable decimal places
- Widely used in Asian markets

**Common Models**:
- CAS CL5000 (Retail)
- CAS MW-II (Precision)
- CAS LP (Platform scales)
- Various generic CAS-compatible scales

### Dibal Protocol

**Manufacturers**: Dibal (Spain), various European retail scales

**Configuration**:
```yaml
protocol: "dibal"
baud_rate: 9600
data_bits: 8
parity: "none"
stop_bits: 1
```

**Features**:
- Continuous output mode (no commands required)
- STX/ETX frame delimiters
- Status indicators (+/- for stable/unstable)
- Supports kg, g, lb, oz units
- Optional checksum validation (D-900 series)
- Configurable decimal places (0, 1, 2, or 3)

**Common Models**:
- Dibal D-900 Series (Retail)
- Dibal D-955 (Retail with printer)
- Dibal M-525 (Industrial)
- Dibal K-280 (Compact retail)

**Example Response Format**:
```
<STX>+00123kg<ETX>  # Stable, 1.23 kg
<STX>-01500g<ETX>   # Unstable, 15.00 g
<STX>O99999kg<ETX>  # Overload
```

### Toledo 8217 Protocol

**Manufacturers**: Toledo (legacy), Mettler Toledo (older models)

**Configuration**:
```yaml
protocol: "toledo"  # or "toledo8217" or "toledo-8217"
baud_rate: 9600
data_bits: 8
parity: "none"
stop_bits: 1
```

**Features**:
- Continuous ASCII output (no commands required)
- Fixed or variable width format
- Status indicators (S/D/M/+/-/?)
- Supports kg, g, lb, oz units
- CR/LF line termination
- Legacy protocol still used in many installations

**Common Models**:
- Toledo 8217 (legacy)
- Toledo 8530 (legacy platform scale)
- Older Mettler Toledo models (pre-MT-SICS)

**Example Response Formats**:
```
S    123.45 kg\r\n      # Variable width, stable
D     50.2 lb\r\n       # Variable width, dynamic
S  00001234 kg\r\n      # Fixed width, 8-digit weight
S M 123.45 kg\r\n       # With motion flag
```

### Generic Protocol

**Manufacturers**: Various manufacturers with simple ASCII output

**Configuration**:
```yaml
protocol: "generic"
baud_rate: 9600
data_bits: 8
parity: "none"
stop_bits: 1
```

**Features**:
- Configurable weight format (regex-based)
- Multiple line terminator support
- Flexible unit parsing
- Suitable for custom or uncommon scales

**Use When**: Your scale doesn't match any of the above protocols but outputs simple ASCII text like "123.45 kg"

## Hardware Requirements

### Scales

Any serial scale that supports one of the protocols listed above. The scale must have:
- RS-232 or RS-485 serial port (or USB with built-in serial adapter)
- Compatible protocol (MT-SICS, CAS, etc.)
- Power supply

### Computer Connection

Most modern scales use one of these connection methods:

1. **USB-to-Serial Adapter (Most Common)**
   - Built into many modern scales
   - Creates virtual COM port
   - No driver needed on modern systems

2. **Native RS-232 Port**
   - Older computers/industrial PCs
   - Requires RS-232 port or PCI/PCIe card

3. **Network-to-Serial Adapter**
   - For remote scale access
   - Requires network configuration

### Cables

- **RS-232**: Standard 9-pin serial cable (null modem if connecting directly)
- **USB**: Type A to Type B (or Mini/Micro USB depending on scale)

## Platform-Specific Setup

### Linux

#### 1. Identify the Serial Port

When you connect your scale, Linux creates a device file. Find it using:

```bash
# List all serial ports
ls /dev/tty{USB,S,ACM}*

# Or watch for new devices when you plug in the scale
dmesg | grep tty
```

Common port paths:
- `/dev/ttyUSB0` - USB-to-Serial adapter
- `/dev/ttyS0` - Native serial port
- `/dev/ttyACM0` - Some USB devices

#### 2. Set Permissions

By default, serial ports require root access. Add your user to the `dialout` group:

```bash
# Add user to dialout group
sudo usermod -a -G dialout $USER

# Verify membership
groups $USER

# Log out and log back in for changes to take effect
```

Alternatively, for testing only (resets on reboot):
```bash
sudo chmod 666 /dev/ttyUSB0
```

#### 3. Test Port Access

Verify you can access the port:

```bash
# Check port is accessible
ls -l /dev/ttyUSB0

# Should show: crw-rw---- 1 root dialout ...
```

#### 4. Configure Device

Create or edit your configuration file:

```yaml
devices:
  - id: my-scale
    name: "Production Scale"
    kind: scale.serial
    enabled: true
    metadata:
      port: "/dev/ttyUSB0"
      baud_rate: 9600
      protocol: "mtsics"
      data_bits: 8
      parity: "none"
      stop_bits: 1
```

### macOS

#### 1. Identify the Serial Port

When you connect your scale:

```bash
# List all serial devices
ls /dev/{cu,tty}.*

# Common paths:
# /dev/cu.usbserial-*
# /dev/tty.usbserial-*
# /dev/cu.SLAB_USBtoUART (for Silicon Labs adapters)
```

**Note**: Use `/dev/cu.*` for outgoing connections (recommended)

#### 2. No Special Permissions Needed

macOS doesn't require special permissions for serial port access.

#### 3. Configure Device

```yaml
devices:
  - id: my-scale
    name: "Production Scale"
    kind: scale.serial
    enabled: true
    metadata:
      port: "/dev/cu.usbserial-1420"
      baud_rate: 9600
      protocol: "mtsics"
```

### Windows

#### 1. Identify the COM Port

When you connect your scale:

1. Open **Device Manager** (Win+X → Device Manager)
2. Expand **Ports (COM & LPT)**
3. Find your USB Serial Port (e.g., "USB Serial Port (COM3)")
4. Note the COM port number

Or use PowerShell:
```powershell
Get-WMIObject Win32_SerialPort | Select-Object DeviceID,Description
```

#### 2. No Special Permissions Needed

Windows doesn't require special permissions for COM port access.

#### 3. Configure Device

```yaml
devices:
  - id: my-scale
    name: "Production Scale"
    kind: scale.serial
    enabled: true
    metadata:
      port: "COM3"            # Use the COM port from Device Manager
      baud_rate: 9600
      protocol: "mtsics"
```

## Configuration

### Basic Configuration

Minimal configuration for a scale:

```yaml
devices:
  - id: scale-01
    name: "My Scale"
    kind: scale.serial
    enabled: true
    metadata:
      port: "/dev/ttyUSB0"    # Adjust for your platform
      baud_rate: 9600
      protocol: "mtsics"
```

### Advanced Configuration

Full configuration with all options:

```yaml
devices:
  - id: scale-advanced
    name: "Advanced Scale"
    kind: scale.serial
    enabled: true
    metadata:
      # Serial Port Settings
      port: "/dev/ttyUSB0"
      baud_rate: 9600         # Common: 9600, 19200, 38400
      data_bits: 8            # Usually 8
      parity: "none"          # Options: "none", "even", "odd"
      stop_bits: 1            # Usually 1

      # Protocol Settings
      protocol: "mtsics"      # Options: "mtsics", "cas"
      read_timeout: 1000      # Milliseconds
      retry_attempts: 3       # Number of retries
      retry_delay: 100        # Delay between retries (ms)

      # Unit Preference
      preferred_unit: "kg"    # Options: "g", "kg", "lb", "oz", "t", "ozt", "dwt", "ct"
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `port` | string | required | Serial port path (e.g., "/dev/ttyUSB0", "COM1") |
| `baud_rate` | int | 9600 | Communication speed (9600, 19200, etc.) |
| `data_bits` | int | 8 | Number of data bits (7 or 8) |
| `parity` | string | "none" | Parity checking ("none", "even", "odd") |
| `stop_bits` | int | 1 | Number of stop bits (1 or 2) |
| `protocol` | string | "mtsics" | Protocol name ("mtsics", "cas") |
| `read_timeout` | int | 1000 | Read timeout in milliseconds |
| `retry_attempts` | int | 3 | Number of retry attempts |
| `retry_delay` | int | 100 | Delay between retries in milliseconds |
| `preferred_unit` | string | "kg" | Preferred weight unit |

## Testing Your Scale

### 1. Start the Bridge

```bash
./bridge --config configs/config.scale.simple.yaml
```

### 2. Verify Device Registration

Check that your scale is registered:

```bash
bridge-cli devices list
```

You should see your scale with status "ready".

### 3. Read Weight

Basic weight reading:

```bash
bridge-cli scale read scale-01
```

Expected output:
```
Weight: 123.4 kg
Stable: true
Timestamp: 2024-01-15T10:30:45Z
```

### 4. Test Zero Operation

Zero the scale (set current reading as zero point):

```bash
bridge-cli scale zero scale-01
```

### 5. Test Tare Operation

Tare the scale (subtract container weight):

```bash
bridge-cli scale tare scale-01
```

## Usage Examples

### gRPC Example (Go)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

func main() {
    // Connect to bridge
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewDeviceBridgeClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Read weight
    resp, err := client.ReadWeight(ctx, &pb.ReadWeightRequest{
        DeviceId: "scale-01",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Weight: %.2f %s\n", resp.Weight, resp.Unit)
    fmt.Printf("Stable: %v\n", resp.Stable)
}
```

### REST API Example (curl)

```bash
# Read weight
curl -X POST http://localhost:8080/api/v1/scale/read \
  -H "Content-Type: application/json" \
  -d '{"device_id": "scale-01"}'

# Response:
# {
#   "weight": 123.4,
#   "unit": "kg",
#   "stable": true,
#   "timestamp": "2024-01-15T10:30:45Z"
# }

# Zero the scale
curl -X POST http://localhost:8080/api/v1/scale/zero \
  -H "Content-Type: application/json" \
  -d '{"device_id": "scale-01"}'

# Tare the scale
curl -X POST http://localhost:8080/api/v1/scale/tare \
  -H "Content-Type: application/json" \
  -d '{"device_id": "scale-01"}'
```

### WebSocket Example (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('Connected to Device Bridge');

    // Read weight
    ws.send(JSON.stringify({
        type: 'scale.read',
        device_id: 'scale-01'
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    if (data.type === 'scale.weight') {
        console.log(`Weight: ${data.weight} ${data.unit}`);
        console.log(`Stable: ${data.stable}`);
    }
};
```

### Python Example

```python
import grpc
import devicebridge_pb2
import devicebridge_pb2_grpc

# Connect to bridge
channel = grpc.insecure_channel('localhost:50051')
client = devicebridge_pb2_grpc.DeviceBridgeStub(channel)

# Read weight
request = devicebridge_pb2.ReadWeightRequest(device_id='scale-01')
response = client.ReadWeight(request)

print(f"Weight: {response.weight} {response.unit}")
print(f"Stable: {response.stable}")

# Zero the scale
zero_request = devicebridge_pb2.ZeroScaleRequest(device_id='scale-01')
client.ZeroScale(zero_request)

# Tare the scale
tare_request = devicebridge_pb2.TareScaleRequest(device_id='scale-01')
client.TareScale(tare_request)
```

## Troubleshooting

### Scale Not Detected

**Problem**: Scale not appearing in device list

**Solutions**:
1. **Check physical connection**
   - Ensure USB cable is securely connected
   - Try a different USB port
   - Check scale power supply

2. **Verify port exists**
   ```bash
   # Linux
   ls /dev/tty{USB,S,ACM}*

   # macOS
   ls /dev/cu.*

   # Windows (PowerShell)
   Get-WMIObject Win32_SerialPort
   ```

3. **Check permissions (Linux only)**
   ```bash
   # Add user to dialout group
   sudo usermod -a -G dialout $USER
   # Log out and log back in
   ```

### Permission Denied Error (Linux)

**Problem**: "permission denied" when opening port

**Solution**:
```bash
# Add user to dialout group
sudo usermod -a -G dialout $USER

# Or temporarily (for testing)
sudo chmod 666 /dev/ttyUSB0

# Verify
ls -l /dev/ttyUSB0
# Should show: crw-rw---- 1 root dialout
```

### Scale Not Responding

**Problem**: Device registered but not responding to commands

**Solutions**:
1. **Verify baud rate**
   - Check scale manual for correct baud rate
   - Common values: 9600, 19200, 38400
   - Try different rates if unsure

2. **Try different protocol**
   ```yaml
   # If mtsics doesn't work, try cas
   protocol: "cas"
   ```

3. **Increase timeout**
   ```yaml
   read_timeout: 2000  # Increase to 2 seconds
   ```

4. **Check cable**
   - Ensure using correct cable type
   - Some cables are charge-only (no data)
   - Try different cable

5. **Check scale settings**
   - Some scales have communication settings menu
   - Verify protocol is enabled
   - Check if scale is in "continuous" or "on-demand" mode

### Incorrect Weight Readings

**Problem**: Getting wrong weight values

**Solutions**:
1. **Check unit preference**
   ```yaml
   preferred_unit: "kg"  # Make sure this matches expected unit
   ```

2. **Verify protocol**
   - MT-SICS vs CAS have different formats
   - Wrong protocol can cause parsing errors

3. **Zero the scale**
   ```bash
   bridge-cli scale zero scale-01
   ```

4. **Check decimal places (CAS only)**
   - CAS protocol has configurable decimal places
   - Default is 1 decimal place (001234 = 123.4)

### Unstable Readings

**Problem**: Weight constantly changing or not marking as stable

**Solutions**:
1. **Environmental factors**
   - Place scale on stable, level surface
   - Avoid vibration sources
   - Shield from air currents
   - Allow scale to warm up

2. **Increase timeout**
   ```yaml
   read_timeout: 3000  # Wait longer for stability
   ```

3. **Calibrate scale**
   - Follow manufacturer's calibration procedure
   - Zero the scale
   - Use tare for container weight

### Driver Creation Failed

**Problem**: "failed to create scale driver" in logs

**Solutions**:
1. **Check configuration syntax**
   - Validate YAML syntax
   - Ensure all required fields present
   - Check for typos in field names

2. **Verify port path**
   - Use absolute path (not relative)
   - Ensure port exists
   - Check for typos

3. **Check protocol name**
   - Must be: "mtsics", "mt-sics", "mettler", or "cas"
   - Case-sensitive in some environments

## Advanced Topics

### Multiple Scales

You can configure multiple scales on one system:

```yaml
devices:
  - id: scale-checkout-1
    name: "Checkout 1"
    kind: scale.serial
    metadata:
      port: "/dev/ttyUSB0"
      protocol: "cas"

  - id: scale-checkout-2
    name: "Checkout 2"
    kind: scale.serial
    metadata:
      port: "/dev/ttyUSB1"
      protocol: "cas"

  - id: scale-produce
    name: "Produce Department"
    kind: scale.serial
    metadata:
      port: "/dev/ttyUSB2"
      protocol: "mtsics"
```

### Unit Conversion

The driver automatically converts between units:

```yaml
# Scale returns grams, but you want kilograms
preferred_unit: "kg"

# Read weight
# Scale: 1500 g
# Returned: 1.5 kg
```

Supported units:
- `g` - Gram
- `kg` - Kilogram
- `lb` - Pound
- `oz` - Ounce
- `t` - Metric Ton
- `ozt` - Troy Ounce (precious metals)
- `dwt` - Pennyweight
- `ct` - Carat

### Protocol-Specific Features

#### MT-SICS Extended Features

MT-SICS protocol supports additional commands:

- **Scale Information**: Query scale model, serial number, version
- **Unit Query**: Check available units
- **Calibration**: Remote calibration commands (if enabled)

#### CAS Protocol Variants

CAS protocol has vendor-specific variants:

- **With Checksum**: Some models append XOR checksum
- **Decimal Places**: Configurable (0, 1, or 2 decimal places)
- **Extended Status**: Some models return additional error codes

### Custom Serial Settings

For scales with non-standard settings:

```yaml
metadata:
  # High-speed communication
  baud_rate: 38400

  # Laboratory scale with even parity for data integrity
  parity: "even"
  data_bits: 7

  # Industrial environment with longer timeouts
  read_timeout: 5000
  retry_attempts: 10
  retry_delay: 500
```

### Debugging

Enable debug logging:

```yaml
logging:
  level: "debug"  # Show detailed protocol communication
  format: "text"
```

View serial communication:

```bash
# Linux - Monitor port with screen
screen /dev/ttyUSB0 9600

# Or use minicom
minicom -D /dev/ttyUSB0 -b 9600

# Exit: Ctrl+A, then K
```

## Next Steps

- [Configuration Examples](../configs/config.scale.example.yaml) - Comprehensive configuration examples
- [Serial Scale Protocols](SERIAL_SCALE_PROTOCOLS.md) - Detailed protocol specifications
- [API Documentation](API.md) - Complete API reference
- [Hardware Integration Tests](../test/hardware/README.md) - Testing with physical scales

## Support

For issues or questions:
- GitHub Issues: https://github.com/Macber-eg/Flutter-Device/issues
- Documentation: https://docs.flutter-device.com
